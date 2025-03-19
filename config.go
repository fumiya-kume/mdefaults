package main

import (
	"fmt"
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
	lines := strings.Split(content, "\n")
	for _, line := range lines {
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
	var content strings.Builder
	for _, config := range configs {
		if config.Value == nil {
			// Include entries without values, but still include the type information
			content.WriteString(fmt.Sprintf("%s %s -%s\n", config.Domain, config.Key, config.Type))
		} else {
			// Include the type and value for complete entries
			content.WriteString(fmt.Sprintf("%s %s -%s %s\n", config.Domain, config.Key, config.Type, *config.Value))
		}
	}
	return content.String()
}

func writeConfigFile(fs FileSystem, configs []Config) error {
	content := generateConfigFileContent(configs)
	return fs.WriteFile(configFilePath, content)
}
