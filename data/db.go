package data

import (
	"fmt"
	"gg/client/startgg"
	"strings"
)

type DBServiceInterface interface {
	IsCharactersLoaded() bool
	GetCharacterName(key int) string
	AddCharacters(characters []startgg.Character)
	SetIsCharactersLoaded()
	AddSets(slug string, setMapping *map[string]string)
	GetSets(slug string) *map[string]string
}

type InMemoryDBService struct {
	Storage map[string]string
}

func NewInMemoryDBService() *InMemoryDBService {
	return &InMemoryDBService{Storage: make(map[string]string, 0)}
}

func (db *InMemoryDBService) IsCharactersLoaded() bool {
	return db.Storage["is_character_loaded"] == "1"
}

func (db *InMemoryDBService) GetCharacterName(key int) string {
	return db.Storage[fmt.Sprintf("character:%v", key)]
}

func (db *InMemoryDBService) AddCharacters(characters []startgg.Character) {
	for _, character := range characters {
		db.Storage[fmt.Sprintf("character:%v", character.Id)] = character.Name
	}
}

func (db *InMemoryDBService) SetIsCharactersLoaded() {
	db.Storage["is_character_loaded"] = "1"
}

func (db *InMemoryDBService) AddSets(slug string, setMapping *map[string]string) {
	for setId, s := range *setMapping {
		db.AddSet(slug, setId, s)
	}
}

func (db *InMemoryDBService) AddSet(slug string, setId string, set string) {
	db.Storage[fmt.Sprintf("%s_%s", slug, setId)] = set
}

func (db *InMemoryDBService) GetSets(slug string) *map[string]string {
	setMapping := make(map[string]string, 0)
	for key, set := range db.Storage {
		parts := strings.Split(key, "_")
		if parts[0] == slug {
			setMapping[parts[1]] = set
		}
	}
	return &setMapping
}
