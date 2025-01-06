package main

import (
	"errors"
	"testing"
)

// func TestPull_Success(t *testing.T) {
// 	mockFS := &MockFileSystem{
// 		homeDir:           "/mock/home",
// 		configFileContent: "com.apple.dock autohide\n",
// 	}

// 	configs := []Config{
// 		{Domain: "com.apple.dock", Key: "autohide"},
// 	}

// 	mockDefaults := &MockDefaultsCommand{ReadResult: "true", ReadError: nil}

// 	pull(mockFS, configs, mockDefaults)

// 	expectedContent := "com.apple.dock autohide true\n"
// 	if mockFS.writeFileContent != expectedContent {
// 		t.Errorf("Expected writeFileContent %q, got %q", expectedContent, mockFS.writeFileContent)
// 	}
// }

func TestPull_ReadError(t *testing.T) {
	mockFS := &MockFileSystem{
		homeDir:           "/mock/home",
		configFileContent: "com.apple.dock autohide\n",
	}

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide"},
	}

	mockDefaults := &MockDefaultsCommand{ReadError: errors.New("read error")}

	pull(mockFS, configs, mockDefaults)

	if configs[0].Value != "" {
		t.Errorf("Expected value '', got %s", configs[0].Value)
	}
}
