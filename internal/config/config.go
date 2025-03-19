package config

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
			value := ""
			if len(parts) == 3 {
				value = parts[2]
			}
			configs = append(configs, Config{Domain: parts[0], Key: parts[1], Value: &value})
		}
	}
	return configs, nil
}

// GenerateConfigFileContent generates the content for the configuration file from a slice of Config.
func GenerateConfigFileContent(configs []Config) string {
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

// WriteConfigFile writes the configs to the configuration file.
func WriteConfigFile(fs FileSystemReader, configs []Config) error {
	content := GenerateConfigFileContent(configs)
	return fs.WriteFile(ConfigFilePath, content)
}
