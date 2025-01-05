package main

import (
	"strings"
)

// Config represents a configuration entry with domain, key, and value.
type Config struct {
	Domain string
	Key    string
	Value  string
}

// readConfigFile reads the configuration file and returns a slice of Config.
func readConfigFile(fs FileSystem) ([]Config, error) {
	content, err := readConfigFileString(fs)
	if err != nil {
		return nil, err
	}
	configs := []Config{}
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			configs = append(configs, Config{Domain: line})
		}
	}
	return configs, nil
}
