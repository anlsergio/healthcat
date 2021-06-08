package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"
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

type testCase struct {
	name         string
	configSource string
	value        string
	want         interface{}
	got          func() interface{}
}

func TestLoadConfigPrecedenceOrder(t *testing.T) {
	cmdArgs := &mainCmdArgs{}

	configFileTestCases := []testCase{
		{
			name:         "listen-address",
			configSource: "file",
			value:        "localhost",
			want:         "localhost",
			got: func() interface{} {
				return cmdArgs.host
			},
		},
		{
			name:         "cluster-id",
			configSource: "file",
			value:        "random_domain.com",
			want:         "random_domain.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name:         "excluded-namespaces",
			configSource: "file",
			value:        "healthcat,monitoring",
			want:         []string{"healthcat", "monitoring"},
			got: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
		},
		{
			name:         "namespaces",
			configSource: "file",
			value:        "nsfromconfig,anotherone,yetanotheronefromconfig",
			want:         []string{"nsfromconfig", "anotherone", "yetanotheronefromconfig"},
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name:         "time-between-hc",
			configSource: "file",
			value:        "5m0s",
			want:         duration("5m0s"),
			got: func() interface{} {
				return cmdArgs.interval
			},
		},
		{
			name:         "successful-hc-cnt",
			configSource: "file",
			value:        "4",
			want:         4,
			got: func() interface{} {
				return cmdArgs.nsuccess
			},
		},
		{
			name:         "failed-hc-cnt",
			configSource: "file",
			value:        "6",
			want:         6,
			got: func() interface{} {
				return cmdArgs.nfailure
			},
		},
		{
			name:         "status-threshold",
			configSource: "file",
			value:        "200",
			want:         200,
			got: func() interface{} {
				return cmdArgs.threshold
			},
		},
		{
			name:         "port",
			configSource: "file",
			value:        "8980",
			want:         8980,
			got: func() interface{} {
				return cmdArgs.port
			},
		},
		{
			name:         "logPreset",
			configSource: "file",
			value:        "prod",
			want:         "prod",
			got: func() interface{} {
				return cmdArgs.logPreset
			},
		},
	}

	envVariableTestCases := []testCase{
		{
			name:         "HEALTHCAT_LISTEN_ADDRESS",
			configSource: "env",
			value:        "localhost_from_env.com",
			want:         "localhost_from_env.com",
			got: func() interface{} {
				return cmdArgs.host
			},
		},
		{
			name:         "HEALTHCAT_CLUSTER_ID",
			configSource: "env",
			value:        "a_not_so_random_domain.com",
			want:         "a_not_so_random_domain.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name:         "HEALTHCAT_NAMESPACES",
			configSource: "env",
			value:        "myapp,anotherfancyapp",
			want:         []string{"myapp", "anotherfancyapp"},
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name:         "HEALTHCAT_EXCLUDED_NAMESPACES",
			configSource: "env",
			value:        "healthcat,monitoring,somethingelse",
			want:         []string{"healthcat", "monitoring", "somethingelse"},
			got: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
		},
		{
			name:         "HEALTHCAT_TIME_BETWEEN_HC",
			configSource: "env",
			value:        "10m0s",
			want:         duration("10m0s"),
			got: func() interface{} {
				return cmdArgs.interval
			},
		},
		{
			name:         "HEALTHCAT_SUCCESSFUL_HC_CNT",
			configSource: "env",
			value:        "8",
			want:         8,
			got: func() interface{} {
				return cmdArgs.nsuccess
			},
		},
		{
			name:         "HEALTHCAT_FAILED_HC_CNT",
			configSource: "env",
			value:        "5",
			want:         5,
			got: func() interface{} {
				return cmdArgs.nfailure
			},
		},
		{
			name:         "HEALTHCAT_STATUS_THRESHOLD",
			configSource: "env",
			value:        "75",
			want:         75,
			got: func() interface{} {
				return cmdArgs.threshold
			},
		},
		{
			name:         "HEALTHCAT_PORT",
			configSource: "env",
			value:        "8585",
			want:         8585,
			got: func() interface{} {
				return cmdArgs.port
			},
		},
		{
			name:         "HEALTHCAT_LOG_PRESET",
			configSource: "env",
			value:        "zzzz",
			want:         "zzzz",
			got: func() interface{} {
				return cmdArgs.logPreset
			},
		},
	}

	flagTestCases := []testCase{
		{
			name:         "--cluster-id",
			configSource: "flag",
			value:        "wiley.com",
			want:         "wiley.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name:         "--namespaces",
			configSource: "flag",
			value:        "anotsofancyapp",
			want:         []string{"anotsofancyapp"},
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name:         "HEALTHCAT_LISTEN_ADDRESS",
			configSource: "env",
			value:        "localhost_from_env.com",
			want:         "localhost_from_env.com",
			got: func() interface{} {
				return cmdArgs.host
			},
		},
		{
			name:         "HEALTHCAT_CLUSTER_ID",
			configSource: "env",
			value:        "a_not_so_random_domain.com",
			want:         "wiley.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name:         "HEALTHCAT_NAMESPACES",
			configSource: "env",
			value:        "myapp,anotherfancyapp",
			want:         []string{"anotsofancyapp"},
			got: func() interface{} {
				return cmdArgs.namespaces
			},
		},
		{
			name:         "HEALTHCAT_EXCLUDED_NAMESPACES",
			configSource: "env",
			value:        "healthcat,monitoring,somethingelse",
			want:         []string{"healthcat", "monitoring", "somethingelse"},
			got: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
		},
		{
			name:         "HEALTHCAT_TIME_BETWEEN_HC",
			configSource: "env",
			value:        "10m0s",
			want:         duration("10m0s"),
			got: func() interface{} {
				return cmdArgs.interval
			},
		},
		{
			name:         "HEALTHCAT_SUCCESSFUL_HC_CNT",
			configSource: "env",
			value:        "8",
			want:         8,
			got: func() interface{} {
				return cmdArgs.nsuccess
			},
		},
		{
			name:         "HEALTHCAT_FAILED_HC_CNT",
			configSource: "env",
			value:        "5",
			want:         5,
			got: func() interface{} {
				return cmdArgs.nfailure
			},
		},
		{
			name:         "HEALTHCAT_STATUS_THRESHOLD",
			configSource: "env",
			value:        "75",
			want:         75,
			got: func() interface{} {
				return cmdArgs.threshold
			},
		},
		{
			name:         "port",
			configSource: "file",
			value:        "8980",
			want:         8980,
			got: func() interface{} {
				return cmdArgs.port
			},
		},
		{
			name:         "log-preset",
			configSource: "file",
			value:        "prod",
			want:         "prod",
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

	_, fileErr := f.Write(yamlConfigFile)
	if fileErr != nil {
		t.Errorf("couldn't write data into the test file: %v", fileErr)
	}

	// Assert that parameters provided by a config file take effect even if a default flag value is set
	t.Run("Config file parameters parsing", func(t *testing.T) {
		flags := []string{
			"--config", "./config.yml",
		}

		runTests(cmdArgs, flags, configFileTestCases, t)
	})

	// Assert that parameters provided as env. variables take effect even if a config file is being used as config source
	t.Run("Env. variables parameters parsing", func(t *testing.T) {
		flags := []string{
			"--config", "./config.yml",
		}

		for _, e := range envVariableTestCases {
			if e.configSource == "env" {
				os.Setenv(e.name, fmt.Sprintf("%v", e.value))
				defer os.Unsetenv(e.name)
			}
		}

		runTests(cmdArgs, flags, envVariableTestCases, t)
	})

	// Assert the precedence order:
	// - Flags take precedence over all
	// - Env. variables take precedence over the config file source
	// - The config file source takes precedence over default values (--config)
	t.Run("Precedence Order", func(t *testing.T) {
		flags := []string{
			"--config", "./config.yml",
		}
		for _, p := range flagTestCases {
			if p.configSource == "env" {
				os.Setenv(p.name, fmt.Sprintf("%v", p.value))
				defer os.Unsetenv(p.name)
			} else if p.configSource == "flag" {
				flags = append(flags, p.name, p.value)
			}
		}

		runTests(cmdArgs, flags, flagTestCases, t)
	})
}

func runTests(cmdArgs *mainCmdArgs, flags []string, testCases []testCase, t *testing.T) {
	cmd := newMainCmd(cmdArgs)
	resetCommand(cmd, flags)
	if err := cmd.Execute(); err != nil {
		t.Errorf("got error: %v", err)
	}

	for _, p := range testCases {
		if got, want := p.got(), p.want; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v, name: %v", got, want, p.name)
		}
	}
}
