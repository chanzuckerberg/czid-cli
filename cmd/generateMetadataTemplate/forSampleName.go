package generateMetadataTemplate

import (
	"errors"

	"github.com/spf13/cobra"
)

var forSampleNameCmd = &cobra.Command{
	Use:   "for-sample-name [sample-name]",
	Short: "Generate a metadata csv template for a sample name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("missing required positional argument: sample-name")
		}
		sampleName := args[0]
		generateMetadataTemplate(cmd, output, []string{sampleName})
		return nil
	},
}

func init() {
	GenerateMetadataTemplateCmd.AddCommand(forSampleNameCmd)
	loadSharedFlags(forSampleNameCmd)
}
