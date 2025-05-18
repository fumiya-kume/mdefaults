package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a configuration entry with domain, key, value and type.
type Config struct {
	Domain string
	Key    string
	Value  *string
	Type   string // Type can be: string, bool, int, float, data, date, array, dict
}

// ConfigFilePath is the default path for the configuration file.
var ConfigFilePath = filepath.Join(os.Getenv("HOME"), ".mdefaults")

// FileSystemReader is a minimal interface for file system operations needed by config package
type FileSystemReader interface {
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
}

// ReadConfigFile reads the configuration file and returns a slice of Config.
func ReadConfigFile(fs FileSystemReader) ([]Config, error) {
	content, err := fs.ReadFile(ConfigFilePath)
	if err != nil {
		return nil, err
	}
	configs := []Config{}
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				log.Printf("Skipping invalid line: %s", line)
				continue
			}

			domain := parts[0]
			key := parts[1]

			// Default type to string
			configType := "string"
			value := ""

			// Check if format includes type
			if len(parts) >= 4 && isValidType(parts[2]) {
				// Format: domain key type value
				configType = parts[2]
				value = strings.Join(parts[3:], " ")
			} else if len(parts) >= 3 {
				// Format: domain key value
				value = strings.Join(parts[2:], " ")
			}

			configs = append(configs, Config{
				Domain: domain,
				Key:    key,
				Value:  &value,
				Type:   configType,
			})
		}
	}
	return configs, nil
}

// isValidType checks if the given string is a valid type for macOS defaults
func isValidType(t string) bool {
	validTypes := []string{"string", "bool", "int", "float", "data", "date", "array", "dict"}
	for _, validType := range validTypes {
		if t == validType {
			return true
		}
	}
	return false
}

// GenerateConfigFileContent generates the content for the configuration file from a slice of Config.
func GenerateConfigFileContent(configs []Config) string {
	content := ""
	for _, config := range configs {
		if config.Value == nil {
			log.Printf("Skipping %s: Value is nil", config.Key)
			continue
		}

		// Include type in the output if it's not the default "string" type
		if config.Type != "" && config.Type != "string" {
			content += fmt.Sprintf("%s %s %s %s\n", config.Domain, config.Key, config.Type, *config.Value)
		} else {
			content += fmt.Sprintf("%s %s %s\n", config.Domain, config.Key, *config.Value)
		}
	}
	return content
}

// WriteConfigFile writes the configs to the configuration file.
func WriteConfigFile(fs FileSystemReader, configs []Config) error {
	content := GenerateConfigFileContent(configs)
	return fs.WriteFile(ConfigFilePath, content)
}
