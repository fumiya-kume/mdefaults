package push

import (
	"context"
	"fmt"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

type defaultsWriter interface {
	Write(ctx context.Context, value string) error
	WriteWithType(ctx context.Context, value, valueType string) error
}

type defaultsWriterFactory func(domain, key string) defaultsWriter

// Push writes the provided configurations to the system defaults.
func Push(configs []config.Config) error {
	return pushImpl(context.Background(), configs, func(domain, key string) defaultsWriter {
		return defaults.NewDefaultsCommandImpl(domain, key)
	})
}

func pushImpl(ctx context.Context, configs []config.Config, newWriter defaultsWriterFactory) error {
	for _, cfg := range configs {
		if cfg.Domain == "" || cfg.Key == "" {
			return fmt.Errorf("write %s %s: domain and key cannot be empty", cfg.Domain, cfg.Key)
		}
		if cfg.Value == nil {
			return fmt.Errorf("write %s %s: value is nil", cfg.Domain, cfg.Key)
		}
		if err := defaults.ValidateValue(*cfg.Value, cfg.Type); err != nil {
			return fmt.Errorf("write %s %s: %w", cfg.Domain, cfg.Key, err)
		}
	}

	for _, cfg := range configs {
		writer := newWriter(cfg.Domain, cfg.Key)
		var err error
		if cfg.Type != "" && cfg.Type != "string" {
			err = writer.WriteWithType(ctx, *cfg.Value, cfg.Type)
		} else {
			err = writer.Write(ctx, *cfg.Value)
		}
		if err != nil {
			return fmt.Errorf("write %s %s: %w", cfg.Domain, cfg.Key, err)
		}
	}

	return nil
}
