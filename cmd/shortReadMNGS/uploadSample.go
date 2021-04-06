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
		metadata := make(idseq.Metadata, len(stringMetadata))
		for k, v := range stringMetadata {
			metadata[k] = v
		}
		r1path := args[0]
		r2path := ""
		if len(args) > 1 {
			r2path = args[1]
		}
		if len(args) > 2 {
			return fmt.Errorf("too many positional arguments (maximum 2), args: %v", args)
		}
		if r1path == r2path {
			return errors.New("r1 and r2 cannot be the same file")
		}
		projectID, err := idseq.GetProjectID(projectName)
		if err != nil {
			log.Fatal(err)
		}

		if sampleName == "" {
			sampleName = idseq.ToSampleName(r1path)
		}

		samplesMetadata := idseq.SamplesMetadata{}
		if metadataCSVPath != "" {
			samplesMetadata, err = idseq.CSVMetadata(metadataCSVPath)
			if err != nil {
				log.Fatal(err)
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

		inputFiles := idseq.SampleFiles{Single: r1path}
		if r2path != "" {
			inputFiles = idseq.SampleFiles{
				R1: r1path,
				R2: r2path,
			}
		}

		credentials, samples, err := idseq.CreateSamples(
			projectID,
			map[string]idseq.SampleFiles{sampleName: inputFiles},
			samplesMetadata,
		)
		if err != nil {
			log.Fatal(err)
		}

		uploader := upload.NewUploader(credentials)
		for _, sample := range samples {
			for _, upload := range sample.InputFiles {
				localPath := r1path
				if filepath.Base(upload.S3Path) == filepath.Base(r2path) {
					localPath = r2path
				}
				err = uploader.UploadFile(localPath, upload.S3Path, upload.MultipartUploadId)
				if err != nil {
					log.Fatal(err)
				}
			}
			err = idseq.MarkSampleUploaded(sample.ID, sampleName)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	},
}

func init() {
	ShortReadMNGSCmd.AddCommand(uploadSampleCmd)
	loadSharedFlags(uploadSampleCmd)
	uploadSampleCmd.Flags().StringVarP(&sampleName, "sample-name", "s", "", "Sample name. Optional, defaults to the base file name of r1path with extensions and _R1 removed")
}
