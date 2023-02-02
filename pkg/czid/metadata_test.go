package czid

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func createDummyMetadataCSV() (string, error) {
	csv, err := os.CreateTemp("", "*.csv")
	if err != nil {
		return csv.Name(), err
	}

	csvData := []byte("Sample Name\u200b,Host Genome,collection_location,Nucleotide Type\n" +
		"sample one,Human,\"California, USA\",DNA\n\n" +
		"sample two,D\u200bog,\"California, USA\",RNA",
	)

	_, err = csv.Write(csvData)
	return csv.Name(), err
}

func TestCSVMetadata(t *testing.T) {
	csvName, err := createDummyMetadataCSV()
	defer os.Remove(csvName)
	if err != nil {
		t.Fatal(err)
	}

	samplesMetadata, err := CSVMetadata(csvName)
	if err != nil {
		t.Fatal(err)
	}

	if !samplesMetadata["sample one"].isHuman() {
		t.Errorf("expected sample one to be human but isHuman was false")
	}

	if samplesMetadata["sample two"].isHuman() {
		t.Error("expected sample two to not be human but isHuman was true")
	}

	if samplesMetadata["sample one"].rawCollectionLocation != "California, USA" {
		t.Errorf("expected rawCollectionLocation to be \"California, USA\" but it was \"%s\"", samplesMetadata["sample one"].rawCollectionLocation)
	}
}

func TestGetCombinedMetadata(t *testing.T) {
	csvName, err := createDummyMetadataCSV()
	defer os.Remove(csvName)
	if err != nil {
		t.Fatal(err)
	}

	dummySampleFiles := SampleFiles{}
	sampleFiles := map[string]SampleFiles{"sample one": dummySampleFiles, "sample two": dummySampleFiles}
	stringMetadata := map[string]string{"Nucleotide Type": "DNA", "Foo": "Bar"}

	samplesMetadata, err := GetCombinedMetadata(sampleFiles, stringMetadata, csvName)
	if err != nil {
		t.Fatal(err)
	}

	if samplesMetadata["sample two"].HostGenome != "D\u200bog" {
		t.Errorf("sample two should have 'Host Genome' from CSV but it was '%s'", samplesMetadata["sample two"].HostGenome)
	}

	if samplesMetadata["sample two"].fields["Nucleotide Type"] != "DNA" {
		t.Errorf("sample two should have 'Nucleotide Type' overwritten with 'DNA' from flags but it was '%s'", samplesMetadata["sample two"].fields["Nucleotide Type"])
	}
}

func TestGetCombinedMetadataWithoutCSV(t *testing.T) {
	dummySampleFiles := SampleFiles{}
	sampleFiles := map[string]SampleFiles{"sample one": dummySampleFiles, "sample two": dummySampleFiles}
	stringMetadata := map[string]string{"Nucleotide Type": "DNA", "Foo": "Bar"}

	samplesMetadata, err := GetCombinedMetadata(sampleFiles, stringMetadata, "")
	if err != nil {
		t.Fatal(err)
	}

	if samplesMetadata["sample one"].fields["Nucleotide Type"] != "DNA" {
		t.Errorf("sample one should have 'Nucleotide Type' 'DNA' from flags but it was '%s'", samplesMetadata["sample one"].fields["Nucleotide Type"])
	}

	if samplesMetadata["sample two"].fields["Nucleotide Type"] != "DNA" {
		t.Errorf("sample two should have 'Nucleotide Type' 'DNA' from flags but it was '%s'", samplesMetadata["sample two"].fields["Nucleotide Type"])
	}
}

func TestGetCombinedMetadataMissingFromCSV(t *testing.T) {
	csvName, err := createDummyMetadataCSV()
	defer os.Remove(csvName)
	if err != nil {
		t.Fatal(err)
	}

	dummySampleFiles := SampleFiles{}
	sampleFiles := map[string]SampleFiles{
		"sample one": dummySampleFiles,
		"sample two": dummySampleFiles,
		"sample thr": dummySampleFiles,
	}
	stringMetadata := map[string]string{"Nucleotide Type": "DNA", "Foo": "Bar"}

	_, err = GetCombinedMetadata(sampleFiles, stringMetadata, csvName)
	if err == nil {
		t.Errorf("expected GetCombinedMetadata to return missing metadata in CSV error if sample is missing from metadata CSV")
	}

	if err != nil && err.Error() != "missing metadata in CSV for samples" {
		t.Fatal(err)
	}
}

func TestFuse(t *testing.T) {
	mOne := NewMetadata(map[string]string{
		"host_genome":         "Koala",
		"collection_location": "California, USA",
		"Water Control":       "No",
		"Nucleotide Type":     "DNA",
	})

	mTwo := NewMetadata(map[string]string{
		"host_genome":   "Human",
		"Water Control": "Yes",
		"Foo":           "Bar",
	})

	m := mOne.Fuse(mTwo)

	if !m.isHuman() {
		t.Error("expected fused metadata to be human but isHuman was false")
	}

	if m.rawCollectionLocation != "California, USA" {
		t.Errorf("expected fused metadata rawCollectionLocation to be \"California, USA\" but it was \"%s\"", m.rawCollectionLocation)
	}

	if m.fields["Water Control"] != "Yes" {
		t.Error("second metadata should have overwritten \"Water Control\" but it did not")
	}

	if _, has := m.fields["Nucleotide Type"]; !has {
		t.Error("fused metadata is missing field \"Nucleotide Type\" from first metadata")
	}

	if _, has := m.fields["Foo"]; !has {
		t.Error("fused metadata is missing field \"Foo\" from second metadata")
	}
}

func TestJSONMarshal(t *testing.T) {
	m := NewMetadata(map[string]string{
		"host_genome":         "Koala",
		"collection_location": "California, USA",
		"Water Control":       "No",
		"Nucleotide Type":     "DNA",
	})

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	mapM := map[string]interface{}{}
	err = json.Unmarshal(b, &mapM)
	if err != nil {
		t.Fatal(err)
	}

	if mapM["host_genome"] != "Koala" {
		t.Errorf("expected \"Host Genome\" to be Koala but it was %s", mapM["host_genome"])
	}

	if _, has := mapM["collection_location"]; has {
		t.Error("expected JSON format to filter out collection_location but it did not")
	}
}

func TestFallbackToRawMetadataWithNoSuggestion(t *testing.T) {
	m := NewMetadata(map[string]string{
		"host_genome":         "Koala",
		"collection_location": "Unknown",
		"Water Control":       "No",
		"Nucleotide Type":     "DNA",
	})

	if m.rawCollectionLocation != "Unknown" {
		t.Fatalf("collection_location 'Unknown' should have been extracted into rawCollectionLocation but it was '%s'", m.rawCollectionLocation)
	}

	if m.CollectionLocation != (GeoSearchSuggestion{}) {
		t.Fatalf("CollectionLocation should not be populated bit it was: %s", m.CollectionLocation)
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "Unknown") {
		t.Fatalf("")
	}
}
