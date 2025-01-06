package main

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfigFile_Success(t *testing.T) {
	fs := MockFileSystem{homeDir: "/mock/home", statError: nil, createErr: nil, configFileContent: "com.apple.dock\nautohide\n"}

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
	fs := MockFileSystem{statError: errors.New("read error")}

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
