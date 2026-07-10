package filesystem_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
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
	if fs.ChmodMode != 0o600 {
		t.Fatalf("Expected private permissions, got %04o", fs.ChmodMode)
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
	if fs.ChmodMode != 0o600 {
		t.Fatalf("Expected existing file permissions to be repaired, got %04o", fs.ChmodMode)
	}
}

func TestSetupConfigFile_PropagatesChmodError(t *testing.T) {
	chmodErr := errors.New("chmod error")
	fs := &filesystem.MockFileSystem{ChmodErr: chmodErr}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if !errors.Is(err, chmodErr) {
		t.Fatalf("Expected chmod error %v, got %v", chmodErr, err)
	}
}

func TestSetupConfigFile_PropagatesStatError(t *testing.T) {
	statErr := errors.New("stat error")
	fs := &filesystem.MockFileSystem{
		StatError: statErr,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if !errors.Is(err, statErr) {
		t.Fatalf("Expected stat error %v, got %v", statErr, err)
	}
}

func TestSetupConfigFile_PropagatesCreateError(t *testing.T) {
	createErr := errors.New("create error")
	fs := &filesystem.MockFileSystem{
		StatError: os.ErrNotExist,
		CreateErr: createErr,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if !errors.Is(err, createErr) {
		t.Fatalf("Expected create error %v, got %v", createErr, err)
	}
}

func TestSetupConfigFile_HandlesConcurrentCreate(t *testing.T) {
	fs := &filesystem.MockFileSystem{
		StatError: os.ErrNotExist,
		CreateErr: os.ErrExist,
	}

	err := filesystem.CreateConfigFileIfMissing(fs)
	if err != nil {
		t.Fatalf("Expected concurrent create to succeed, got %v", err)
	}
	if fs.ChmodMode != 0o600 {
		t.Fatalf("Expected concurrent file permissions to be repaired, got %04o", fs.ChmodMode)
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

func TestOSFileSystemFileLifecycle(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	fs := filesystem.NewOSFileSystem()

	actualHome, err := fs.UserHomeDir()
	if err != nil {
		t.Fatalf("expected home directory, got %v", err)
	}
	if actualHome != homeDir {
		t.Fatalf("expected home %q, got %q", homeDir, actualHome)
	}

	path := filepath.Join(homeDir, "settings")
	if _, err := fs.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected missing file, got %v", err)
	}
	if err := fs.Create(path); err != nil {
		t.Fatalf("expected file creation to succeed, got %v", err)
	}
	if _, err := fs.Stat(path); err != nil {
		t.Fatalf("expected created file to exist, got %v", err)
	}
	if err := fs.Create(path); !os.IsExist(err) {
		t.Fatalf("expected exclusive create error, got %v", err)
	}

	const content = "configuration\n"
	if err := fs.WriteFile(path, content); err != nil {
		t.Fatalf("expected write to succeed, got %v", err)
	}
	actualContent, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("expected read to succeed, got %v", err)
	}
	if actualContent != content {
		t.Fatalf("expected content %q, got %q", content, actualContent)
	}

	if _, err := fs.ReadFile(filepath.Join(homeDir, "missing")); !os.IsNotExist(err) {
		t.Fatalf("expected missing read error, got %v", err)
	}
	if err := fs.WriteFile(homeDir, "cannot replace a directory"); err == nil {
		t.Fatal("expected writing to a directory to fail")
	}
}

func TestCreateConfigFileIfMissingWithOSFileSystemPreservesExistingContent(t *testing.T) {
	originalPath := config.ConfigFilePath
	config.ConfigFilePath = filepath.Join(t.TempDir(), ".mdefaults")
	t.Cleanup(func() {
		config.ConfigFilePath = originalPath
	})

	fs := filesystem.NewOSFileSystem()
	if err := filesystem.CreateConfigFileIfMissing(fs); err != nil {
		t.Fatalf("expected missing config creation to succeed, got %v", err)
	}
	info, err := os.Stat(config.ConfigFilePath)
	if err != nil {
		t.Fatalf("failed to stat created config file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected created config mode 0600, got %04o", info.Mode().Perm())
	}
	const sentinel = "do not truncate\n"
	if err := fs.WriteFile(config.ConfigFilePath, sentinel); err != nil {
		t.Fatalf("failed to seed config file: %v", err)
	}
	if err := os.Chmod(config.ConfigFilePath, 0o644); err != nil {
		t.Fatalf("set legacy config mode: %v", err)
	}

	if err := filesystem.CreateConfigFileIfMissing(fs); err != nil {
		t.Fatalf("expected existing config check to succeed, got %v", err)
	}
	actual, err := fs.ReadFile(config.ConfigFilePath)
	if err != nil {
		t.Fatalf("failed to read preserved config file: %v", err)
	}
	if actual != sentinel {
		t.Fatalf("expected existing content %q, got %q", sentinel, actual)
	}
	info, err = os.Stat(config.ConfigFilePath)
	if err != nil {
		t.Fatalf("failed to stat migrated config file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected migrated config mode 0600, got %04o", info.Mode().Perm())
	}
}

func TestMockFileSystemAccessors(t *testing.T) {
	mockFS := &filesystem.MockFileSystem{
		HomeDir:           "/mock/home",
		ConfigFileContent: "content",
	}

	home, err := mockFS.UserHomeDir()
	if err != nil {
		t.Fatalf("expected no home error, got %v", err)
	}
	if home != mockFS.HomeDir {
		t.Fatalf("expected home %q, got %q", mockFS.HomeDir, home)
	}
	if _, err := mockFS.Stat("ignored"); !os.IsNotExist(err) {
		t.Fatalf("expected mock missing status for populated content, got %v", err)
	}
	if err := mockFS.WriteFile("ignored", "updated"); err != nil {
		t.Fatalf("expected mock write to succeed, got %v", err)
	}
	if mockFS.WriteFileContent != "updated" {
		t.Fatalf("expected recorded write, got %q", mockFS.WriteFileContent)
	}
	if err := mockFS.Chmod("settings", 0o600); err != nil {
		t.Fatalf("expected mock chmod to succeed, got %v", err)
	}
	if mockFS.ChmodName != "settings" || mockFS.ChmodMode != 0o600 {
		t.Fatalf("unexpected recorded chmod: %q %04o", mockFS.ChmodName, mockFS.ChmodMode)
	}
}
