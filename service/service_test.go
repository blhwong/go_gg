package service

import (
	"encoding/json"
	"gg/client/startgg"
	"gg/db"
	"gg/domain"
	"gg/mapper"
	"log"
	"os"
	"slices"
	"testing"
)

type FakeStartGGClient struct{}

func (f *FakeStartGGClient) GetCharacters(slug string) startgg.CharactersResponse {
	data, err := os.ReadFile("../db/characters.json")
	if err != nil {
		log.Fatalf("Error while reading file. e=%s\n", err)
	}
	var charactersResponse startgg.CharactersResponse
	if err := json.Unmarshal(data, &charactersResponse); err != nil {
		log.Fatalf("Error while unmarshaling characters. e=%s\n", err)
	}
	return charactersResponse
}

func (f *FakeStartGGClient) GetEvent(slug string, page int) (*startgg.EventResponse, error) {
	return &startgg.EventResponse{}, nil
}

type FakeFileReaderWriter struct{}

func (f *FakeFileReaderWriter) ReadFile(fileName string) []byte {
	data, err := os.ReadFile("../db/test_data.json")
	if err != nil {
		log.Fatalf("Error while reading test_data. e=%s\n", err)
	}
	return data
}

func (f *FakeFileReaderWriter) WriteString(filename, data string) {
}

var fakeStartGGClient startgg.ClientInterface = &FakeStartGGClient{}
var fakeFileReaderWriter FileInterface = &FakeFileReaderWriter{}

var service = NewService(
	db.NewInMemoryDBService(),
	fakeStartGGClient,
	fakeFileReaderWriter,
)

var slug = "tournament/smash-factor-x/event/smash-bros-ultimate-singles"

func TestServiceSetsFromFile(t *testing.T) {
	service.Process(
		slug,
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"db/startgg_data.json",
		"game/ultimate",
	)
}

func TestServiceSetsFromAPI(t *testing.T) {
	service.Process(
		slug,
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"",
		"game/ultimate",
	)
}

func TestSort(t *testing.T) {
	var winners []domain.UpsetThreadItem = []domain.UpsetThreadItem{
		{UpsetFactor: 5, WinnersName: "f"},
		{UpsetFactor: 6, WinnersName: "c"},
		{UpsetFactor: 7, WinnersName: "b"},
		{UpsetFactor: 6, WinnersName: "d"},
		{UpsetFactor: 5, WinnersName: "e"},
		{UpsetFactor: 8, WinnersName: "a"},
	}

	slices.SortFunc(winners, func(i, j domain.UpsetThreadItem) int {
		return defaultSort(winners, i, j)
	})

	var expectedWinners []domain.UpsetThreadItem = []domain.UpsetThreadItem{
		{UpsetFactor: 8, WinnersName: "a"},
		{UpsetFactor: 7, WinnersName: "b"},
		{UpsetFactor: 6, WinnersName: "c"},
		{UpsetFactor: 6, WinnersName: "d"},
		{UpsetFactor: 5, WinnersName: "e"},
		{UpsetFactor: 5, WinnersName: "f"},
	}

	for i, winner := range winners {
		expected := expectedWinners[i]
		if expected.UpsetFactor != winner.UpsetFactor {
			t.Errorf("Failed upset factor. Expected %v, got %v", expected.UpsetFactor, winner.UpsetFactor)
		}
		if expected.WinnersName != winner.WinnersName {
			t.Errorf("Failed winners name. Expected %v, got %v", expected.WinnersName, winner.WinnersName)
		}
	}
}

func TestMarkdownMapper(t *testing.T) {
	upsetThread := service.Process(
		slug,
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"db/startgg_data.json",
		"game/ultimate",
	)

	mapper.ToMarkdown(upsetThread, slug)

}

func TestHTMLMapper(t *testing.T) {
	upsetThread := service.Process(
		slug,
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"db/startgg_data.json",
		"game/ultimate",
	)

	mapper.ToHTML(upsetThread, slug)
}
