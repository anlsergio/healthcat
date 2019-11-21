package cmd

import (
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestFlags(t *testing.T) {
	var config *Config
	flags := []struct {
		names      []string
		arg        string
		required   bool
		value      func() interface{}
		defaultVal interface{}
	}{
		{
			[]string{"-i", "--cluster-id"}, "zzz", true,
			func() interface{} { return config.clusterID }, "",
		},
		{
			[]string{"-N", "--excluded-namespaces"}, "e-n-1,e-n-2", false,
			func() interface{} { return config.excludedNamespaces },
			[]string{"kube-system", "default", "kube-public", "istio-system", "monitoring"},
		},
		{
			[]string{"-f", "--failed-hc-cnt"}, "7", false,
			func() interface{} { return config.nfailure }, 2,
		},
		{
			[]string{"-t", "--time-between-hc"}, "2m30s", false,
			func() interface{} { return config.interval }, mustParseDuration("1m"),
		},
		{
			[]string{"-l", "--listen-address"}, "hostname", false,
			func() interface{} { return config.listenAddress }, "*",
		},
		{
			[]string{"-p", "--port"}, "8585", false,
			func() interface{} { return config.port }, 8080,
		},
		{
			[]string{"-n", "--namespaces"}, "n-1,n-2,n-3", false,
			func() interface{} { return config.namespaces }, []string{},
		},
		{
			[]string{"-s", "--successful-hc-cnt"}, "25", false,
			func() interface{} { return config.nsuccess }, 1,
		},
		{
			[]string{"-P", "--status-threshold"}, "77", false,
			func() interface{} { return config.threshold }, 100,
		},
	}

	var required []string
	for _, flag := range flags {
		if flag.required {
			required = append(required, flag.names[0], flag.arg)
		}
	}

	t.Run("SetVal", func(t *testing.T) {
		for _, flag := range flags {
			for _, name := range flag.names {
				var args []string
				if !flag.required {
					args = append(args, required...)
				}
				args = append(args, name, flag.arg)

				config = &Config{}
				cmd := newMainCommand(config)
				resetForTesting(cmd, args)

				if err := cmd.Execute(); err != nil {
					t.Errorf("got error: %v", err)
				}

				value := flag.value()
				if want := toValueType(flag.arg, value, t); !reflect.DeepEqual(want, value) {
					t.Errorf("want %v, got %v, args: %v", want, value, args)
				}
			}
		}
	})

	t.Run("Defaults", func(t *testing.T) {
		var args []string
		args = append(args, required...)

		config = &Config{}
		cmd := newMainCommand(config)
		resetForTesting(cmd, args)

		if err := cmd.Execute(); err != nil {
			t.Errorf("got error: %v", err)
		}

		for _, f := range flags {
			if f.required {
				continue
			}
			if got, want := f.value(), f.defaultVal; !reflect.DeepEqual(got, want) {
				t.Errorf("%q: want %v, got %v", f.names[0], want, got)
			}
		}
	})
}

// resetForTesting resets rootCmd for testing purposes
func resetForTesting(cmd *cobra.Command, args []string) {
	cmd.Run = func(*cobra.Command, []string) {}
	cmd.SetErr(ioutil.Discard)
	cmd.SetOutput(ioutil.Discard)
	cmd.SetUsageFunc(func(*cobra.Command) error { return nil })
	cmd.SetArgs(args)
}

func toValueType(val string, target interface{}, t *testing.T) interface{} {
	var targetVal interface{}
	var err error

	switch target.(type) {
	case int:
		targetVal, err = strconv.Atoi(val)
	case string:
		targetVal = val
	case []string:
		targetVal = strings.Split(val, ",")
	case time.Duration:
		targetVal, err = time.ParseDuration(val)
	default:
		t.Fatalf("unexpected type: %T", target)
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return targetVal
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}
