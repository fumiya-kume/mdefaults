package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

type commandOptions struct {
	command string
	version bool
	verbose bool
	yes     bool
	help    bool
}

func parseArguments(args []string) (commandOptions, error) {
	var options commandOptions
	var longVersion bool
	var shortVersion bool
	var flagArgs []string

	for _, arg := range args {
		switch {
		case arg == "--":
			return commandOptions{}, fmt.Errorf("flag terminator is not supported")
		case arg == "-h" || arg == "--help":
			options.help = true
		case strings.HasPrefix(arg, "-"):
			flagArgs = append(flagArgs, arg)
		case options.command == "":
			options.command = arg
		default:
			return commandOptions{}, fmt.Errorf("unexpected argument %q", arg)
		}
	}

	flags := flag.NewFlagSet("mdefaults", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&longVersion, "version", false, "Print version information")
	flags.BoolVar(&shortVersion, "v", false, "Print version information")
	flags.BoolVar(&options.verbose, "verbose", false, "Enable verbose logging")
	flags.BoolVar(&options.yes, "y", false, "Automatically confirm prompts")
	if err := flags.Parse(flagArgs); err != nil {
		return commandOptions{}, err
	}
	if flags.NArg() != 0 {
		return commandOptions{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	options.version = longVersion || shortVersion
	return options, nil
}
