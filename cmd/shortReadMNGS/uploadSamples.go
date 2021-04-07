package shortReadMNGS

import (
	"errors"
	"fmt"
	"log"
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
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if projectName == "" {
			return errors.New("missing required argument: project")
		}
		if len(args) == 0 {
			return errors.New("missing required positional argument: directory")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many positional arguments, (maximum 1), args: %v", args)
		}
		directory := args[0]
		metadata := idseq.NewMetadata(stringMetadata)
		sampleFiles, err := idseq.SamplesFromDir(directory, verbose)
		if err != nil {
			log.Fatal(err)
		}

		projectID, err := idseq.GetProjectID(projectName)
		if err != nil {
			log.Fatal(err)
		}

		samplesMetadata := idseq.SamplesMetadata{}
		if metadataCSVPath != "" {
			samplesMetadata, err = idseq.CSVMetadata(metadataCSVPath)
			if err != nil {
				log.Fatal(err)
			}
			for sampleName := range samplesMetadata {
				if _, hasSampleName := sampleFiles[sampleName]; !hasSampleName {
					delete(samplesMetadata, sampleName)
				}
			}
		}
		for sampleName := range sampleFiles {
			if _, hasMetadata := samplesMetadata[sampleName]; !hasMetadata {
				samplesMetadata[sampleName] = idseq.Metadata{}
			}
		}
		for sampleName, m := range samplesMetadata {
			samplesMetadata[sampleName] = m.Fuse(metadata)
		}
		err = idseq.GeoSearchSuggestions(&samplesMetadata)
		if err != nil {
			log.Fatal(err)
		}
		err = idseq.ValidateSamplesMetadata(projectID, samplesMetadata)
		if err != nil {
			if err.Error() == "metadata validation failed" {
				os.Exit(1)
			}
			log.Fatal(err)
		}

		credentials, samples, err := idseq.CreateSamples(projectID, sampleFiles, samplesMetadata)
		if err != nil {
			log.Fatal(err)
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
					log.Fatal(err)
				}
			}
			err := idseq.MarkSampleUploaded(sample.ID, sample.Name)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	},
}

func init() {
	ShortReadMNGSCmd.AddCommand(uploadSamplesCmd)
	loadSharedFlags(uploadSamplesCmd)
}
