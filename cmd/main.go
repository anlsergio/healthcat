package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"wiley.com/healthcat/checker"
	"wiley.com/healthcat/k8s"
	"wiley.com/healthcat/server"
)

const (
	defaultAddress    = "*"
	defaultExcludes   = "kube-system,default,kube-public,istio-system,monitoring"
	defaultThreshold  = 100
	defaultNSuccess   = 1
	defaultNFailure   = 2
	defaultInterval   = "1m"
	defaultPort       = 8080
	defaultLogPreset  = "dev"
	defaultConfigFile = "./config/config.yml"
)

type mainCmdArgs struct {
	host               string
	clusterID          string
	namespaces         []string
	excludedNamespaces []string
	interval           time.Duration
	nsuccess           int
	nfailure           int
	threshold          int
	port               int
	logPreset          string
	configFile         string
}

func newMainCmd(mainArgs *mainCmdArgs) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "healthcat",
		Long: `Healthcat - K8S Cluster Health Check

Provides HTTP status (200 OK|5xx Failed) of a k8s cluster based on the
percentage of healthy services (--status-threshold) in the monitored
namespaces (--namespaces).  

To be included in the health status of the cluster, a healthy service
must provide API /healthz that returns HTTP 200 OK and use
“healthcat.wiley.com/healthz: enable” annotation. Healthcat will scan regularly
included services (--time-between-hc).

A service will be in a failed state if it fails predefined number of
consecutive health-checks (--failed-hc-cnt), and in healthy state if
it passes predefined number of successful health-checks
(--successful-hc-cnt).  Excluded namespaces (--excluded-namespaces)
will not be monitored by healthcat.  Cluster ID (--cluster-id) is a unique
cluster identifier that will be included in all healthcat reports.`,
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if mainArgs.configFile != "" {
				abs, err := filepath.Abs(mainArgs.configFile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not get the valid path to the file: %v", err)
				}
				fileName := filepath.Base(abs)
				fileLocation := filepath.Dir(abs)

				return LoadConfig(cmd, fileLocation, fileName)
			}
			return nil
		},
		RunE: func(*cobra.Command, []string) error {
			return runServer(mainArgs)
		},
	}

	flags := rootCmd.Flags()
	flags.StringVarP(&mainArgs.clusterID, "cluster-id", "i", "", "cluster ID")
	flags.StringVarP(&mainArgs.host, "listen-address", "l", defaultAddress, "bind address")
	flags.IntVarP(&mainArgs.port, "port", "p", defaultPort, "bind port")
	flags.StringSliceVarP(&mainArgs.namespaces, "namespaces", "n", []string{}, "list of namespaces to watch")
	flags.StringSliceVarP(&mainArgs.excludedNamespaces, "excluded-namespaces", "N", strings.Split(defaultExcludes, ","),
		"list of namespaces to exclude")
	flags.DurationVarP(&mainArgs.interval, "time-between-hc", "t", duration(defaultInterval), "time between two consecutive health checks")
	flags.IntVarP(&mainArgs.nsuccess, "successful-hc-cnt", "s", defaultNSuccess, "number of successful consecutive health checks counts")
	flags.IntVarP(&mainArgs.nfailure, "failed-hc-cnt", "F", defaultNFailure, "number of failed consecutive health checks counts")
	flags.IntVarP(&mainArgs.threshold, "status-threshold", "P", defaultThreshold, "percentage of successful health checks to set cluster status OK")
	flags.StringVar(&mainArgs.logPreset, "log-preset", defaultLogPreset, "Log preset config (dev|prod)")
	flags.StringVarP(&mainArgs.configFile, "config", "f", defaultConfigFile, "/path/to/config.yml")

	rootCmd.MarkFlagRequired("cluster-id")

	return rootCmd
}

// Execute executes the root command
func Execute() error {
	cmd := newMainCmd(&mainCmdArgs{})
	return cmd.Execute()
}

func runServer(cmdArgs *mainCmdArgs) error {
	var host string
	if cmdArgs.host != "*" {
		host = cmdArgs.host
	}

	var log *zap.Logger
	var errLog error

	switch cmdArgs.logPreset {
	case "dev":
		log, errLog = zap.NewDevelopment()
	case "prod":
		log, errLog = zap.NewProduction()
	default:
		log, errLog = zap.NewDevelopment()
		log.Sugar().Infof("Log preset not provided. Using Development preset.")
	}

	if errLog != nil {
		panic(errLog)
	}

	defer log.Sync()

	checker := &checker.Checker{
		ClusterID:        cmdArgs.clusterID,
		Interval:         cmdArgs.interval,
		FailureThreshold: cmdArgs.nfailure,
		SuccessThreshold: cmdArgs.nsuccess,
		StateThreshold:   cmdArgs.threshold,
		Logger:           log,
	}
	if err := checker.Run(); err != nil {
		return err
	}

	eventSource := &k8s.EventSource{
		Logger:             log,
		Namespaces:         cmdArgs.namespaces,
		ExcludedNamespaces: cmdArgs.excludedNamespaces,
		Registry:           checker,
	}
	if err := eventSource.Start(); err != nil {
		return err
	}

	server := &server.Server{
		Address: fmt.Sprintf("%s:%d", host, cmdArgs.port),
		Checker: checker,
		Logger:  log,
	}
	server.Run()
	return nil
}

func duration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}
