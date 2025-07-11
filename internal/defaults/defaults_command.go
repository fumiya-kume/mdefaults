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
	Write(ctx context.Context, value string) error
	WriteWithType(ctx context.Context, value string, valueType string) error
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
	output, err := exec.CommandContext(ctx, "defaults", "read", d.domain, d.key).Output()
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
	output, err := exec.CommandContext(ctx, "defaults", "read-type", d.domain, d.key).Output()
	if err != nil {
		return "string", nil
	}
	typeOutput := strings.TrimSpace(string(output))
	if strings.Contains(typeOutput, "Type is ") {
		parts := strings.Split(typeOutput, "Type is ")
		if len(parts) > 1 {
			return mapMacOSTypeToInternal(strings.TrimSpace(parts[1])), nil
		}
	}
	return "string", nil
}

// Write executes a command to write a default setting.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string) error {
	if d.domain == "" || d.key == "" {
		return fmt.Errorf("domain and key cannot be empty")
	}
	_, err := exec.CommandContext(ctx, "defaults", "write", d.domain, d.key, value).Output()
	if err != nil {
		return err
	}
	return nil
}

func (d *DefaultsCommandImpl) WriteWithType(ctx context.Context, value string, valueType string) error {
	if d.domain == "" || d.key == "" {
		return fmt.Errorf("domain and key cannot be empty")
	}
	
	typeFlag := mapInternalTypeToFlag(valueType)
	var args []string
	if typeFlag == "" || valueType == "string" {
		args = []string{"defaults", "write", d.domain, d.key, value}
	} else {
		args = []string{"defaults", "write", d.domain, d.key, typeFlag, value}
	}
	
	_, err := exec.CommandContext(ctx, args[0], args[1:]...).Output()
	if err != nil {
		return err
	}
	return nil
}

func mapMacOSTypeToInternal(macOSType string) string {
	switch strings.ToLower(macOSType) {
	case "integer":
		return "integer"
	case "boolean":
		return "boolean"
	case "string":
		return "string"
	case "float", "real":
		return "float"
	case "date":
		return "date"
	case "array":
		return "array"
	case "dictionary":
		return "dict"
	case "data":
		return "data"
	default:
		return "string"
	}
}

func mapInternalTypeToFlag(internalType string) string {
	switch internalType {
	case "integer":
		return "-int"
	case "boolean":
		return "-bool"
	case "float":
		return "-float"
	case "date":
		return "-date"
	case "array":
		return "-array"
	case "dict":
		return "-dict"
	case "data":
		return "-data"
	case "string":
		return ""
	default:
		return ""
	}
}
