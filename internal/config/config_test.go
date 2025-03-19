package config

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfigFile_Success(t *testing.T) {
	fs := &MockFileSystem{HomeDir: "/mock/home", StatError: nil, CreateErr: nil, ConfigFileContent: "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	value1 := "1"
	value2 := "true"
	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
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
	}
}

func TestReadConfigFile_Error(t *testing.T) {
	fs := &MockFileSystem{StatError: errors.New("read error")}

	_, err := ReadConfigFile(fs)
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
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestGenerateConfigFileContent_Empty(t *testing.T) {
	configs := []Config{}

	expectedContent := ""
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestWriteConfigFile_Success(t *testing.T) {
	mockFS := &MockFileSystem{
		HomeDir:          "/mock/home",
		WriteFileContent: "",
		WriteFileErr:     nil,
	}
	value1 := "1"
	value2 := "true"

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
	}

	err := WriteConfigFile(mockFS, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	if mockFS.WriteFileContent != expectedContent {
		t.Errorf("Expected WriteFileContent %q, got %q", expectedContent, mockFS.WriteFileContent)
	}
}

func TestWriteConfigFile_Error(t *testing.T) {
	mockFS := &MockFileSystem{
		HomeDir:      "/mock/home",
		WriteFileErr: errors.New("write error"),
	}

	value1 := "1"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
	}

	err := WriteConfigFile(mockFS, configs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.WriteFileErr) {
		t.Errorf("Expected error %v, got %v", mockFS.WriteFileErr, err)
	}
}
