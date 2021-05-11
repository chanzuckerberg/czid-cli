package idseq

import (
	"bytes"
	"io/ioutil"
	"net/http"
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

type mockHTTPClientGetTemplateCSV struct {
	calls []*http.Request
}

func (c *mockHTTPClientGetTemplateCSV) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	body := ioutil.NopCloser(bytes.NewReader([]byte("hello world")))
	return &http.Response{StatusCode: 200, Body: body}, nil
}

func TestGetTemplateCSV(t *testing.T) {
	type mockHTTPClient struct {
		calls []*http.Request
	}
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &mockHTTPClientGetTemplateCSV{calls: []*http.Request{}},
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
