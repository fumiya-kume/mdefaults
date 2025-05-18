package push

import (
	"context"
	"log"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

// Push writes the provided configurations to the system defaults.
func Push(configs []config.Config) {
	for _, cfg := range configs {
		if cfg.Value == nil {
			log.Printf("Skipping %s: Value is nil", cfg.Key)
			continue
		}

		defaults := defaults.NewDefaultsCommandImpl(cfg.Domain, cfg.Key)

		// Use the config's type if available, otherwise default to string
		valueType := cfg.Type
		if valueType == "" {
			valueType = "string"
		}

		if err := defaults.Write(context.Background(), *cfg.Value, valueType); err != nil {
			log.Printf("Failed to write defaults for %s: %v", cfg.Key, err)
		}
	}
}
