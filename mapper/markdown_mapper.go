package mapper

import (
	"fmt"
	"gg/domain"
	"strings"
	"time"
)

func toLineItem(item domain.UpsetThreadItem) string {
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
		words = append(words, fmt.Sprintf("%v, out at %v", losersSeed, item.LosersPlacement))
	}
	if item.UpsetFactor > 0 {
		words = append(words, fmt.Sprintf("- Upset Factor %v", item.UpsetFactor))
	}
	text := strings.Join(words, " ")
	if item.UpsetFactor >= 4 {
		return fmt.Sprintf("**%s**", text)
	}
	return text
}

func toDQLineItem(item domain.UpsetThreadItem) string {
	return item.LosersName
}

func ToMarkdown(upsetThread *domain.UpsetThread, slug string) string {
	var winnersItems, losersItems, notablesItems, dqItems []string
	for _, s := range upsetThread.Winners {
		winnersItems = append(winnersItems, toLineItem(s))
	}
	for _, s := range upsetThread.Losers {
		losersItems = append(losersItems, toLineItem(s))
	}
	for _, s := range upsetThread.Notables {
		notablesItems = append(notablesItems, toLineItem(s))
	}
	for _, s := range upsetThread.DQs {
		dqItems = append(dqItems, toDQLineItem(s))
	}
	winners := strings.Join(winnersItems, "\n")
	losers := strings.Join(losersItems, "\n")
	notables := strings.Join(notablesItems, "\n")
	dqs := strings.Join(dqItems, "\n")
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	currentTime := time.Now().In(location)
	t := currentTime.Format("01/02/2006 03:04pm MST")
	return fmt.Sprintf(`[Bracket](https;//start.gg/%s)

# Winners
%s

# Losers
%s

# Notables
%s

# DQs
%s

*Last updated at: %s*
`, slug, winners, losers, notables, dqs, t)
}
