package main

import (
	"context"
	"fmt"
	"os/exec"
)

// DefaultsCommand interface defines methods for reading and writing defaults.
type DefaultsCommand interface {
	Read(ctx context.Context) (string, error)
	Write(ctx context.Context, value string) error
}

// DefaultsCommandImpl is an implementation of the DefaultsCommand interface.
type DefaultsCommandImpl struct {
	domain string
	key    string
}

// Read executes a command to read a default setting.
func (d *DefaultsCommandImpl) Read(ctx context.Context) (string, error) {
	command := fmt.Sprintf("defaults read %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Write executes a command to write a default setting.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string) error {
	command := fmt.Sprintf("defaults write %s %s %s", d.domain, d.key, value)
	_, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return err
	}
	return nil
}
