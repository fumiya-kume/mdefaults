package filesystem

import (
	"fmt"
	"os"

	"github.com/fumiya-kume/mdefaults/internal/config"
	apperrors "github.com/fumiya-kume/mdefaults/internal/errors"
)

// OSFileSystem is a concrete implementation of FileSystem using the os package
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
}

// OSFileSystem is a concrete implementation of the FileSystem interface
type OSFileSystem struct{}

// NewOSFileSystem creates a new instance of OSFileSystem
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) UserHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.FileReadError, "failed to get user home directory")
	}
	return dir, nil
}

func (f *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apperrors.Wrap(err, apperrors.FileNotFound, fmt.Sprintf("file not found: %s", name))
		}
		return nil, apperrors.Wrap(err, apperrors.FileReadError, fmt.Sprintf("failed to stat file: %s", name))
	}
	return info, nil
}

func (f *OSFileSystem) Create(name string) (*os.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.FileWriteError, fmt.Sprintf("failed to create file: %s", name))
	}
	return file, nil
}

func (f *OSFileSystem) WriteFile(name string, content string) error {
	err := os.WriteFile(name, []byte(content), 0644)
	if err != nil {
		return apperrors.Wrap(err, apperrors.FileWriteError, fmt.Sprintf("failed to write file: %s", name))
	}
	return nil
}

func (f *OSFileSystem) ReadFile(name string) (string, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.FileReadError, fmt.Sprintf("failed to read file: %s", name))
	}
	return string(content), nil
}

// createConfigFileIfMissing checks for the existence of the config file and creates it if it doesn't exist
func CreateConfigFileIfMissing(fs FileSystem) error {
	if _, err := fs.Stat(config.ConfigFilePath); os.IsNotExist(err) {
		file, err := fs.Create(config.ConfigFilePath)
		if err != nil {
			return apperrors.Wrap(err, apperrors.FileWriteError, fmt.Sprintf("failed to create config file: %s", config.ConfigFilePath))
		}
		defer func() {
			_ = file.Close()
		}()
	}
	return nil
}

// readConfigFileString reads the config file and returns its content as a string
func ReadConfigFileString(fs FileSystem) (string, error) {
	content, err := fs.ReadFile(config.ConfigFilePath)
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.FileReadError, fmt.Sprintf("failed to read config file: %s", config.ConfigFilePath))
	}
	return content, nil
}
