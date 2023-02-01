package czid

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
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
	if m.CollectionLocation != (GeoSearchSuggestion{}) {
		interfaceMap["Collection Location"] = m.CollectionLocation
	} else if m.rawCollectionLocation != "" {
		interfaceMap["Collection Location"] = m.rawCollectionLocation
	}

	return json.Marshal(interfaceMap)
}

func (m Metadata) isHuman() bool {
	return strings.ToLower(m.HostGenome) == "human"
}

func trimInvisible(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsGraphic(r) || !unicode.IsPrint(r)
	})
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
	for i, header := range headers {
		headers[i] = trimInvisible(header)
	}

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
		if len(row) == 0 || (len(row) == 1 && trimInvisible(row[0]) == "") {
			continue
		}

		sampleName := ""
		metadata := make(map[string]string, len(headers))
		for i, header := range headers {
			if trimInvisible(header) == "Sample Name" {
				if i >= len(row) {
					return samplesMetadata, fmt.Errorf("row %d is missing 'Sample Name'", rowNum)
				}
				sampleName = trimInvisible(row[i])
			} else {
				if i >= len(row) {
					metadata[header] = ""
				} else {
					metadata[header] = trimInvisible(row[i])
				}
			}
		}
		samplesMetadata[sampleName] = NewMetadata(metadata)
	}
	return samplesMetadata, nil
}

// GetCombinedMetadata parses the metadata CSV, validates it, then fuses it with flag-based metadata
func GetCombinedMetadata(sampleFiles map[string]SampleFiles, stringMetadata map[string]string, metadataCSVPath string) (SamplesMetadata, error) {
	metadata := NewMetadata(stringMetadata)
	hasMetadataCSV := metadataCSVPath != ""

	samplesMetadata := SamplesMetadata{}
	if hasMetadataCSV {
		var err error
		samplesMetadata, err = CSVMetadata(metadataCSVPath)
		if err != nil {
			return samplesMetadata, err
		}

		for sampleName := range samplesMetadata {
			if _, hasSampleName := sampleFiles[sampleName]; !hasSampleName {
				delete(samplesMetadata, sampleName)
			}
		}
	}
	missing := false
	for sampleName := range sampleFiles {
		if _, hasMetadata := samplesMetadata[sampleName]; !hasMetadata {
			if !hasMetadataCSV {
				samplesMetadata[sampleName] = NewMetadata(map[string]string{})
			} else {
				log.Printf("missing metadata in metadata CSV for sample name '%s'\n", sampleName)
				missing = true
			}
		}
	}
	if missing {
		return samplesMetadata, fmt.Errorf("missing metadata in CSV for samples")
	}

	for sampleName, m := range samplesMetadata {
		samplesMetadata[sampleName] = m.Fuse(metadata)
	}

	return samplesMetadata, nil
}

func GeoSearchSuggestions(samplesMetadata *SamplesMetadata) error {
	remapping := make(map[string]GeoSearchSuggestion, len(*samplesMetadata))
	for sampleName, metadata := range *samplesMetadata {
		if c, has := remapping[metadata.rawCollectionLocation]; has {
			metadata.CollectionLocation = c
			(*samplesMetadata)[sampleName] = metadata
			continue
		}
		suggestion, err := DefaultClient.GetGeoSearchSuggestion(
			metadata.rawCollectionLocation,
			metadata.isHuman(),
		)
		if err != nil {
			return err
		}
		metadata.CollectionLocation = suggestion
		(*samplesMetadata)[sampleName] = metadata
		remapping[metadata.rawCollectionLocation] = suggestion
	}

	for o, n := range remapping {
		if o != n.String() && n != (GeoSearchSuggestion{}) {
			fmt.Printf("  replacing location \"%s\" with \"%s\"\n", o, n.String())
		}
	}

	return nil
}
