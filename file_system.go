package main

import (
	"os"
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
	if _, err := fs.Stat(configFilePath); os.IsNotExist(err) {
		file, err := fs.Create(configFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func readConfigFileString(fs FileSystem) (string, error) {
	content, err := fs.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
