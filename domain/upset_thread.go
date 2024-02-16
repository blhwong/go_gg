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
	Winners  []UpsetThreadItem
	Losers   []UpsetThreadItem
	Notables []UpsetThreadItem
	DQs      []UpsetThreadItem
	Other    []UpsetThreadItem
}
