package shortReadMNGS

import (
	"errors"
	"os"

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
			return errors.New("missing required argument project")
		}
		if len(args) == 0 {
			return errors.New("missing required argument r1path")
		}
		metadata := make(idseq.Metadata, len(stringMetadata))
		for k, v := range stringMetadata {
			metadata[k] = v
		}
		r1path := args[0]
		r2path := ""
		if len(args) > 1 {
			r2path = args[2]
		}
		if len(args) > 2 {
			return errors.New("too many input files (maximum 2)")
		}
		if r1path == r2path {
			return errors.New("r1 and r2 cannot be the same file")
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

		err = idseq.GeoSearchSuggestions(&samplesMetadata)
		if err != nil {
			return err
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
		inputFiles := []string{r1path}
		if r2path != "" {
			inputFiles = append(inputFiles, r2path)
		}

		inputFileAttributes := make([]idseq.InputFileAttribute, len(inputFiles))
		for i, inputFile := range inputFiles {
			inputFileAttributes[i] = idseq.NewInputFile(inputFile)
		}

		uploadableSamples := []idseq.UploadableSample{{
			Name:                sampleName,
			ProjectID:           projectID,
			HostGenomeName:      samplesMetadata[sampleName]["Host Organism"].(string),
			InputFileAttributes: inputFileAttributes,
			Status:              "created",
		}}

		r, err := idseq.UploadSample(uploadableSamples, samplesMetadata)
		if err != nil {
			return err
		}

		if sampleName == "" {
			sampleName = idseq.ToSampleName(r1path)
		}

		uploader := upload.NewUploader(r.Credentials)
		err = uploader.UploadFile(r1path, r.Samples[0].InputFiles[0].S3Path, r.Samples[0].InputFiles[0].MultipartUploadId)
		if err != nil {
			return err
		}

		if r2path != "" {
			err = uploader.UploadFile(r2path, r.Samples[0].InputFiles[1].S3Path, r.Samples[0].InputFiles[1].MultipartUploadId)
			if err != nil {
				return err
			}
		}

		return idseq.MarkSampleUploaded(r.Samples[0].ID, sampleName)
	},
}

func init() {
	ShortReadMNGSCmd.AddCommand(uploadSampleCmd)
	loadSharedFlags(uploadSampleCmd)
	uploadSampleCmd.Flags().StringVarP(&sampleName, "sample-name", "s", "", "Sample name. Optional, defaults to the base file name of r1path with extensions and _R1 removed")
}
