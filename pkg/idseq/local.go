package idseq

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

var inputExp = regexp.MustCompile("\\.(fasta|fa|fastq)(\\.gz)?$")

func IsInput(path string) bool {
	return inputExp.MatchString(path)
}

var sampleNameExp = regexp.MustCompile("(_R[12])?\\.(fasta|fa|fastq)(\\.gz)?$")

func ToSampleName(path string) string {
	return sampleNameExp.ReplaceAllString(filepath.Base(path), "")
}

var r1Exp = regexp.MustCompile("_R1\\.(fasta|fa|fastq)(\\.gz)?$")

func IsR1(path string) bool {
	return r1Exp.MatchString(path)
}

var r2Exp = regexp.MustCompile("_R2\\.(fasta|fa|fastq)(\\.gz)?$")

func IsR2(path string) bool {
	return r2Exp.MatchString(path)
}

type SampleFiles struct {
	R1     string
	R2     string
	Single string
}

func SamplesFromDir(directory string) (map[string]SampleFiles, error) {
	pairs := make(map[string]SampleFiles)
	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if match := IsInput(path); match {
			sampleName := ToSampleName(path)
			sampleFiles := pairs[sampleName]

			if IsR1(path) {
				if sampleFiles.Single != "" {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}
				if sampleFiles.R1 != "" {
					return fmt.Errorf("found multiple R1 files for sample '%s': %s, %s", sampleName, path, sampleFiles.R1)
				}
				sampleFiles.R1 = path
			} else if IsR2(path) {
				if sampleFiles.Single != "" {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}
				if sampleFiles.R2 != "" {
					return fmt.Errorf("found multiple R2 files for sample '%s': %s, %s", sampleName, path, sampleFiles.R2)
				}
				sampleFiles.R2 = path
			} else {
				if sampleFiles.R1 != "" {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R1)
				}
				if sampleFiles.R2 != "" {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R2)
				}
				if sampleFiles.Single != "" {
					return fmt.Errorf("found multiple single end files for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}
				sampleFiles.Single = path
			}
			pairs[sampleName] = sampleFiles
		}
		return err
	})
	return pairs, err
}
