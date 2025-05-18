package defaults

import (
	"context"
	"fmt"
	"os/exec"
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

// NewDefaultsCommandImpl creates a new DefaultsCommandImpl with the given domain and key.
func NewDefaultsCommandImpl(domain, key string) *DefaultsCommandImpl {
	return &DefaultsCommandImpl{
		domain: domain,
		key:    key,
	}
}

func (d *DefaultsCommandImpl) Domain() string {
	return d.domain
}

func (d *DefaultsCommandImpl) Key() string {
	return d.key
}

// Read executes a command to read a default setting.
func (d *DefaultsCommandImpl) Read(ctx context.Context) (string, error) {
	if d.domain == "" || d.key == "" {
		return "", fmt.Errorf("domain and key cannot be empty")
	}
	command := fmt.Sprintf("defaults read %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ReadType executes a command to read the type of a default setting.
func (d *DefaultsCommandImpl) ReadType(ctx context.Context) (string, error) {
	if d.domain == "" || d.key == "" {
		return "", fmt.Errorf("domain and key cannot be empty")
	}
	command := fmt.Sprintf("defaults read-type %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "", err
	}

	// Parse the output to extract the type
	typeValue := strings.TrimSpace(string(output))

	// Output format is "Type is <type>", so we extract just the type
	if strings.HasPrefix(typeValue, "Type is ") {
		typeValue = typeValue[8:] // Remove "Type is " prefix
	}

	// Convert the macOS type name to our internal representation
	switch typeValue {
	case "boolean":
		return "bool", nil
	case "integer":
		return "int", nil
	case "float":
		return "float", nil
	case "string":
		return "string", nil
	default:
		// Return as-is for other types (data, date, array, dict)
		return typeValue, nil
	}
}

// Write executes a command to write a default setting with the specified type.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string, valueType string) error {
	if d.domain == "" || d.key == "" {
		return fmt.Errorf("domain and key cannot be empty")
	}

	var command string

	// Use the appropriate type flag based on valueType
	switch valueType {
	case "bool":
		// For bool values, we need to handle "true"/"false" specifically
		if value == "true" || value == "1" {
			command = fmt.Sprintf("defaults write %s %s -bool true", d.domain, d.key)
		} else {
			command = fmt.Sprintf("defaults write %s %s -bool false", d.domain, d.key)
		}
	case "int":
		command = fmt.Sprintf("defaults write %s %s -int %s", d.domain, d.key, value)
	case "float":
		command = fmt.Sprintf("defaults write %s %s -float %s", d.domain, d.key, value)
	case "string":
		command = fmt.Sprintf("defaults write %s %s -string %s", d.domain, d.key, value)
	default:
		// Use string as default if type is not recognized or empty
		command = fmt.Sprintf("defaults write %s %s %s", d.domain, d.key, value)
	}

	_, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return err
	}
	return nil
}
