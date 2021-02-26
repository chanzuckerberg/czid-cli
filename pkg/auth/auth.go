package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/util"
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

const refreshTokenKey = "SECRET"
const accessTokenKey = "ACCESS_TOKEN"
const expiresAtKey = "EXPIRES_AT"

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
		viper.Set(refreshTokenKey, t.RefreshToken)
	}
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	cache, err := util.ViperCache("auth")
	if t.AccessToken != "" {
		cache.Set("access_token", t.AccessToken)
	}
	cache.Set("expires_at", t.ExpiresAt)
	return cache.WriteConfig()
}

type errorResponse struct {
	ErrorType        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *errorResponse) Error() string {
	return fmt.Sprintf("authentication error: %s", e.ErrorDescription)
}

func formPost(endpoint string, params map[string]string, r interface{}) error {
	var eR errorResponse
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		json.Unmarshal(body, &eR)
		return &eR
	} else {
		return json.Unmarshal(body, &r)
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

func requestDeviceCode(persistent bool) (deviceCodeResponse, error) {
	var d deviceCodeResponse
	endpoint := "https://czi-idseq-dev.auth0.com/oauth/device/code"
	params := map[string]string{
		"client_id": clientID,
		"scope":     "openid",
		"audience":  "https://czi-idseq-dev.auth0.com/api/v2/",
	}
	if persistent {
		params["scope"] = "openid offline_access"
	}
	timeFetched := time.Now()
	err := formPost(endpoint, params, &d)
	d.ExpiresAt = addSeconds(timeFetched, d.ExpiresIn)
	return d, err
}

func promptDeviceActivation(verificantionURIComplete string, headless bool) {
	if headless {
		fmt.Printf("please navigate to %s and authenticate\n", verificantionURIComplete)
	} else {
		fmt.Printf("directing you to authenticate at %s\n", verificantionURIComplete)
		time.Sleep(2 * time.Second)
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
	err := formPost(endpoint, params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

func pollForTokens(interval time.Duration, expiresAt time.Time, deviceCode string) (tokenResponse, error) {
	var tR tokenResponse
	var err error
	for t := range time.Tick(interval) {
		if t.After(expiresAt) {
			return tR, errors.New("expired token")
		}
		tR, err = requestToken(deviceCode)
		if err != nil {
			serr, ok := err.(*errorResponse)
			if ok && serr.ErrorType == "authorization_pending" {
				fmt.Println("waiting for authentication in browser...")
				continue
			} else {
				return tR, err
			}
		}
		break
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
	err := formPost(endpoint, params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

func AccessToken() (string, error) {
	cache, err := util.ViperCache("auth")
	if err != nil {
		return "", nil
	}
	accessToken := cache.GetString(accessTokenKey)
	expiresAt := cache.GetTime(expiresAtKey)
	if accessToken != "" && expiresAt.After(time.Now()) {
		return accessToken, nil
	}
	refreshToken := viper.GetString(refreshTokenKey)
	if refreshToken != "" {
		t, err := refreshAccessToken(refreshToken)
		writeErr := t.writeToConfig()
		if writeErr != nil {
			log.Printf("warning: credential cache failed")
		}
		return t.AccessToken, err
	}
	return "", fmt.Errorf("not authenticated, try running `idseq login` or adding your `secret` to %s manually", viper.GetViper().ConfigFileUsed())
}

func Login(headless bool, persistent bool) error {
	d, err := requestDeviceCode(persistent)
	if err != nil {
		return err
	}
	promptDeviceActivation(d.VerificationURIComplete, headless)
	t, err := pollForTokens(time.Duration(d.Interval)*time.Second, d.ExpiresAt, d.DeviceCode)
	if err != nil {
		return err
	}
	err = t.writeToConfig()
	if err != nil {
		return err
	}
	if persistent {
		// refresh the access token to make sure the user is authenticated
		_, err = refreshAccessToken(t.RefreshToken)
	}
	return err
}
