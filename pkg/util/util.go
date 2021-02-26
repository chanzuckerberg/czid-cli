package util

import (
	"fmt"
	"os"
	"path"
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
		return v, nil
	}
	return v, err
}
