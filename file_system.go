package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OSFileSystem is a concrete implementation of FileSystem using the os package
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

type fileSystem struct{}

func (f *fileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (f *fileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *fileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// createConfigFileIfMissing checks for the existence of the config file and creates it if it doesn't exist
func createConfigFileIfMissing(fs FileSystem) error {
	home, err := fs.UserHomeDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(home, ".config", ".mdefaults")
	if _, err := fs.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Config file not found, creating it")
		fs.Create(configFile)
	}
	return nil
}

func readConfigFileString(fs FileSystem) (string, error) {
	home, err := fs.UserHomeDir()
	if err != nil {
		return "", err
	}
	configFile := filepath.Join(home, ".config", ".mdefaults")
	content, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type Config struct {
	Domain string
	Key    string
	Value  string
}

func readConfigFile(fs FileSystem) ([]Config, error) {
	content, err := readConfigFileString(fs)
	if err != nil {
		return nil, err
	}
	configs := []Config{}
	for _, line := range strings.Split(content, "\n") {
		configs = append(configs, Config{Domain: line})
	}
	return configs, nil
}
