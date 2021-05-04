package cmd

import (
	"encoding/csv"
	"errors"
	"io"
	"log"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var generateMetadataTemplateCmd = &cobra.Command{
	Use:   "generate-metadata-template [directory]",
	Short: "Generates a template metadata csv file",
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return errors.New("missing required positional argument: directory")
		}
		directory := args[0]

		stringMetadata, err := cmd.Flags().GetStringToString("metadatum")
		if err != nil {
			return err
		}
		metadata := idseq.NewMetadata(stringMetadata)

		sampleFiles, err := idseq.SamplesFromDir(directory, verbose)
		if err != nil {
			log.Fatal(err)
		}

		sampleNames := make([]string, len(sampleFiles))
		i := 0
		for sampleName := range sampleFiles {
			sampleNames[i] = sampleName
			i++
		}

		templateCSV, err := idseq.DefaultClient.GetTemplateCSV(sampleNames, metadata.HostGenome)
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

		writer := csv.NewWriter(cmd.OutOrStdout())
		err = writer.Write(fieldNames)
		if err != nil {
			return err
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
				return err
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(generateMetadataTemplateCmd)
	generateMetadataTemplateCmd.Flags().StringToStringP("metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
}
