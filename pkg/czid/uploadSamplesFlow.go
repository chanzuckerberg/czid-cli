package czid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/chanzuckerberg/czid-cli/pkg/upload"
)

func UploadSamplesFlow(
	sampleFiles map[string]SampleFiles,
	stringMetadata map[string]string,
	projectName string,
	metadataCSVPath string,
	workflow string,
	sampleOptions SampleOptions,
	disableBuffer bool,
) error {
	projectID, err := DefaultClient.GetProjectID(projectName)
	if err != nil {
		log.Fatal(err)
	}

	samplesMetadata, err := GetCombinedMetadata(sampleFiles, stringMetadata, metadataCSVPath)
	if err != nil {
		log.Fatal(err)
	}

	sampleNames := make([]string, 0, len(sampleFiles))
	for sampleName := range samplesMetadata {
		sampleNames = append(sampleNames, sampleName)
	}
	newSampleNames, err := DefaultClient.ValidateSampleNames(sampleNames, projectID)
	if err != nil {
		log.Fatal(err)
	}
	if len(sampleNames) != len(newSampleNames) {
		log.Fatal("error validating sample names")
	}
	for i := range sampleNames {
		if newSampleNames[i] != sampleNames[i] {
			samplesMetadata[newSampleNames[i]] = samplesMetadata[sampleNames[i]]
			delete(samplesMetadata, sampleNames[i])
			sampleFiles[newSampleNames[i]] = sampleFiles[sampleNames[i]]
			delete(sampleFiles, sampleNames[i])
		}
	}

	err = GeoSearchSuggestions(&samplesMetadata)
	if err != nil {
		log.Fatal(err)
	}
	err = DefaultClient.ValidateSamplesMetadata(projectID, samplesMetadata)
	if err != nil {
		if err.Error() == "metadata validation failed" {
			os.Exit(1)
		}
		log.Fatal(err)
	}

	samples, err := DefaultClient.CreateSamples(
		projectID,
		sampleFiles,
		samplesMetadata,
		workflow,
		sampleOptions,
	)
	if err != nil {
		log.Fatal(err)
	}

	var credentials aws.Credentials
	for _, sample := range samples {
		credentials, err = DefaultClient.GetUploadCredentials(sample.ID)
		if err != nil {
			log.Fatal(err)
		}
		u := upload.NewUploader(credentials, disableBuffer)
		sF := sampleFiles[sample.Name]
		for _, inputFile := range sample.InputFiles {
			var filenames []string
			if len(sF.R1) > 0 && filepath.Base(StripLaneNumber(sF.R1[0])) == filepath.Base(inputFile.S3Path) {
				filenames = sF.R1
			} else if len(sF.R2) > 0 && filepath.Base(StripLaneNumber(sF.R2[0])) == filepath.Base(inputFile.S3Path) {
				filenames = sF.R2
			} else if len(sF.Single) > 0 && filepath.Base(sF.Single[0]) == filepath.Base(inputFile.S3Path) {
				filenames = sF.Single
			} else if len(sF.ReferenceFasta) > 0 && filepath.Base(sF.ReferenceFasta[0]) == filepath.Base(inputFile.S3Path) {
				filenames = sF.ReferenceFasta
			} else if len(sF.PrimerBed) > 0 && filepath.Base(sF.PrimerBed[0]) == filepath.Base(inputFile.S3Path) {
				filenames = sF.PrimerBed
			} else {
				allFilenames := []string{}
				if len(sF.R1) > 0 {
					allFilenames = append(allFilenames, StripLaneNumber(sF.R1[0]))
				}
				if len(sF.R2) > 0 {
					allFilenames = append(allFilenames, StripLaneNumber(sF.R2[0]))
				}
				if len(sF.Single) > 0 {
					allFilenames = append(allFilenames, StripLaneNumber(sF.Single[0]))
				}
				if len(sF.ReferenceFasta) > 0 {
					allFilenames = append(allFilenames, StripLaneNumber(sF.ReferenceFasta[0]))
				}
				if len(sF.PrimerBed) > 0 {
					allFilenames = append(allFilenames, StripLaneNumber(sF.PrimerBed[0]))
				}

				return fmt.Errorf("s3 path %s did not match any of %s", inputFile.S3Path, strings.Join(allFilenames, ", "))
			}
			err := u.UploadFiles(filenames, inputFile.S3Path, inputFile.MultipartUploadId)
			if err != nil {
				log.Fatal(err)
			}
		}
		err := DefaultClient.MarkSampleUploaded(sample.ID, sample.Name)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
