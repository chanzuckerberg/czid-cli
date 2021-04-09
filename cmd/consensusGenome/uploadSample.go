package consensusGenome

import (
	"errors"
	"fmt"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
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
			sampleName = idseq.ToSampleName(r1path)
		}

		sampleFiles := map[string]idseq.SampleFiles{
			sampleName: {Single: r1path},
		}

		if len(args) > 1 {
			r2path = args[1]
			sampleFiles[sampleName] = idseq.SampleFiles{R1: r1path, R2: r2path}
		}
		if len(args) > 2 {
			return fmt.Errorf("too many positional arguments (maximum 2), args: %v", args)
		}
		if r1path == r2path {
			return errors.New("r1 and r2 cannot be the same file")
		}
		if technology == "" {
			return errors.New("missing required argument: sequencing-platform")
		}
		if technology == "Illumina" && wetlabProtocol == "" {
			return errors.New("missing required argument: wetlab-protocol")
		}
		if technology == "Nanopore" && wetlabProtocol != "" {
			return errors.New("wetlab-protocol not supported for Nanopore")
		}

		return idseq.UploadSamplesFlow(
			sampleFiles,
			stringMetadata,
			projectName,
			metadataCSVPath,
			"consensus-genome",
			technologies[technology],
			wetlabProtocols[wetlabProtocol],
		)
	},
}

func init() {
	ConsensusGenomeCmd.AddCommand(uploadSampleCmd)
	loadSharedFlags(uploadSampleCmd)
	uploadSampleCmd.Flags().StringVarP(&sampleName, "sample-name", "s", "", "Sample name. Optional, defaults to the base file name of r1path with extensions and _R1 removed")
}
