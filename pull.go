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
		value, err := defaults[i].Read(context.Background())
		if err != nil {
			continue
		}
		value = strings.TrimSpace(strings.ReplaceAll(value, "\n", ""))

		// Read the value type
		valueType, err := defaults[i].ReadType(context.Background())
		if err != nil {
			valueType = "string" // Default to string type if we can't determine the type
		}

		// Handle boolean values (convert 1/0 to true/false)
		if valueType == "boolean" {
			if value == "1" {
				value = "true"
			} else if value == "0" {
				value = "false"
			}
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
