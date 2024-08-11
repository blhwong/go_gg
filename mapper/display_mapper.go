package mapper

import (
	"gg/domain"
	"log"
	"strconv"
	"strings"
	"time"
)

func toLineItemDisplay(item domain.UpsetThreadItem) *domain.UpsetThreadItemDisplay {
	words := []string{item.WinnersName}
	if len(item.WinnersCharacters) > 0 {
		words = append(words, "("+item.WinnersCharacters+")")
	}
	words = append(words, "(seed "+strconv.Itoa(item.WinnersSeed)+")")
	words = append(words, *item.Score)
	words = append(words, item.LosersName)
	if len(item.LosersCharacters) > 0 {
		words = append(words, "("+item.LosersCharacters+")")
	}
	losersSeed := "(seed " + strconv.Itoa(item.LosersSeed) + ")"
	if item.IsWinnersBracket {
		words = append(words, losersSeed)
	} else {
		words = append(words, losersSeed+", out at "+getOrdinal(item.LosersPlacement))
	}
	if item.UpsetFactor > 0 {
		words = append(words, "- Upset Factor "+strconv.Itoa(item.UpsetFactor))
	}
	content := strings.Join(words, " ")
	var bold bool
	if item.UpsetFactor >= 4 {
		bold = true
	}
	return &domain.UpsetThreadItemDisplay{
		Content: content,
		Bold:    bold,
	}
}

func toDQLineItemDisplay(item domain.UpsetThreadItem) *domain.UpsetThreadItemDisplay {
	return &domain.UpsetThreadItemDisplay{
		Content: item.LosersName,
	}
}

func ToDisplay(upsetThread *domain.UpsetThread, host string) *domain.UpsetThreadDisplay {
	var winners, losers, notables, dqs []*domain.UpsetThreadItemDisplay
	for _, s := range upsetThread.Winners {
		winners = append(winners, toLineItemDisplay(s))
	}
	for _, s := range upsetThread.Losers {
		losers = append(losers, toLineItemDisplay(s))
	}
	for _, s := range upsetThread.Notables {
		notables = append(notables, toLineItemDisplay(s))
	}
	for _, s := range upsetThread.DQs {
		dqs = append(dqs, toDQLineItemDisplay(s))
	}
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatalf("Error while loading location. e=%s\n", err)
	}
	currentTime := time.Now().In(location)
	lastUpdatedAt := currentTime.Format("01/02/2006 03:04pm MST")
	return &domain.UpsetThreadDisplay{
		Host:          host,
		Title:         upsetThread.Title,
		Slug:          upsetThread.Slug,
		LastUpdatedAt: lastUpdatedAt,
		Winners:       winners,
		Losers:        losers,
		Notables:      notables,
		DQs:           dqs,
	}
}
