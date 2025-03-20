package filesystem_test

import (
	"errors"
	"os"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/filesystem"
)

func TestSetupConfigFile_CreatesFileIfNotExist(t *testing.T) {
	fs := &filesystem.MockFileSystem{
		HomeDir:   "/mock/home",
		StatError: os.ErrNotExist,
		CreateErr: nil,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func TestSetupConfigFile_DoesNotCreateFileIfExists(t *testing.T) {
	fs := &filesystem.MockFileSystem{
		HomeDir:   "/mock/home",
		StatError: nil,
		CreateErr: nil,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func TestSetupConfigFile_HandleUserHomeDirError(t *testing.T) {
	fs := &filesystem.MockFileSystem{
		HomeDir:   "",
		StatError: nil,
		CreateErr: nil,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

func TestReadConfigFileString_Success(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{
		ConfigFileContent: "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n",
	}

	content, err := filesystem.ReadConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Empty(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{
		ConfigFileContent: "",
	}

	content, err := filesystem.ReadConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := ""
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileString_Error(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{
		StatError: errors.New("read error"),
	}

	_, err := filesystem.ReadConfigFileString(mockFS)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.StatError) {
		t.Errorf("Expected error %v, got %v", mockFS.StatError, err)
	}
}

func TestReadConfigFileString_MalformedContent(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{
		ConfigFileContent: "com.apple.dock autohide\nmalformed line without key\ncom.apple.finder ShowPathbar true\n",
	}

	content, err := filesystem.ReadConfigFileString(mockFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide\nmalformed line without key\ncom.apple.finder ShowPathbar true\n"
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}
