package idseq

import (
	"testing"
)

type Auth0 interface {
	IDToken() (string, error)
	Login(headless bool, persistent bool) error
	Secret() (string, bool)
}

type mockAuth0Client struct{}

func (c *mockAuth0Client) IDToken() (string, error) {
	return "id", nil
}

func (c *mockAuth0Client) Login(headless bool, persistent bool) error {
	return nil
}

func (c *mockAuth0Client) Secret() (string, bool) {
	return "secret", true
}

func TestGetTemplateCSV(t *testing.T) {
	httpClient := newMockHTTPClient([]byte("hello world"))
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	csv, err := apiClient.GetTemplateCSV([]string{"sample name"}, "human")
	if err != nil {
		t.Fatal(err)
	}

	head, err := csv.Read()
	if err != nil {
		t.Fatal(err)
	}
	if head[0] != "hello world" {
		t.Error("")
	}
}
