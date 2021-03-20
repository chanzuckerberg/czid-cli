package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/upload"
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

		uploadableSamples := []idseq.UploadableSample{}
		for sampleName, sF := range sampleFiles {
			inputFileAttributes := []idseq.InputFileAttribute{}
			if sF.Single != "" {
				inputFileAttributes = append(inputFileAttributes, idseq.NewInputFile(sF.Single))
			} else {
				inputFileAttributes = append(inputFileAttributes, idseq.NewInputFile(sF.R1))
				inputFileAttributes = append(inputFileAttributes, idseq.NewInputFile(sF.R2))
			}

			hostGenome := samplesMetadata[sampleName]["Host Organism"]
			uploadableSamples = append(uploadableSamples, idseq.UploadableSample{
				Name:                sampleName,
				ProjectID:           projectID,
				HostGenomeName:      hostGenome,
				InputFileAttributes: inputFileAttributes,
				Status:              "created",
			})
		}

		r, err := idseq.UploadSample(uploadableSamples, samplesMetadata)
		if err != nil {
			return err
		}

		u := upload.NewUploader(r.Credentials)
		for _, sample := range r.Samples {
			sF := sampleFiles[sample.Name]
			for _, inputFile := range sample.InputFiles {
				filename := ""
				if filepath.Base(sF.R1) == filepath.Base(inputFile.S3Path) {
					filename = sF.R1
				} else if filepath.Base(sF.R2) == filepath.Base(inputFile.S3Path) {
					filename = sF.R2
				} else {
					filename = sF.Single
				}
				err := u.UploadFile(filename, inputFile.S3Path, inputFile.MultipartUploadId)
				if err != nil {
					return err
				}
			}
			err := idseq.MarkSampleUploaded(sample.ID, sample.Name)
			if err != nil {
				return err
			}
		}
		return err
	},
}

func init() {
	shortReadMNGSCmd.AddCommand(uploadSamplesCmd)

	uploadSamplesCmd.Flags().StringToStringVarP(&metadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	uploadSamplesCmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	uploadSamplesCmd.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
}
