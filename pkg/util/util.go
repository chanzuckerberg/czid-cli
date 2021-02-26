package util

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

const AppName = "idseq-cli-v2"

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

func GetConfigDir() (string, error) {
	userCacheDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil
	}
	cacheDir := path.Join(userCacheDir, AppName)
	return cacheDir, MkdirIfNotExists(cacheDir)
}

func GetCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", nil
	}
	cacheDir := path.Join(userCacheDir, AppName)
	return cacheDir, MkdirIfNotExists(cacheDir)
}

var cacheCache = make(map[string]*viper.Viper)

func ViperCache(name string) (*viper.Viper, error) {
	if cache, hasCache := cacheCache[name]; hasCache {
		return cache, nil
	}
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigFile(path.Join(cacheDir, fmt.Sprintf("%s.yaml", name)))
	cacheCache[name] = v
	return v, v.SafeWriteConfig()
}
