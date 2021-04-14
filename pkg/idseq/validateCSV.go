package idseq

import (
	"encoding/json"
	"errors"
	"fmt"
)

// validateCSVReq

type validateCSVReqMetadata struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type validateCSVReqSample struct {
	Name      string `json:"name"`
	ProjectID int    `json:"project_id"`
}

type validateCSVReq struct {
	Metadata validateCSVReqMetadata `json:"metadata"`
	Samples  []validateCSVReqSample `json:"samples"`
}

// validateCSVRes

type validateCSVResIssue struct {
	StringError   string
	DetailedIssue struct {
		Caption string
		Rows    [][]string
		Headers []string
		IsGroup bool
	}
}

func (mI *validateCSVResIssue) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case string:
		mI.StringError = v
	case map[string]interface{}:
		mI.DetailedIssue.Caption = v["caption"].(string)
		mI.DetailedIssue.IsGroup = v["isGroup"].(bool)

		headers := v["headers"].([]interface{})
		mI.DetailedIssue.Headers = make([]string, len(headers))
		for i, header := range headers {
			mI.DetailedIssue.Headers[i] = header.(string)
		}

		rows := v["rows"].([]interface{})
		mI.DetailedIssue.Rows = make([][]string, len(rows))
		for i, row := range rows {
			row := row.([]interface{})
			mI.DetailedIssue.Rows[i] = make([]string, len(row))
			for j, val := range row {
				mI.DetailedIssue.Rows[i][j] = fmt.Sprint(val)
			}
		}
	default:
		return errors.New("unable to parse metadata issue")
	}
	return nil
}

func (m validateCSVResIssue) friendlyPrint() {
	if m.StringError != "" {
		fmt.Printf("  %s\n", m.StringError)
	} else {
		fmt.Printf("  %s\n", m.DetailedIssue.Caption)
		for _, row := range m.DetailedIssue.Rows {
			for i, header := range m.DetailedIssue.Headers {
				fmt.Printf("      %s: %s\n", header, row[i])
			}
			fmt.Println("")
		}
	}
	fmt.Println("")
}

type validateCSVResIssues struct {
	Errors   []validateCSVResIssue `json:"errors"`
	Warnings []validateCSVResIssue `json:"warnings"`
}

func (i validateCSVResIssues) friendlyPrint() {
	if len(i.Errors) == 0 && len(i.Warnings) == 0 {
		return
	}
	fmt.Printf("found %d errors and %d warnings\n\n", len(i.Errors), len(i.Warnings))
	if len(i.Errors) > 0 {
		fmt.Println("errors:")
		for _, issue := range i.Errors {
			issue.friendlyPrint()
		}
	}
	if len(i.Warnings) > 0 {
		fmt.Println("warnings:")
		for _, issue := range i.Warnings {
			issue.friendlyPrint()
		}
	}

}

type validateCSVResHostGenome struct {
	Name         string `json:"name"`
	ShowAsOption bool   `json:"showAsOption"`
}

type validateCSVRes struct {
	Status         string                     `json:"status"`
	Issues         validateCSVResIssues       `json:"issues"`
	NewHostGenomes []validateCSVResHostGenome `json:"newHostGenomes"`
}

func ValidateSamplesMetadata(projectID int, samplesMetadata SamplesMetadata) error {
	req := validateCSVReq{}
	for sampleName := range samplesMetadata {
		req.Samples = append(req.Samples, validateCSVReqSample{
			Name:      sampleName,
			ProjectID: projectID,
		})
	}

	headerIndexes := map[string]int{"Sample Name": 0, "Collection Location": 1}
	req.Metadata.Headers = []string{"Sample Name", "Collection Location"}
	req.Metadata.Rows = make([][]interface{}, len(samplesMetadata))
	for sampleName, row := range samplesMetadata {
		validatorRow := make([]interface{}, len(req.Metadata.Headers))
		validatorRow[0] = sampleName
		if row.CollectionLocation != (GeoSearchSuggestion{}) {
			validatorRow[1] = row.CollectionLocation
		} else {
			validatorRow[1] = ""
		}
		for name, value := range row.fields {
			headerIndex, seenHeader := headerIndexes[name]
			if !seenHeader {
				req.Metadata.Headers = append(req.Metadata.Headers, name)
				headerIndexes[name] = len(headerIndexes)
				validatorRow = append(validatorRow, value)
			} else {
				validatorRow[headerIndex] = value
			}
		}
		req.Metadata.Rows = append(req.Metadata.Rows, validatorRow)
	}

	var res validateCSVRes
	err := request("POST", "metadata/validate_csv_for_new_samples.json", "", req, &res)
	if err != nil {
		return err
	}

	res.Issues.friendlyPrint()

	// HACK: new host genomes is a misnomer, all host genomes are returned
	//   new ones will have ShowAsOption = false
	hasNewHostGenomes := false
	for _, hostGenome := range res.NewHostGenomes {
		if !hostGenome.ShowAsOption {
			hasNewHostGenomes = true
			break
		}
	}
	if hasNewHostGenomes {
		fmt.Println(`some of your host organisms were not found in IDSeq
host filtering will only filter out ERCC reads
confirm these host organisms are correct:`)
		for _, hostGenome := range res.NewHostGenomes {
			if !hostGenome.ShowAsOption {
				fmt.Printf("  %s\n", hostGenome.Name)
			}
		}
	}

	if len(res.Issues.Errors) > 0 {
		return errors.New("metadata validation failed")
	}
	return nil
}
