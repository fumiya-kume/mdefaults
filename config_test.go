package main

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfigFile_Success(t *testing.T) {
	fs := &MockFileSystem{homeDir: "/mock/home", statError: nil, createErr: nil, configFileContent: "com.apple.dock autohide -boolean 1\ncom.apple.finder ShowPathbar -boolean true\n"}

	configs, err := readConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	value1 := "1"
	value2 := "true"
	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2, Type: "boolean"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Fatalf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, config := range configs {
		if config.Domain != expectedConfigs[i].Domain {
			t.Errorf("Expected domain %s, got %s", expectedConfigs[i].Domain, config.Domain)
		}
		if config.Key != expectedConfigs[i].Key {
			t.Errorf("Expected key %s, got %s", expectedConfigs[i].Key, config.Key)
		}
		if *config.Value != *expectedConfigs[i].Value {
			t.Errorf("Expected value %s, got %s", *expectedConfigs[i].Value, *config.Value)
		}
		if config.Type != expectedConfigs[i].Type {
			t.Errorf("Expected type %s, got %s", expectedConfigs[i].Type, config.Type)
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
	value1 := "1"
	value2 := "true"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2, Type: "boolean"},
	}

	expectedContent := "com.apple.dock autohide -boolean 1\ncom.apple.finder ShowPathbar -boolean true\n"
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

func TestGenerateConfigFileContent_NilValue(t *testing.T) {
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: nil, Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: nil, Type: "string"},
	}

	// The expected content should include the type information even for nil values
	expectedContent := "com.apple.dock autohide -boolean\ncom.apple.finder ShowPathbar -string\n"
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
	value1 := "1"
	value2 := "true"

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2, Type: "boolean"},
	}

	err := writeConfigFile(mockFS, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide -boolean 1\ncom.apple.finder ShowPathbar -boolean true\n"
	if mockFS.writeFileContent != expectedContent {
		t.Errorf("Expected writeFileContent %q, got %q", expectedContent, mockFS.writeFileContent)
	}
}

func TestWriteConfigFile_Error(t *testing.T) {
	mockFS := &MockFileSystem{
		homeDir:      "/mock/home",
		writeFileErr: errors.New("write error"),
	}

	value1 := "1"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
	}

	err := writeConfigFile(mockFS, configs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.writeFileErr) {
		t.Errorf("Expected error %v, got %v", mockFS.writeFileErr, err)
	}
}
