package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ClientInterface interface {
	Query(query string, variables interface{}) ([]byte, error)
}

type HttpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	url        string
	apiToken   string
	httpClient HttpClientInterface
}

type Payload struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func (client *Client) Query(query string, variables interface{}) ([]byte, error) {
	payload := Payload{query, variables}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error while marshaling. e=%s\n", err)
	}
	req, err := http.NewRequest("POST", client.url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error while creating new request. e=%s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.apiToken)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		log.Fatalf("Error on http client. e=%s\n", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error on io read. e=%s\n", err)
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("api returned a status >= 500. status_code=%d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		log.Fatalf("API returned a status between [400, 500). status_code=%d\n", resp.StatusCode)
	}

	return respBody, nil
}

func NewClient(url, apiToken string, httpClient HttpClientInterface) *Client {
	return &Client{
		url:        url,
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}
