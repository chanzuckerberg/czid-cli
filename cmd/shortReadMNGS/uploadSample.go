package shortReadMNGS

import (
	"errors"
	"fmt"

	"github.com/chanzuckerberg/czid-cli/pkg/czid"
	"github.com/spf13/cobra"
)

var sampleName string

// uploadSampleCmd represents the uploadSample command
var uploadSampleCmd = &cobra.Command{
	Use:   "upload-sample [r1path] [r2path]?",
	Short: "Upload a single sample",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectName == "" {
			return errors.New("missing required argument: project")
		}
		if len(args) == 0 {
			return errors.New("missing required argument: r1path")
		}

		r1path := args[0]
		r2path := ""

		if sampleName == "" {
			sampleName = czid.ToSampleName(r1path)
		}

		sampleFiles := map[string]czid.SampleFiles{
			sampleName: {Single: []string{r1path}},
		}

		if len(args) > 1 {
			r2path = args[1]
			sampleFiles[sampleName] = czid.SampleFiles{R1: []string{r1path}, R2: []string{r2path}}
		}
		if len(args) > 2 {
			return fmt.Errorf("too many positional arguments (maximum 2), args: %v", args)
		}
		if r1path == r2path {
			return errors.New("r1 and r2 cannot be the same file")
		}

		return czid.UploadSamplesFlow(
			sampleFiles,
			stringMetadata,
			projectName,
			metadataCSVPath,
			"short-read-mngs",
			czid.SampleOptions{},
			disableBuffer,
		)
	},
}

func init() {
	ShortReadMNGSCmd.AddCommand(uploadSampleCmd)
	loadSharedFlags(uploadSampleCmd)
	uploadSampleCmd.Flags().StringVarP(&sampleName, "sample-name", "s", "", "Sample name. Optional, defaults to the base file name of r1path with extensions and _R1 removed")
}
