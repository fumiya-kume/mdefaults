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
		// Skip if domain or key is empty
		if defaults[i].Domain() == "" || defaults[i].Key() == "" {
			continue
		}

		// Create a base config entry with domain and key
		config := Config{
			Domain: defaults[i].Domain(),
			Key:    defaults[i].Key(),
			Type:   "string", // Default type
		}

		// Try to read value and type
		value, err := defaults[i].Read(context.Background())
		if err != nil {
			// If we can't read the value, still keep the key in the config but without a value
			updatedConfigs = append(updatedConfigs, config)
			continue
		}

		value = strings.TrimSpace(strings.ReplaceAll(value, "\n", ""))
		config.Value = &value

		// Read the value type
		if valueType, typeErr := defaults[i].ReadType(context.Background()); typeErr == nil {
			config.Type = valueType
		}

		updatedConfigs = append(updatedConfigs, config)
	}
	return updatedConfigs, nil
}
