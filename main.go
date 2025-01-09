package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 0
	}

	command := os.Args[1]
	flag.Parse()
	fs := &fileSystem{}
	createConfigFileIfMissing(fs)
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

	log.Println("Invalid command")
	return 1
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

func handlePull(configs []Config, fs *fileSystem) int {
	updatedConfigs, err := pull(configs)
	if err != nil {
		log.Fatal(err)
	}
	writeConfigFile(fs, updatedConfigs)
	return 0
}

func handlePush(configs []Config) int {
	push(configs)
	return 0
}

func main() {
	os.Exit(run())
}
