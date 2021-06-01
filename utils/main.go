package utils

import (
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads system parameters from a config file
func LoadConfig(filePath string, fileName string) error {
	viper.SetConfigName(strings.Split(fileName, ".")[0])
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filePath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
