package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func run() int {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mdefaults [command]")
		fmt.Println("Commands:")
		fmt.Println("  pull    - Retrieve and update configuration values.")
		fmt.Println("  push    - Write configuration values.")
		fmt.Println("Hey, let's call with pull or push.")
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
	for i := 0; i < len(configs); i++ {
		fmt.Printf("- %s %s\n", configs[i].Domain, configs[i].Key)
	}

	switch command {
	case "pull":
		updatedConfigs, err := pull(configs)
		if err != nil {
			log.Fatal(err)
		}
		writeConfigFile(fs, updatedConfigs)
		return 0
	case "push":
		push(configs)
		return 0
	default:
		log.Fatal("Invalid command")
	}
	return 1
}

func main() {
	os.Exit(run())
}
