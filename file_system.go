package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// OSFileSystem is a concrete implementation of FileSystem using the os package
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
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

func (f *fileSystem) WriteFile(name string, content string) error {
	return os.WriteFile(name, []byte(content), 0644)
}

func (f *fileSystem) ReadFile(name string) (string, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(content), nil
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
	content, err := fs.ReadFile(configFile)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
