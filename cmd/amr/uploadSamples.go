package amr

import (
	"errors"
	"fmt"
	"log"

	"github.com/chanzuckerberg/czid-cli/pkg/czid"
	"github.com/spf13/cobra"
)

// uploadSamplesCmd represents the uploadSamples command
var uploadSamplesCmd = &cobra.Command{
	Use:   "upload-samples [directory]",
	Short: "Bulk upload many samples",
	Long:  "Bulk upload many samples",
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if err := validateCommonArgs(); err != nil {
			return err
		}
		if len(args) == 0 {
			return errors.New("missing required positional argument: directory")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many positional arguments, (maximum 1), args: %v", args)
		}
		directory := args[0]
		sampleFiles, err := czid.SamplesFromDir(directory, verbose)
		if err != nil {
			log.Fatal(err)
		}

		options := czid.SampleOptions{}

		return czid.UploadSamplesFlow(
			sampleFiles,
			stringMetadata,
			projectName,
			metadataCSVPath,
			"amr",
			options,
			disableBuffer,
		)
	},
}

func init() {
	AmrCmd.AddCommand(uploadSamplesCmd)
	loadSharedFlags(uploadSamplesCmd)
}
