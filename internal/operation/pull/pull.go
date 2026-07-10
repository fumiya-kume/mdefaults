package pull

import (
	"context"
	"fmt"
	"strings"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func Pull(configs []config.Config) ([]config.Config, error) {
	defaultsCmds := make([]defaults.DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaultsCmds = append(defaultsCmds, defaults.NewDefaultsCommandImpl(configs[i].Domain, configs[i].Key))
	}
	return PullImpl(defaultsCmds)
}

func PullImpl(defaultsCmds []defaults.DefaultsCommand) ([]config.Config, error) {
	updatedConfigs := make([]config.Config, 0, len(defaultsCmds))
	for i := 0; i < len(defaultsCmds); i++ {
		command := defaultsCmds[i]
		value, err := command.Read(context.Background())
		if err != nil {
			return nil, fmt.Errorf("read %s %s: %w", command.Domain(), command.Key(), err)
		}
		value = strings.TrimSuffix(value, "\n")

		valueType, err := command.ReadType(context.Background())
		if err != nil {
			return nil, fmt.Errorf("read type %s %s: %w", command.Domain(), command.Key(), err)
		}

		updatedConfigs = append(updatedConfigs, config.Config{
			Domain: command.Domain(),
			Key:    command.Key(),
			Value:  &value,
			Type:   valueType,
		})
	}
	return updatedConfigs, nil
}
