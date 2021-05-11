package generateMetadataTemplate

import (
	"errors"
	"log"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/spf13/cobra"
)

var forSampleDirectoryCmd = &cobra.Command{
	Use:   "for-sample-directory [directory]",
	Short: "Generate a metadata csv template for a directory of sample files",
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return errors.New("missing required positional argument: directory")
		}
		directory := args[0]

		sampleFiles, err := idseq.SamplesFromDir(directory, verbose)
		if err != nil {
			log.Fatal(err)
		}

		sampleNames := make([]string, 0, len(sampleFiles))
		for sampleName := range sampleFiles {
			sampleNames = append(sampleNames, sampleName)
		}
		generateMetadataTemplate(cmd, output, sampleNames)
		return nil
	},
}

func init() {
	GenerateMetadataTemplateCmd.AddCommand(forSampleDirectoryCmd)
	loadSharedFlags(forSampleDirectoryCmd)
}
