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
