package push

import (
	"context"
	"log"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
	apperrors "github.com/fumiya-kume/mdefaults/internal/errors"
)

// Push writes the provided configurations to the system defaults.
func Push(configs []config.Config) {
	for _, cfg := range configs {
		if cfg.Value == nil {
			log.Printf("Skipping %s: Value is nil", cfg.Key)
			continue
		}
		defaults := defaults.NewDefaultsCommandImpl(cfg.Domain, cfg.Key)
		if err := defaults.Write(context.Background(), *cfg.Value); err != nil {
			// Extract error code if it's our error type
			code := apperrors.GetErrorCode(err)
			log.Printf("[ERROR-%04d] Failed to write defaults for %s: %v", code, cfg.Key, err)
		}
	}
}
