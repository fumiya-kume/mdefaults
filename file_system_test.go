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
	mockFS := &MockFileSystem{
		configFileContent: "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n",
	}

	content, err := readConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Empty(t *testing.T) {
	mockFS := &MockFileSystem{
		configFileContent: "",
	}

	content, err := readConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := ""
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Error(t *testing.T) {
	mockFS := &MockFileSystem{
		statError: errors.New("read error"),
	}

	_, err := readConfigFileString(mockFS)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.statError) {
		t.Errorf("Expected error %v, got %v", mockFS.statError, err)
	}
}

func TestReadConfigFileString_MalformedContent(t *testing.T) {
	mockFS := &MockFileSystem{
		configFileContent: "com.apple.dock autohide\nmalformed line without key\ncom.apple.finder ShowPathbar true\n",
	}

	content, err := readConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide\nmalformed line without key\ncom.apple.finder ShowPathbar true\n"
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}
