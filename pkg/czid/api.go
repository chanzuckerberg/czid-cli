package czid

// This file is for interracting with the czid API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/chanzuckerberg/czid-cli/pkg/auth0"
)

var defaultCZIDBaseURL = ""

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	auth0      auth0.Auth0
	httpClient HTTPClient
}

var DefaultClient = &Client{
	auth0:      auth0.DefaultClient,
	httpClient: http.DefaultClient,
}

func (c *Client) authorizedRequest(req *http.Request) (*http.Response, error) {
	token, err := c.auth0.IDToken()
	if err != nil {
		return nil, err
	}

	baseURLString := defaultCZIDBaseURL
	if viper.IsSet("czid_base_url") {
		baseURLString = viper.GetString("czid_base_url")
	}
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = baseURL.Scheme
	req.URL.Host = baseURL.Host

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return res, err
	}

	// TODO: don't exit, return an error type
	if res.StatusCode == 401 || res.StatusCode == 403 {
		fmt.Println("not authenticated with czid try running `czid login`")
		os.Exit(1)
	}
	if res.StatusCode == 426 {
		fmt.Println("czid-cli version out of date, please install the latest version here: `https://github.com/chanzuckerberg/czid-cli`")
		os.Exit(1)
	}
	if res.StatusCode >= 400 {
		return res, fmt.Errorf("czid API responded with error code %d", res.StatusCode)
	}

	return res, nil
}

func (c *Client) request(method string, path string, query string, reqBody interface{}, resBody interface{}) error {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	u := url.URL{
		Path:     path,
		RawQuery: query,
	}
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(reqBodyBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.authorizedRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBodyBytes, err := io.ReadAll(res.Body)
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

func (c *Client) MarkSampleUploaded(sampleId int, sampleName string) error {
	req := updateRequest{
		Sample: updateRequestSample{
			Id:     sampleId,
			Name:   sampleName,
			Status: "uploaded",
		},
	}

	var res updateRequest
	return c.request("PUT", fmt.Sprintf("/samples/%d.json", sampleId), "", req, &res)
}

type listProjectsRes struct{}
type project struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}
type listProjectsResp struct {
	Projects []project `json:"projects"`
}

func (c *Client) GetProjectID(projectName string) (int, error) {
	query := url.Values{"basic": []string{"true"}}
	var resp listProjectsResp
	err := c.request("GET", "/projects.json", query.Encode(), listProjectsRes{}, &resp)
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

func (g GeoSearchSuggestion) String() string {
	places := []string{}
	if g.CityName != "" {
		places = append(places, g.CityName)
	}
	if g.SubdivisionName != "" {
		places = append(places, g.SubdivisionName)
	}
	if g.StateName != "" {
		places = append(places, g.StateName)
	}
	if g.CountryName != "" {
		places = append(places, g.CountryName)
	}
	return strings.Join(places, ", ")
}

func (c *Client) GetGeoSearchSuggestion(queryStr string, isHuman bool) (GeoSearchSuggestion, error) {
	query := url.Values{"query": []string{queryStr}, "limit": []string{"1"}}
	resp := []GeoSearchSuggestion{}
	err := c.request(
		"GET",
		"/locations/external_search",
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
