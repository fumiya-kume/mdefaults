package main

import (
	"context"
	"log"
)

func push(configs []Config) {
	for _, config := range configs {
		defaults := DefaultsCommandImpl{
			domain: config.Domain,
			key:    config.Key,
		}
		if err := defaults.Write(context.Background(), config.Value); err != nil {
			log.Printf("Failed to write defaults for %s: %v", config.Key, err)
		}
	}
}
