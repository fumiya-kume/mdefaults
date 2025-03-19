package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// DefaultsCommand interface defines methods for reading and writing defaults.
type DefaultsCommand interface {
	Read(ctx context.Context) (string, error)
	ReadType(ctx context.Context) (string, error)
	Write(ctx context.Context, value string, valueType string) error
	Domain() string
	Key() string
}

// DefaultsCommandImpl is an implementation of the DefaultsCommand interface.
type DefaultsCommandImpl struct {
	domain string
	key    string
}

func (d *DefaultsCommandImpl) Domain() string {
	return d.domain
}

func (d *DefaultsCommandImpl) Key() string {
	return d.key
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

// ReadType executes a command to read the type of a default setting.
func (d *DefaultsCommandImpl) ReadType(ctx context.Context) (string, error) {
	command := fmt.Sprintf("defaults read-type %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "string", err // Default to string type on error
	}
	// Parse output which is in format "Type is <type>"
	outputStr := strings.TrimSpace(string(output))

	// Use a regular expression to match "Type is" with any number of spaces
	re := regexp.MustCompile(`^Type\s+is\s+(.+)$`)
	matches := re.FindStringSubmatch(outputStr)

	if len(matches) > 1 {
		typeStr := strings.TrimSpace(matches[1])
		return typeStr, nil
	}
	return "string", nil // Default to string type if parsing fails
}

// Write executes a command to write a default setting with the specified type.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string, valueType string) error {
	var command string

	// Format command based on the value type
	switch valueType {
	case "string":
		command = fmt.Sprintf(`defaults write %s %s -string "%s"`, d.domain, d.key, value)
	case "integer", "int":
		command = fmt.Sprintf("defaults write %s %s -int %s", d.domain, d.key, value)
	case "float", "real":
		command = fmt.Sprintf("defaults write %s %s -float %s", d.domain, d.key, value)
	case "bool", "boolean":
		command = fmt.Sprintf("defaults write %s %s -bool %s", d.domain, d.key, value)
	case "date":
		command = fmt.Sprintf(`defaults write %s %s -date "%s"`, d.domain, d.key, value)
	case "data":
		command = fmt.Sprintf(`defaults write %s %s -data %s`, d.domain, d.key, value)
	case "array":
		command = fmt.Sprintf("defaults write %s %s -array %s", d.domain, d.key, value)
	case "dict", "dictionary":
		command = fmt.Sprintf("defaults write %s %s -dict %s", d.domain, d.key, value)
	default:
		// If we don't recognize the type, fall back to not specifying a type
		command = fmt.Sprintf("defaults write %s %s %s", d.domain, d.key, value)
	}

	_, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	return err
}
