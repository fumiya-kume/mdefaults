package config

import (
	"testing"
)

func TestReadConfigFileWithTypes(t *testing.T) {
	mockFS := &MockFileSystem{
		ConfigFileContent: "com.apple.dock autohide 1 boolean\ncom.apple.finder ShowPathbar true boolean\ncom.apple.trackpad ClickThreshold 2 integer",
	}

	configs, err := ReadConfigFile(mockFS)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: stringPtr("1"), Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: stringPtr("true"), Type: "boolean"},
		{Domain: "com.apple.trackpad", Key: "ClickThreshold", Value: stringPtr("2"), Type: "integer"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Errorf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, expected := range expectedConfigs {
		if i >= len(configs) {
			t.Errorf("Missing config at index %d", i)
			continue
		}
		actual := configs[i]
		if actual.Domain != expected.Domain || actual.Key != expected.Key || 
		   *actual.Value != *expected.Value || actual.Type != expected.Type {
			t.Errorf("Config mismatch at index %d: expected %+v, got %+v", i, expected, actual)
		}
	}
}

func TestReadConfigFileBackwardCompatibility(t *testing.T) {
	mockFS := &MockFileSystem{
		ConfigFileContent: "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true",
	}

	configs, err := ReadConfigFile(mockFS)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: stringPtr("1"), Type: "string"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: stringPtr("true"), Type: "string"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Errorf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, expected := range expectedConfigs {
		if i >= len(configs) {
			t.Errorf("Missing config at index %d", i)
			continue
		}
		actual := configs[i]
		if actual.Domain != expected.Domain || actual.Key != expected.Key || 
		   *actual.Value != *expected.Value || actual.Type != expected.Type {
			t.Errorf("Config mismatch at index %d: expected %+v, got %+v", i, expected, actual)
		}
	}
}

func TestGenerateConfigFileContentWithTypes(t *testing.T) {
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: stringPtr("1"), Type: "boolean"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: stringPtr("true"), Type: "boolean"},
		{Domain: "com.apple.trackpad", Key: "ClickThreshold", Value: stringPtr("2"), Type: "integer"},
	}

	expectedContent := "com.apple.dock autohide 1 boolean\ncom.apple.finder ShowPathbar true boolean\ncom.apple.trackpad ClickThreshold 2 integer\n"
	actualContent := GenerateConfigFileContent(configs)

	if actualContent != expectedContent {
		t.Errorf("Expected content:\n%s\nGot content:\n%s", expectedContent, actualContent)
	}
}

func TestGenerateConfigFileContentWithEmptyType(t *testing.T) {
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: stringPtr("1"), Type: ""},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: stringPtr("true"), Type: "boolean"},
	}

	expectedContent := "com.apple.dock autohide 1 string\ncom.apple.finder ShowPathbar true boolean\n"
	actualContent := GenerateConfigFileContent(configs)

	if actualContent != expectedContent {
		t.Errorf("Expected content:\n%s\nGot content:\n%s", expectedContent, actualContent)
	}
}

func stringPtr(s string) *string {
	return &s
}
