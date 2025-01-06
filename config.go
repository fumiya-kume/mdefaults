package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a configuration entry with domain, key, and value.
type Config struct {
	Domain string
	Key    string
	Value  string
}

var configFilePath = filepath.Join(os.Getenv("HOME"), ".mdefaults")

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

// generateConfigFileContent generates the content for the configuration file from a slice of Config.
func generateConfigFileContent(configs []Config) string {
	content := ""
	for _, config := range configs {
		content += fmt.Sprintf("%s %s %s\n", config.Domain, config.Key, config.Value)
	}
	return content
}

func writeConfigFile(fs FileSystem, configs []Config) error {
	content := generateConfigFileContent(configs)
	return fs.WriteFile(configFilePath, content)
}
