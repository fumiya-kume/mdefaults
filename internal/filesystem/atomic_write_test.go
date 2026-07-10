package filesystem

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type faultAtomicFile struct {
	atomicFile
	writeErr  error
	syncErr   error
	closeErr  error
	zeroWrite bool
	maxWrite  int
}

func (f *faultAtomicFile) Write(content []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	if f.zeroWrite {
		return 0, nil
	}
	if f.maxWrite > 0 && len(content) > f.maxWrite {
		content = content[:f.maxWrite]
	}
	return f.atomicFile.Write(content)
}

func (f *faultAtomicFile) Sync() error {
	if f.syncErr != nil {
		return f.syncErr
	}
	return f.atomicFile.Sync()
}

func (f *faultAtomicFile) Close() error {
	underlyingErr := f.atomicFile.Close()
	if underlyingErr != nil {
		return underlyingErr
	}
	return f.closeErr
}

func TestOSFileSystemWriteFileAtomicallyReplacesContent(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "settings")
	if err := os.WriteFile(target, []byte("old content\n"), 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := os.Chmod(target, 0o644); err != nil {
		t.Fatalf("set permissive target mode: %v", err)
	}

	operations := defaultAtomicWriteOperations()
	defaultCreateTemp := operations.createTemp
	var actualDirectory string
	var actualPattern string
	operations.createTemp = func(dir string, pattern string) (atomicFile, error) {
		actualDirectory = dir
		actualPattern = pattern
		file, err := defaultCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return &faultAtomicFile{atomicFile: file, maxWrite: 3}, nil
	}
	fs := &OSFileSystem{writeOperations: operations}

	const content = "new content with unicode: 日本語\nand a second line\n"
	if err := fs.WriteFile(target, content); err != nil {
		t.Fatalf("write target: %v", err)
	}

	actualContent, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(actualContent) != content {
		t.Fatalf("expected content %q, got %q", content, actualContent)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %04o", info.Mode().Perm())
	}
	if actualDirectory != directory {
		t.Fatalf("expected temporary file in %q, got %q", directory, actualDirectory)
	}
	if actualPattern != ".settings.tmp-*" {
		t.Fatalf("expected temporary pattern %q, got %q", ".settings.tmp-*", actualPattern)
	}
	assertNoAtomicTemporaryFiles(t, target)
}

func TestOSFileSystemWriteFilePreservesTargetAndCleansUpOnFailure(t *testing.T) {
	sentinelErr := errors.New("injected failure")
	tests := []struct {
		name         string
		configure    func(*atomicWriteOperations)
		expectedErr  error
		expectRemove bool
	}{
		{
			name: "create temporary file",
			configure: func(operations *atomicWriteOperations) {
				operations.createTemp = func(string, string) (atomicFile, error) {
					return nil, sentinelErr
				}
			},
			expectedErr: sentinelErr,
		},
		{
			name: "write",
			configure: func(operations *atomicWriteOperations) {
				wrapCreatedFile(operations, func(file atomicFile) atomicFile {
					return &faultAtomicFile{atomicFile: file, writeErr: sentinelErr}
				})
			},
			expectedErr:  sentinelErr,
			expectRemove: true,
		},
		{
			name: "zero length write",
			configure: func(operations *atomicWriteOperations) {
				wrapCreatedFile(operations, func(file atomicFile) atomicFile {
					return &faultAtomicFile{atomicFile: file, zeroWrite: true}
				})
			},
			expectedErr:  io.ErrShortWrite,
			expectRemove: true,
		},
		{
			name: "sync",
			configure: func(operations *atomicWriteOperations) {
				wrapCreatedFile(operations, func(file atomicFile) atomicFile {
					return &faultAtomicFile{atomicFile: file, syncErr: sentinelErr}
				})
			},
			expectedErr:  sentinelErr,
			expectRemove: true,
		},
		{
			name: "close",
			configure: func(operations *atomicWriteOperations) {
				wrapCreatedFile(operations, func(file atomicFile) atomicFile {
					return &faultAtomicFile{atomicFile: file, closeErr: sentinelErr}
				})
			},
			expectedErr:  sentinelErr,
			expectRemove: true,
		},
		{
			name: "rename",
			configure: func(operations *atomicWriteOperations) {
				operations.rename = func(string, string) error {
					return sentinelErr
				}
			},
			expectedErr:  sentinelErr,
			expectRemove: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			target := filepath.Join(directory, "settings")
			const originalContent = "content that must survive\n"
			if err := os.WriteFile(target, []byte(originalContent), 0o600); err != nil {
				t.Fatalf("seed target: %v", err)
			}

			operations := defaultAtomicWriteOperations()
			removeCalls := 0
			defaultRemove := operations.remove
			operations.remove = func(path string) error {
				removeCalls++
				return defaultRemove(path)
			}
			test.configure(&operations)
			fs := &OSFileSystem{writeOperations: operations}

			err := fs.WriteFile(target, "replacement that must not appear\n")
			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("expected error %v, got %v", test.expectedErr, err)
			}
			if test.expectRemove && removeCalls != 1 {
				t.Fatalf("expected one temporary-file removal, got %d", removeCalls)
			}
			if !test.expectRemove && removeCalls != 0 {
				t.Fatalf("expected no temporary-file removal, got %d", removeCalls)
			}

			actualContent, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatalf("read preserved target: %v", readErr)
			}
			if string(actualContent) != originalContent {
				t.Fatalf("expected preserved content %q, got %q", originalContent, actualContent)
			}
			assertNoAtomicTemporaryFiles(t, target)
		})
	}
}

func TestOSFileSystemWriteFileReportsCleanupFailures(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "settings")
	if err := os.WriteFile(target, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("seed target: %v", err)
	}

	writeErr := errors.New("write failure")
	closeErr := errors.New("cleanup close failure")
	removeErr := errors.New("cleanup remove failure")
	operations := defaultAtomicWriteOperations()
	wrapCreatedFile(&operations, func(file atomicFile) atomicFile {
		return &faultAtomicFile{atomicFile: file, writeErr: writeErr, closeErr: closeErr}
	})
	defaultRemove := operations.remove
	operations.remove = func(path string) error {
		if err := defaultRemove(path); err != nil {
			return err
		}
		return removeErr
	}
	fs := &OSFileSystem{writeOperations: operations}

	err := fs.WriteFile(target, "new\n")
	for _, expectedErr := range []error{writeErr, closeErr, removeErr} {
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected joined error to contain %v, got %v", expectedErr, err)
		}
	}
	assertNoAtomicTemporaryFiles(t, target)
}

func TestZeroValueOSFileSystemWritesAtomically(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings")
	var fs OSFileSystem

	if err := fs.WriteFile(target, "content\n"); err != nil {
		t.Fatalf("write with zero-value file system: %v", err)
	}
	actual, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(actual) != "content\n" {
		t.Fatalf("expected content %q, got %q", "content\n", actual)
	}
	assertNoAtomicTemporaryFiles(t, target)
}

func wrapCreatedFile(operations *atomicWriteOperations, wrap func(atomicFile) atomicFile) {
	defaultCreateTemp := operations.createTemp
	operations.createTemp = func(dir string, pattern string) (atomicFile, error) {
		file, err := defaultCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return wrap(file), nil
	}
}

func assertNoAtomicTemporaryFiles(t *testing.T, target string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatalf("read target directory: %v", err)
	}
	prefix := "." + filepath.Base(target) + ".tmp-"
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			t.Fatalf("unexpected temporary file left behind: %s", entry.Name())
		}
	}
}
