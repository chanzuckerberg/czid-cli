package cmd

import (
	"errors"
	"os"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/upload"
	"github.com/spf13/cobra"
)

var metadata idseq.Metadata
var projectName string
var sampleName string
var r1 string
var r2 string
var metadataCSVPath string

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
		if projectName == "" {
			return errors.New("missing required argument project")
		}
		if r1 == "" {
			return errors.New("missing required argument r1")
		}
		if r1 == r2 {
			return errors.New("r1 and r2 cannot have the same filename")
		}
		projectID, err := idseq.GetProjectID(projectName)
		if err != nil {
			return err
		}

		samples := []idseq.Sample{{
			Name:      sampleName,
			ProjectID: projectID,
		}}
		samplesMetadata := idseq.SamplesMetadata{}
		if metadataCSVPath != "" {
			samplesMetadata, err = idseq.CSVMetadata(metadataCSVPath)
			if err != nil {
				return err
			}
			for sN := range samplesMetadata {
				if sN != sampleName {
					delete(samplesMetadata, sN)
				}
			}
		}
		if samplesMetadata[sampleName] == nil {
			samplesMetadata[sampleName] = metadata
		} else {
			for name, value := range metadata {
				samplesMetadata[sampleName][name] = value
			}
		}
		vm := idseq.ToValidateForm(samplesMetadata)
		validationResp, err := idseq.ValidateCSV(samples, vm)
		if err != nil {
			return err
		}
		validationResp.Issues.FriendlyPrint()
		if len(validationResp.Issues.Errors) > 0 {
			os.Exit(1)
		}
		inputFiles := []string{r1}
		if r2 != "" {
			inputFiles = append(inputFiles, r2)
		}
		r, err := idseq.UploadSample(sampleName, samplesMetadata, inputFiles)
		if err != nil {
			return err
		}

		uploader := upload.NewUploader(r.Credentials)
		err = uploader.UploadFile(r1, r.Samples[0].InputFiles[0].S3Path, r.Samples[0].InputFiles[0].MultipartUploadId)
		if err != nil {
			return err
		}

		if r2 != "" {
			err = uploader.UploadFile(r2, r.Samples[0].InputFiles[1].S3Path, r.Samples[0].InputFiles[1].MultipartUploadId)
			if err != nil {
				return err
			}
		}

		return idseq.MarkSampleUploaded(r.Samples[0].ID, sampleName)
	},
}

func init() {
	RootCmd.AddCommand(uploadSampleCmd)
	uploadSampleCmd.Flags().StringToStringVarP(&metadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	uploadSampleCmd.Flags().StringVarP(&sampleName, "sample-name", "s", "", "sample name")
	uploadSampleCmd.Flags().StringVar(&r1, "r1", "", "Read 1 file path")
	uploadSampleCmd.Flags().StringVar(&r2, "r2", "", "Read 2 file path (optional)")
	uploadSampleCmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	uploadSampleCmd.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
}
