package main

import (
	"context"
	"strings"

	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func pull(configs []Config) ([]Config, error) {
	defaultsCmds := make([]defaults.DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaultsCmds = append(defaultsCmds, defaults.NewDefaultsCommandImpl(configs[i].Domain, configs[i].Key))
	}
	return pullImpl(defaultsCmds)
}

func pullImpl(defaultsCmds []defaults.DefaultsCommand) ([]Config, error) {
	updatedConfigs := make([]Config, 0, len(defaultsCmds))
	for i := 0; i < len(defaultsCmds); i++ {
		value, err := defaultsCmds[i].Read(context.Background())
		if err != nil {
			continue
		}
		value = strings.ReplaceAll(value, "\n", "")
		updatedConfigs = append(updatedConfigs, Config{
			Domain: defaultsCmds[i].Domain(),
			Key:    defaultsCmds[i].Key(),
			Value:  &value,
		})
	}
	return updatedConfigs, nil
}
