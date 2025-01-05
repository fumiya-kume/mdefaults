package main

import (
	"errors"
	"os"
	"testing"
)

// MockFileSystem is a mock implementation of the FileSystem interface
// MockFileSystem inherits from OSFileSystem
type MockFileSystem struct {
	homeDir           string
	statError         error
	createErr         error
	configFileContent string
}

func (m MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}

func (m MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statError != nil {
		return nil, m.statError
	}
	if m.configFileContent != "" {
		return nil, os.ErrNotExist
	}
	return nil, nil
}

func (m MockFileSystem) Create(name string) (*os.File, error) {
	return nil, m.createErr
}

func (m MockFileSystem) ReadFile(name string) (string, error) {
	if m.statError != nil {
		return "", m.statError
	}
	return m.configFileContent, nil
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

func TestReadConfigFileString_Success(t *testing.T) {
	expectedContent := "com.apple.dock\nautohide\n"
	fs := MockFileSystem{homeDir: "/mock/home", configFileContent: expectedContent}

	content, err := readConfigFileString(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Error(t *testing.T) {
	fs := MockFileSystem{homeDir: "/mock/home", configFileContent: "", statError: errors.New("read error")}

	_, err := readConfigFileString(fs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, fs.statError) {
		t.Errorf("Expected error %v, got %v", fs.statError, err)
	}
}
