package main

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfigFile_Success(t *testing.T) {
	fs := &MockFileSystem{homeDir: "/mock/home", statError: nil, createErr: nil, configFileContent: "com.apple.dock\nautohide\n"}

	configs, err := readConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedConfigs := []Config{
		{Domain: "com.apple.dock"},
		{Domain: "autohide"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Fatalf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, config := range configs {
		if config.Domain != expectedConfigs[i].Domain {
			t.Errorf("Expected domain %s, got %s", expectedConfigs[i].Domain, config.Domain)
		}
	}
}

func TestReadConfigFile_Error(t *testing.T) {
	fs := &MockFileSystem{statError: errors.New("read error")}

	_, err := readConfigFile(fs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "read error") {
		t.Errorf("Expected error to contain 'read error', got %v", err)
	}
}

func TestGenerateConfigFileContent(t *testing.T) {
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: "1"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: "true"},
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	content := generateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestGenerateConfigFileContent_Empty(t *testing.T) {
	configs := []Config{}

	expectedContent := ""
	content := generateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestWriteConfigFile_Success(t *testing.T) {
	mockFS := &MockFileSystem{
		homeDir:          "/mock/home",
		writeFileContent: "",
		writeFileErr:     nil,
	}

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: "1"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: "true"},
	}

	err := writeConfigFile(mockFS, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	if mockFS.writeFileContent != expectedContent {
		t.Errorf("Expected writeFileContent %q, got %q", expectedContent, mockFS.writeFileContent)
	}
}

func TestWriteConfigFile_Error(t *testing.T) {
	mockFS := &MockFileSystem{
		homeDir:      "/mock/home",
		writeFileErr: errors.New("write error"),
	}

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: "1"},
	}

	err := writeConfigFile(mockFS, configs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.writeFileErr) {
		t.Errorf("Expected error %v, got %v", mockFS.writeFileErr, err)
	}
}
