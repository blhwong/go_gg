package db

import (
	"context"
	"gg/client/startgg"
	"testing"

	"github.com/go-redis/redismock/v9"
)

var db, mock = redismock.NewClientMock()

var redisDBService = NewRedisDBService(*db, context.TODO())

func TestIsCharactersLoadedNotFound(t *testing.T) {
	mock.ExpectHGet("characters:game/ultimate", "is_character_loaded").RedisNil()
	isLoaded := redisDBService.IsCharactersLoaded("game/ultimate")

	if isLoaded {
		t.Errorf("Expected isLoaded=false, got %v\n", isLoaded)
	}
}

func TestIsCharactersLoadedFound(t *testing.T) {
	mock.ExpectHGet("characters:game/ultimate", "is_character_loaded").SetVal("1")
	isLoaded := redisDBService.IsCharactersLoaded("game/ultimate")

	if !isLoaded {
		t.Errorf("Expected isLoaded=true, got %v\n", isLoaded)
	}
}

func TestGetCharacterName(t *testing.T) {
	mock.ExpectHGet("characters:game/ultimate", "character:123").SetVal("Cloud")
	character := redisDBService.GetCharacterName(123, "game/ultimate")

	if character != "Cloud" {
		t.Errorf("Expected character=Cloud, got %v\n", character)
	}
}

func TestAddCharacters(t *testing.T) {
	mock.ExpectHSet("characters:game/ultimate", "character:123", "Cloud").SetVal(1)
	redisDBService.AddCharacters([]startgg.Character{{Id: 123, Name: "Cloud"}}, "game/ultimate")
}

func TestSetIsCharactersLoaded(t *testing.T) {
	mock.ExpectHSet("characters:game/ultimate", "is_character_loaded", "1").SetVal(1)
	redisDBService.SetIsCharactersLoaded("game/ultimate")
}

func TestAddSets(t *testing.T) {
	sets := map[string]string{
		"123": "hello_how_are_you",
		"456": "fine_how_about_you",
	}
	mock.ExpectHSet("event:tournament/supernova-2024/event/ultimate-1v1-singles_sets", "123", "hello_how_are_you").SetVal(1)
	mock.ExpectHSet("event:tournament/supernova-2024/event/ultimate-1v1-singles_sets", "456", "fine_how_about_you").SetVal(1)
	redisDBService.AddSets("tournament/supernova-2024/event/ultimate-1v1-singles", &sets)
}

func TestGetSets(t *testing.T) {
	storedSets := map[string]string{
		"123": "hello_how_are_you",
		"456": "fine_how_about_you",
	}
	mock.ExpectHGetAll("event:tournament/supernova-2024/event/ultimate-1-1-singles_sets").SetVal(storedSets)
	sets := redisDBService.GetSets("tournament/supernova-2024/event/ultimate-1v1-singles")

	for key, val := range *sets {
		if val != storedSets[key] {
			t.Errorf("Expected val=%s for key=%s, got %s", storedSets[key], key, val)
		}
	}
}
