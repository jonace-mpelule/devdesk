package main

import (
	"fmt"
	"os"

	"github.com/jonacempelule/devdesk/internal/devdesk"
)

func main() {
	if err := devdesk.RunCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
