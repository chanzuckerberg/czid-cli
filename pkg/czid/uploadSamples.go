package czid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/czid-cli/pkg"
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
	Workflows           []string                    `json:"workflows"`
	Technology          string                      `json:"technology"`
	WetlabProtocol      string                      `json:"wetlab_protocol"`
	MedakaModel         *string                     `json:"medaka_model,omitempty"`
	ClearLabs           *bool                       `json:"clearlabs,omitempty"`
}

type samplesReq struct {
	Client   string                   `json:"client"`
	Metadata SamplesMetadata          `json:"metadata"`
	Samples  []createSamplesReqSample `json:"samples"`
}

type createSamplesResSample struct {
	Name       string       `json:"name"`
	ID         int          `json:"id"`
	InputFiles []UploadInfo `json:"input_files"`
}

type createSamplesRes struct {
	Samples []createSamplesResSample `json:"samples"`
	Errors  []string                 `json:"errors"`
}

// UploadInfo stores the data necessary to upload a file to s3
type UploadInfo struct {
	MultipartUploadId *string `json:"multipart_upload_id"`
	S3Path            string  `json:"s3_path"`
}

// CreateSamples creates samples on the back end and returns the necessary information to upload their files
func (c *Client) CreateSamples(
	projectID int,
	sampleFiles map[string]SampleFiles,
	samplesMetadata SamplesMetadata,
	workflow string,
	technology string,
	wetlabProtocol string,
	medakaModel string,
	clearlabs bool,
) ([]createSamplesResSample, error) {
	req := samplesReq{
		Metadata: samplesMetadata,
		Client:   pkg.VersionNumber(),
	}

	for sampleName := range samplesMetadata {
		files := sampleFiles[sampleName]
		var filenames []string
		// TODO concat files instead of using first
		if len(files.Single) > 0 {
			filenames = []string{files.Single[0]}
		} else {
			filenames = []string{files.R1[0], files.R2[0]}
		}

		sample := createSamplesReqSample{
			HostGenomeName:      samplesMetadata[sampleName].HostGenome,
			InputFileAttributes: make([]createSamplesReqInputFile, len(filenames)),
			Name:                sampleName,
			ProjectID:           projectID,
			Status:              "created",
			Workflows:           []string{workflow},
			Technology:          technology,
			WetlabProtocol:      wetlabProtocol,
		}

		if medakaModel != "" {
			sample.MedakaModel = &medakaModel
		}

		if clearlabs {
			sample.ClearLabs = &clearlabs
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
	err := c.request("POST", "/samples/bulk_upload_with_metadata.json", "", req, &res)

	if len(res.Errors) > 0 {
		fmt.Println("encountered errors while uploading")
		for _, e := range res.Errors {
			fmt.Printf("  %s\n", e)
		}
		os.Exit(1)
	}

	return res.Samples, err
}
