package main

import (
	"errors"
	"os"
	"testing"
)

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
	fs := &MockFileSystem{
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
	fs := &MockFileSystem{
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
	fs := &MockFileSystem{homeDir: "/mock/home", configFileContent: expectedContent}

	content, err := readConfigFileString(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Error(t *testing.T) {
	fs := &MockFileSystem{homeDir: "/mock/home", configFileContent: "", statError: errors.New("read error")}

	_, err := readConfigFileString(fs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, fs.statError) {
		t.Errorf("Expected error %v, got %v", fs.statError, err)
	}
}
