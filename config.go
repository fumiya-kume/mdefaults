package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// OSFileSystem is a concrete implementation of FileSystem using the os package
type OSFileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

type osFileSystem struct{}

func (f *osFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (f *osFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *osFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// setupConfigFile checks for the existence of the config file and creates it if it doesn't exist
func setupConfigFile(fs OSFileSystem) {
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
