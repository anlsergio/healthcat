package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

var yamlConfigFile = []byte(`
listen-address: localhost
cluster-id: random_domain.com
namespaces: nsfromconfig,anotherone,yetanotheronefromconfig
excluded-namespaces: healthcat,monitoring
time-between-hc: 5m
successful-hc-cnt: 4
failed-hc-cnt: 6
status-threshold: 200
port: 8980
log-preset: prod
`)

func TestLoadConfigPrecedenceOrder(t *testing.T) {
	var cmdArgs *mainCmdArgs

	configFileTestCases := []struct {
		name string
		want interface{}
		got  func() interface{}
	}{
		{
			name: "listen-address",
			want: "localhost",
			got: func() interface{} {
				return cmdArgs.host
			},
		},
		{
			name: "cluster-id",
			want: "random_domain.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name: "excluded-namespaces",
			want: "healthcat,monitoring",
			got: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
		},
		{
			name: "namespaces",
			want: "nsfromconfig,anotherone,yetanotheronefromconfig",
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name: "time-between-hc",
			want: "5m0s",
			got: func() interface{} {
				return time.Duration.String(cmdArgs.interval)
			},
		},
		{
			name: "successful-hc-cnt",
			want: 4,
			got: func() interface{} {
				return cmdArgs.nsuccess
			},
		},
		{
			name: "failed-hc-cnt",
			want: 6,
			got: func() interface{} {
				return cmdArgs.nfailure
			},
		},
		{
			name: "status-threshold",
			want: 200,
			got: func() interface{} {
				return cmdArgs.threshold
			},
		},
		{
			name: "port",
			want: 8980,
			got: func() interface{} {
				return cmdArgs.port
			},
		},
		{
			name: "logPreset",
			want: "prod",
			got: func() interface{} {
				return cmdArgs.logPreset
			},
		},
	}

	envVariableTestCases := []struct {
		name string
		want interface{}
		got func () interface{}
	}{
		{
			name: "HEALTHCAT_NAMESPACES",
			want: "myapp,anotherfancyapp",
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name: "HEALTHCAT_EXCLUDED_NAMESPACES",
			want: "healthcat,monitoring,somethingelse",
			got: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
		},
		{
			name: "HEALTHCAT_FAILED_HC_CNT",
			want: 5,
			got: func() interface{} {
				return cmdArgs.nfailure
			},
		},
		{
			name: "HEALTHCAT_SUCCESSFUL_HC_CNT",
			want: 8,
			got: func() interface{} {
				return cmdArgs.nsuccess
			},
		},
		{
			name: "HEALTHCAT_STATUS_THRESHOLD",
			want: 75,
			got: func() interface{} {
				return cmdArgs.threshold
			},
		},
		{
			name: "HEALTHCAT_PORT",
			want: 8585,
			got: func() interface{} {
				return cmdArgs.port
			},
		},
		{
			name: "HEALTHCAT_TIME_BETWEEN_HC",
			want: "10m0s",
			got: func() interface{} {
				return time.Duration.String(cmdArgs.interval)
			},
		},
		{
			name: "HEALTHCAT_LOG_PRESET",
			want: "prod",
			got: func() interface{} {
				return cmdArgs.logPreset
			},
		},
	}

	tmpDir := t.TempDir()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("couldn't get the current working directory: %v", err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Errorf("couldn't change to the temporary test directory: %v", err)
	}

	f, err := os.Create("./config.yml")
	if err != nil {
		t.Errorf("couldn't create the test config file: %v", err)
	}
	defer f.Close()

	_, err2 := f.Write(yamlConfigFile)
	if err2 != nil {
		t.Errorf("couldn't write data into the test file: %v", err)
	}

	t.Run("ConfigFile", func(t *testing.T) {
		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommandWithoutArgs(cmd)
		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		for _, p := range configFileTestCases {
			if strings.Contains(fmt.Sprintf("%v", p.want), ",") {
				p.want = strings.Split(fmt.Sprintf("%v", p.want), ",")
			}
			if got, want := p.got(), p.want; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v, name: %v", got, want, p.name)
			}
		}
	})

	t.Run("Env. Variable", func(t *testing.T) {
		for _, p := range envVariableTestCases {
			os.Setenv(p.name, fmt.Sprintf("%v", p.want))
			defer os.Unsetenv(p.name)
		}

		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommandWithoutArgs(cmd)
		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		for _, p := range envVariableTestCases {
			if strings.Contains(fmt.Sprintf("%v", p.want), ",") {
				p.want = strings.Split(fmt.Sprintf("%v", p.want), ",")
			}
			if got, want := p.got(), p.want; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v, name: %v", got, want, p.name)
			}
		}
	})
}

func resetCommandWithoutArgs(cmd *cobra.Command) {
	cmd.RunE = func(*cobra.Command, []string) error { return nil }
	cmd.SetErr(ioutil.Discard)
	cmd.SetOutput(ioutil.Discard)
	cmd.SetUsageFunc(func(*cobra.Command) error { return nil })
	cmd.SetArgs([]string{
		// "--cluster-id", "wiley.com",
		"--config", "./config.yml",
	})
}
