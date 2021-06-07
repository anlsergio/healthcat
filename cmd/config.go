package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

	v.SetConfigName(strings.Split(fileName, ".")[0])
	v.SetConfigType(configFileType)
	v.AddConfigPath(filePath)

	v.AutomaticEnv()
	v.SetEnvPrefix(configEnvPrefix)

	if err := v.ReadInConfig(); err != nil {
		// if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// 	return err
		// }
		return err
	} else {
		fmt.Fprintf(os.Stdout, "configuration file found. Loading system parameters from file %s\n", fileName)
	}

	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		fmt.Printf("Flag: '%v', corresponding value from source: '%v'\n", f.Name, v.Get(f.Name))
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", configEnvPrefix, envVarSuffix))
			fmt.Printf("Changing variables: %s_%s\n", configEnvPrefix, envVarSuffix)
		}

		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			fmt.Println("Setting unset parameters from configuration source: ", val)
		}
	})
	// v.WriteConfigAs("./config_generated_from_viper.yml")
}

// RootDir returns the root directory of the project
func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}
