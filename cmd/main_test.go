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
			[]string{"-i", "--cluster-id"}, "zzz", true,
			"zzz",
			func() interface{} { return cmdArgs.clusterID }, "",
		},
		{
			[]string{"-N", "--excluded-namespaces"}, "e-n-1,e-n-2", false,
			[]string{"e-n-1", "e-n-2"},
			func() interface{} { return cmdArgs.excludedNamespaces },
			[]string{"kube-system", "default", "kube-public", "istio-system", "monitoring"},
		},
		{
			[]string{"-F", "--failed-hc-cnt"}, "7", false,
			7,
			func() interface{} { return cmdArgs.nfailure }, 2,
		},
		{
			[]string{"-t", "--time-between-hc"}, "2m30s", false,
			duration("2m30s"),
			func() interface{} { return cmdArgs.interval }, duration("1m"),
		},
		{
			[]string{"-l", "--listen-address"}, "hostname", false,
			"hostname",
			func() interface{} { return cmdArgs.host }, "*",
		},
		{
			[]string{"-p", "--port"}, "8585", false,
			8585,
			func() interface{} { return cmdArgs.port }, 8080,
		},
		{
			[]string{"-n", "--namespaces"}, "n-1,n-2,n-3", false,
			[]string{"n-1", "n-2", "n-3"},
			func() interface{} { return cmdArgs.namespaces }, []string{},
		},
		{
			[]string{"-s", "--successful-hc-cnt"}, "25", false,
			25,
			func() interface{} { return cmdArgs.nsuccess }, 1,
		},
		{
			[]string{"-P", "--status-threshold"}, "77", false,
			77,
			func() interface{} { return cmdArgs.threshold }, 100,
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
