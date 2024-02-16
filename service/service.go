package service

import (
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"gg/mapper"
	"math"
	"sort"
	"strconv"
)

type ServiceInterface interface {
	ToDomainSet(node startgg.Node) domain.Set
	GetEvent()
	GetUpsetThread(sets []domain.Set) *domain.UpsetThread
	GetUpsetThreadDB()
	SubmitToSubreddit()
}

type Service struct {
	DBService     data.DBServiceInterface
	StartGGClient startgg.ClientInterface
}

func toDomainEntrant(entrant startgg.Entrant) domain.Entrant {
	return domain.Entrant{
		Id:          entrant.Id,
		Name:        entrant.Name,
		InitialSeed: entrant.InitialSeedNum,
		Placement:   entrant.Standing.Placement,
		IsFinal:     entrant.Standing.IsFinal,
	}
}

func (s *Service) getCharacterName(key int) string {
	if !s.DBService.IsCharactersLoaded() {
		res := s.StartGGClient.GetCharacters()
		s.DBService.AddCharacters(res.Data.VideoGame.Characters)
		s.DBService.SetIsCharactersLoaded()
	}
	return s.DBService.GetCharacterName(key)
}

func (s *Service) toDomainCharacter(selectionType string, value int) *domain.Character {
	if selectionType != "CHARACTER" {
		return nil
	}
	return &domain.Character{
		Value: value,
		Name:  s.getCharacterName(value),
	}
}

func (s *Service) toDomainSelection(selection startgg.Selection) domain.Selection {
	return domain.Selection{
		Entrant:   toDomainEntrant(selection.Entrant),
		Character: s.toDomainCharacter(selection.SelectionType, selection.SelectionValue),
	}
}

func (s *Service) toDomainGame(game startgg.Game) domain.Game {
	var selections []domain.Selection
	if game.Selections != nil {
		for _, selection := range game.Selections {
			selections = append(selections, s.toDomainSelection(selection))
		}
	}
	return domain.Game{
		Id:         game.Id,
		WinnerId:   game.WinnerId,
		Selections: selections,
	}
}

func (s *Service) ToDomainSet(node startgg.Node) domain.Set {
	entrants := make([]domain.Entrant, 0)
	for _, slot := range node.Slots {
		entrants = append(entrants, toDomainEntrant(slot.Entrant))
	}
	var games []domain.Game
	if node.Games != nil {
		for _, game := range node.Games {
			games = append(games, s.toDomainGame(game))
		}
	}
	return *domain.NewSet(
		strconv.Itoa(node.Id),
		node.DisplayScore,
		&node.FullRoundText,
		node.TotalGames,
		node.Round,
		node.LPlacement,
		node.WinnerId,
		entrants,
		&games,
		node.CompletedAt,
	)
}

func (s *Service) GetEvent() {

}

func applyFilter(upsetFactor, winnerInitialSeed, loserInitialSeed int, isDQ bool, score *string, minUpsetFactor, maxSeed int, includeDQ bool) bool {
	fulfillsMinUpsetFactor := upsetFactor >= minUpsetFactor
	fulfillsNotDQ := !isDQ || includeDQ
	fulfillsMaxSeed := winnerInitialSeed <= maxSeed || loserInitialSeed <= maxSeed
	return fulfillsMinUpsetFactor && fulfillsNotDQ && fulfillsMaxSeed && score != nil
}

func (s *Service) GetUpsetThread(sets []domain.Set) *domain.UpsetThread {
	var winners, losers, notables, dqs, other []domain.Set
	for _, set := range sets {
		if set.IsWinnersBracket() && applyFilter(
			set.UpsetFactor,
			set.Winner.InitialSeed,
			set.Loser.InitialSeed,
			set.IsDQ(),
			set.Score,
			1,
			50,
			false,
		) {
			winners = append(winners, set)
		} else if !set.IsWinnersBracket() && applyFilter(
			set.UpsetFactor,
			set.Winner.InitialSeed,
			set.Loser.InitialSeed,
			set.IsDQ(),
			set.Score,
			1,
			50,
			false,
		) {
			losers = append(losers, set)
		} else if set.IsDQAndOut() && applyFilter(
			set.UpsetFactor,
			set.Winner.InitialSeed,
			set.Loser.InitialSeed,
			true,
			set.Score,
			math.MinInt,
			math.MaxInt,
			true,
		) {
			dqs = append(dqs, set)
		} else if set.IsNotable() && applyFilter(
			-set.UpsetFactor,
			set.Winner.InitialSeed,
			set.Loser.InitialSeed,
			set.IsDQ(),
			set.Score,
			3,
			50,
			false,
		) {
			notables = append(notables, set)
		} else {
			other = append(other, set)
		}
	}
	sort.Slice(notables, func(i, j int) bool {
		return notables[i].UpsetFactor < notables[j].UpsetFactor
	})
	var winnersUpsetThreadItems, losersUpsetThreadItems, notablesUpsetThreadItems, dqsUpsetThreadItems, otherUpsetThreadItems []domain.UpsetThreadItem
	for _, set := range winners {
		winnersUpsetThreadItems = append(winnersUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "winners"))
	}
	for _, set := range losers {
		losersUpsetThreadItems = append(losersUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "winners"))
	}
	for _, set := range notables {
		notablesUpsetThreadItems = append(notablesUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "winners"))
	}
	for _, set := range dqs {
		dqsUpsetThreadItems = append(dqsUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "winners"))
	}
	for _, set := range other {
		otherUpsetThreadItems = append(otherUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "winners"))
	}
	return &domain.UpsetThread{
		Winners:  winnersUpsetThreadItems,
		Losers:   losersUpsetThreadItems,
		Notables: notablesUpsetThreadItems,
		DQs:      dqsUpsetThreadItems,
		Other:    otherUpsetThreadItems,
	}
}

func (s *Service) GetUpsetThreadDB() {

}
func (s *Service) SubmitToSubreddit() {

}
