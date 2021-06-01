package utils

import (
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads system parameters from a config file and from enviroment variables if they are defined
func LoadConfig(filePath string, fileName string) error {
	viper.SetConfigName(strings.Split(fileName, ".")[0])
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filePath)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("healthcat")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// RootDir returns the root directory of the project
func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}
