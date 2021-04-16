package idseq

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/url"
)

func (c *Client) GetTemplateCSV(sampleNames []string, hostGenome string) (*csv.Reader, error) {
	query := url.Values{
		"new_sample_names": sampleNames,
		"host_genomes":     []string{hostGenome},
	}

	url := url.URL{
		Path:     "metadata/metadata_template_csv",
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequest("GET", url.String(), bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	res, err := c.authorizedRequest(req)
	if err != nil {
		return nil, err
	}

	return csv.NewReader(res.Body), nil
}
