package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClientInterface interface {
	Query(query string, variables interface{}) []byte
}

type Client struct {
	Url      string
	ApiToken string
}

type Payload struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

func (client Client) Query(query string, variables interface{}) []byte {
	payload := Payload{query, variables}
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", client.Url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiToken))
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
