package pull

import (
	"context"
	"fmt"
	"strings"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
	apperrors "github.com/fumiya-kume/mdefaults/internal/errors"
)

// Pull retrieves the current values of the specified configurations from the system defaults
func Pull(configs []config.Config) ([]config.Config, error) {
	defaultsCmds := make([]defaults.DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaultsCmds = append(defaultsCmds, defaults.NewDefaultsCommandImpl(configs[i].Domain, configs[i].Key))
	}
	return PullImpl(defaultsCmds)
}

// PullImpl is the implementation of the Pull function that works with DefaultsCommand interfaces
func PullImpl(defaultsCmds []defaults.DefaultsCommand) ([]config.Config, error) {
	updatedConfigs := make([]config.Config, 0, len(defaultsCmds))
	var lastError error

	for i := 0; i < len(defaultsCmds); i++ {
		value, err := defaultsCmds[i].Read(context.Background())
		if err != nil {
			// Store the last error to return it later
			lastError = apperrors.Wrap(err, apperrors.DefaultsReadError,
				fmt.Sprintf("failed to read defaults for domain '%s' and key '%s'",
					defaultsCmds[i].Domain(), defaultsCmds[i].Key()))
			continue
		}
		value = strings.ReplaceAll(value, "\n", "")
		updatedConfigs = append(updatedConfigs, config.Config{
			Domain: defaultsCmds[i].Domain(),
			Key:    defaultsCmds[i].Key(),
			Value:  &value,
		})
	}

	return updatedConfigs, lastError
}
