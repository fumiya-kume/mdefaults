package main

import (
	"context"
	"strings"
)

func pull(configs []Config) ([]Config, error) {
	defaults := make([]DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaults = append(defaults, &DefaultsCommandImpl{
			domain: configs[i].Domain,
			key:    configs[i].Key,
		})
	}
	return pullImpl(defaults)
}

func pullImpl(defaults []DefaultsCommand) ([]Config, error) {
	updatedConfigs := make([]Config, 0, len(defaults))
	for i := 0; i < len(defaults); i++ {
		// Read the value
		value, err := defaults[i].Read(context.Background())
		if err != nil {
			continue
		}
		value = strings.ReplaceAll(value, "\n", "")

		// Read the value type
		valueType, err := defaults[i].ReadType(context.Background())
		if err != nil {
			// If we can't determine the type, default to string
			valueType = "string"
		}

		updatedConfigs = append(updatedConfigs, Config{
			Domain: defaults[i].Domain(),
			Key:    defaults[i].Key(),
			Value:  &value,
			Type:   valueType,
		})
	}
	return updatedConfigs, nil
}
