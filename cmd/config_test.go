package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
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
	name string
	want interface{}
	got  func() interface{}
}

func TestLoadConfigPrecedenceOrder(t *testing.T) {
	var cmdArgs *mainCmdArgs

	configFileTestCases := []testCase {
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

	envVariableTestCases := []testCase {
		{
			name: "HEALTHCAT_LISTEN_ADDRESS",
			want: "localhost_from_env.com",
			got: func() interface{} {
				return cmdArgs.host
			},
		},
		{
			name: "HEALTHCAT_CLUSTER_ID",
			want: "a_not_so_random_domain.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
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
			name: "HEALTHCAT_TIME_BETWEEN_HC",
			want: "10m0s",
			got: func() interface{} {
				return time.Duration.String(cmdArgs.interval)
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
			name: "HEALTHCAT_FAILED_HC_CNT",
			want: 5,
			got: func() interface{} {
				return cmdArgs.nfailure
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
			name: "HEALTHCAT_LOG_PRESET",
			want: "zzzz",
			got: func() interface{} {
				return cmdArgs.logPreset
			},
		},
	}

	flagTestCases := []testCase {
		{
			name: "cluster-id",
			want: "wiley.com",
			got: func() interface{} {
				return cmdArgs.clusterID
			},
		},
		{
			name: "namespaces",
			want: "anotsofancyapp",
			got: func() interface{} {
				return cmdArgs.namespaces
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

	// Assert that parameters provided by a config file takes effect even if a default flag value is set
	t.Run("Config file takes precedence", func(t *testing.T) {
		args := []string{
			"--config", "./config.yml",
		}

		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommand(cmd, args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		runTests(configFileTestCases, t)
	})

	// Assert that parameters provided as env. variables take precedence over the config file ones
	t.Run("Env. variable takes precedence", func(t *testing.T) {
		args := []string{
			"--config", "./config.yml",
		}

		for _, e := range envVariableTestCases {
			os.Setenv(e.name, fmt.Sprintf("%v", e.want))
			defer os.Unsetenv(e.name)
		}

		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommand(cmd, args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		runTests(envVariableTestCases, t)
	})

	// Assert that the parameters provided as flags take precedence over all
	t.Run("Flags takes precedence", func(t *testing.T) {
		args := []string{
			"--cluster-id", "wiley.com",
			"--namespaces", "anotsofancyapp",
			"--config", "./config.yml",
		}

		for _, e := range envVariableTestCases {
			os.Setenv(e.name, fmt.Sprintf("%v", e.want))
			defer os.Unsetenv(e.name)
		}

		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommand(cmd, args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		runTests(flagTestCases, t)
	})
}

func runTests(testCases []testCase, t *testing.T) {
	for _, p := range testCases {
		if strings.Contains(fmt.Sprintf("%v", p.want), ",") {
			p.want = strings.Split(fmt.Sprintf("%v", p.want), ",")
		} else if reflect.TypeOf(p.got()).Kind() == reflect.Slice && reflect.TypeOf(p.want).Kind() == reflect.String {
			p.want = []string{fmt.Sprintf("%v", p.want)}
		}
		if got, want := p.got(), p.want; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v, name: %v", got, want, p.name)
		}
	}
}
