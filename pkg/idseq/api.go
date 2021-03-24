package idseq

// This file is for interracting with the idseq API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0"
)

var baseURL = ""

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

type updateRequestSample struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type updateRequest struct {
	Sample updateRequestSample `json:"sample"`
}

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

type GetGeoSearchSuggestionReq struct{}

type GeoSearchSuggestion struct {
	Name            string `json:"name"`
	GeoLevel        string `json:"geo_level"`
	CountryName     string `json:"country_name"`
	StateName       string `json:"state_name"`
	SubdivisionName string `json:"subdivision_name"`
	CityName        string `json:"city_name"`
	// Lat             float64 `json:"lat"`
	// Lng             float64 `json:"lng"`
	CountryCode string `json:"country_code"`
	// OSMID           int64   `json:"osm_id"`
	// OSMType         string  `json:"osm_type"`
	// LocationID      int64   `json:"locationiq_id"`
}

func GetGeoSearchSuggestion(queryStr string, isHuman bool) (GeoSearchSuggestion, error) {
	query := url.Values{"query": []string{queryStr}, "limit": []string{"1"}}
	resp := []GeoSearchSuggestion{}
	err := request(
		"GET",
		"locations/external_search",
		query.Encode(),
		GetGeoSearchSuggestionReq{},
		&resp,
	)
	result := GeoSearchSuggestion{}
	if len(resp) > 0 {
		result = resp[0]
	}
	if isHuman && result.GeoLevel == "city" {
		if result.SubdivisionName == result.CityName {
			result.SubdivisionName = ""
		}

		result.CityName = ""
		result.Name = ""
		for _, s := range []string{result.SubdivisionName, result.StateName, result.CountryName} {
			if s != "" {
				if len(result.Name) > 0 {
					result.Name += ", "
				}
				result.Name += s
			}
		}

		if result.SubdivisionName != "" {
			result.GeoLevel = "subdivision"
		} else if result.StateName != "" {
			result.GeoLevel = "state"
		} else if result.CountryName != "" {
			result.GeoLevel = "country"
		}
	}

	return result, err
}
