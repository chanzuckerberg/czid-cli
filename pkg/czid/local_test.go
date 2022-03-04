package czid

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSamplesFromDir(t *testing.T) {
	dirname, err := ioutil.TempDir(".", "samples")
	defer os.RemoveAll(dirname)
	if err != nil {
		t.Error(err)
	}

	filenames := []string{"ABC_L001_R1.fasta", "ABC_L002_R1.fasta", "ABC_L002_R2.fasta", "ABC_L001_R2.fasta", "DEF.fasta"}
	for _, filename := range filenames {
		err := ioutil.WriteFile(path.Join(dirname, filename), []byte{}, fs.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}

	samples, err := SamplesFromDir(dirname, false)

	if err != nil {
		t.Fatal(err)
	}

	if samples["ABC"].R1[0] != path.Join(dirname, "ABC_L001_R1.fasta") {
		t.Fatalf("%s != %s", samples["ABC"].R1[0], path.Join(dirname, "ABC_L001_R1.fasta"))
	}

	if samples["ABC"].R1[1] != path.Join(dirname, "ABC_L002_R1.fasta") {
		t.Fatalf("%s != %s", samples["ABC"].R1[1], path.Join(dirname, "ABC_L002_R1.fasta"))
	}

	if samples["ABC"].R2[0] != path.Join(dirname, "ABC_L001_R2.fasta") {
		t.Fatalf("%s != %s", samples["ABC"].R2[0], path.Join(dirname, "ABC_L001_R2.fasta"))
	}

	if samples["ABC"].R2[1] != path.Join(dirname, "ABC_L002_R2.fasta") {
		t.Fatalf("%s != %s", samples["ABC"].R2[1], path.Join(dirname, "ABC_L002_R2.fasta"))
	}

	if samples["DEF"].Single[0] != path.Join(dirname, "DEF.fasta") {
		t.Fatalf("%s != %s", samples["ABC"].Single[0], path.Join(dirname, "DEF.fasta"))
	}
}

func TestSamplesFromDirMissingLane(t *testing.T) {
	dirname, err := ioutil.TempDir(".", "samples")
	defer os.RemoveAll(dirname)
	if err != nil {
		t.Error(err)
	}

	filenames := []string{"ABC_L001_R1.fasta", "ABC_L002_R1.fasta", "ABC_L002_R2.fasta"}
	for _, filename := range filenames {
		err := ioutil.WriteFile(path.Join(dirname, filename), []byte{}, fs.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}

	_, err = SamplesFromDir(dirname, false)

	if err.Error() != "missmatch in R1 and R2 file count for sample name 'ABC' 2 != 1" {
		t.Fatal(err)
	}
}

func TestSamplesFromDirMissingPair(t *testing.T) {
	dirname, err := ioutil.TempDir(".", "samples")
	defer os.RemoveAll(dirname)
	if err != nil {
		t.Error(err)
	}

	filenames := []string{"ABC_L001_R1.fasta"}
	for _, filename := range filenames {
		err := ioutil.WriteFile(path.Join(dirname, filename), []byte{}, fs.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}

	_, err = SamplesFromDir(dirname, false)

	if err.Error() != "missmatch in R1 and R2 file count for sample name 'ABC' 1 != 0" {
		t.Fatal(err)
	}
}

func TestSamplesFromDirPairAndSingle(t *testing.T) {
	dirname, err := ioutil.TempDir(".", "samples")
	defer os.RemoveAll(dirname)
	if err != nil {
		t.Error(err)
	}

	filenames := []string{"ABC_L001_R1.fasta", "ABC_L001.fasta"}
	for _, filename := range filenames {
		err := ioutil.WriteFile(path.Join(dirname, filename), []byte{}, fs.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}

	_, err = SamplesFromDir(dirname, false)

	if err.Error() != fmt.Sprintf("found R1 file and single end file for sample 'ABC': %s, [%s]", path.Join(dirname, "ABC_L001_R1.fasta"), path.Join(dirname, "ABC_L001.fasta")) {
		t.Fatal(err)
	}
}

func TestStripLaneNumber(t *testing.T) {
	newPath := StripLaneNumber("ABC_L001_R1.fasta")

	if newPath != "ABC_R1.fasta" {
		t.Errorf("'%s' != '%s'", newPath, "ABC_R1.fasta")
	}
}
