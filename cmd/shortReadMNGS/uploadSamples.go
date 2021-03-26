package shortReadMNGS

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
		fmt.Printf("args: %v\n", args)
		if projectName == "" {
			return errors.New("missing required argument project")
		}
		if len(args) == 0 {
			return errors.New("missing required positional argument directory")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many positional arguments %d found 1 expected", len(args))
		}
		metadata := make(idseq.Metadata, len(stringMetadata))
		for k, v := range stringMetadata {
			metadata[k] = v
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
		err = idseq.GeoSearchSuggestions(&samplesMetadata)
		if err != nil {
			return err
		}
		err = idseq.ValidateSamplesMetadata(projectID, samplesMetadata)
		if err != nil {
			if err.Error() == "metadata validation failed" {
				os.Exit(1)
			}
			return err
		}

		credentials, samples, err := idseq.CreateSamples(projectID, sampleFiles, samplesMetadata)
		if err != nil {
			return err
		}

		u := upload.NewUploader(credentials)
		for _, sample := range samples {
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
	ShortReadMNGSCmd.AddCommand(uploadSamplesCmd)
	loadSharedFlags(uploadSamplesCmd)
}
