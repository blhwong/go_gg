package db

import (
	"gg/client/startgg"
)

type DBServiceInterface interface {
	IsCharactersLoaded(slug string) bool
	GetCharacterName(key int, slug string) string
	AddCharacters(characters []startgg.Character, slug string)
	SetIsCharactersLoaded(slug string)
	AddSets(slug string, setMapping *map[string]string)
	GetSets(slug string) *map[string]string
}
