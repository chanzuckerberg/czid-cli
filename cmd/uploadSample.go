package cmd

import (
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/upload"
	"github.com/spf13/cobra"
)

// uploadSampleCmd represents the uploadSample command
var uploadSampleCmd = &cobra.Command{
	Use:   "upload-sample",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO proper sample name
		sampleName := args[0]
		r, err := idseq.UploadSample(sampleName)
		if err != nil {
			return err
		}

		uploader := upload.NewUploader(r.Credentials)
		err = uploader.UploadFile("test.fasta", r.Samples[0].InputFiles[0].S3Path, r.Samples[0].InputFiles[0].MultipartUploadId)
		if err != nil {
			return err
		}

		return idseq.MarkSampleUploaded(r.Samples[0].ID, sampleName)
	},
}

func init() {
	RootCmd.AddCommand(uploadSampleCmd)
}
