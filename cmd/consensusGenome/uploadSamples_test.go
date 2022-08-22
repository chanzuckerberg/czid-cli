package consensusGenome

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestClearLabsWithIncorrectParams(t *testing.T) {
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

func TestMain(m *testing.M) {
	viper.Set("ACCEPTED_USER_AGREEMENT", "Y")
	ConsensusGenomeCmd.PersistentFlags().Bool("verbose", false, "")
	code := m.Run()
	os.Exit(code)
}
