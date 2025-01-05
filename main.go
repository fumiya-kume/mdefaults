package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello, World!")

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	configFile := filepath.Join(home, ".config", ".mdefaults")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Config file not found, creating it")
		os.Create(configFile)
	}

	defaults := &DefaultsCommandImpl{
		domain: "com.apple.dock",
		key:    "autohide",
	}

	result, err := defaults.Read(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
