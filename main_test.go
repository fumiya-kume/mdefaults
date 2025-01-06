package main

import (
	"errors"
	"fmt"
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

	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadError: errors.New("read error")},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}
