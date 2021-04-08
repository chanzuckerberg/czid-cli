package idseq

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCSVMetadata(t *testing.T) {
	csv, err := ioutil.TempFile("", "*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(csv.Name())

	csvData := []byte(`Sample Name,Host Genome,collection_location,Nucleotide Type
sample one,Human,"California, USA",DNA
sample two,Dog,"California, USA",RNA
`)

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
		t.Errorf("expected rawCollectionLocation to be \"California, USA\" but it was %s", samplesMetadata["sample one"].rawCollectionLocation)
	}
}
