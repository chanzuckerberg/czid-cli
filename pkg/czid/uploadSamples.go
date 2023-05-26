package czid

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/czid-cli/pkg"
)

type createSamplesReqInputFile struct {
	FileType	 string `json:"file_type"`
	Name         string `json:"name"`
	Parts        string `json:"parts"`
	Source       string `json:"source"`
	SourceType   string `json:"source_type"`
	UploadClient string `json:"upload_client"`
}

type CreateSamplesReqSample struct {
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
	ReferenceAccession  *string                     `json:"accession_id,omitempty"`
	ReferenceFasta      *string                     `json:"ref_fasta,omitempty"`
	PrimerBed           *string                     `json:"primer_bed,omitempty"`
}

type samplesReq struct {
	Client   string                   `json:"client"`
	Metadata SamplesMetadata          `json:"metadata"`
	Samples  []CreateSamplesReqSample `json:"samples"`
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

// Must match file type constants present in CZID's InputFile model
type inputFileType string
const (
	FASTQFileType inputFileType = "fastq"
	PrimerBedFileType inputFileType = "primer_bed"
	ReferenceSequenceFileType inputFileType = "reference_sequence"
)

type inputFileMetadata struct {
	Filename string
	FileType inputFileType
}

// CreateSamples creates samples on the back end and returns the necessary information to upload their files
func (c *Client) CreateSamples(
	projectID int,
	sampleFiles map[string]SampleFiles,
	samplesMetadata SamplesMetadata,
	workflow string,
	sampleOptions SampleOptions,
) ([]createSamplesResSample, error) {
	req := samplesReq{
		Metadata: samplesMetadata,
		Client:   pkg.VersionNumber(),
	}

	for sampleName := range samplesMetadata {
		files := sampleFiles[sampleName]
		var filesMetadata []inputFileMetadata
		if len(files.Single) > 0 {
			metadata := inputFileMetadata {
				Filename: StripLaneNumber(files.Single[0]),
				FileType: FASTQFileType,
			}
			filesMetadata = []inputFileMetadata{metadata}
		} else {
			metadataR1 := inputFileMetadata {
				Filename: StripLaneNumber(files.R1[0]),
				FileType: FASTQFileType,
			}
			metadataR2 := inputFileMetadata {
				Filename: StripLaneNumber(files.R2[0]),
				FileType: FASTQFileType,
			}
			filesMetadata = []inputFileMetadata {metadataR1, metadataR2}
		}

		if len(files.ReferenceFasta) > 0 {
			metadata := inputFileMetadata { 
				Filename: files.ReferenceFasta[0],
				FileType: ReferenceSequenceFileType,
			}
			filesMetadata = append(filesMetadata, metadata)
		}
		if len(files.PrimerBed) > 0 {
			metadata := inputFileMetadata {
				Filename: files.PrimerBed[0],
				FileType: PrimerBedFileType,
			}
			filesMetadata = append(filesMetadata, metadata)
		}

		sample := CreateSamplesReqSample{
			HostGenomeName:      samplesMetadata[sampleName].HostGenome,
			InputFileAttributes: make([]createSamplesReqInputFile, len(filesMetadata)),
			Name:                sampleName,
			ProjectID:           projectID,
			Status:              "created",
			Workflows:           []string{workflow},
		}

		if sampleOptions.Technology != "" {
			sample.Technology = sampleOptions.Technology
		}

		if sampleOptions.WetlabProtocol != "" {
			sample.WetlabProtocol = sampleOptions.WetlabProtocol
		}

		if sampleOptions.MedakaModel != "" {
			sample.MedakaModel = &sampleOptions.MedakaModel
		}

		if sampleOptions.ClearLabs {
			sample.ClearLabs = &sampleOptions.ClearLabs
		}

		if sampleOptions.ReferenceAccession != "" {
			sample.ReferenceAccession = &sampleOptions.ReferenceAccession
		}

		if sampleOptions.ReferenceFasta != "" {
			referenceFasta := filepath.Base(sampleOptions.ReferenceFasta)
			sample.ReferenceFasta = &referenceFasta
		}

		if sampleOptions.PrimerBed != "" {
			primerBed := filepath.Base(sampleOptions.PrimerBed)
			sample.PrimerBed = &primerBed
		}

		for i, metadata := range filesMetadata {
			sample.InputFileAttributes[i] = createSamplesReqInputFile{
				FileType:	  string(metadata.FileType),
				Name:         filepath.Base(metadata.Filename),
				Parts:        filepath.Base(metadata.Filename),
				Source:       filepath.Base(metadata.Filename),
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