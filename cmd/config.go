package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configFileType  = "yaml"
	configEnvPrefix = "HEALTHCAT"
)

// LoadConfig loads system parameters from a config file and from enviroment variables if they are defined
func LoadConfig(cmd *cobra.Command, filePath string, fileName string) error {
	v := viper.New()

	configName := fileName
	if i := strings.LastIndex(fileName, "."); i >= 0 {
		configName = fileName[:i]
	}

	v.SetConfigName(configName)
	v.SetConfigType(configFileType)
	v.AddConfigPath(filePath)

	v.AutomaticEnv()
	v.SetEnvPrefix(configEnvPrefix)

	replacer := strings.NewReplacer("-", "_")
	v.SetEnvKeyReplacer(replacer)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		fmt.Fprintf(os.Stderr, "[WARNING] couldn't load system parameters from the provided file %s. Falling back to the default parameters respecting the order of precedence instead.\n", fileName)
	} else {
		fmt.Fprintf(os.Stdout, "[INFO] configuration file found. Loading system parameters from file %s\n", fileName)
	}

	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
