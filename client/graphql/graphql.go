package graphql

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
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

const (
	MAX_RETRIES = 10
	BASE_DELAY  = 1 * time.Second
)

func (client *Client) query(query string, variables interface{}) ([]byte, error) {
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
	req.Header.Set("Authorization", "Bearer "+client.apiToken)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("API returned a non-200 response. status_code=" + strconv.Itoa(resp.StatusCode))
	}

	return respBody, nil
}

func (client *Client) Query(query string, variables interface{}) ([]byte, error) {
	var body []byte
	var err error

	for i := 0; i < MAX_RETRIES; i++ {
		body, err = client.query(query, variables)
		if err == nil {
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
	return body, nil
}

func NewClient(url, apiToken string, httpClient HttpClientInterface) *Client {
	return &Client{
		url:        url,
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}
