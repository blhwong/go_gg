package service

import (
	"encoding/json"
	"fmt"
	"gg/client/startgg"
	"gg/domain"
	"gg/mapper"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
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

type InMemoryDBService struct {
	storage map[string]string
}

func NewInMemoryDBService() *InMemoryDBService {
	return &InMemoryDBService{storage: make(map[string]string, 0)}
}

func (db *InMemoryDBService) IsCharactersLoaded(slug string) bool {
	fmt.Println("Inside is characters loaded")
	return db.storage[slug+"_is_character_loaded"] == "1"
}

func (db *InMemoryDBService) GetCharacterName(key int, slug string) string {
	return db.storage[slug+"_"+strconv.Itoa(key)]
}

func (db *InMemoryDBService) AddCharacters(characters []startgg.Character, slug string) {
	for _, character := range characters {
		db.storage[slug+"_"+strconv.Itoa(character.Id)] = character.Name
	}
}

func (db *InMemoryDBService) SetIsCharactersLoaded(slug string) {
	db.storage[slug+"_is_character_loaded"] = "1"
}

func (db *InMemoryDBService) AddSets(slug string, setMapping *map[string]string) {
	for setId, s := range *setMapping {
		db.AddSet(slug, setId, s)
	}
}

func (db *InMemoryDBService) AddSet(slug string, setId string, set string) {
	db.storage[slug+"_"+setId] = set
}

func (db *InMemoryDBService) GetSets(slug string) *map[string]string {
	setMapping := make(map[string]string, 0)
	for key, set := range db.storage {
		parts := strings.Split(key, "_")
		if parts[0] == slug {
			setMapping[parts[1]] = set
		}
	}
	return &setMapping
}

var service = NewService(
	NewInMemoryDBService(),
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

func TestDisplayMapper(t *testing.T) {
	upsetThread := service.Process(
		slug,
		"Smash Factor X Ultimate Singles Upset Thread",
		"",
		"db/startgg_data.json",
		"game/ultimate",
	)

	mapper.ToDisplay(upsetThread, slug)
}
