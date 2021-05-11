package idseq

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type mockHTTPClientGetMetadataForHost struct {
	calls []*http.Request
}

func (c *mockHTTPClientGetMetadataForHost) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	body := ioutil.NopCloser(bytes.NewReader([]byte(`{
  "display_name": "test",
  "description": "desc",
  "examples": "{\"all\": [\"example\"] }"
}`)))
	return &http.Response{StatusCode: 200, Body: body}, nil
}

func TestGetMetadataForHost(t *testing.T) {
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &mockHTTPClientGetMetadataForHost{calls: []*http.Request{}},
	}

	fields, err := apiClient.GetMetadataForHostGenome("human")
	if err != nil {
		t.Fatal(err)
	}

	if len(fields) != 1 {
		t.Errorf("expected 1 field but got %d", len(fields))
	}

	if fields[0].Name != "test" {
		t.Errorf("expected name to be 'test' but it was '%s'", fields[0].Name)
	}

	if len(fields[0].Example.All) != 1 {
		t.Errorf("expected one option in All but got %d", 1)
	}

	if fields[0].Example.All[0] != "example" {
		t.Errorf("expected option to be 'example' but it was '%s'", fields[0].Example.All[0])
	}
}
