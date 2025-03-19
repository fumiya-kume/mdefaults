package main

import (
	"context"
	"log"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func push(configs []config.Config) {
	for _, cfg := range configs {
		if cfg.Value == nil {
			log.Printf("Skipping %s: Value is nil", cfg.Key)
			continue
		}
		defaults := defaults.NewDefaultsCommandImpl(cfg.Domain, cfg.Key)
		if err := defaults.Write(context.Background(), *cfg.Value); err != nil {
			log.Printf("Failed to write defaults for %s: %v", cfg.Key, err)
		}
	}
}
