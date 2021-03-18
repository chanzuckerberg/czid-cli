package idseq

// This file is for interracting with the idseq API

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0"
)

var baseURL = ""
var roleARN = ""

func request(method string, path string, query string, reqBody interface{}, resBody interface{}) error {
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
	u.RawQuery = query

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

type Sample struct {
	Name      string `json:"name"`
	ProjectID int    `json:"project_id"`
}

type validationMetadata struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type validateCSVReq struct {
	Metadata validationMetadata `json:"metadata"`
	Samples  []Sample           `json:"samples"`
}

type metadataIssueRow struct {
	items []string
}

func (m *metadataIssueRow) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	m.items = make([]string, len(v))
	for i, item := range v {
		switch item := item.(type) {
		case string:
			m.items[i] = item
		case float64:
			m.items[i] = fmt.Sprint(item)
		default:
			return fmt.Errorf("expected elements of metadata issue rows to be strings or numbers but %s contained an element of a different type", string(data))
		}
	}

	return nil
}

type detailedMetadataIssue struct {
	Caption string             `json:"caption"`
	Rows    []metadataIssueRow `json:"rows"`
	Headers []string           `json:"headers"`
	isGroup bool
}

type metadataIssue struct {
	StringError   string
	DetailedIssue detailedMetadataIssue
}

func (mI *metadataIssue) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case string:
		mI.StringError = v
	case map[string]interface{}:
		var dI detailedMetadataIssue
		if err := json.Unmarshal(data, &dI); err != nil {
			return err
		}
		mI.DetailedIssue = dI
	default:
		return errors.New("unable to parse metadataIssue")
	}
	return nil
}

func (m metadataIssue) FriendlyPrint() {
	if m.StringError != "" {
		fmt.Println(m.StringError)
	} else {
		fmt.Printf("  %s\n", m.DetailedIssue.Caption)
		for _, row := range m.DetailedIssue.Rows {
			for i, header := range m.DetailedIssue.Headers {
				fmt.Printf("      %s: %s\n", header, row.items[i])
			}
			fmt.Println("")
		}
	}
	fmt.Println("")
}

type Issues struct {
	Errors   []metadataIssue `json:"errors"`
	Warnings []metadataIssue `json:"warnings"`
}

func (i Issues) FriendlyPrint() {
	if len(i.Errors) == 0 && len(i.Warnings) == 0 {
		return
	}
	fmt.Printf("found %d errors and %d warnings\n\n", len(i.Errors), len(i.Warnings))
	if len(i.Errors) > 0 {
		fmt.Println("errors:")
		for _, issue := range i.Errors {
			issue.FriendlyPrint()
		}
	}
	if len(i.Warnings) > 0 {
		fmt.Println("warnings:")
		for _, issue := range i.Warnings {
			issue.FriendlyPrint()
		}
	}

}

type validateCSVRes struct {
	Status string `json:"status"`
	Issues Issues `json:"issues"`
}

func ValidateCSV(samples []Sample, vM validationMetadata) (validateCSVRes, error) {
	reqBody := validateCSVReq{
		Metadata: vM,
		Samples:  samples,
	}

	var resBody validateCSVRes

	err := request("POST", "metadata/validate_csv_for_new_samples.json", "", reqBody, &resBody)
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

func UploadSample(sampleName string, samplesMetadata SamplesMetadata, inputFiles []string) (samplesRes, error) {
	reqBody := samplesReq{
		Samples: []uploadableSample{{
			Name:                sampleName,
			ProjectID:           6,
			InputFileAttributes: make([]inputFileAttribute, len(inputFiles)),
			HostGenomeName:      "human",
			Status:              "created",
		}},
		Metadata: samplesMetadata,
		Client:   "0.8.10",
	}
	for i, inputFile := range inputFiles {
		reqBody.Samples[0].InputFileAttributes[i] = inputFileAttribute{
			Name:       inputFile,
			Source:     inputFile,
			SourceType: "local",
			Parts:      "test.fasta",
		}
	}
	var resBody samplesRes
	err := request("POST", "samples/bulk_upload_with_metadata.json", "", reqBody, &resBody)
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
	return request("PUT", fmt.Sprintf("samples/%d.json", sampleId), "", req, &res)
}

type listProjectsRes struct{}
type project struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}
type listProjectsResp struct {
	Projects []project `json:"projects"`
}

func GetProjectID(projectName string) (int, error) {
	query := url.Values{"basic": []string{"true"}}
	var resp listProjectsResp
	err := request("GET", "projects.json", query.Encode(), listProjectsRes{}, &resp)
	if err != nil {
		return 0, err
	}
	var projectId int
	found := false
	for _, project := range resp.Projects {
		if project.Name == projectName {
			projectId = project.Id
			found = true
			break
		}
	}
	if !found {
		return projectId, fmt.Errorf("project '%s' not found", projectName)
	}
	return projectId, nil
}
