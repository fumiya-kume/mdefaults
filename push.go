package main

import (
	"context"
	"log"
	"strconv"
	"strings"
)

func push(configs []Config) {
	for _, config := range configs {
		if config.Value == nil {
			log.Printf("Skipping %s: Value is nil", config.Key)
			continue
		}
		defaults := DefaultsCommandImpl{
			domain: config.Domain,
			key:    config.Key,
		}

		// Use the type specified in the config if available and not the default
		valueType := config.Type

		// If type is the default "string" or empty, try to infer the type from the value
		if valueType == "" || valueType == "string" {
			// Try to infer the type from the value
			value := strings.TrimSpace(*config.Value)

			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || 
			   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
				valueType = "string"
			} else if value == "true" || value == "false" || value == "1" || value == "0" {
				// Boolean values
				valueType = "boolean"
			} else if _, err := strconv.Atoi(value); err == nil {
				// Integer values
				valueType = "integer"
			} else if _, err := strconv.ParseFloat(value, 64); err == nil {
				// Float values
				valueType = "float"
			} else {
				// Default to string for everything else
				valueType = "string"
			}
		}

		if valueType == "" {
			valueType = "string"
		}

		if err := defaults.Write(context.Background(), *config.Value, valueType); err != nil {
			log.Printf("Failed to write defaults for %s: %v", config.Key, err)
		}
	}
}
