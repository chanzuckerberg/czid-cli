package idseq

import "fmt"

type validateSampleNamesRequest struct {
	SampleNames      []string `json:"sample_names"`
	IgnoreUnuploaded bool     `json:"ignore_unuploaded"`
}

func (c *Client) ValidateSampleNames(sampleNames []string, projectID int) ([]string, error) {
	var res []string
	err := c.request(
		"POST",
		fmt.Sprintf("/projects/%d/validate_sample_names", projectID),
		"",
		validateSampleNamesRequest{
			SampleNames:      sampleNames,
			IgnoreUnuploaded: true,
		},
		&res,
	)

	return res, err
}
