package generateMetadataTemplate

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/czid"
	"github.com/spf13/cobra"
)

var stringMetadata map[string]string
var output string

func generateMetadataTemplate(cmd *cobra.Command, output string, sampleNames []string) {
	var writer *csv.Writer
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		writer = csv.NewWriter(f)
	} else {
		writer = csv.NewWriter(cmd.OutOrStdout())
	}

	metadata := czid.NewMetadata(stringMetadata)
	templateCSV, err := czid.DefaultClient.GetTemplateCSV(sampleNames, metadata.HostGenome)
	templateCSV.LazyQuotes = true
	if err != nil {
		log.Fatal(err)
	}

	fieldNames, err := templateCSV.Read()
	if err != nil {
		log.Fatal(err)
	}

	fieldNameSet := make(map[string]bool, len(fieldNames))
	for _, header := range fieldNames {
		fieldNameSet[header] = true
	}

	for name := range stringMetadata {
		if _, alreadyHas := fieldNameSet[name]; !alreadyHas && name != "Host Organism" {
			fieldNames = append(fieldNames, name)
		}
	}

	err = writer.Write(fieldNames)
	if err != nil {
		log.Fatal(err)
	}

	fieldNameToIdx := make(map[string]int, len(fieldNames))
	for idx, name := range fieldNames {
		fieldNameToIdx[name] = idx
	}

	for readRow, err := templateCSV.Read(); err == nil || !errors.Is(err, io.EOF); readRow, err = templateCSV.Read() {
		if err != nil {
			log.Fatal(err)
		}
		writeRow := make([]string, len(fieldNames))
		for name, idx := range fieldNameToIdx {
			if idx < len(readRow) {
				writeRow[idx] = readRow[idx]
			}
			if val, has := stringMetadata[name]; has {
				writeRow[idx] = val
			}
		}
		err = writer.Write(writeRow)
		if err != nil {
			log.Fatal(err)
		}
	}
	writer.Flush()
}

var GenerateMetadataTemplateCmd = &cobra.Command{
	Use:   "generate-metadata-template",
	Short: "Commands related to generating metadata template csvs",
}

func loadSharedFlags(c *cobra.Command) {
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "Metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVarP(&output, "output", "o", "", "Output file path (optional, by default prints to stdout)")
}
