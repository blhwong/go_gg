package mapper

import (
	"fmt"
	"gg/domain"
	"strings"
	"time"
)

func toLineItemHTML(item domain.UpsetThreadItem) *domain.UpsetThreadItemHTML {
	words := []string{item.WinnersName}
	if len(item.WinnersCharacters) > 0 {
		words = append(words, fmt.Sprintf("(%s)", item.WinnersCharacters))
	}
	words = append(words, fmt.Sprintf("(seed %v)", item.WinnersSeed))
	words = append(words, *item.Score)
	words = append(words, item.LosersName)
	if len(item.LosersCharacters) > 0 {
		words = append(words, fmt.Sprintf("(%s)", item.LosersCharacters))
	}
	losersSeed := fmt.Sprintf("(seed %v)", item.LosersSeed)
	if item.IsWinnersBracket {
		words = append(words, losersSeed)
	} else {
		words = append(words, fmt.Sprintf("%v, out at %s", losersSeed, getOrdinal(item.LosersPlacement)))
	}
	if item.UpsetFactor > 0 {
		words = append(words, fmt.Sprintf("- Upset Factor %v", item.UpsetFactor))
	}
	content := strings.Join(words, " ")
	var bold bool
	if item.UpsetFactor >= 4 {
		bold = true
	}
	return &domain.UpsetThreadItemHTML{
		Content: content,
		Bold:    bold,
	}
}

func toDQLineItemHTML(item domain.UpsetThreadItem) *domain.UpsetThreadItemHTML {
	return &domain.UpsetThreadItemHTML{
		Content: item.LosersName,
	}
}

func ToHTML(upsetThread *domain.UpsetThread, host string) *domain.UpsetThreadHTML {
	var winners, losers, notables, dqs []*domain.UpsetThreadItemHTML
	for _, s := range upsetThread.Winners {
		winners = append(winners, toLineItemHTML(s))
	}
	for _, s := range upsetThread.Losers {
		losers = append(losers, toLineItemHTML(s))
	}
	for _, s := range upsetThread.Notables {
		notables = append(notables, toLineItemHTML(s))
	}
	for _, s := range upsetThread.DQs {
		dqs = append(dqs, toDQLineItemHTML(s))
	}
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	currentTime := time.Now().In(location)
	lastUpdatedAt := currentTime.Format("01/02/2006 03:04pm MST")
	return &domain.UpsetThreadHTML{
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
