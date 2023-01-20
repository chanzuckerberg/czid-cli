package metagenomics

import (
	"bytes"
	"io"
	"testing"

	"github.com/spf13/viper"
)

func TestWithoutProjectID(t *testing.T) {
	viper.Set("ACCEPTED_USER_AGREEMENT", "Y")
	MetagenomicsCmd.PersistentFlags().Bool("verbose", false, "")
	b := bytes.NewBufferString("")
	e := bytes.NewBufferString("")
	MetagenomicsCmd.SetOut(b)
	MetagenomicsCmd.SetErr(e)
	MetagenomicsCmd.SetArgs([]string{"upload-samples"})
	err := MetagenomicsCmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	errOut, err := io.ReadAll(e)
	if err != nil {
		t.Fatal(err)
	}
	if string(errOut) != "Error: missing required argument: project\n" {
		t.Fatalf("expected a missing project error but error was: %s", string(errOut))
	}
}
