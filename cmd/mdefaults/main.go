package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/filesystem"
	pullop "github.com/fumiya-kume/mdefaults/internal/operation/pull"
	pushop "github.com/fumiya-kume/mdefaults/internal/operation/push"
	"github.com/fumiya-kume/mdefaults/internal/printer"
)

var (
	version      string
	architecture string
)

type appDependencies struct {
	newFileSystem func() filesystem.FileSystem
	pull          func([]config.Config) ([]config.Config, error)
	push          func([]config.Config) error
	stdin         io.Reader
	stdout        io.Writer
	logger        *log.Logger
	goos          string
	printSuccess  func(string)
	printError    func(string)
	printWarning  func(string)
}

func defaultAppDependencies() appDependencies {
	return appDependencies{
		newFileSystem: func() filesystem.FileSystem { return filesystem.NewOSFileSystem() },
		pull:          pullop.Pull,
		push:          pushop.Push,
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		logger:        log.Default(),
		goos:          runtime.GOOS,
		printSuccess:  printer.PrintSuccess,
		printError:    printer.PrintError,
		printWarning:  func(message string) { color.Yellow("%s", message) },
	}
}

func setupLogging(path string, logger *log.Logger) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	if err := logFile.Chmod(0600); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	logger.SetOutput(logFile)
	logger.SetFlags(log.LstdFlags | log.Lshortfile)
	return logFile, nil
}

func run(args []string, deps appDependencies) int {
	options, err := parseArguments(args)
	if err != nil {
		deps.logger.Printf("Failed to parse command line arguments: %v", err)
		deps.printError("Invalid command line arguments")
		return 1
	}

	if deps.goos == "linux" || deps.goos == "windows" {
		if err := writeOutput(deps.stdout, "Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows.\n"); err != nil {
			return handleOutputError(err, deps)
		}
	}

	if options.help {
		if err := printUsage(deps.stdout); err != nil {
			return handleOutputError(err, deps)
		}
		return 0
	}
	if options.version {
		if err := printVersionInfo(deps.stdout); err != nil {
			return handleOutputError(err, deps)
		}
		return 0
	}
	if options.command == "" {
		if err := printUsage(deps.stdout); err != nil {
			return handleOutputError(err, deps)
		}
		return 0
	}
	if options.verbose {
		deps.logger.SetFlags(log.LstdFlags | log.Lshortfile)
		deps.logger.Println("Verbose mode enabled")
	}

	switch options.command {
	case "debug":
		deps.logger.Println("Debug command executed")
		return 0
	case "pull", "push", "config":
		// Continue below after loading the configuration.
	default:
		deps.logger.Printf("Unknown command: %s", options.command)
		deps.printError("Unknown command")
		return 1
	}

	fs := deps.newFileSystem()
	if err := filesystem.CreateConfigFileIfMissing(fs); err != nil {
		deps.logger.Printf("Failed to create config file: %v", err)
		deps.printError("Failed to create config file")
		return 1
	}
	configs, err := config.ReadConfigFile(fs)
	if err != nil {
		deps.logger.Printf("Failed to read config file: %v", err)
		deps.printError("Failed to read config file")
		return 1
	}

	if options.command == "config" {
		if err := printConfigs(deps.stdout, configs); err != nil {
			return handleOutputError(err, deps)
		}
		return 0
	}
	if options.command == "pull" {
		return handlePull(configs, fs, options.yes, deps)
	}
	if err := printConfigs(deps.stdout, configs); err != nil {
		return handleOutputError(err, deps)
	}
	return handlePush(configs, deps)
}

func handlePull(configs []config.Config, fs filesystem.FileSystem, skipConfirmation bool, deps appDependencies) int {
	if err := writeOutput(deps.stdout, "Current Configuration:\n"); err != nil {
		return handleOutputError(err, deps)
	}
	if err := printConfigs(deps.stdout, configs); err != nil {
		return handleOutputError(err, deps)
	}
	if err := writeOutput(deps.stdout, "macOS Configuration:\n"); err != nil {
		return handleOutputError(err, deps)
	}

	macOSConfigs, err := deps.pull(configs)
	if err != nil {
		deps.logger.Printf("Failed to pull configurations: %v", err)
		deps.printError("Failed to pull configurations")
		return 1
	}
	if err := printConfigs(deps.stdout, macOSConfigs); err != nil {
		return handleOutputError(err, deps)
	}

	if !skipConfirmation {
		deps.printWarning("Warning: mdefaults will override your configuration file (~/.mdefaults). Proceed with caution.")
		if err := writeOutput(deps.stdout, "Do you want to continue? (yes/no): "); err != nil {
			return handleOutputError(err, deps)
		}
		var response string
		if _, err := fmt.Fscan(deps.stdin, &response); err != nil {
			deps.logger.Printf("Failed to read confirmation: %v", err)
			if writeErr := writeOutput(deps.stdout, "Failed to read input, operation cancelled.\n"); writeErr != nil {
				return handleOutputError(writeErr, deps)
			}
			return 1
		}
		if response != "yes" {
			if err := writeOutput(deps.stdout, "Operation cancelled.\n"); err != nil {
				return handleOutputError(err, deps)
			}
			return 0
		}
	}

	if err := config.WriteConfigFile(fs, macOSConfigs); err != nil {
		deps.logger.Printf("Failed to write config file: %v", err)
		deps.printError("Failed to write config file")
		return 1
	}
	deps.printSuccess("Configurations pulled successfully")
	return 0
}

func handlePush(configs []config.Config, deps appDependencies) int {
	if err := deps.push(configs); err != nil {
		deps.logger.Printf("Failed to push configurations: %v", err)
		deps.printError("Failed to push configurations")
		return 1
	}
	deps.printSuccess("Configurations pushed successfully")
	return 0
}

func handleOutputError(err error, deps appDependencies) int {
	deps.logger.Printf("Failed to write output: %v", err)
	deps.printError("Failed to write output")
	return 1
}

func printUsage(output io.Writer) error {
	return writeOutput(output, "Usage: mdefaults [flags] [command]\n"+
		"Commands:\n"+
		"  pull    - Retrieve and update configuration values.\n"+
		"  push    - Write configuration values.\n"+
		"  config  - Print the configuration file.\n")
}

func printConfigs(output io.Writer, configs []config.Config) error {
	var content strings.Builder
	for _, cfg := range configs {
		content.WriteString("- ")
		content.WriteString(cfg.Domain)
		content.WriteByte(' ')
		content.WriteString(cfg.Key)
		if cfg.Value == nil {
			content.WriteByte('\n')
			continue
		}
		content.WriteByte(' ')
		content.WriteString(*cfg.Value)
		content.WriteByte(' ')
		content.WriteString(cfg.Type)
		content.WriteByte('\n')
	}
	return writeOutput(output, content.String())
}

func printVersionInfo(output io.Writer) error {
	return writeOutput(output, fmt.Sprintf("Version: %s\nArchitecture: %s\n", version, architecture))
}

func writeOutput(output io.Writer, value string) error {
	_, err := io.WriteString(output, value)
	return err
}

func main() {
	logFile, err := setupLogging("mdefaults.log", log.Default())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to open log file: %v\n", err)
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	exitCode := run(os.Args[1:], defaultAppDependencies())
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close log file: %v\n", err)
		}
	}
	os.Exit(exitCode)
}
