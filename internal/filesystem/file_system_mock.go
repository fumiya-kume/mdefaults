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

func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.HomeDir, nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatError != nil {
		return nil, m.StatError
	}
	if m.ConfigFileContent != "" {
		return nil, os.ErrNotExist
	}
	return nil, nil
}

func (m *MockFileSystem) Create(name string) (*os.File, error) {
	return nil, m.CreateErr
}

func (m *MockFileSystem) ReadFile(name string) (string, error) {
	if m.StatError != nil {
		return "", m.StatError
	}
	return m.ConfigFileContent, nil
}

func (m *MockFileSystem) WriteFile(name string, content string) error {
	m.WriteFileContent = content
	return m.WriteFileErr
}