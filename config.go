package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a configuration entry with domain, key, value, and type.
type Config struct {
	Domain string
	Key    string
	Value  *string
	Type   string // Added Type field to store the value type
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
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				config := Config{Domain: parts[0], Key: parts[1], Type: "string"} // Default to string type

				// Check if type information is available (format: domain key -type value)
				if len(parts) >= 4 && strings.HasPrefix(parts[2], "-") {
					config.Type = strings.TrimPrefix(parts[2], "-")
					if len(parts) > 3 {
						value := strings.Join(parts[3:], " ")
						config.Value = &value
					}
				} else if len(parts) > 2 {
					// Old format without type information
					value := strings.Join(parts[2:], " ")
					config.Value = &value
				}

				configs = append(configs, config)
			}
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
		// Include the type in the config file format: domain key -type value
		content += fmt.Sprintf("%s %s -%s %s\n", config.Domain, config.Key, config.Type, *config.Value)
	}
	return content
}

func writeConfigFile(fs FileSystem, configs []Config) error {
	content := generateConfigFileContent(configs)
	return fs.WriteFile(configFilePath, content)
}
