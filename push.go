package main

import (
	"context"
	"log"
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

		// Use the stored type when writing the value, or default to string if not specified
		valueType, err := defaults.ReadType(context.Background())
		if err != nil {
			log.Printf("Failed to read type for %s: %v", config.Key, err)
			continue
		}
		if valueType == "" {
			valueType = "string"
		}

		if err := defaults.Write(context.Background(), *config.Value, valueType); err != nil {
			log.Printf("Failed to write defaults for %s: %v", config.Key, err)
		}
	}
}
