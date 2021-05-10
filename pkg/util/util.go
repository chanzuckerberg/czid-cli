package util

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/spf13/viper"
)

const AppName = "idseq-cli-v2"

// MkdirIfNotExists makes a directory if it doesn't exist
func MkdirIfNotExists(dirname string) error {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.Mkdir(dirname, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// GetConfigDir gets the config directory for this application
func GetConfigDir() (string, error) {
	userCacheDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil
	}
	cacheDir := path.Join(userCacheDir, AppName)
	return cacheDir, MkdirIfNotExists(cacheDir)
}

// GetCacheDir gets the cache directory for this application
func GetCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", nil
	}
	cacheDir := path.Join(userCacheDir, AppName)
	return cacheDir, MkdirIfNotExists(cacheDir)
}

// TrimLeadingSlash trims the leading slash from a string if one is present
func TrimLeadingSlash(s string) string {
	if len(s) == 0 || s[0] != '/' {
		return s
	}
	return s[1:]
}

var viperCache *viper.Viper
var viperCacheMut sync.Mutex

// ViperCache returns a viper instance backed by a file in the cache directory.
// If no such file or directory exists, one is created. It is safe to call
// this concurrently.
func ViperCache() (*viper.Viper, error) {
	viperCacheMut.Lock()
	defer viperCacheMut.Unlock()
	if viperCache != nil {
		return viperCache, nil
	}
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.AddConfigPath(cacheDir)
	v.SetConfigName("cache")
	v.SetConfigType("yaml")
	err = v.SafeWriteConfig()
	_, alreadyExists := err.(viper.ConfigFileAlreadyExistsError)
	if alreadyExists {
		return v, v.ReadInConfig()
	}
	return v, err
}

func Tune() {
	fmt.Println(runtime.NumCPU())
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("%v\n", memStats.TotalAlloc)
}

func StringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func StringMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
