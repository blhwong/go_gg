package db

import (
	"context"
	"gg/client/startgg"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisDBService struct {
	rdb redis.Client
	ctx context.Context
}

func NewRedisDBService() *RedisDBService {
	return &RedisDBService{
		rdb: *redis.NewClient(&redis.Options{Addr: os.Getenv("localhost:6379")}),
		ctx: context.Background(),
	}
}

func (r *RedisDBService) IsCharactersLoaded(slug string) bool {
	val, err := r.rdb.HGet(r.ctx, "characters:"+slug, "is_character_loaded").Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		log.Fatalf("Error on getting character is loaded. e=%s\n", err)
	}
	return val == "1"
}

func (r *RedisDBService) GetCharacterName(key int, slug string) string {
	val, err := r.rdb.HGet(r.ctx, "characters:"+slug, "character:"+strconv.Itoa(key)).Result()
	if err != nil {
		log.Fatalf("Error while getting character name. e=%s\n", err)
	}
	return val
}

func (r *RedisDBService) AddCharacters(characters []startgg.Character, slug string) {
	for _, character := range characters {
		err := r.rdb.HSet(r.ctx, "characters:"+slug, "character:"+strconv.Itoa(character.Id), character.Name).Err()
		if err != nil {
			log.Fatalf("Error while adding character. e=%s\n", err)
		}
	}
}

func (r *RedisDBService) SetIsCharactersLoaded(slug string) {
	err := r.rdb.HSet(r.ctx, "characters:"+slug, "is_character_loaded", "1").Err()
	if err != nil {
		log.Fatalf("Error while setting character is loaded. e=%s\n", err)
	}
}

func (r *RedisDBService) AddSets(slug string, setMapping *map[string]string) {
	for setId, s := range *setMapping {
		r.AddSet(slug, setId, s)
	}
}

func (r *RedisDBService) AddSet(slug string, setId string, set string) {
	err := r.rdb.HSet(r.ctx, "event:"+slug+"_sets", setId, set).Err()
	if err != nil {
		log.Fatalf("Error while adding set. e=%s\n", err)
	}
}

func (r *RedisDBService) GetSets(slug string) *map[string]string {
	setMapping := r.rdb.HGetAll(r.ctx, "event:"+slug+"_sets").Val()
	return &setMapping
}
