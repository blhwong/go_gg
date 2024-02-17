package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ClientInterface interface {
	Query(query string, variables interface{}) []byte
}

type Client struct {
	url      string
	apiToken string
}

type Payload struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func (client *Client) Query(query string, variables interface{}) []byte {
	payload := Payload{query, variables}
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", client.url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiToken))
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return respBody
}

func NewClient() *Client {
	return &Client{
		url:      os.Getenv("START_GG_API_URL"),
		apiToken: os.Getenv("START_GG_API_KEY"),
	}
}
