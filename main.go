package main

import (
	"os"

	"wiley.com/do-k8s-cluster-health-check/cmd"
)

func main() {
	if cmd.Execute() != nil {
		os.Exit(1)
	}
}
