package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func mockFormPost(path string, params map[string]string, r interface{}) error {
	if path == "oauth/device/code" {
		deviceCode := "device_code"
		if strings.Contains(params["scope"], "offline_access") {
			deviceCode = "device_code_refresh"
		}
		resp := fmt.Sprintf(`{
          "device_code":"%s",
          "user_code":"",
          "verification_uri":"",
          "expires_in":900,
          "interval":1,
          "verification_uri_complete":"verification_uri_complete"
        }`, deviceCode)
		return json.Unmarshal([]byte(resp), r)
	}
	if path == "oauth/token" {
		refreshToken := ""
		if params["device_code"] == "device_code_refresh" {
			refreshToken = `"refresh_token": "refresh_token",`
		}
		resp := fmt.Sprintf(`{
          "access_token":"access_token",
          %s
          "id_token":"id_token",
          "scope":"openid",
          "expires_in": 800
        }`, refreshToken)
		return json.Unmarshal([]byte(resp), r)
	}
	return errors.New("URL not recognized")
}

func TestLoginPersistent(t *testing.T) {
	viperF, v, err := tempViper()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(viperF.Name())
	cacheF, cache, err := tempViper()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cacheF.Name())
	c := client{
		formPost: mockFormPost,
		viper:    v,
		cache:    cache,
	}
	err = c.login(true, true)
	if err != nil {
		t.Fatal(err)
	}
	if v.GetString("secret") != "refresh_token" {
		t.Errorf("secret should be 'refresh_token' but it was '%s'", v.GetString("secret"))
	}
	if cache.GetString("TOKEN") != "id_token" {
		t.Errorf("expected cached token to be 'id_token' but it was '%s'", cache.GetString("TOKEN"))
	}
}

func tempViper() (*os.File, *viper.Viper, error) {
	tmpfile, err := ioutil.TempFile("", "*.yaml")
	v := viper.New()
	v.SetConfigFile(tmpfile.Name())
	return tmpfile, v, err
}

func TestLogin(t *testing.T) {
	viperF, v, err := tempViper()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(viperF.Name())
	cacheF, cache, err := tempViper()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(cacheF.Name())
	c := client{
		formPost: mockFormPost,
		viper:    v,
		cache:    cache,
	}
	err = c.login(true, false)
	if err != nil {
		t.Fatal(err)
	}
	if v.IsSet("secret") {
		t.Error("secret should not be defined if persistent is false")
	}
	if cache.GetString("TOKEN") != "id_token" {
		t.Errorf("expected cached token to be 'id_token' but it was '%s'", cache.GetString("TOKEN"))
	}
}

func TestSecret(t *testing.T) {
	f, v, err := tempViper()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	v.Set("secret", "shh")
	c := client{
		viper: v,
	}
	secret, hasSecret := c.secret()
	if !hasSecret {
		t.Fatalf("should have secret but secret() returned false")
	}
	if secret != "shh" {
		t.Fatalf("expected secret 'shh' but it was %s", secret)
	}
}
