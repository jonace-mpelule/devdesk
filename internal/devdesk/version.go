package devdesk

import (
	"fmt"
	"io"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func PrintVersion(stdout io.Writer) {
	fmt.Fprintf(stdout, "devdesk %s\ncommit: %s\nbuilt: %s\n", Version, Commit, Date)
}
