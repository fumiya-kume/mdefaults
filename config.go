package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a configuration entry with domain, key, and value.
type Config struct {
	Domain string
	Key    string
	Value  *string
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
			parts := strings.Split(line, " ")
			value := ""
			if len(parts) == 3 {
				value = parts[2]
			}
			configs = append(configs, Config{Domain: parts[0], Key: parts[1], Value: &value})
		}
	}
	return configs, nil
}

// generateConfigFileContent generates the content for the configuration file from a slice of Config.
func generateConfigFileContent(configs []Config) string {
	content := ""
	for _, config := range configs {
		if config.Value == nil {
			log.Printf("Skipping %s: Value is nil", config.Key)
			continue
		}
		content += fmt.Sprintf("%s %s %s\n", config.Domain, config.Key, *config.Value)
	}
	return content
}

func writeConfigFile(fs FileSystem, configs []Config) error {
	content := generateConfigFileContent(configs)
	return fs.WriteFile(configFilePath, content)
}
