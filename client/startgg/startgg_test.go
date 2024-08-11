package startgg

import "testing"

type FakeGraphQLClient struct {
	queryMethodCalled bool
	returnValue       []byte
}

func (client *FakeGraphQLClient) Query(string, interface{}) ([]byte, error) {
	client.queryMethodCalled = true
	return client.returnValue, nil
}

func TestGetEvent(t *testing.T) {
	fakeGraphQLClient := FakeGraphQLClient{returnValue: []byte(`{ "data": { "event": { "videogame": {}, "sets": {} } } }`)}
	client := NewClient(&fakeGraphQLClient)
	client.GetEvent("slug", 1)
	if !fakeGraphQLClient.queryMethodCalled {
		t.Errorf("Expected query method to be called")
	}
}

func TestGetCharacters(t *testing.T) {
	fakeGraphQLClient := FakeGraphQLClient{returnValue: []byte(`{ "data": { "videogame": {} } }`)}
	client := NewClient(&fakeGraphQLClient)
	client.GetCharacters("slug")
	if !fakeGraphQLClient.queryMethodCalled {
		t.Errorf("Expected query method to be called")
	}
}
