package mapper

import "gg/domain"

func SetToUpsetThreadItem(set domain.Set, category string) domain.UpsetThreadItem {
	return domain.UpsetThreadItem{
		Id:                set.Id,
		WinnersName:       set.Winner.Name,
		WinnersCharacters: set.GetWinnerCharacterSelections(),
		WinnersSeed:       set.Winner.InitialSeed,
		Score:             set.Score,
		LosersName:        set.Loser.Name,
		LosersCharacters:  set.GetLoserCharacterSelections(),
		IsWinnersBracket:  set.IsWinnersBracket(),
		LosersSeed:        set.Loser.InitialSeed,
		LosersPlacement:   set.LosersPlacement,
		UpsetFactor:       set.UpsetFactor,
		CompletedAt:       set.CompletedAt,
		Category:          category,
	}
}
