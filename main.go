package main

import (
	"context"
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("Hello, World!")
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
