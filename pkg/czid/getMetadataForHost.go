package idseq

import (
	"encoding/json"
	"net/url"
)

type getMetadataForHostGenomeReq struct{}

type getMetadataForHostGenomeMetadataField struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Examples    string `json:"examples"`
}

type Example struct {
	All []string `json:"all"`
	One []string `json:"1"`
}

type MetadataField struct {
	Name        string
	Description string
	Example     Example
}

func (c *Client) GetMetadataForHostGenome(hostGenome string) ([]MetadataField, error) {
	query := url.Values{
		"name": []string{hostGenome},
	}

	var res []getMetadataForHostGenomeMetadataField
	err := c.request(
		"GET",
		"/metadata/metadata_for_host_genome.json",
		query.Encode(),
		getMetadataForHostGenomeReq{},
		&res,
	)

	metadataFields := make([]MetadataField, len(res))
	for i, f := range res {
		var ex Example
		err = json.Unmarshal([]byte(f.Examples), &ex)
		if err != nil {
			return []MetadataField{}, err
		}
		metadataFields[i] = MetadataField{
			Name:        f.DisplayName,
			Description: f.Description,
			Example:     ex,
		}
	}

	return metadataFields, err
}
