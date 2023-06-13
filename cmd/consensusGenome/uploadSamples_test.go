package consensusGenome

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetVars() {
	projectName = ""
	stringMetadata = map[string]string{}
	metadataCSVPath = ""
	technology = ""
	wetlabProtocol = ""
	medakaModel = ""
	clearLabs = false
	disableBuffer = false
	technologyOptionsString = ""
	wetlabProtocolOptionsString = ""
	nanoporeWetlabProtocolOptionsString = ""
	medakaModelsString = ""
	referenceAccession = ""
	referenceFasta = ""
	primerBed = ""
}

func TestClearLabsWithIncorrectParams(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Nanopore", "--clearlabs", "--wetlab-protocol", "Midnight"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if string(errOut) != "Error: wetlab-protocol ARTIC v3 is required with clearlabs\n" {
		t.Fatalf("expected a wetlab protocol error but error was: %s", string(errOut))
	}
}

func TestNanoporeWithIncorrectParams(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Nanopore", "--wetlab-protocol", "MSSPE"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(errOut), "Error: wetlab protocol \"MSSPE\" not supported, please choose one of: ") {
		t.Fatalf("expected a wetlab protocol error but error was: %s", string(errOut))
	}
}

func TestReferenceFastaWithNanopore(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Nanopore", "--reference-fasta", "foo.fasta"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if string(errOut) != "Error: reference-accession, reference-fasta, and primer-bed require sequencing-platform 'Illumina'\n" {
		t.Fatalf("expected a sequencing-platform error but error was: %s", string(errOut))
	}
}

func TestReferenceFastaWetlabProtocol(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Illumina", "--reference-fasta", "foo.fasta", "--wetlab-protocol", "SNAP"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if string(errOut) != "Error: wetlab-protocol is not supported with reference-accession, reference-fasta, or primer-bed\n" {
		t.Fatalf("expected a sequencing-platform error but error was: %s", string(errOut))
	}
}

func TestReferenceReferenceFastaMissingPrimerBed(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Illumina", "--reference-fasta", "foo.fasta"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		return
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
		t.Fatalf("Unexpected %s", string(errOut))
	}
}

func TestReferencePrimerBedMissingReferenceFasta(t *testing.T) {
	resetVars()
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	ConsensusGenomeCmd.SetOut(b)
	ConsensusGenomeCmd.SetErr(e)
	ConsensusGenomeCmd.SetArgs([]string{"upload-samples", "-p", "test", "--sequencing-platform", "Illumina", "--primer-bed", "primer.bed"})
	err := ConsensusGenomeCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if string(errOut) != "Error: primer-bed requires reference-accession or reference-fasta\n" {
		t.Fatalf("expected a sequencing-platform error but error was: %s", string(errOut))
	}
}

func TestMain(m *testing.M) {
	viper.Set("ACCEPTED_USER_AGREEMENT", "Y")
	ConsensusGenomeCmd.PersistentFlags().Bool("verbose", false, "")
	code := m.Run()
	os.Exit(code)
}
