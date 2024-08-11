package data

import (
	"gg/client/startgg"
	"strconv"
	"strings"
)

type DBServiceInterface interface {
	IsCharactersLoaded(slug string) bool
	GetCharacterName(key int, slug string) string
	AddCharacters(characters []startgg.Character, slug string)
	SetIsCharactersLoaded(slug string)
	AddSets(slug string, setMapping *map[string]string)
	GetSets(slug string) *map[string]string
}

type InMemoryDBService struct {
	storage map[string]string
}

func NewInMemoryDBService() *InMemoryDBService {
	return &InMemoryDBService{storage: make(map[string]string, 0)}
}

func (db *InMemoryDBService) IsCharactersLoaded(slug string) bool {
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
