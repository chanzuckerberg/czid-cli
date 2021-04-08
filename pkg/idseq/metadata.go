package idseq

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Metadata struct {
	HostGenome            string
	rawCollectionLocation string
	CollectionLocation    GeoSearchSuggestion
	fields                map[string]string
}

var hostGenomeAliases map[string]bool = map[string]bool{
	"host_genome":   true,
	"Host Genome":   true,
	"Host genome":   true,
	"host genome":   true,
	"host_organism": true,
	"Host Organism": true,
	"Host organism": true,
	"host organism": true,
}

var collectionLocationAliases map[string]bool = map[string]bool{
	"collection location": true,
	"Collection Location": true,
	"Collection location": true,
	"collection_location": true,
}

func (m Metadata) update(fields map[string]string) Metadata {
	for k, v := range fields {
		if hostGenomeAliases[k] {
			m.HostGenome = v
		}

		if collectionLocationAliases[k] {
			m.rawCollectionLocation = v
		} else {
			m.fields[k] = v
		}
	}
	return m
}

func NewMetadata(m map[string]string) Metadata {
	metadata := Metadata{fields: make(map[string]string)}
	return metadata.update(m)
}

func (a Metadata) Fuse(b Metadata) Metadata {
	c := a.update(b.fields)

	if b.rawCollectionLocation != "" {
		c.rawCollectionLocation = b.rawCollectionLocation
	}

	if b.CollectionLocation != (GeoSearchSuggestion{}) {
		c.CollectionLocation = b.CollectionLocation
	}

	return c
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	interfaceMap := make(map[string]interface{}, len(m.fields)+1)
	for k, v := range m.fields {
		interfaceMap[k] = v
	}
	interfaceMap["Collection Location"] = m.CollectionLocation
	return json.Marshal(interfaceMap)
}

func (m Metadata) isHuman() bool {
	return strings.ToLower(m.HostGenome) == "human"
}

func (m Metadata) withSuggestedLocation() (Metadata, error) {
	suggestion, err := GetGeoSearchSuggestion(m.rawCollectionLocation, m.isHuman())
	m.CollectionLocation = suggestion
	return m, err
}

type SamplesMetadata = map[string]Metadata

func CSVMetadata(csvpath string) (SamplesMetadata, error) {
	samplesMetadata := SamplesMetadata{}
	f, err := os.Open(csvpath)
	if err != nil {
		return samplesMetadata, err
	}
	defer f.Close()
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
		metadata := make(map[string]string, len(headers))
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
		samplesMetadata[sampleName] = NewMetadata(metadata)
	}
	return samplesMetadata, nil
}

func GeoSearchSuggestions(samplesMetadata *SamplesMetadata) error {
	for sampleName, metadata := range *samplesMetadata {
		m, err := metadata.withSuggestedLocation()
		if err != nil {
			return err
		}
		(*samplesMetadata)[sampleName] = m
	}
	return nil
}
