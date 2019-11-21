package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"wiley.com/do-k8s-cluster-health-check/server"
)

const (
	defaultAddress   = "*"
	defaultExcludes  = "kube-system,default,kube-public,istio-system,monitoring"
	defaultThreshold = 100
	defaultNSuccess  = 1
	defaultNFailure  = 2
	defaultInterval  = "1m"
	defaultPort      = 8080
)

// Config contains configuration properties
type Config struct {
	listenAddress      string
	clusterID          string
	namespaces         []string
	excludedNamespaces []string
	interval           time.Duration
	nsuccess           int
	nfailure           int
	threshold          int
	port               int
}

func newMainCommand(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "chc",
		Long: `CHC - Cluster Health Check

Provides HTTP status (200 OK|5xx Failed) of a k8s cluster based on the
percentage of healthy services (--status-threshold) in the monitored
namespaces (--namespaces).  

To be included in the health status of the cluster, a healthy service
must provide API /healthz that returns HTTP 200 OK and use
“chc.wiley.com/healthz: enable” annotation.  CHC will scan regularly
included services (--time-between-hc).

A service will be in a failed state if it fails predefined number of
consecutive health-checks (--failed-hc-cnt), and in healthy state if
it passes predefined number of successful health-checks
(--successful-hc-cnt).  Excluded namespaces (--excluded-namespaces)
will not be monitored by CHC.  Cluster ID (--cluster-id) is a unique
cluster identifier that will be included in all CHC reports.`,
		Args: cobra.NoArgs,
		Run: func(*cobra.Command, []string) {
			runServer(config)
		},
	}

	addFlags(cmd, config)

	return cmd
}

// Execute executes the root command
func Execute() error {
	cmd := newMainCommand(&Config{})
	return cmd.Execute()
}

func runServer(config *Config) {
	var host string
	if config.listenAddress != "*" {
		host = config.listenAddress
	}

	server := &server.Server{
		Address: fmt.Sprintf("%s:%d", host, config.port),
	}
	server.Run()
}

func addFlags(cmd *cobra.Command, config *Config) {
	flags := cmd.Flags()

	flags.StringVarP(&config.clusterID, "cluster-id", "i", "", "cluster ID")
	flags.StringVarP(&config.listenAddress, "listen-address", "l", defaultAddress, "bind address")
	flags.IntVarP(&config.port, "port", "p", defaultPort, "bind port")
	flags.StringSliceVarP(&config.namespaces, "namespaces", "n", []string{}, "list of namespaces to watch")
	flags.StringSliceVarP(&config.excludedNamespaces, "excluded-namespaces", "N", strings.Split(defaultExcludes, ","),
		"list of namespaces to exclude")

	interval, err := time.ParseDuration(defaultInterval)
	if err != nil {
		panic(err)
	}
	flags.DurationVarP(&config.interval, "time-between-hc", "t", interval, "time between two consecutive health checks")

	flags.IntVarP(&config.nsuccess, "successful-hc-cnt", "s", defaultNSuccess, "number of successful consecutive health checks counts")
	flags.IntVarP(&config.nfailure, "failed-hc-cnt", "f", defaultNFailure, "number of failed consecutive health checks counts")
	flags.IntVarP(&config.threshold, "status-threshold", "P", defaultThreshold, "percentage of successful health checks to set cluster status OK")

	cmd.MarkFlagRequired("cluster-id")
}
