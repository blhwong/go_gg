package domain

type UpsetThreadItem struct {
	Id, WinnersName, WinnersCharacters                    string
	WinnersSeed                                           int
	Score                                                 *string
	LosersName, LosersCharacters                          string
	IsWinnersBracket                                      bool
	LosersSeed, LosersPlacement, UpsetFactor, CompletedAt int
	Category                                              string
}

type UpsetThread struct {
	Title    string
	Slug     string
	Winners  []UpsetThreadItem
	Losers   []UpsetThreadItem
	Notables []UpsetThreadItem
	DQs      []UpsetThreadItem
	Other    []UpsetThreadItem
}

type UpsetThreadItemHTML struct {
	Content string
	Bold    bool
}

type UpsetThreadHTML struct {
	Host          string
	Title         string
	Slug          string
	LastUpdatedAt string
	Winners       []*UpsetThreadItemHTML
	Losers        []*UpsetThreadItemHTML
	Notables      []*UpsetThreadItemHTML
	DQs           []*UpsetThreadItemHTML
}
