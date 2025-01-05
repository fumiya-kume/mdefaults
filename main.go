package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// 1. make the config file at ~/.config/.mdefaults if missing
// 2. print the config file

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

type DefaultsCommand interface {
	Read(ctx context.Context) (string, error)
	Write(ctx context.Context, value string) error
}

type DefaultsCommandImpl struct {
	domain string
	key    string
}

func (d *DefaultsCommandImpl) Read(ctx context.Context) (string, error) {
	command := fmt.Sprintf("defaults read %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (d *DefaultsCommandImpl) Write(ctx context.Context, value string) error {
	command := fmt.Sprintf("defaults write %s %s %s", d.domain, d.key, value)
	_, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return err
	}
	return nil
}
