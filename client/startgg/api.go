package startgg

import (
	"encoding/json"
	"fmt"
	"gg/client/graphql"
)

type ClientInterface interface {
	GetEvent(slug string, page int) EventResponse
	GetCharacters() CharactersResponse
}

type Client struct {
	graphQLClient graphql.ClientInterface
}

type Entrant struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	InitialSeedNum int    `json:"initialSeedNum"`
	Standing       struct {
		IsFinal   bool `json:"isFinal"`
		Placement int  `json:"placement"`
	} `json:"standing"`
}

type Selection struct {
	OrderNum       int     `json:"orderNum"`
	SelectionType  string  `json:"selectionType"`
	SelectionValue int     `json:"selectionValue"`
	Entrant        Entrant `json:"entrant"`
}

type Game struct {
	Id         int         `json:"id"`
	WinnerId   int         `json:"winnerId"`
	OrderNum   int         `json:"orderNum"`
	Selections []Selection `json:"selections"`
}

type Node struct {
	Id            int    `json:"id"`
	CompletedAt   int    `json:"completedAt"`
	Games         []Game `json:"games"`
	Identifier    string `json:"identifier"`
	DisplayScore  string `json:"displayScore"`
	FullRoundText string `json:"fullRoundText"`
	TotalGames    int    `json:"totalGames"`
	LPlacement    int    `json:"lPlacement"`
	WPlacement    int    `json:"wPlacement"`
	WinnerId      int    `json:"winnerId"`
	State         int    `json:"state"`
	SetGamesType  int    `json:"setGamesType"`
	Round         int    `json:"round"`
	PhaseGroup    struct {
		DisplayIdentifier string `json:"displayIdentifier"`
	} `json:"phaseGroup"`
	Slots []struct {
		Entrant Entrant `json:"entrant"`
	} `json:"slots"`
}

type EventResponse struct {
	Data struct {
		Event struct {
			Id        int    `json:"id"`
			Slug      string `json:"slug"`
			UpdatedAt int    `json:"updatedAt"`
			Sets      struct {
				PageInfo struct {
					Total      int    `json:"total"`
					TotalPages int    `json:"totalPages"`
					Page       int    `json:"page"`
					PerPage    int    `json:"perPage"`
					SortBy     string `json:"sortBy"`
					Filter     string `json:"filter"`
				} `json:"pageInfo"`
				Nodes []Node `json:"nodes"`
			} `json:"sets"`
		} `json:"event"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *Client) GetEvent(slug string, page int) EventResponse {
	fmt.Printf("Getting event. slug: %s, page: %v\n", slug, page)
	type filters struct {
		State int `json:"state"`
	}
	type variables struct {
		Slug     string  `json:"slug"`
		Page     int     `json:"page"`
		Filters  filters `json:"filters"`
		SortType string  `json:"sortType"`
	}
	resp := client.graphQLClient.Query(eventsQuery, variables{slug, page, filters{3}, "RECENT"})
	var eventResponse EventResponse
	if err := json.Unmarshal(resp, &eventResponse); err != nil {
		panic(err)
	}
	return eventResponse
}

type Character struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CharactersResponse struct {
	Data struct {
		VideoGame struct {
			Id         int         `json:"id"`
			Slug       string      `json:"slug"`
			Characters []Character `json:"characters"`
		} `json:"videogame"`
	} `json:"data"`
}

func (client *Client) GetCharacters() CharactersResponse {
	fmt.Println("Getting characters")
	type variables struct {
		Slug string `json:"slug"`
	}
	resp := client.graphQLClient.Query(charactersQuery, variables{"game/ultimate"})
	var charactersResponse CharactersResponse
	if err := json.Unmarshal(resp, &charactersResponse); err != nil {
		panic(err)
	}
	return charactersResponse
}

func NewClient() *Client {
	return &Client{graphQLClient: graphql.NewClient()}
}
