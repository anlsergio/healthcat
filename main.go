package main

import (
	"os"

	"wiley.com/do-k8s-cluster-health-check/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
