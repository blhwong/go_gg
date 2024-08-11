package mapper

import (
	"encoding/json"
	"gg/domain"
	"log"
)

func DBSetToUpsetThreadItem(setId, set string) *domain.UpsetThreadItem {
	arr := []interface{}{}
	err := json.Unmarshal([]byte(set), &arr)
	if err != nil {
		log.Fatalf("Error while unmarshaling to upset thread item. e=%s\n", err)
	}
	score := arr[3].(string)
	return &domain.UpsetThreadItem{
		WinnersName:       arr[0].(string),
		WinnersCharacters: arr[1].(string),
		WinnersSeed:       int(arr[2].(float64)),
		Score:             &score,
		LosersName:        arr[4].(string),
		LosersCharacters:  arr[5].(string),
		IsWinnersBracket:  arr[6].(bool),
		LosersSeed:        int(arr[7].(float64)),
		LosersPlacement:   int(arr[8].(float64)),
		UpsetFactor:       int(arr[9].(float64)),
		CompletedAt:       int(arr[10].(float64)),
		Category:          arr[11].(string),
	}
}

func UpsetThreadItemToDBSet(item domain.UpsetThreadItem) string {
	res, err := json.Marshal([]interface{}{
		item.WinnersName,
		item.WinnersCharacters,
		item.WinnersSeed,
		item.Score,
		item.LosersName,
		item.LosersCharacters,
		item.IsWinnersBracket,
		item.LosersSeed,
		item.LosersPlacement,
		item.UpsetFactor,
		item.CompletedAt,
		item.Category,
	})
	if err != nil {
		log.Fatalf("Error while marshaling to db set. e=%s\n", err)
	}
	return string(res)
}
