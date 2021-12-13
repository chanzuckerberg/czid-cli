package idseq

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/url"
)

func (c *Client) GetTemplateCSV(sampleNames []string, hostGenome string) (*csv.Reader, error) {
	query := url.Values{
		"new_sample_names[]": sampleNames,
	}

	if hostGenome != "" {
		query["host_genomes[]"] = make([]string, len(sampleNames))
		for i := range sampleNames {
			query["host_genomes[]"][i] = hostGenome
		}
	}

	url := url.URL{
		Path:     "/metadata/metadata_template_csv",
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequest("GET", url.String(), bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	res, err := c.authorizedRequest(req)
	if err != nil {
		return nil, err
	}

	return csv.NewReader(res.Body), nil
}
