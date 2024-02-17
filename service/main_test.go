package service

import (
	"encoding/json"
	"gg/client/startgg"
	"gg/data"
	"os"
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
