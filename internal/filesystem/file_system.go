package filesystem

import (
	"os"

	"github.com/fumiya-kume/mdefaults/internal/config"
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
	return os.UserHomeDir()
}

func (f *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (f *OSFileSystem) WriteFile(name string, content string) error {
	return os.WriteFile(name, []byte(content), 0644)
}

func (f *OSFileSystem) ReadFile(name string) (string, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// createConfigFileIfMissing checks for the existence of the config file and creates it if it doesn't exist
func CreateConfigFileIfMissing(fs FileSystem) error {
	if _, err := fs.Stat(config.ConfigFilePath); os.IsNotExist(err) {
		file, err := fs.Create(config.ConfigFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

// readConfigFileString reads the config file and returns its content as a string
func ReadConfigFileString(fs FileSystem) (string, error) {
	return fs.ReadFile(config.ConfigFilePath)
}
