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

// ReadConfigFile reads the configuration file and returns a slice of Config.
func ReadConfigFile(fs FileSystem) ([]Config, error) {
	content, err := ReadConfigFileString(fs)
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

func WriteConfigFile(fs FileSystem, configs []Config) error {
	content := GenerateConfigFileContent(configs)
	return fs.WriteFile(ConfigFilePath, content)
}

// FileSystem interface for file operations
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
}

// CreateConfigFileIfMissing checks for the existence of the config file and creates it if it doesn't exist
func CreateConfigFileIfMissing(fs FileSystem) error {
	if _, err := fs.Stat(ConfigFilePath); os.IsNotExist(err) {
		file, err := fs.Create(ConfigFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func ReadConfigFileString(fs FileSystem) (string, error) {
	content, err := fs.ReadFile(ConfigFilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
