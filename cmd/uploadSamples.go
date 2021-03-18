package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/spf13/cobra"
)

// uploadSamplesCmd represents the uploadSamples command
var uploadSamplesCmd = &cobra.Command{
	Use:   "upload-samples [directory]",
	Short: "Bulk upload many samples",
	Long:  "Bulk upload many samples",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectName == "" {
			return errors.New("missing required argument project")
		}
		if len(args) == 0 {
			return errors.New("missing required positional argument directory")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many positional arguments %d found 1 expected", len(args))
		}
		directory := args[0]
		sampleFiles, err := idseq.SamplesFromDir(directory)
		if err != nil {
			return err
		}

		projectID, err := idseq.GetProjectID(projectName)
		if err != nil {
			return err
		}

		samples := []idseq.Sample{}
		for sampleName := range sampleFiles {
			samples = append(samples, idseq.Sample{
				Name:      sampleName,
				ProjectID: projectID,
			})
		}

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
		for sampleName := range sampleFiles {
			if _, hasMetadata := samplesMetadata[sampleName]; !hasMetadata {
				samplesMetadata[sampleName] = idseq.Metadata{}
			}
		}
		for sampleName := range samplesMetadata {
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

		r, err := idseq.UploadSample(sampleName, samplesMetadata, inputFiles)
		return err
	},
}

func init() {
	RootCmd.AddCommand(uploadSamplesCmd)

	uploadSamplesCmd.Flags().StringToStringVarP(&metadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	uploadSamplesCmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	uploadSamplesCmd.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
}
