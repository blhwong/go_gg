package mapper

import (
	"fmt"
	"gg/domain"
	"strconv"
	"strings"
	"time"
)

func toLineItem(item domain.UpsetThreadItem) string {
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
	text := strings.Join(words, " ")
	if item.UpsetFactor >= 4 {
		return "**" + text + "**"
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
	winners := strings.Join(winnersItems, "  \n")
	losers := strings.Join(losersItems, "  \n")
	notables := strings.Join(notablesItems, "  \n")
	dqs := strings.Join(dqItems, "  \n")
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
