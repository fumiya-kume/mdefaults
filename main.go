package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
)

var (
	version      string
	architecture string
)

func initFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.BoolVar(&versionFlag, "version", false, "Print version information")
	flag.BoolVar(&vFlag, "v", false, "Print version information")
}

var (
	versionFlag bool
	vFlag       bool
)

func run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 0
	}

	command := os.Args[1]
	flag.Parse()
	fs := &fileSystem{}
	if err := createConfigFileIfMissing(fs); err != nil {
		log.Printf("Failed to create config file: %v", err)
	}
	configs, err := readConfigFile(fs)
	if err != nil {
		fmt.Println(err)
	}
	printConfigs(configs)

	switch command {
	case "pull":
		return handlePull(configs, fs)
	case "push":
		return handlePush(configs)
	default:
		log.Println("Error:", err)
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

func printConfigs(configs []Config) {
	for i := 0; i < len(configs); i++ {
		fmt.Printf("- %s %s\n", configs[i].Domain, configs[i].Key)
	}
}

func printError(message string) {
	color.Red("Error: %s", message)
}

func printWarning(message string) {
	color.Yellow("Warning: %s", message)
}

func printSuccess(message string) {
	color.Green("Success: %s", message)
}

func handlePull(configs []Config, fs *fileSystem) int {
	updatedConfigs, err := pull(configs)
	if err != nil {
		printError("Failed to pull configurations")
		return 1
	}
	printSuccess("Configurations pulled successfully")
	if err := writeConfigFile(fs, updatedConfigs); err != nil {
		log.Printf("Failed to write config file: %v", err)
	}
	return 0
}

func handlePush(configs []Config) int {
	push(configs)
	printSuccess("Configurations pushed successfully")
	return 0
}

func main() {
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
