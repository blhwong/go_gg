package graphql

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type FakeHttpClient struct {
	doMethodCalled bool
}

func (client *FakeHttpClient) Do(*http.Request) (*http.Response, error) {
	client.doMethodCalled = true
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}, nil
}

func TestQuery(t *testing.T) {
	fakeHttpClient := FakeHttpClient{}
	client := Client{"url", "apiToken", &fakeHttpClient}
	query := `
		query CharactersQuery(
			$slug: String
		) {
			videogame(slug: $slug) {
				id
				slug
				characters {
					id
					name
				}
			}
		}
	`
	type variables struct {
		slug string
	}
	client.Query(query, variables{"game/ultimate"})
	if !fakeHttpClient.doMethodCalled {
		t.Fatalf("httpClient Do method was not called.")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("url", "apiToken", &FakeHttpClient{})
	if client.url != "url" {
		t.Fatalf("Result %s, expected %s", client.url, "url")
	}
	if client.apiToken != "apiToken" {
		t.Fatalf("Result %s, expected %s", client.apiToken, "apiToken")
	}
}
