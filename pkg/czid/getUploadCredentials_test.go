package czid

import (
	"testing"
)

func TestGetUploadCredentials(t *testing.T) {
	response := []byte(`{
      "access_key_id": "access_key_id_123",
      "expiration": "2021-06-01T00:00:00Z",
      "secret_access_key": "secret_access_key_123",
      "session_token": "session_token_123"
    }`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	creds, err := apiClient.GetUploadCredentials(1)
	if err != nil {
		t.Fatal(err)
	}

	if creds.AccessKeyID != "access_key_id_123" {
		t.Errorf("access key %s did not equal '%s'", creds.AccessKeyID, "access_key_id_123")
	}
}
