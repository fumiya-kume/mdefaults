package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fumiya-kume/mdefaults/internal/config"
)

// OSFileSystem is a concrete implementation of FileSystem using the os package
type FileSystem interface {
	UserHomeDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) error
	Chmod(name string, mode os.FileMode) error
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
}

type atomicFile interface {
	Write(p []byte) (int, error)
	Sync() error
	Close() error
	Name() string
}

type atomicWriteOperations struct {
	createTemp func(dir string, pattern string) (atomicFile, error)
	rename     func(oldPath string, newPath string) error
	remove     func(path string) error
}

// OSFileSystem is a concrete implementation of the FileSystem interface
type OSFileSystem struct {
	writeOperations atomicWriteOperations
}

// NewOSFileSystem creates a new instance of OSFileSystem
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{writeOperations: defaultAtomicWriteOperations()}
}

func (f *OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (f *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFileSystem) Create(name string) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	return file.Close()
}

func (f *OSFileSystem) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

func (f *OSFileSystem) WriteFile(name string, content string) (err error) {
	operations := f.effectiveWriteOperations()
	directory := filepath.Dir(name)
	pattern := "." + filepath.Base(name) + ".tmp-*"
	temporaryFile, err := operations.createTemp(directory, pattern)
	if err != nil {
		return fmt.Errorf("create temporary file for %q: %w", name, err)
	}

	temporaryPath := temporaryFile.Name()
	closed := false
	renamed := false
	defer func() {
		if !closed {
			closeErr := temporaryFile.Close()
			if closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close temporary file %q during cleanup: %w", temporaryPath, closeErr))
			}
		}
		if !renamed {
			removeErr := operations.remove(temporaryPath)
			if removeErr != nil && !os.IsNotExist(removeErr) {
				err = errors.Join(err, fmt.Errorf("remove temporary file %q: %w", temporaryPath, removeErr))
			}
		}
	}()

	if err := writeAll(temporaryFile, []byte(content)); err != nil {
		return fmt.Errorf("write temporary file %q: %w", temporaryPath, err)
	}
	if err := temporaryFile.Sync(); err != nil {
		return fmt.Errorf("sync temporary file %q: %w", temporaryPath, err)
	}
	closed = true
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("close temporary file %q: %w", temporaryPath, err)
	}
	if err := operations.rename(temporaryPath, name); err != nil {
		return fmt.Errorf("replace %q with temporary file: %w", name, err)
	}
	renamed = true
	return nil
}

func defaultAtomicWriteOperations() atomicWriteOperations {
	return atomicWriteOperations{
		createTemp: func(dir string, pattern string) (atomicFile, error) {
			return os.CreateTemp(dir, pattern)
		},
		rename: os.Rename,
		remove: os.Remove,
	}
}

func (f *OSFileSystem) effectiveWriteOperations() atomicWriteOperations {
	operations := f.writeOperations
	defaults := defaultAtomicWriteOperations()
	if operations.createTemp == nil {
		operations.createTemp = defaults.createTemp
	}
	if operations.rename == nil {
		operations.rename = defaults.rename
	}
	if operations.remove == nil {
		operations.remove = defaults.remove
	}
	return operations
}

func writeAll(file atomicFile, content []byte) error {
	for len(content) > 0 {
		written, err := file.Write(content)
		if err != nil {
			return err
		}
		if written <= 0 || written > len(content) {
			return io.ErrShortWrite
		}
		content = content[written:]
	}
	return nil
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
	_, err := fs.Stat(config.ConfigFilePath)
	if err == nil {
		return fs.Chmod(config.ConfigFilePath, 0o600)
	}
	if !os.IsNotExist(err) {
		return err
	}
	err = fs.Create(config.ConfigFilePath)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return fs.Chmod(config.ConfigFilePath, 0o600)
}

// readConfigFileString reads the config file and returns its content as a string
func ReadConfigFileString(fs FileSystem) (string, error) {
	return fs.ReadFile(config.ConfigFilePath)
}
