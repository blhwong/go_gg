package data

import (
	"context"
	"gg/client/startgg"
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

func (r *RedisDBService) IsCharactersLoaded() bool {
	val, err := r.rdb.Get(r.ctx, "is_character_loaded").Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		panic(err)
	}
	return val == "1"
}

func (r *RedisDBService) GetCharacterName(key int) string {
	val, err := r.rdb.Get(r.ctx, "character:"+strconv.Itoa(key)).Result()
	if err != nil {
		panic(err)
	}
	return val
}

func (r *RedisDBService) AddCharacters(characters []startgg.Character) {
	for _, character := range characters {
		err := r.rdb.Set(r.ctx, "character:"+strconv.Itoa(character.Id), character.Name, 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func (r *RedisDBService) SetIsCharactersLoaded() {
	err := r.rdb.Set(r.ctx, "is_character_loaded", "1", 0).Err()
	if err != nil {
		panic(err)
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
		panic(err)
	}
}

func (r *RedisDBService) GetSets(slug string) *map[string]string {
	setMapping := r.rdb.HGetAll(r.ctx, "event:"+slug+"_sets").Val()
	return &setMapping
}
