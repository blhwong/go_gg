package startgg

import (
	"encoding/json"
	"errors"
	"gg/client/graphql"
	"log"
	"math"
	"time"
)

const (
	MAX_RETRIES = 10
	BASE_DELAY  = 1 * time.Second
)

var ErrorGreaterthan10KEntry = errors.New("cannot query more than 10,000th entry")

type ClientInterface interface {
	GetEvent(slug string, page int) (*EventResponse, error)
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

func (client *Client) getEvent(slug string, page int) (*EventResponse, error, bool) {
	type filters struct {
		State int `json:"state"`
	}
	type variables struct {
		Slug     string  `json:"slug"`
		Page     int     `json:"page"`
		Filters  filters `json:"filters"`
		SortType string  `json:"sortType"`
	}
	resp, err := client.graphQLClient.Query(eventsQuery, variables{slug, page, filters{3}, "RECENT"})
	if err != nil {
		return nil, err, true
	}
	var eventResponse EventResponse
	if err := json.Unmarshal(resp, &eventResponse); err != nil {
		log.Fatalf("Error while unmarshaling event. e=%s\n", err)
	}
	if eventResponse.Errors != nil {
		if eventResponse.Errors[0].Message == "Cannot query more than the 10,000th entry" {
			return &eventResponse, ErrorGreaterthan10KEntry, false
		}
		return &eventResponse, errors.New(eventResponse.Errors[0].Message), true
	}
	return &eventResponse, nil, false
}

func (client *Client) GetEvent(slug string, page int) (*EventResponse, error) {
	var eventResponse *EventResponse
	var err error
	var retryable bool

	for i := 0; i < MAX_RETRIES; i++ {
		eventResponse, err, retryable = client.getEvent(slug, page)
		if err == nil || !retryable {
			break
		}

		secRetry := math.Pow(2, float64(i))
		delay := time.Duration(secRetry) * BASE_DELAY
		log.Printf("Error: %s. Retrying %d of %d\n in %v seconds", err, i+1, MAX_RETRIES, delay.Seconds())
		time.Sleep(delay)
	}

	if err != nil {
		return nil, err
	}
	return eventResponse, nil
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
	log.Println("Getting characters")
	type variables struct {
		Slug string `json:"slug"`
	}
	resp, _ := client.graphQLClient.Query(charactersQuery, variables{"game/street-fighter-6"})
	var charactersResponse CharactersResponse
	if err := json.Unmarshal(resp, &charactersResponse); err != nil {
		log.Fatalf("Error while marshaling characters. e=%s\n", err)
	}
	return charactersResponse
}

func NewClient(url, apiToken string, httpClient graphql.HttpClientInterface) *Client {
	return &Client{graphQLClient: graphql.NewClient(url, apiToken, httpClient)}
}
