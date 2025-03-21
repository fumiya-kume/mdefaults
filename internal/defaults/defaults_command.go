package defaults

import (
	"context"
	"fmt"
	"os/exec"

	apperrors "github.com/fumiya-kume/mdefaults/internal/errors"
)

// DefaultsCommand interface defines methods for reading and writing defaults.
type DefaultsCommand interface {
	Read(ctx context.Context) (string, error)
	Write(ctx context.Context, value string) error
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
		return "", apperrors.New(apperrors.DefaultsEmptyDomainOrKey, "domain and key cannot be empty")
	}
	command := fmt.Sprintf("defaults read %s %s", d.domain, d.key)
	output, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.DefaultsReadError, fmt.Sprintf("failed to read defaults for domain '%s' and key '%s'", d.domain, d.key))
	}
	return string(output), nil
}

// Write executes a command to write a default setting.
func (d *DefaultsCommandImpl) Write(ctx context.Context, value string) error {
	if d.domain == "" || d.key == "" {
		return apperrors.New(apperrors.DefaultsEmptyDomainOrKey, "domain and key cannot be empty")
	}
	command := fmt.Sprintf("defaults write %s %s %s", d.domain, d.key, value)
	_, err := exec.CommandContext(ctx, "bash", "-c", command).Output()
	if err != nil {
		return apperrors.Wrap(err, apperrors.DefaultsWriteError, fmt.Sprintf("failed to write defaults for domain '%s' and key '%s'", d.domain, d.key))
	}
	return nil
}
