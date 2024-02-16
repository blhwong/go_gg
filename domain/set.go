package domain

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type UpsetFactorTable struct {
	Storage [][]int
}

func NewUpsetFactorTable() *UpsetFactorTable {
	offset := 0
	var storage [][]int
	for i := 0; i < 20; i++ {
		var row []int
		for j := 0; j < 20; j++ {
			row = append(row, -j+offset)
		}
		storage = append(storage, row)
		offset++
	}
	return &UpsetFactorTable{Storage: storage}
}

var upsetFactorTable *UpsetFactorTable = NewUpsetFactorTable()

func getTableIdx(seed int) int {
	seeds := []int{769, 513, 385, 257, 193, 129, 97, 65, 49, 33, 25, 17, 13, 9, 7, 5, 4, 3, 2, 1}
	for i, s := range seeds {
		if seed >= s {
			return 19 - i
		}
	}
	return 0
}

func (upsetFactorTable *UpsetFactorTable) GetUpsetFactor(winnerSeed, loserSeed int) int {
	winnerIdx, loserIdx := getTableIdx(winnerSeed), getTableIdx(loserSeed)
	return upsetFactorTable.Storage[winnerIdx][loserIdx]
}

type Entrant struct {
	Id          int
	Name        string
	InitialSeed int
	Placement   int
	IsFinal     bool
}

type Character struct {
	Value int
	Name  string
}

type Selection struct {
	Entrant   Entrant
	Character Character
}

type Game struct {
	Id         int
	WinnerId   int
	Selections []Selection
}

func initSlots(winnerId int, entrants []Entrant) (Entrant, Entrant) {
	winner, loser := entrants[0], entrants[1]
	if winnerId == loser.Id {
		winner, loser = loser, winner
	}
	return winner, loser
}

func initUpsetFactor(winnerSeed, loserSeed int) int {
	return upsetFactorTable.GetUpsetFactor(winnerSeed, loserSeed)
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func initScore(games *[]Game, displayScore string, winner, loser Entrant, totalGames int) *string {
	fmt.Printf("Initializing score. games=%v displayScore=%s totalGames=%v\n", games, displayScore, totalGames)
	if displayScore == "DQ" {
		return &displayScore
	}
	var scoresFromGames *string
	scoreFromDisplayScore := displayScore
	scoreFromDisplayScore = strings.Replace(scoreFromDisplayScore, winner.Name, "", 1)
	scoreFromDisplayScore = strings.Replace(scoreFromDisplayScore, loser.Name, "", 1)
	scoreFromDisplayScore = strings.Replace(scoreFromDisplayScore, " ", "", -1)
	if slices.Contains([]string{"0-2", "0-3", "1-2", "1-3", "2-3"}, scoreFromDisplayScore) {
		scoreFromDisplayScore = reverseString(scoreFromDisplayScore)
	}
	if games != nil {
		winnerScore, loserScore := 0, 0
		for _, game := range *games {
			if game.WinnerId == winner.Id {
				winnerScore++
			} else {
				loserScore++
			}
			s := fmt.Sprintf("%v-%v", winnerScore, loserScore)
			scoresFromGames = &s
		}
	}
	if len(scoreFromDisplayScore) > 0 && scoresFromGames != nil && len(*scoresFromGames) > 0 {
		fmt.Printf("scoreFromDisplayScore=%s scoreFromGames=%s\n", scoreFromDisplayScore, *scoresFromGames)
		numFromDisplayScore, err := strconv.Atoi(string(scoreFromDisplayScore[0]))
		if err != nil {
			panic(err)
		}
		gameScore := *scoresFromGames
		numFromGameScore, err := strconv.Atoi(string(gameScore[0]))
		if err != nil {
			panic(err)
		}
		if numFromDisplayScore > numFromGameScore {
			return &scoreFromDisplayScore
		}
		return scoresFromGames
	}
	if scoresFromGames != nil {
		return scoresFromGames
	}
	return &scoreFromDisplayScore
}

type Set struct {
	Id              string
	DisplayScore    string
	FullRoundText   *string
	Round           int
	LosersPlacement int
	TotalGames      int
	Games           *[]Game
	CompletedAt     int
	Winner          Entrant
	Loser           Entrant
	UpsetFactor     int
	Score           *string
}

func NewSet(identifier string, displayScore string, fullRoundText *string, totalGames int, roundNum int, losersPlacement int, winnerId int, entrants []Entrant, games *[]Game, completedAt int) *Set {
	winner, loser := initSlots(winnerId, entrants)
	upsetFactor := initUpsetFactor(winner.InitialSeed, loser.InitialSeed)
	score := initScore(games, displayScore, winner, loser, totalGames)
	return &Set{
		Id:              identifier,
		DisplayScore:    displayScore,
		FullRoundText:   fullRoundText,
		Round:           roundNum,
		LosersPlacement: losersPlacement,
		TotalGames:      totalGames,
		Games:           games,
		CompletedAt:     completedAt,
		Winner:          winner,
		Loser:           loser,
		UpsetFactor:     upsetFactor,
		Score:           score,
	}
}

func (s *Set) IsWinnersBracket() bool {
	return s.Round > 0
}

func (s *Set) IsDQ() bool {
	return s.DisplayScore == "DQ"
}

func (s *Set) IsDQAndOut() bool {
	return s.IsWinnersBracket() && s.IsDQ()
}

func (s *Set) IsNotable() bool {
	if s.Score == nil {
		return false
	}
	return slices.Contains([]string{"3-2", "2-1"}, *s.Score)
}

func (s *Set) GetCharacterSelections(entrantId int) string {
	if s.Games == nil {
		return ""
	}
	set := make(map[string]bool, 0)
	for _, game := range *s.Games {
		for _, selection := range game.Selections {
			if selection.Entrant.Id == entrantId {
				set[selection.Character.Name] = true
			}
		}
	}
	var selections []string
	for key := range set {
		selections = append(selections, key)
	}
	sort.Strings(selections)
	return strings.Join(selections, ", ")
}
func (s *Set) GetWinnerCharacterSelections() string {
	return s.GetCharacterSelections(s.Winner.Id)
}
func (s *Set) GetLoserCharacterSelections() string {
	return s.GetCharacterSelections(s.Loser.Id)
}
