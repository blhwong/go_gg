package service

import (
	"fmt"
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"gg/mapper"
	"math"
	"sort"
	"strconv"
	"time"
)

type ServiceInterface interface {
	ToDomainSet(node startgg.Node) domain.Set
	GetSetsFromAPI(slug string) *[]domain.Set
	GetUpsetThread(sets []domain.Set) *domain.UpsetThread
	GetUpsetThreadDB(slug string) *domain.UpsetThread
	SubmitToSubreddit()
	AddSets(slug string, upsetThread *domain.UpsetThread)
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
	lPlacement := 0
	for _, entrant := range entrants {
		if entrant.Id != node.WinnerId {
			lPlacement = entrant.Placement
		}
	}
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
		lPlacement,
		node.WinnerId,
		entrants,
		&games,
		node.CompletedAt,
	)
}

func (s *Service) GetSetsFromAPI(slug string) *[]domain.Set {
	page := 1
	var sets []domain.Set
	for {
		time.Sleep(800 * time.Millisecond)
		res := s.StartGGClient.GetEvent(slug, page)
		if res.Errors != nil {
			panic(res.Errors)
		}
		totalPages := res.Data.Event.Sets.PageInfo.TotalPages
		fmt.Printf("page: %v totalPage: %v\n", page, totalPages)
		if page > totalPages {
			break
		}
		page++
		for _, node := range res.Data.Event.Sets.Nodes {
			sets = append(sets, s.ToDomainSet(node))
		}
	}
	return &sets
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
		losersUpsetThreadItems = append(losersUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "losers"))
	}
	for _, set := range notables {
		notablesUpsetThreadItems = append(notablesUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "notables"))
	}
	for _, set := range dqs {
		dqsUpsetThreadItems = append(dqsUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "dqs"))
	}
	for _, set := range other {
		otherUpsetThreadItems = append(otherUpsetThreadItems, mapper.SetToUpsetThreadItem(set, "other"))
	}
	return &domain.UpsetThread{
		Winners:  winnersUpsetThreadItems,
		Losers:   losersUpsetThreadItems,
		Notables: notablesUpsetThreadItems,
		DQs:      dqsUpsetThreadItems,
		Other:    otherUpsetThreadItems,
	}
}

func (s *Service) GetUpsetThreadDB(slug string) *domain.UpsetThread {
	setMapping := s.DBService.GetSets(slug)
	var winners, losers, notables, dqs, other []domain.UpsetThreadItem
	for setId, set := range *setMapping {
		upsetThreadItem := mapper.DBSetToUpsetThreadItem(setId, set)
		category := upsetThreadItem.Category
		if category == "winners" {
			winners = append(winners, *upsetThreadItem)
		} else if category == "losers" {
			losers = append(losers, *upsetThreadItem)
		} else if category == "notables" {
			notables = append(notables, *upsetThreadItem)
		} else if category == "dqs" {
			dqs = append(dqs, *upsetThreadItem)
		} else {
			other = append(other, *upsetThreadItem)
		}
	}
	sort.Slice(winners, func(i, j int) bool {
		return winners[i].UpsetFactor > winners[j].UpsetFactor
	})
	sort.Slice(losers, func(i, j int) bool {
		return losers[i].UpsetFactor > losers[j].UpsetFactor
	})
	sort.Slice(notables, func(i, j int) bool {
		return notables[i].UpsetFactor < notables[j].UpsetFactor
	})
	sort.Slice(dqs, func(i, j int) bool {
		return dqs[i].UpsetFactor > dqs[j].UpsetFactor
	})
	sort.Slice(other, func(i, j int) bool {
		return other[i].UpsetFactor > other[j].UpsetFactor
	})
	return &domain.UpsetThread{
		Winners:  winners,
		Losers:   losers,
		Notables: notables,
		DQs:      dqs,
		Other:    other,
	}
}

func (s *Service) SubmitToSubreddit() {

}

func (s *Service) AddSets(slug string, upsetThread *domain.UpsetThread) {
	setMapping := make(map[string]string, 0)
	for _, s := range upsetThread.Winners {
		setMapping[s.Id] = mapper.UpsetThreadItemToDBSet(s)
	}
	for _, s := range upsetThread.Losers {
		setMapping[s.Id] = mapper.UpsetThreadItemToDBSet(s)
	}
	for _, s := range upsetThread.Notables {
		setMapping[s.Id] = mapper.UpsetThreadItemToDBSet(s)
	}
	for _, s := range upsetThread.DQs {
		setMapping[s.Id] = mapper.UpsetThreadItemToDBSet(s)
	}
	for _, s := range upsetThread.Other {
		setMapping[s.Id] = mapper.UpsetThreadItemToDBSet(s)
	}
	s.DBService.AddSets(slug, &setMapping)
}
