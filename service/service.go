package service

import (
	"cmp"
	"encoding/json"
	"fmt"
	"gg/client/startgg"
	"gg/data"
	"gg/domain"
	"gg/mapper"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"time"
)

type FileInterface interface {
	ReadFile(fileName string) []byte
	WriteString(fileName, data string)
}

type FileReaderWriter struct{}

func (f *FileReaderWriter) ReadFile(fileName string) []byte {
	file, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return file
}

func (f *FileReaderWriter) WriteString(fileName, data string) {
	outputFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	l, err := outputFile.WriteString(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v bytes written\n", l)
}

type ServiceInterface interface {
	toDomainSet(node startgg.Node) domain.Set
	getSetsFromAPI(slug string) *[]domain.Set
	getUpsetThread(sets []domain.Set) *domain.UpsetThread
	submitToSubreddit()
	addSets(slug string, upsetThread *domain.UpsetThread)
	GetUpsetThreadDB(slug, title string) *domain.UpsetThread
	Process(slug, title, subreddit, file string) *domain.UpsetThread
}

type Service struct {
	dbService     data.DBServiceInterface
	startGGClient startgg.ClientInterface
	file          FileInterface
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
	if !s.dbService.IsCharactersLoaded() {
		res := s.startGGClient.GetCharacters()
		s.dbService.AddCharacters(res.Data.VideoGame.Characters)
		s.dbService.SetIsCharactersLoaded()
	}
	return s.dbService.GetCharacterName(key)
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

func (s *Service) toDomainSet(node startgg.Node) domain.Set {
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

func (s *Service) getSetsFromAPI(slug string) *[]domain.Set {
	page := 1
	var sets []domain.Set
	for {
		time.Sleep(800 * time.Millisecond)
		res := s.startGGClient.GetEvent(slug, page)
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
			sets = append(sets, s.toDomainSet(node))
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

func (s *Service) getUpsetThread(sets []domain.Set) *domain.UpsetThread {
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

func defaultSort(entry []domain.UpsetThreadItem, i, j domain.UpsetThreadItem) int {
	return cmp.Or(
		cmp.Compare(j.UpsetFactor, i.UpsetFactor),
		cmp.Compare(i.WinnersName, j.WinnersName),
	)
}

func notablesSort(entry []domain.UpsetThreadItem, i, j domain.UpsetThreadItem) int {
	return cmp.Or(
		cmp.Compare(i.UpsetFactor, j.UpsetFactor),
		cmp.Compare(i.WinnersName, j.WinnersName),
	)
}

func (s *Service) GetUpsetThreadDB(slug, title string) *domain.UpsetThread {
	setMapping := s.dbService.GetSets(slug)
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
	slices.SortFunc(winners, func(i, j domain.UpsetThreadItem) int {
		return defaultSort(winners, i, j)
	})
	slices.SortFunc(losers, func(i, j domain.UpsetThreadItem) int {
		return defaultSort(losers, i, j)
	})
	slices.SortFunc(notables, func(i, j domain.UpsetThreadItem) int {
		return notablesSort(notables, i, j)
	})
	slices.SortFunc(dqs, func(i, j domain.UpsetThreadItem) int {
		return defaultSort(dqs, i, j)
	})
	slices.SortFunc(other, func(i, j domain.UpsetThreadItem) int {
		return defaultSort(other, i, j)
	})
	return &domain.UpsetThread{
		Slug:     slug,
		Title:    title,
		Winners:  winners,
		Losers:   losers,
		Notables: notables,
		DQs:      dqs,
		Other:    other,
	}
}

func (s *Service) submitToSubreddit() {

}

func (s *Service) addSets(slug string, upsetThread *domain.UpsetThread) {
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
	s.dbService.AddSets(slug, &setMapping)
}

func (s *Service) Process(slug, title, subreddit, file string) *domain.UpsetThread {
	var sets []domain.Set

	if file != "" {
		fmt.Println("Using file data", file)
		storedFile := s.file.ReadFile(file)
		var nodes []startgg.Node
		if err := json.Unmarshal(storedFile, &nodes); err != nil {
			panic(err)
		}
		for _, node := range nodes {
			sets = append(sets, s.toDomainSet(node))
		}
	} else {
		fmt.Println("Fetching data from startgg")
		sets = *s.getSetsFromAPI(slug)
	}
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].UpsetFactor > sets[j].UpsetFactor
	})
	upsetThread := s.getUpsetThread(sets)
	s.addSets(slug, upsetThread)
	savedUpsetThread := s.GetUpsetThreadDB(slug, title)
	// md := mapper.ToMarkdown(savedUpsetThread, slug)
	// outputName := fmt.Sprintf("output/%v %s.md", time.Now().UnixMilli(), title)
	// s.file.WriteString(outputName, md)
	return savedUpsetThread
}

func NewService(dbService data.DBServiceInterface, startGGClient startgg.ClientInterface, file FileInterface) *Service {
	return &Service{
		dbService:     dbService,
		startGGClient: startGGClient,
		file:          file,
	}
}
