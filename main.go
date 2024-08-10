package main

import (
	"flag"
	"fmt"
	"gg/client/startgg"
	"gg/data"
	"gg/mapper"
	"gg/service"
	"log"
	"net/http"
	"os"
	"strconv"
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

	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
)

var (
	addr                = flag.String("addr", ":8080", "http service address")
	homeTempl           = template.Must(template.New("").Parse(homeHTML))
	upsetThreadTemplate = template.Must(template.ParseFiles("template/upset-thread.html"))
	filename            = "output/1723309578401 Supernova Ultimate Singles Upset Thread.md"
	upgrader            = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type HomeHandler struct {
	service      service.ServiceInterface
	slugPtr      *string
	titlePtr     *string
	subredditPtr *string
	filePtr      *string
}

type TestHandler struct{}

func main() {
	slugPtr := flag.String("slug", "", "Slug.")
	titlePtr := flag.String("title", "", "Title.")
	subredditPtr := flag.String("subreddit", "", "Subreddit.")
	filePtr := flag.String("file", "", "File.")
	frequencyMinutesPtr := flag.Int("frequency_minutes", 0, "Frequency minutes.")
	flag.Parse()

	fmt.Printf("slugPtr: %s, titlePtr: %s, subredditPtr: %s, filePtr: %s, frequencyMinutesPtr: %v\n", *slugPtr, *titlePtr, *subredditPtr, *filePtr, *frequencyMinutesPtr)
	var service service.ServiceInterface = service.NewService(
		data.NewRedisDBService(),
		startgg.NewClient(os.Getenv("START_GG_API_URL"), os.Getenv("START_GG_API_KEY"), &http.Client{}),
		&service.FileReaderWriter{},
	)
	homeHandler := HomeHandler{
		service:      service,
		slugPtr:      slugPtr,
		titlePtr:     titlePtr,
		subredditPtr: subredditPtr,
		filePtr:      filePtr,
	}

	// testHandler := TestHandler{}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", &homeHandler)
	// http.Handle("/", &testHandler)
	// http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	http.ListenAndServe(":8080", nil)
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	upsetThread := h.service.Process(*h.slugPtr, *h.titlePtr, *h.subredditPtr, *h.filePtr)
	htmlUpsetThread := mapper.ToHTML(upsetThread, r.Host)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	upsetThreadTemplate.Execute(w, &htmlUpsetThread)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p, lastMod, err := readFileIfModified(time.Time{})
	if err != nil {
		p = []byte(err.Error())
		lastMod = time.Unix(0, 0)
	}
	var v = struct {
		Host    string
		Data    string
		LastMod string
	}{
		r.Host,
		string(p),
		strconv.FormatInt(lastMod.UnixNano(), 16),
	}
	homeTempl.Execute(w, &v)
}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p, lastMod, err := readFileIfModified(time.Time{})
	if err != nil {
		p = []byte(err.Error())
		lastMod = time.Unix(0, 0)
	}
	var v = struct {
		Host    string
		Data    string
		LastMod string
	}{
		r.Host,
		string(p),
		strconv.FormatInt(lastMod.UnixNano(), 16),
	}
	homeTempl.Execute(w, &v)
}

func readFileIfModified(lastMod time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, lastMod, err
	}
	if !fi.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}
	p, err := os.ReadFile(filename)
	if err != nil {
		return nil, fi.ModTime(), err
	}
	return p, fi.ModTime(), nil
}

func writer(ws *websocket.Conn, lastMod time.Time) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
			var p []byte
			var err error

			p, lastMod, err = readFileIfModified(lastMod)

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
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	var lastMod time.Time
	if n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64); err == nil {
		lastMod = time.Unix(0, n)
	}

	go writer(ws, lastMod)
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>WebSocket Example</title>
    </head>
    <body>
        <pre id="fileData">{{.Data}}</pre>
        <script type="text/javascript">
            (function() {
                var data = document.getElementById("fileData");
                var conn = new WebSocket("ws://{{.Host}}/ws?lastMod={{.LastMod}}");
                conn.onclose = function(evt) {
                    data.textContent = 'Connection closed';
                }
                conn.onmessage = function(evt) {
                    console.log('file updated');
                    data.textContent = evt.data;
                }
            })();
        </script>
    </body>
</html>
`
