// Package auth0 provides methods to authenticate with auth0
package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/chanzuckerberg/czid-cli/pkg/util"
	"github.com/spf13/viper"
)

var defaultClientID string = ""
var defaultAuth0Host string = ""

func clientID() string {
	if viper.IsSet("auth0_client_id") {
		return viper.GetString("auth0_client_id")
	}
	return defaultClientID
}

func auth0Host() string {
	if viper.IsSet("auth0_host") {
		return viper.GetString("auth0_host")
	}
	return defaultAuth0Host
}

const refreshTokenKey = "SECRET"
const idTokenKey = "TOKEN"
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

func (c *Client) saveToken(t tokenResponse) error {
	if t.RefreshToken != "" {
		c.viper.Set(refreshTokenKey, t.RefreshToken)
	}
	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}
	cache, err := c.getCache()
	if err != nil {
		return err
	}
	if t.IdToken != "" {
		cache.Set(idTokenKey, t.IdToken)
	}
	if t.ExpiresIn != 0 {
		cache.Set(expiresAtKey, t.ExpiresAt)
	}
	return cache.WriteConfig()
}

type errorResponse struct {
	ErrorType        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *errorResponse) Error() string {
	return fmt.Sprintf("authentication error: %s", e.ErrorDescription)
}

func formPost(path string, params map[string]string, r interface{}) error {
	endpoint := url.URL{
		Scheme: "https",
		Host:   auth0Host(),
		Path:   path,
	}
	var eR errorResponse
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", endpoint.String(), payload)
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		err := json.Unmarshal(body, &eR)
		if err != nil {
			return err
		}
		return &eR
	} else {
		return json.Unmarshal(body, &r)
	}
}

type Auth0 interface {
	IDToken() (string, error)
	Login(headless bool, persistent bool) error
	Secret() (string, bool)
}

type Client struct {
	formPost func(path string, params map[string]string, r interface{}) error
	viper    *viper.Viper
	cache    *viper.Viper
}

func (c *Client) getCache() (*viper.Viper, error) {
	if c.cache != nil {
		return c.cache, nil
	}
	return util.ViperCache()
}

var DefaultClient = &Client{
	formPost: formPost,
	viper:    viper.GetViper(),
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

func (c *Client) requestDeviceCode(persistent bool) (deviceCodeResponse, error) {
	var d deviceCodeResponse
	audience := url.URL{
		Scheme: "https",
		Host:   auth0Host(),
		Path:   "api/v2/",
	}
	params := map[string]string{
		"client_id": clientID(),
		"scope":     "email openid",
		"audience":  audience.String(),
	}
	if persistent {
		params["scope"] = "email openid offline_access"
	}
	timeFetched := time.Now()
	err := c.formPost("oauth/device/code", params, &d)
	d.ExpiresAt = addSeconds(timeFetched, d.ExpiresIn)
	return d, err
}

func promptDeviceActivation(verificantionURIComplete string, headless bool) {
	if headless {
		fmt.Printf("please navigate to %s and authenticate\n", verificantionURIComplete)
	} else {
		fmt.Printf("directing you to authenticate at %s\n", verificantionURIComplete)
		time.Sleep(2 * time.Second)
		err := openBrowser(verificantionURIComplete)
		if err != nil {
			fmt.Printf("error directing you to %s, please navigate to the URL manually", verificantionURIComplete)
		}
	}
}

func (c *Client) requestToken(deviceCode string) (tokenResponse, error) {
	var t tokenResponse
	params := map[string]string{
		"client_id":   clientID(),
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}
	timeFetched := time.Now()
	err := c.formPost("oauth/token", params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

func (c *Client) pollForTokens(interval time.Duration, expiresAt time.Time, deviceCode string) (tokenResponse, error) {
	var tR tokenResponse
	var err error
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for t := range ticker.C {
		if t.After(expiresAt) {
			return tR, errors.New("expired token")
		}
		tR, err = c.requestToken(deviceCode)
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

func (c *Client) refreshIdToken(refreshToken string) (tokenResponse, error) {
	var t tokenResponse
	params := map[string]string{
		"client_id":     clientID(),
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	timeFetched := time.Now()
	err := c.formPost("oauth/token", params, &t)
	t.ExpiresAt = addSeconds(timeFetched, t.ExpiresIn)
	return t, err
}

// IDToken returns a valid auth0 access token
// If a non-expired access token is found in the cache
// that token is returned. Otherwise the secret/refresh
// token from the application config is used to fetch
// a fresh one. If there is no secret configured this
// function errors.
func (c *Client) IDToken() (string, error) {
	cache, err := c.getCache()
	if err != nil {
		return "", nil
	}
	idToken := cache.GetString(idTokenKey)
	expiresAt := cache.GetTime(expiresAtKey)
	if cache.IsSet(idTokenKey) && cache.IsSet(expiresAtKey) && expiresAt.After(time.Now()) {
		return idToken, nil
	}
	if c.viper.IsSet(refreshTokenKey) {
		refreshToken := c.viper.GetString(refreshTokenKey)
		t, err := c.refreshIdToken(refreshToken)
		writeErr := c.saveToken(t)
		if writeErr != nil {
			fmt.Println("warning: credential cache failed")
		}
		return t.IdToken, err
	}
	return "", fmt.Errorf("not authenticated, try running `czid login` or adding your `secret` to %s manually", viper.GetViper().ConfigFileUsed())
}

// Login performs the auth0 device authorization flow:
// https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow
// This function prompts the user to navigate to a URL or
// directs the user there.
func (c *Client) Login(headless bool, persistent bool) error {
	d, err := c.requestDeviceCode(persistent)
	if err != nil {
		return err
	}
	promptDeviceActivation(d.VerificationURIComplete, headless)
	t, err := c.pollForTokens(time.Duration(d.Interval)*time.Second, d.ExpiresAt, d.DeviceCode)
	if err != nil {
		return err
	}
	err = c.saveToken(t)
	if err != nil {
		return err
	}
	if persistent {
		// refresh the access token to make sure the user is authenticated
		_, err = c.refreshIdToken(t.RefreshToken)
	}
	return err
}

// Secret returns the auth0 secret/refresh token and a boolean representing
// whether the secret is defined.
func (c *Client) Secret() (string, bool) {
	return c.viper.GetString(refreshTokenKey), c.viper.IsSet(refreshTokenKey)
}
