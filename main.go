package main

import (
	"os"

	"wiley.com/healthcat/cmd"
)

func main() {
	if cmd.Execute() != nil {
		os.Exit(1)
	}
}
