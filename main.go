package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gg/client/graphql"
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"gg/mapper"
	"gg/service"
	"os"
	"sort"
)

func main() {
	slugPtr := flag.String("slug", "", "Slug.")
	titlePtr := flag.String("title", "", "Title.")
	subredditPtr := flag.String("subreddit", "", "Subreddit.")
	filePtr := flag.String("file", "", "File.")
	frequencyMinutesPtr := flag.Int("frequency_minutes", 0, "Frequency minutes.")
	flag.Parse()

	fmt.Printf("slugPtr: %s, titlePtr: %s, subredditPtr: %s, filePtr: %s, frequencyMinutesPtr: %v\n", *slugPtr, *titlePtr, *subredditPtr, *filePtr, *frequencyMinutesPtr)

	var service service.ServiceInterface = service.Service{
		DBService: data.NewInMemoryDBService(),
		StartGGClient: startgg.Client{
			GraphQLClient: graphql.Client{
				Url:      os.Getenv("START_GG_API_URL"),
				ApiToken: os.Getenv("START_GG_API_KEY"),
			},
		},
	}

	sets := make([]domain.Set, 0)

	if filePtr != nil {
		fmt.Println("Using file data")
		file, err := os.ReadFile(*filePtr)
		if err != nil {
			panic(err)
		}
		var nodes []startgg.Node
		if err := json.Unmarshal(file, &nodes); err != nil {
			panic(err)
		}
		for _, node := range nodes {
			sets = append(sets, service.ToDomainSet(node))
		}
	} else {
		fmt.Println("Fetching data from startgg")
	}
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].UpsetFactor > sets[j].UpsetFactor
	})
	upsetThread := service.GetUpsetThread(sets)
	md := mapper.ToMarkdown(upsetThread, *slugPtr)

	fmt.Println(md)
}
