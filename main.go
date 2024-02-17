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
	"time"
)

func process(slug, title, subreddit, file string) {
	var service service.ServiceInterface = &service.Service{
		DBService: data.NewInMemoryDBService(),
		StartGGClient: &startgg.Client{
			GraphQLClient: &graphql.Client{
				Url:      os.Getenv("START_GG_API_URL"),
				ApiToken: os.Getenv("START_GG_API_KEY"),
			},
		},
	}

	var sets []domain.Set

	if file != "" {
		fmt.Println("Using file data")
		file, err := os.ReadFile(file)
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
		sets = *service.GetSetsFromAPI(slug)
	}
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].UpsetFactor > sets[j].UpsetFactor
	})
	upsetThread := service.GetUpsetThread(sets)
	service.AddSets(slug, upsetThread)
	savedUpsetThread := service.GetUpsetThreadDB(slug)
	md := mapper.ToMarkdown(savedUpsetThread, slug)
	outputName := fmt.Sprintf("output/%v %s.md", time.Now().UnixMilli(), title)
	outputFile, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	l, err := outputFile.WriteString(md)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v bytes written\n", l)

}

func main() {
	slugPtr := flag.String("slug", "", "Slug.")
	titlePtr := flag.String("title", "", "Title.")
	subredditPtr := flag.String("subreddit", "", "Subreddit.")
	filePtr := flag.String("file", "", "File.")
	frequencyMinutesPtr := flag.Int("frequency_minutes", 0, "Frequency minutes.")
	flag.Parse()

	fmt.Printf("slugPtr: %s, titlePtr: %s, subredditPtr: %s, filePtr: %s, frequencyMinutesPtr: %v\n", *slugPtr, *titlePtr, *subredditPtr, *filePtr, *frequencyMinutesPtr)
	process(*slugPtr, *titlePtr, *subredditPtr, *filePtr)
}
