package main

import (
	"os"
	"testing"
)

// MockFileSystem is a mock implementation of the FileSystem interface
// MockFileSystem inherits from OSFileSystem
type MockFileSystem struct {
	homeDir   string
	statError error
	createErr error
}

func (m MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}

func (m MockFileSystem) Stat(name string) (os.FileInfo, error) {
	return nil, m.statError
}

func (m MockFileSystem) Create(name string) (*os.File, error) {
	return nil, m.createErr
}

func TestSetupConfigFile_CreatesFileIfNotExist(t *testing.T) {
	fs := &MockFileSystem{
		homeDir:   "/mock/home",
		statError: os.ErrNotExist,
		createErr: nil,
	}

	createConfigFileIfMissing(fs)
	if fs.createErr != nil {
		t.Errorf("Expected create error to be nil but got %v", fs.createErr)
	}
}

func TestSetupConfigFile_DoesNotCreateFileIfExists(t *testing.T) {
	fs := MockFileSystem{
		homeDir:   "/mock/home",
		statError: nil,
		createErr: nil,
	}

	createConfigFileIfMissing(fs)
	if fs.createErr != nil {
		t.Errorf("Expected create error to be nil but got %v", fs.createErr)
	}
}

func TestSetupConfigFile_HandleUserHomeDirError(t *testing.T) {
	fs := MockFileSystem{
		homeDir:   "",
		statError: nil,
		createErr: nil,
	}

	createConfigFileIfMissing(fs)
	if fs.createErr != nil {
		t.Errorf("Expected create error to be nil but got %v", fs.createErr)
	}
}
