package idseq

import (
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg"
)

type createSamplesReqInputFile struct {
	Name         string `json:"name"`
	Parts        string `json:"parts"`
	Source       string `json:"source"`
	SourceType   string `json:"source_type"`
	UploadClient string `json:"upload_client"`
}

type createSamplesReqSample struct {
	HostGenomeName      string                      `json:"host_genome_name"`
	InputFileAttributes []createSamplesReqInputFile `json:"input_files_attributes"`
	Name                string                      `json:"name"`
	ProjectID           int                         `json:"project_id"`
	Status              string                      `json:"status"`
}

type samplesReq struct {
	Client   string                   `json:"client"`
	Metadata SamplesMetadata          `json:"metadata"`
	Samples  []createSamplesReqSample `json:"samples"`
}

type createSamplesResCredentials struct {
	AccessKeyID     string    `json:"access_key_id"`
	Expiration      time.Time `json:"expiration"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
}

type createSamplesResSample struct {
	Name       string       `json:"name"`
	ID         int          `json:"id"`
	InputFiles []UploadInfo `json:"input_files"`
}

type createSamplesRes struct {
	Credentials createSamplesResCredentials `json:"credentials"`
	Samples     []createSamplesResSample    `json:"samples"`
}

// UploadInfo stores the data necessary to upload a file to s3
type UploadInfo struct {
	MultipartUploadId *string `json:"multipart_upload_id"`
	S3Path            string  `json:"s3_path"`
}

// CreateSamples creates samples on the back end and returns the necessary information to upload their files
func CreateSamples(projectID int, sampleFiles map[string]SampleFiles, samplesMetadata SamplesMetadata) (aws.Credentials, []createSamplesResSample, error) {
	req := samplesReq{
		Metadata: samplesMetadata,
		Client:   pkg.VersionNumber(),
	}

	for sampleName := range samplesMetadata {
		files := sampleFiles[sampleName]
		filenames := []string{files.Single}
		if sampleFiles[sampleName].Single == "" {
			filenames = []string{files.R1, files.R2}
		}

		sample := createSamplesReqSample{
			HostGenomeName:      samplesMetadata[sampleName]["Host Organism"].(string),
			InputFileAttributes: make([]createSamplesReqInputFile, len(filenames)),
			Name:                sampleName,
			ProjectID:           projectID,
			Status:              "created",
		}

		for i, filename := range filenames {
			sample.InputFileAttributes[i] = createSamplesReqInputFile{
				Name:         filepath.Base(filename),
				Parts:        filepath.Base(filename),
				Source:       filepath.Base(filename),
				SourceType:   "local",
				UploadClient: "cli",
			}
		}
		req.Samples = append(req.Samples, sample)
	}

	res := createSamplesRes{}
	err := request("POST", "samples/bulk_upload_with_metadata.json", "", req, &res)

	credentials := aws.Credentials{
		AccessKeyID:     res.Credentials.AccessKeyID,
		Expires:         res.Credentials.Expiration,
		SecretAccessKey: res.Credentials.SecretAccessKey,
		SessionToken:    res.Credentials.SessionToken,
	}
	return credentials, res.Samples, err
}
