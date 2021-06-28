package cmd

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestFlags(t *testing.T) {
	var cmdArgs *mainCmdArgs
	flags := []struct {
		names      []string
		arg        string
		required   bool
		want       interface{}
		value      func() interface{}
		defaultVal interface{}
	}{
		{
			names:    []string{"-i", "--cluster-id"},
			arg:      "zzz",
			required: true,
			want:     "zzz",
			value: func() interface{} {
				return cmdArgs.clusterID
			},
			defaultVal: "",
		},
		{
			names:    []string{"-N", "--excluded-namespaces"},
			arg:      "e-n-1,e-n-2",
			required: false,
			want:     []string{"e-n-1", "e-n-2"},
			value: func() interface{} {
				return cmdArgs.excludedNamespaces
			},
			defaultVal: []string{"kube-system", "default", "kube-public", "istio-system", "monitoring"},
		},
		{
			names:    []string{"-F", "--failed-hc-cnt"},
			arg:      "7",
			required: false,
			want:     7,
			value: func() interface{} {
				return cmdArgs.nfailure
			},
			defaultVal: 2,
		},
		{
			names:    []string{"-t", "--time-between-hc"},
			arg:      "2m30s",
			required: false,
			want:     duration("2m30s"),
			value: func() interface{} {
				return cmdArgs.interval
			},
			defaultVal: duration("1m"),
		},
		{
			names:    []string{"-l", "--listen-address"},
			arg:      "hostname",
			required: false,
			want:     "hostname",
			value: func() interface{} {
				return cmdArgs.host
			},
			defaultVal: "*",
		},
		{
			names:    []string{"-p", "--port"},
			arg:      "8585",
			required: false,
			want:     8585,
			value: func() interface{} {
				return cmdArgs.port
			},
			defaultVal: 8080,
		},
		{
			names:    []string{"-n", "--namespaces"},
			arg:      "n-1,n-2,n-3",
			required: false,
			want:     []string{"n-1", "n-2", "n-3"},
			value: func() interface{} {
				return cmdArgs.namespaces
			},
			defaultVal: []string{},
		},
		{
			names:    []string{"-s", "--successful-hc-cnt"},
			arg:      "25",
			required: false,
			want:     25,
			value: func() interface{} {
				return cmdArgs.nsuccess
			},
			defaultVal: 1,
		},
		{
			names:    []string{"-P", "--status-threshold"},
			arg:      "77",
			required: false,
			want:     77,
			value: func() interface{} {
				return cmdArgs.threshold
			},
			defaultVal: 100,
		},
		{
			names:    []string{"-f", "--config"},
			arg:      "../config/config.yml",
			required: false,
			want:     "../config/config.yml",
			value: func() interface{} {
				return cmdArgs.configFile
			},
			defaultVal: "./config/config.yml",
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

				cmdArgs = &mainCmdArgs{}
				cmd := newMainCmd(cmdArgs)
				resetCommand(cmd, args)

				if err := cmd.Execute(); err != nil {
					t.Errorf("got error: %v", err)
				}

				if got, want := flag.value(), flag.want; !reflect.DeepEqual(got, want) {
					t.Errorf("got %v, want %v, args: %v", got, want, args)
				}
			}
		}
	})

	t.Run("Defaults", func(t *testing.T) {
		var args []string
		args = append(args, required...)

		cmdArgs = &mainCmdArgs{}
		cmd := newMainCmd(cmdArgs)
		resetCommand(cmd, args)

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

// resetCommand resets rootCmd for testing purposes
func resetCommand(cmd *cobra.Command, args []string) {
	cmd.RunE = func(*cobra.Command, []string) error { return nil }
	cmd.SetErr(ioutil.Discard)
	cmd.SetOutput(ioutil.Discard)
	cmd.SetUsageFunc(func(*cobra.Command) error { return nil })
	cmd.SetArgs(args)
}
