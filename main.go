package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// FileSystem interface for file operations
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

// OSFileSystem is a concrete implementation of FileSystem using the os package
type OSFileSystem struct{}

func (OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (OSFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func main() {
	fs := OSFileSystem{}
	setupConfigFile(fs)

	defaults := &DefaultsCommandImpl{
		domain: "com.apple.dock",
		key:    "autohide",
	}

	result, err := defaults.Read(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

func setupConfigFile(fs FileSystem) {
	home, err := fs.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	configFile := filepath.Join(home, ".config", ".mdefaults")
	if _, err := fs.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Config file not found, creating it")
		fs.Create(configFile)
	}
}
