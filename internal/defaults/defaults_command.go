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
	run    commandRunner
}

type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// NewDefaultsCommandImpl creates a new DefaultsCommandImpl with the given domain and key.
func NewDefaultsCommandImpl(domain, key string) *DefaultsCommandImpl {
	return newDefaultsCommandImpl(domain, key, runCommand)
}

func newDefaultsCommandImpl(domain, key string, run commandRunner) *DefaultsCommandImpl {
	return &DefaultsCommandImpl{
		domain: domain,
		key:    key,
		run:    run,
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
	if err := validateDomainAndKey(d.domain, d.key); err != nil {
		return "", err
	}
	output, err := d.run(ctx, "defaults", "read", d.domain, d.key)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ReadType executes a command to read the type of a default setting.
func (d *DefaultsCommandImpl) ReadType(ctx context.Context) (string, error) {
	if err := validateDomainAndKey(d.domain, d.key); err != nil {
		return "", err
	}
	output, err := d.run(ctx, "defaults", "read-type", d.domain, d.key)
	if err != nil {
		return "", err
	}
	return parseMacOSTypeOutput(string(output))
}

// Write executes a command to write a default setting.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string) error {
	if err := validateDomainAndKey(d.domain, d.key); err != nil {
		return err
	}
	_, err := d.run(ctx, "defaults", "write", d.domain, d.key, value)
	if err != nil {
		return err
	}
	return nil
}

func (d *DefaultsCommandImpl) WriteWithType(ctx context.Context, value string, valueType string) error {
	if err := validateDomainAndKey(d.domain, d.key); err != nil {
		return err
	}

	normalizedValue, err := normalizeValue(value, valueType)
	if err != nil {
		return err
	}

	typeFlag := mapInternalTypeToFlag(valueType)
	var args []string
	if typeFlag == "" {
		args = []string{"write", d.domain, d.key, normalizedValue}
	} else {
		args = []string{"write", d.domain, d.key, typeFlag, normalizedValue}
	}

	_, err = d.run(ctx, "defaults", args...)
	if err != nil {
		return err
	}
	return nil
}

func validateDomainAndKey(domain, key string) error {
	if domain == "" || key == "" {
		return fmt.Errorf("domain and key cannot be empty")
	}
	return nil
}

func parseMacOSTypeOutput(output string) (string, error) {
	const typePrefix = "Type is "

	output = strings.TrimSpace(output)
	if !strings.HasPrefix(output, typePrefix) {
		return "", fmt.Errorf("unexpected defaults type output %q", output)
	}
	return mapMacOSTypeToInternal(strings.TrimSpace(strings.TrimPrefix(output, typePrefix)))
}

func mapMacOSTypeToInternal(macOSType string) (string, error) {
	switch strings.ToLower(macOSType) {
	case "integer":
		return "integer", nil
	case "boolean":
		return "boolean", nil
	case "string":
		return "string", nil
	case "float", "real":
		return "float", nil
	case "date":
		return "date", nil
	case "array", "dictionary", "data":
		return "", fmt.Errorf("unsupported macOS defaults type %q", macOSType)
	default:
		return "", fmt.Errorf("unknown macOS defaults type %q", macOSType)
	}
}

// ValidateType reports whether mdefaults can safely write the supplied type.
func ValidateType(valueType string) error {
	if valueType == "" || valueType == "string" || mapInternalTypeToFlag(valueType) != "" {
		return nil
	}
	return fmt.Errorf("unsupported defaults type %q", valueType)
}

// ValidateValue reports whether mdefaults can safely write the value as the supplied type.
func ValidateValue(value, valueType string) error {
	_, err := normalizeValue(value, valueType)
	return err
}

func normalizeValue(value, valueType string) (string, error) {
	if err := ValidateType(valueType); err != nil {
		return "", err
	}
	if valueType != "boolean" {
		return value, nil
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes":
		return "true", nil
	case "0", "false", "no":
		return "false", nil
	default:
		return "", fmt.Errorf("invalid boolean value %q: expected 1, 0, true, false, yes, or no", value)
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
	case "string":
		return ""
	default:
		return ""
	}
}
