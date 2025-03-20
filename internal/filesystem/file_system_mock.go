package filesystem

import (
	"os"
)

// MockFileSystem is a mock implementation of the FileSystem interface
type MockFileSystem struct {
	HomeDir           string
	StatError         error
	CreateErr         error
	ConfigFileContent string
	WriteFileErr      error
	WriteFileContent  string
}

// UserHomeDir returns the home directory for the mock file system.
func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.HomeDir, nil
}

// Stat returns file info for the given file name in the mock file system.
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatError != nil {
		return nil, m.StatError
	}
	if m.ConfigFileContent != "" {
		return nil, os.ErrNotExist
	}
	return nil, nil
}

// Create creates a new file in the mock file system.
func (m *MockFileSystem) Create(name string) (*os.File, error) {
	return nil, m.CreateErr
}

// ReadFile reads a file from the mock file system.
func (m *MockFileSystem) ReadFile(name string) (string, error) {
	if m.StatError != nil {
		return "", m.StatError
	}
	return m.ConfigFileContent, nil
}

// WriteFile writes content to a file in the mock file system.
func (m *MockFileSystem) WriteFile(name string, content string) error {
	m.WriteFileContent = content
	return m.WriteFileErr
}
