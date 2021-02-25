package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var clientID string = ""

type deviceCodeResponse struct {
	DeviceCode              string    `json:"device_code"`
	UserCode                string    `json:"user_code"`
	VerificationURI         string    `json:"verification_uri"`
	VerificationURIComplete string    `json:"verification_uri_complete"`
	ExpiresIn               int       `json:"expires_in"`
	ExpiresAt               time.Time `json:"-"`
	Interval                int       `json:"interval"`
}

type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IdToken      string    `json:"id_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"-"`
	Scope        string    `json:"scope"`
}

func (t tokenResponse) writeToConfig() error {
	if t.RefreshToken != "" {
		viper.Set("refresh_token", t.RefreshToken)
	}
	if t.AccessToken != "" {
		viper.Set("access_token", t.AccessToken)
	}
	viper.Set("expires_at", t.ExpiresAt)
	return viper.WriteConfig()
}

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func formPost(endpoint string, params map[string]string, r interface{}) (int, errorResponse, error) {
	var eR errorResponse
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		return 0, eR, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return res.StatusCode, eR, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, eR, err
	}

	if res.StatusCode < 400 {
		return res.StatusCode, eR, json.Unmarshal(body, r)
	} else {
		json.Unmarshal(body, &eR)
		return res.StatusCode, eR, nil
	}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func addSeconds(t time.Time, s int) time.Time {
	return t.Add(time.Duration(s) * time.Second)
}

func requestDeviceCode() (deviceCodeResponse, error) {
	var d deviceCodeResponse
	endpoint := "https://czi-idseq-dev.auth0.com/oauth/device/code"
	params := map[string]string{
		"client_id": clientID,
		"scope":     "openid offline_access",
		"audience":  "https://czi-idseq-dev.auth0.com/api/v2/",
	}
	timeFetched := time.Now()
	_, _, err := formPost(endpoint, params, &d)
	d.ExpiresAt = addSeconds(timeFetched, d.ExpiresIn)
	return d, err
}

func requestDeviceActivation(verificantionURIComplete string, headless bool) {
	if headless {
		fmt.Printf("please navigate to %s and authenticate\n", verificantionURIComplete)
	} else {
		fmt.Printf("directing you to authenticate at %s\n", verificantionURIComplete)
		time.Sleep(5 * time.Second)
		openBrowser(verificantionURIComplete)
	}
}

func requestToken(deviceCode string) (tokenResponse, error) {
	endpoint := "https://czi-idseq-dev.auth0.com/oauth/token"
	var t tokenResponse
	params := map[string]string{
		"client_id":   clientID,
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}
	timeFetched := time.Now()
	_, _, err := formPost(endpoint, params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

func pollForTokens(interval time.Duration, expiresAt time.Time, deviceCode string) (tokenResponse, error) {
	var tR tokenResponse
	for t := range time.Tick(interval) {
		if t.After(expiresAt) {
			return tR, errors.New("expired token")
		}
		tR, err := requestToken(deviceCode)
		if err == nil {
			return tR, err
		}
		fmt.Println(err.Error())
		fmt.Println("waiting for authentication in browser...")
	}
	return tR, nil
}

func refreshAccessToken(refreshToken string) (tokenResponse, error) {
	var t tokenResponse
	endpoint := "https://czi-idseq-dev.auth0.com/oauth/token"
	params := map[string]string{
		"client_id":     clientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	timeFetched := time.Now()
	_, _, err := formPost(endpoint, params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

func getAccessToken() string {
	accessToken := viper.GetString("access_token")
	expiresAt := viper.GetTime("expires_at")
	if accessToken != "" && expiresAt.After(time.Now()) {
		return accessToken
	}
	refreshToken := viper.GetString("refresh_token")
	if refreshToken != "" {
		t, _ := refreshAccessToken(refreshToken)
		t.writeToConfig()
		return t.AccessToken
	} else {
		log.Fatalf("not authenticated, try running `idseq login` or adding your `refresh_token` to %s manually", viper.GetViper().ConfigFileUsed())
		return ""
	}
}

func Login() {
	d, err := requestDeviceCode()
	if err != nil {
		panic(err)
	}
	requestDeviceActivation(d.VerificationURIComplete, false)
	t, err := pollForTokens(time.Duration(d.Interval)*time.Second, d.ExpiresAt, d.DeviceCode)
	t.writeToConfig()
}
