package main

import (
	"context"
)

func push(configs []Config) {
	for _, config := range configs {
		defaults := DefaultsCommandImpl{
			domain: config.Domain,
			key:    config.Key,
		}
		defaults.Write(context.Background(), config.Value)
	}
}
