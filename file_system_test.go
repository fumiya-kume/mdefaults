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

	err := createConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func TestSetupConfigFile_DoesNotCreateFileIfExists(t *testing.T) {
	fs := &MockFileSystem{
		homeDir:   "/mock/home",
		statError: nil,
		createErr: nil,
	}

	err := createConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func TestSetupConfigFile_HandleUserHomeDirError(t *testing.T) {
	fs := &MockFileSystem{
		homeDir:   "",
		statError: nil,
		createErr: nil,
	}

	err := createConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
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

// TestFileSystemCreateConfigFileIfMissingError tests error handling when creating a config file fails
func TestFileSystemCreateConfigFileIfMissingError(t *testing.T) {
	fs := &MockFileSystem{
		homeDir:   "/mock/home",
		statError: os.ErrNotExist,
		createErr: errors.New("create error"),
	}

	err := createConfigFileIfMissing(fs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, fs.createErr) {
		t.Errorf("Expected error %v, got %v", fs.createErr, err)
	}
}

// TestFileSystemImpl tests the fileSystem implementation
func TestFileSystemImpl(t *testing.T) {
	// Skip these tests in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping file system implementation tests in CI environment")
	}

	fs := &fileSystem{}

	// Test UserHomeDir
	t.Run("UserHomeDir", func(t *testing.T) {
		homeDir, err := fs.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir failed: %v", err)
		}
		if homeDir == "" {
			t.Error("Expected non-empty home directory")
		}
	})

	// Test Stat
	t.Run("Stat", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "mdefaults-test")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Test Stat on the temporary file
		fileInfo, err := fs.Stat(tempFile.Name())
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if fileInfo.Name() == "" {
			t.Error("Expected non-empty file name")
		}
	})

	// Test Create, WriteFile, and ReadFile
	t.Run("CreateWriteRead", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "mdefaults-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Test Create
		tempFilePath := tempDir + "/test.txt"
		file, err := fs.Create(tempFilePath)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		file.Close()

		// Test WriteFile
		testContent := "test content"
		err = fs.WriteFile(tempFilePath, testContent)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		// Test ReadFile
		content, err := fs.ReadFile(tempFilePath)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if content != testContent {
			t.Errorf("Expected content %q, got %q", testContent, content)
		}
	})
}
