package main

import (
	"flag"
	"os"
)

var (
	versionFlag bool
	vFlag       bool
	verboseFlag bool
	yesFlag     bool
)

// initFlags initializes command-line flags
func initFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.BoolVar(&versionFlag, "version", false, "Print version information")
	flag.BoolVar(&vFlag, "v", false, "Print version information")
	flag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&yesFlag, "y", false, "Automatically confirm prompts")
}
