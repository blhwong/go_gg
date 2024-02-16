package data

import (
	"fmt"
	"gg/client/startgg"
)

type DBServiceInterface interface {
	IsCharactersLoaded() bool
	GetCharacterName(key int) string
	AddCharacters(characters []startgg.Character)
	SetIsCharactersLoaded()
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
