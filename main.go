package main

import (
	"flag"
	"fmt"
	"gg/client/startgg"
	"gg/data"
	"gg/service"
	"net/http"
	"os"
)

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
	service.Process(*slugPtr, *titlePtr, *subredditPtr, *filePtr)
}
