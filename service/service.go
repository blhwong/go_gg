package service

import (
	"cmp"
	"encoding/json"
	"gg/client/startgg"
	"gg/db"
	"gg/domain"
	"gg/mapper"
	"log"
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
		log.Fatalf("Error while reading file. e=%s\n", err)
	}
	return file
}

func (f *FileReaderWriter) WriteString(fileName, data string) {
	outputFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error while creating file. e=%s\n", err)
	}
	defer outputFile.Close()
	l, err := outputFile.WriteString(data)
	if err != nil {
		log.Fatalf("Error while writing to file. e=%s\n", err)
	}
	log.Printf("%v bytes written\n", l)
}

type ServiceInterface interface {
	toDomainSet(node startgg.Node, slug string) domain.Set
	getSetsFromAPI(slug string) *[]domain.Set
	getUpsetThread(sets []domain.Set) *domain.UpsetThread
	submitToSubreddit()
	addSets(slug string, upsetThread *domain.UpsetThread)
	GetUpsetThreadDB(slug, title string) *domain.UpsetThread
	Process(slug, title, subreddit, file, gameSlug string) *domain.UpsetThread
}

type Service struct {
	dbService     db.DBServiceInterface
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

func (s *Service) getCharacterName(key int, slug string) string {
	if !s.dbService.IsCharactersLoaded(slug) {
		res := s.startGGClient.GetCharacters(slug)
		s.dbService.AddCharacters(res.Data.VideoGame.Characters, slug)
		s.dbService.SetIsCharactersLoaded(slug)
	}
	return s.dbService.GetCharacterName(key, slug)
}

func (s *Service) toDomainCharacter(selectionType string, value int, slug string) *domain.Character {
	if selectionType != "CHARACTER" {
		return nil
	}
	return &domain.Character{
		Value: value,
		Name:  s.getCharacterName(value, slug),
	}
}

func (s *Service) toDomainSelection(selection startgg.Selection, slug string) domain.Selection {
	return domain.Selection{
		Entrant:   toDomainEntrant(selection.Entrant),
		Character: s.toDomainCharacter(selection.SelectionType, selection.SelectionValue, slug),
	}
}

func (s *Service) toDomainGame(game startgg.Game, slug string) domain.Game {
	var selections []domain.Selection
	if game.Selections != nil {
		for _, selection := range game.Selections {
			selections = append(selections, s.toDomainSelection(selection, slug))
		}
	}
	return domain.Game{
		Id:         game.Id,
		WinnerId:   game.WinnerId,
		Selections: selections,
	}
}

func (s *Service) toDomainSet(node startgg.Node, slug string) domain.Set {
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
			games = append(games, s.toDomainGame(game, slug))
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
		res, err := s.startGGClient.GetEvent(slug, page)
		if err == startgg.ErrorGreaterthan10KEntry {
			log.Println("Finishing because cannot query more than 10,000th entry.")
			break
		}
		if err != nil {
			log.Fatalf("Something went wrong getting event. e=%s\n", err)
		}
		if res.Errors != nil {
			log.Fatalf("Response contains errors. e=%s\n", res.Errors)
		}
		totalPages := res.Data.Event.Sets.PageInfo.TotalPages
		log.Printf("Event received. slug=%s page=%v totalPage=%v\n", slug, page, totalPages)
		if page > totalPages {
			break
		}
		page++
		for _, node := range res.Data.Event.Sets.Nodes {
			sets = append(sets, s.toDomainSet(node, res.Data.Event.Videogame.Slug))
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

func (s *Service) Process(slug, title, subreddit, file, gameSlug string) *domain.UpsetThread {
	var sets []domain.Set
	if file != "" {
		log.Println("Using file data", file)
		storedFile := s.file.ReadFile(file)
		var nodes []startgg.Node
		if err := json.Unmarshal(storedFile, &nodes); err != nil {
			log.Fatalf("Error while unmarshaling node. e=%s\n", err)
		}
		for _, node := range nodes {
			sets = append(sets, s.toDomainSet(node, gameSlug))
		}
	} else {
		log.Println("Fetching data from startgg")
		sets = *s.getSetsFromAPI(slug)
	}
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].UpsetFactor > sets[j].UpsetFactor
	})
	upsetThread := s.getUpsetThread(sets)
	s.addSets(slug, upsetThread)
	savedUpsetThread := s.GetUpsetThreadDB(slug, title)
	return savedUpsetThread
}

func NewService(dbService db.DBServiceInterface, startGGClient startgg.ClientInterface, file FileInterface) *Service {
	return &Service{
		dbService:     dbService,
		startGGClient: startGGClient,
		file:          file,
	}
}
