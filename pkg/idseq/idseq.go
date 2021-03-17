package idseq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0"
)

var baseURL = ""
var roleARN = ""

func request(method string, path string, reqBody interface{}, resBody interface{}) error {
	token, err := auth0.IdToken()
	if err != nil {
		return err
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	u.Path = path

	req, err := http.NewRequest(method, u.String(), bytes.NewReader(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("idseq API responded with error code %d", res.StatusCode)
	}
	defer res.Body.Close()

	resBodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(resBodyBytes, resBody)
}

type sample struct {
	Name      string `json:"name"`
	ProjectID int    `json:"project_id"`
}

type metadata struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type validateCSVReq struct {
	Metadata metadata `json:"metadata"`
	Samples  []sample `json:"samples"`
}

// TODO: handle issue groups
type issues struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type validateCSVRes struct {
	Status string `json:"status"`
	Issues issues `json:"issues"`
}

func ValidateCSV() (validateCSVRes, error) {
	reqBody := validateCSVReq{
		Metadata: metadata{
			Headers: []string{},
			Rows:    [][]string{},
		},
		Samples: []sample{{
			Name:      "foo",
			ProjectID: 6,
		}},
	}

	var resBody validateCSVRes

	err := request("POST", "metadata/validate_csv_for_new_samples.json", reqBody, &resBody)
	if err != nil {
		return resBody, err
	}

	return resBody, nil
}

type inputFileAttribute struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	Parts      string `json:"parts"`
}

type uploadableSample struct {
	Name                string               `json:"name"`
	ProjectID           int                  `json:"project_id"`
	InputFileAttributes []inputFileAttribute `json:"input_files_attributes"`
	HostGenomeName      string               `json:"host_genome_name"`
	Status              string               `json:"status"`
}

type samplesReq struct {
	Samples  []uploadableSample           `json:"samples"`
	Metadata map[string]map[string]string `json:"metadata"`
	Client   string                       `json:"client"`
}

type inputFile struct {
	S3Path            string  `json:"s3_path"`
	MultipartUploadId *string `json:"multipart_upload_id"`
}

type sampleRes struct {
	ID         int         `json:"id"`
	InputFiles []inputFile `json:"input_files"`
}

type samplesRes struct {
	Credentials aws.Credentials `json:"credentials"`
	Samples     []sampleRes     `json:"samples"`
}

func UploadSample(sampleName string) (samplesRes, error) {
	reqBody := samplesReq{
		Samples: []uploadableSample{{
			Name:      sampleName,
			ProjectID: 6,
			InputFileAttributes: []inputFileAttribute{{
				Name:       "test.fasta",
				Source:     "test.fasta",
				SourceType: "local",
				Parts:      "test.fasta",
			}},
			HostGenomeName: "human",
			Status:         "created",
		}},
		Metadata: map[string]map[string]string{
			sampleName: {
				"Collection Date":     "2021-03",
				"Collection Location": "California, USA",
				"Nucleotide Type":     "DNA",
				"Sample Type":         "Cerebrospinal fluid",
				"Water Control":       "No",
			},
		},
		Client: "0.8.10",
	}
	var resBody samplesRes
	err := request("POST", "samples/bulk_upload_with_metadata.json", reqBody, &resBody)
	return resBody, err
}

type updateRequestSample struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type updateRequest struct {
	Sample updateRequestSample `json:"sample"`
}

type updateResponse struct{}

func MarkSampleUploaded(sampleId int, sampleName string) error {
	req := updateRequest{
		Sample: updateRequestSample{
			Id:     sampleId,
			Name:   sampleName,
			Status: "uploaded",
		},
	}

	var res updateRequest
	return request("PUT", fmt.Sprintf("samples/%d.json", sampleId), req, &res)
}
