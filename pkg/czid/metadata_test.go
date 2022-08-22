package czid

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCSVMetadata(t *testing.T) {
	csv, err := os.CreateTemp("", "*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(csv.Name())

	csvData := []byte("Sample Name\u200b,Host Genome,collection_location,Nucleotide Type\n" +
		"sample one,Human,\"California, USA\",DNA\n\n" +
		"sample two,D\u200bog,\"California, USA\",RNA",
	)

	_, err = csv.Write(csvData)
	if err != nil {
		t.Fatal(err)
	}

	samplesMetadata, err := CSVMetadata(csv.Name())
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
