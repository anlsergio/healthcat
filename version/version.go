package version

import (
	"fmt"
	"os"
)

// ChcVersion for JSON encoding
type ChcVersion struct {
	Version   string
	GitCommit string
	BuildDate string
}

// ChcVer instance of JSON version
var ChcVer ChcVersion = ChcVersion{
	Version:   Version,
	GitCommit: GitCommit,
	BuildDate: BuildDate,
}

// BuildDate is the date when the binary was built
var BuildDate string

// GitCommit is the commit hash when the binary was built
var GitCommit string

// Version is the binary version
var Version string

// PrintVersion prints the version and exits
// TODO: Consider adding version cmd
func PrintVersion() {
	fmt.Printf("Version: %s - Commit: %s - Date: %s\n", Version,
		GitCommit,
		BuildDate)
	os.Exit(0)
}
