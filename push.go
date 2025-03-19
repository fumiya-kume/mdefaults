package main

import (
	"context"
	"log"

	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func push(configs []Config) {
	for _, config := range configs {
		if config.Value == nil {
			log.Printf("Skipping %s: Value is nil", config.Key)
			continue
		}
		defaults := defaults.NewDefaultsCommandImpl(config.Domain, config.Key)
		if err := defaults.Write(context.Background(), *config.Value); err != nil {
			log.Printf("Failed to write defaults for %s: %v", config.Key, err)
		}
	}
}
