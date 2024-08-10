package service

import (
	"encoding/json"
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"os"
	"slices"
	"testing"
)

type FakeStartGGClient struct{}

func (f *FakeStartGGClient) GetCharacters() startgg.CharactersResponse {
	data, err := os.ReadFile("../data/characters.json")
	if err != nil {
		panic(err)
	}
	var charactersResponse startgg.CharactersResponse
	if err := json.Unmarshal(data, &charactersResponse); err != nil {
		panic(err)
	}
	return charactersResponse
}

func (f *FakeStartGGClient) GetEvent(slug string, page int) startgg.EventResponse {
	return startgg.EventResponse{}
}

type FakeFileReaderWriter struct{}

func (f *FakeFileReaderWriter) ReadFile(fileName string) []byte {
	data, err := os.ReadFile("../data/test_data.json")
	if err != nil {
		panic(err)
	}
	return data
}

func (f *FakeFileReaderWriter) WriteString(filename, data string) {
}

var fakeStartGGClient startgg.ClientInterface = &FakeStartGGClient{}
var fakeFileReaderWriter FileInterface = &FakeFileReaderWriter{}

func TestServiceSetsFromFile(t *testing.T) {
	service := NewService(
		data.NewInMemoryDBService(),
		fakeStartGGClient,
		fakeFileReaderWriter,
	)
	service.Process(
		"tournament/smash-factor-x/event/smash-bros-ultimate-singles",
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"data/startgg_data.json",
	)
}

func TestServiceSetsFromAPI(t *testing.T) {
	service := NewService(
		data.NewInMemoryDBService(),
		fakeStartGGClient,
		fakeFileReaderWriter,
	)
	service.Process(
		"tournament/smash-factor-x/event/smash-bros-ultimate-singles",
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"",
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
