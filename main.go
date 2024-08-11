package main

import (
	"bytes"
	"flag"
	"gg/client/graphql"
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"gg/mapper"
	"gg/service"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	addr                    = flag.String("addr", ":8080", "http service address")
	slug                    = flag.String("slug", "", "Slug.")
	title                   = flag.String("title", "", "Title.")
	subreddit               = flag.String("subreddit", "", "Subreddit.")
	file                    = flag.String("file", "", "File.")
	upsetThreadTemplate     = template.Must(template.ParseFiles("template/upset-thread.tmpl"))
	upsetThreadHTMLTemplate = template.Must(template.ParseFiles("template/upset-thread.html"))
	upgrader                = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type IndexHandler struct {
	service service.ServiceInterface
}

type WebSockerHandler struct {
	service         service.ServiceInterface
	upsetThreadChan chan *domain.UpsetThreadHTML
}

func main() {
	flag.Parse()

	var service service.ServiceInterface = service.NewService(
		data.NewRedisDBService(),
		startgg.NewClient(graphql.NewClient(os.Getenv("START_GG_API_URL"), os.Getenv("START_GG_API_KEY"), &http.Client{})),
		&service.FileReaderWriter{},
	)
	indexHandler := IndexHandler{
		service: service,
	}

	webSocketHandler := WebSockerHandler{
		service:         service,
		upsetThreadChan: make(chan *domain.UpsetThreadHTML),
	}

	go webSocketHandler.getUpsetThreadHTML()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", &indexHandler)
	http.Handle("/ws", &webSocketHandler)
	http.ListenAndServe(*addr, nil)
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	upsetThread := h.service.GetUpsetThreadDB(*slug, *title)
	htmlUpsetThread := mapper.ToHTML(upsetThread, r.Host)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	upsetThreadHTMLTemplate.Execute(w, &htmlUpsetThread)
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WebSockerHandler) getUpsetThreadHTML() {
	for {
		upsetThread := h.service.Process(*slug, *title, *subreddit, *file, "")
		htmlUpsetThread := mapper.ToHTML(upsetThread, "")
		h.upsetThreadChan <- htmlUpsetThread
	}
}

func (h *WebSockerHandler) writer(ws *websocket.Conn) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case htmlUpsetThread := <-h.upsetThreadChan:
			var p []byte
			var err error

			var buff bytes.Buffer
			upsetThreadTemplate.Execute(&buff, htmlUpsetThread)
			p = buff.Bytes()

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func (h *WebSockerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go h.writer(ws)
	reader(ws)
}
