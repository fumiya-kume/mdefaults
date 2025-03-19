package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/filesystem"
	pullop "github.com/fumiya-kume/mdefaults/internal/operation/pull"
	pushop "github.com/fumiya-kume/mdefaults/internal/operation/push"
)

var (
	version      string
	architecture string
)

func initFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.BoolVar(&versionFlag, "version", false, "Print version information")
	flag.BoolVar(&vFlag, "v", false, "Print version information")
	flag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&yesFlag, "y", false, "Automatically confirm prompts")
}

var (
	versionFlag bool
	vFlag       bool
	verboseFlag bool
	yesFlag     bool
)

func setupLogging() {
	logFile, err := os.OpenFile("mdefaults.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 0
	}

	command := os.Args[1]

	// Parse flags after the command
	flag.CommandLine.Parse(os.Args[2:])

	fs := filesystem.NewOSFileSystem()
	if err := filesystem.CreateConfigFileIfMissing(fs); err != nil {
		log.Printf("Failed to create config file: %v", err)
	}
	configs, err := config.ReadConfigFile(fs)
	if err != nil {
		fmt.Println(err)
	}

	if verboseFlag {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose mode enabled")
	}

	switch command {
	case "pull":
		fmt.Println("Current Configuration:")
		printConfigs(configs)
		fmt.Println("macOS Configuration:")
		macOSConfigs, err := pullop.Pull(configs)
		if err != nil {
			printError("Failed to pull configurations")
			return 1
		}
		printConfigs(macOSConfigs)

		if !yesFlag {
			color.Yellow("Warning: mdefaults will override your configuration file (~/.mdefaults). Proceed with caution.")
			fmt.Print("Do you want to continue? (yes/no): ")
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				fmt.Println("Failed to read input, operation cancelled.")
				return 1
			}
			if response != "yes" {
				fmt.Println("Operation cancelled.")
				return 0
			}
		}

		printSuccess("Configurations pulled successfully")
		if err := config.WriteConfigFile(fs, macOSConfigs); err != nil {
			log.Printf("Failed to write config file: %v", err)
			return 1
		}
		return 0
	case "push":
		printConfigs(configs)
		return handlePush(configs)
	case "debug":
		log.Println("Debug command executed")
		// Add more debug information here
		return 0
	default:
		log.Println("Error: Unknown command")
		return 1
	}
}

func printUsage() {
	fmt.Println("Usage: mdefaults [command]")
	fmt.Println("Commands:")
	fmt.Println("  pull    - Retrieve and update configuration values.")
	fmt.Println("  push    - Write configuration values.")
	fmt.Println("Hey, let's call with pull or push.")
}

func printConfigs(configs []config.Config) {
	for i := 0; i < len(configs); i++ {
		fmt.Printf("- %s %s %s\n", configs[i].Domain, configs[i].Key, *configs[i].Value)
	}
}

func printError(message string) {
	color.Red("Error: %s", message)
}

func printSuccess(message string) {
	color.Green("Success: %s", message)
}

func handlePush(configs []config.Config) int {
	pushop.Push(configs)
	printSuccess("Configurations pushed successfully")
	return 0
}

func main() {
	setupLogging()
	osType := runtime.GOOS
	if osType == "linux" || osType == "windows" {
		fmt.Println("Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows.")
	}
	initFlags()
	flag.Parse()

	if versionFlag || vFlag {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Architecture: %s\n", architecture)
		return
	}

	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Architecture: %s\n", architecture)
	os.Exit(run())
}
