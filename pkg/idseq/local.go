package idseq

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
)

type Metadata = map[string]string
type SamplesMetadata = map[string]Metadata

func ToValidateForm(m SamplesMetadata) validationMetadata {
	headerIndexes := map[string]int{"Sample Name": 0}
	vM := validationMetadata{
		Headers: []string{"Sample Name"},
		Rows:    make([][]string, len(m)),
	}

	for sampleName, row := range m {
		validatorRow := make([]string, len(vM.Headers))
		validatorRow[0] = sampleName
		for name, value := range row {
			headerIndex, seenHeader := headerIndexes[name]
			if !seenHeader {
				vM.Headers = append(vM.Headers, name)
				headerIndexes[name] = len(headerIndexes)
				validatorRow = append(validatorRow, value)
			} else {
				validatorRow[headerIndex] = value
			}
		}
		vM.Rows = append(vM.Rows, validatorRow)
	}
	return vM
}

func CSVMetadata(csvpath string) (SamplesMetadata, error) {
	samplesMetadata := SamplesMetadata{}
	f, err := os.Open(csvpath)
	if err != nil {
		return samplesMetadata, err
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		return samplesMetadata, err
	}
	if len(rows) < 2 {
		return samplesMetadata, nil
	}
	headers := rows[0]
	hasSampleName := false
	for _, header := range headers {
		if header == "Sample Name" {
			hasSampleName = true
			break
		}
	}
	if !hasSampleName {
		return samplesMetadata, errors.New("column 'Sample Name' is required but it was not found")
	}
	for rowNum, row := range rows[1:] {
		sampleName := ""
		metadata := make(Metadata, len(headers))
		for i, header := range headers {
			if header == "Sample Name" {
				if i >= len(row) {
					return samplesMetadata, fmt.Errorf("row %d is missing 'Sample Name'", rowNum)
				}
				sampleName = row[i]
			} else {
				if i >= len(row) {
					metadata[header] = ""
				} else {
					metadata[header] = row[i]
				}
			}
		}
		samplesMetadata[sampleName] = metadata
	}
	return samplesMetadata, nil
}
