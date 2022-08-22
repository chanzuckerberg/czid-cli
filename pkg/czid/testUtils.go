package czid

import (
	"bytes"
	"io"
	"net/http"
)

type mockHTTPClient struct {
	calls    []*http.Request
	response []byte
}

func newMockHTTPClient(response []byte) mockHTTPClient {
	return mockHTTPClient{
		calls:    []*http.Request{},
		response: response,
	}
}

func (c *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	body := io.NopCloser(bytes.NewReader(c.response))
	return &http.Response{StatusCode: 200, Body: body}, nil
}
