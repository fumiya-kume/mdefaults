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

func TestReadConfigFile_VariousInputTypes(t *testing.T) {
	// Test with various input types
	configContent := `com.apple.dock autohide 1
com.apple.finder ShowPathbar true
com.example.app floatValue 3.14
com.example.app negativeValue -42
com.example.app zeroValue 0
com.example.app specialChars !@#$%^&*()
com.example.app emptyValue 
com.example.app longValue 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890
`
	fs := &MockFileSystem{HomeDir: "/mock/home", ConfigFileContent: configContent}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Expected values
	value1 := "1"
	value2 := "true"
	value3 := "3.14"
	value4 := "-42"
	value5 := "0"
	value6 := "!@#$%^&*()"
	value7 := ""
	value8 := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
		{Domain: "com.example.app", Key: "floatValue", Value: &value3},
		{Domain: "com.example.app", Key: "negativeValue", Value: &value4},
		{Domain: "com.example.app", Key: "zeroValue", Value: &value5},
		{Domain: "com.example.app", Key: "specialChars", Value: &value6},
		{Domain: "com.example.app", Key: "emptyValue", Value: &value7},
		{Domain: "com.example.app", Key: "longValue", Value: &value8},
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

func TestGenerateConfigFileContent_VariousInputTypes(t *testing.T) {
	// Test with various input types
	value1 := "3.14"
	value2 := "-42"
	value3 := "0"
	value4 := "!@#$%^&*()"
	value5 := ""
	value6 := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

	configs := []Config{
		{Domain: "com.example.app", Key: "floatValue", Value: &value1},
		{Domain: "com.example.app", Key: "negativeValue", Value: &value2},
		{Domain: "com.example.app", Key: "zeroValue", Value: &value3},
		{Domain: "com.example.app", Key: "specialChars", Value: &value4},
		{Domain: "com.example.app", Key: "emptyValue", Value: &value5},
		{Domain: "com.example.app", Key: "longValue", Value: &value6},
	}

	expectedContent := "com.example.app floatValue 3.14\ncom.example.app negativeValue -42\ncom.example.app zeroValue 0\ncom.example.app specialChars !@#$%^&*()\ncom.example.app emptyValue \ncom.example.app longValue 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890\n"
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}
