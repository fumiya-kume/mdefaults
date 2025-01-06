package main

import (
	"os"
)

// MockFileSystem is a mock implementation of the FileSystem interface
type MockFileSystem struct {
	homeDir           string
	statError         error
	createErr         error
	configFileContent string
	writeFileErr      error
	writeFileContent  string
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statError != nil {
		return nil, m.statError
	}
	if m.configFileContent != "" {
		return nil, os.ErrNotExist
	}
	return nil, nil
}

func (m *MockFileSystem) Create(name string) (*os.File, error) {
	return nil, m.createErr
}

func (m *MockFileSystem) ReadFile(name string) (string, error) {
	if m.statError != nil {
		return "", m.statError
	}
	return m.configFileContent, nil
}

func (m *MockFileSystem) WriteFile(name string, content string) error {
	m.writeFileContent = content
	return m.writeFileErr
}
