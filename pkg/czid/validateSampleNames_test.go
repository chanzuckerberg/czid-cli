package czid

import (
	"testing"
)

func TestValidateSampleNames(t *testing.T) {
	response := []byte(`["sample one", "sample one_1"]`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	sample_names, err := apiClient.ValidateSampleNames([]string{"sample one", "sample one"}, 123)
	if err != nil {
		t.Errorf("encountered an error")
	}

	if sample_names[0] != "sample one" && sample_names[1] != "sample one_1" {
		t.Errorf("incorrect response %s", sample_names)
	}
}
