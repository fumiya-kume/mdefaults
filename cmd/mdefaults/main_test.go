package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/filesystem"
)

type fakeFileSystem struct {
	statErr     error
	createErr   error
	chmodErr    error
	readErr     error
	writeErr    error
	content     string
	written     string
	createCalls int
	chmodCalls  int
	readCalls   int
	writeCalls  int
}

type failingWriter struct {
	err error
}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, f.err
}

type failOnCallWriter struct {
	failAt int
	calls  int
	err    error
}

func (f *failOnCallWriter) Write(value []byte) (int, error) {
	f.calls++
	if f.calls == f.failAt {
		return 0, f.err
	}
	return len(value), nil
}

func (f *fakeFileSystem) UserHomeDir() (string, error) {
	return "/fake/home", nil
}

func (f *fakeFileSystem) Stat(string) (os.FileInfo, error) {
	return nil, f.statErr
}

func (f *fakeFileSystem) Create(string) error {
	f.createCalls++
	return f.createErr
}

func (f *fakeFileSystem) Chmod(string, os.FileMode) error {
	f.chmodCalls++
	return f.chmodErr
}

func (f *fakeFileSystem) ReadFile(string) (string, error) {
	f.readCalls++
	return f.content, f.readErr
}

func (f *fakeFileSystem) WriteFile(_ string, content string) error {
	f.writeCalls++
	f.written = content
	return f.writeErr
}

type appTestHarness struct {
	fs              *fakeFileSystem
	deps            appDependencies
	stdout          bytes.Buffer
	logs            bytes.Buffer
	successMessages []string
	errorMessages   []string
	warnings        []string
	fsCreations     int
}

func newAppTestHarness() *appTestHarness {
	harness := &appTestHarness{fs: &fakeFileSystem{}}
	harness.deps = appDependencies{
		newFileSystem: func() filesystem.FileSystem {
			harness.fsCreations++
			return harness.fs
		},
		pull: func(configs []config.Config) ([]config.Config, error) {
			return configs, nil
		},
		push:   func([]config.Config) error { return nil },
		stdin:  strings.NewReader(""),
		stdout: &harness.stdout,
		logger: log.New(&harness.logs, "", 0),
		goos:   "darwin",
		printSuccess: func(message string) {
			harness.successMessages = append(harness.successMessages, message)
		},
		printError: func(message string) {
			harness.errorMessages = append(harness.errorMessages, message)
		},
		printWarning: func(message string) {
			harness.warnings = append(harness.warnings, message)
		},
	}
	return harness
}

func TestDefaultAppDependencies(t *testing.T) {
	deps := defaultAppDependencies()
	if deps.newFileSystem == nil || deps.pull == nil || deps.push == nil || deps.stdin == nil || deps.stdout == nil || deps.logger == nil {
		t.Fatal("default dependencies are incomplete")
	}
	if deps.printSuccess == nil || deps.printError == nil || deps.printWarning == nil || deps.goos == "" {
		t.Fatal("default output dependencies are incomplete")
	}
}

func TestSetupLogging(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "mdefaults.log")
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatalf("create existing log: %v", err)
	}
	if err := os.Chmod(logPath, 0o644); err != nil {
		t.Fatalf("set existing log mode: %v", err)
	}
	logger := log.New(io.Discard, "", 0)

	logFile, err := setupLogging(logPath, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	logger.Print("covered")
	if err := logFile.Close(); err != nil {
		t.Fatalf("failed to close log: %v", err)
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if !bytes.Contains(content, []byte("covered")) {
		t.Fatalf("expected log content, got %q", content)
	}
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat log: %v", err)
	}
	if actual := info.Mode().Perm(); actual != 0o600 {
		t.Fatalf("expected private log permissions, got %o", actual)
	}
}

func TestSetupLoggingFailure(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "missing", "mdefaults.log")
	if _, err := setupLogging(logPath, log.New(io.Discard, "", 0)); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRunMetaCommands(t *testing.T) {
	originalVersion := version
	originalArchitecture := architecture
	version = "1.2.3"
	architecture = "arm64"
	t.Cleanup(func() {
		version = originalVersion
		architecture = originalArchitecture
	})

	testCases := []struct {
		name        string
		args        []string
		goos        string
		exitCode    int
		output      string
		logText     string
		errorText   string
		wantFSCalls int
	}{
		{name: "no command", exitCode: 0, output: "Usage:"},
		{name: "help", args: []string{"--help"}, exitCode: 0, output: "config  - Print"},
		{name: "version", args: []string{"--version"}, exitCode: 0, output: "Version: 1.2.3"},
		{name: "linux warning", args: []string{"-v"}, goos: "linux", exitCode: 0, output: "Work In Progress"},
		{name: "debug verbose", args: []string{"--verbose", "debug"}, exitCode: 0, logText: "Verbose mode enabled"},
		{name: "unknown command", args: []string{"unknown"}, exitCode: 1, logText: "Unknown command", errorText: "Unknown command"},
		{name: "invalid flag", args: []string{"--invalid"}, exitCode: 1, logText: "Failed to parse", errorText: "Invalid command line arguments"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			harness := newAppTestHarness()
			if tc.goos != "" {
				harness.deps.goos = tc.goos
			}

			actualExitCode := run(tc.args, harness.deps)
			if actualExitCode != tc.exitCode {
				t.Fatalf("expected exit code %d, got %d", tc.exitCode, actualExitCode)
			}
			if tc.output != "" && !strings.Contains(harness.stdout.String(), tc.output) {
				t.Fatalf("expected output %q, got %q", tc.output, harness.stdout.String())
			}
			if tc.logText != "" && !strings.Contains(harness.logs.String(), tc.logText) {
				t.Fatalf("expected log %q, got %q", tc.logText, harness.logs.String())
			}
			if tc.errorText != "" && !containsString(harness.errorMessages, tc.errorText) {
				t.Fatalf("expected error %q, got %v", tc.errorText, harness.errorMessages)
			}
			if harness.fsCreations != tc.wantFSCalls {
				t.Fatalf("expected %d filesystem creations, got %d", tc.wantFSCalls, harness.fsCreations)
			}
		})
	}
}

func TestRunConfig(t *testing.T) {
	harness := newAppTestHarness()
	harness.fs.content = "com.example.app target\ncom.example.app value \"hello world\" string\n"

	exitCode := run([]string{"config"}, harness.deps)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	output := harness.stdout.String()
	if !strings.Contains(output, "- com.example.app target\n") || !strings.Contains(output, "- com.example.app value hello world string\n") {
		t.Fatalf("unexpected config output: %q", output)
	}
}

func TestRunReportsOutputFailure(t *testing.T) {
	writeErr := errors.New("output failed")
	harness := newAppTestHarness()
	harness.deps.stdout = failingWriter{err: writeErr}

	if exitCode := run([]string{"--version"}, harness.deps); exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !containsString(harness.errorMessages, "Failed to write output") {
		t.Fatalf("expected output error, got %v", harness.errorMessages)
	}
	if !strings.Contains(harness.logs.String(), writeErr.Error()) {
		t.Fatalf("expected output error log, got %q", harness.logs.String())
	}
}

func TestRunReportsOutputFailureAtEveryStage(t *testing.T) {
	writeErr := errors.New("output failed")
	testCases := []struct {
		name   string
		args   []string
		input  string
		goos   string
		failAt int
		config string
	}{
		{name: "platform warning", args: []string{"--version"}, goos: "linux", failAt: 1},
		{name: "help", args: []string{"--help"}, failAt: 1},
		{name: "usage", failAt: 1},
		{name: "config output", args: []string{"config"}, failAt: 1, config: "com.example.app key\n"},
		{name: "push config output", args: []string{"push"}, failAt: 1, config: "com.example.app key \"value\" string\n"},
		{name: "pull current heading", args: []string{"pull", "-y"}, failAt: 1, config: "com.example.app key\n"},
		{name: "pull current config", args: []string{"pull", "-y"}, failAt: 2, config: "com.example.app key\n"},
		{name: "pull macOS heading", args: []string{"pull", "-y"}, failAt: 3, config: "com.example.app key\n"},
		{name: "pull updated config", args: []string{"pull", "-y"}, failAt: 4, config: "com.example.app key\n"},
		{name: "pull prompt", args: []string{"pull"}, input: "yes\n", failAt: 5, config: "com.example.app key\n"},
		{name: "pull EOF message", args: []string{"pull"}, failAt: 6, config: "com.example.app key\n"},
		{name: "pull cancellation message", args: []string{"pull"}, input: "no\n", failAt: 6, config: "com.example.app key\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			harness := newAppTestHarness()
			harness.fs.content = tc.config
			harness.deps.stdin = strings.NewReader(tc.input)
			if tc.goos != "" {
				harness.deps.goos = tc.goos
			}
			writer := &failOnCallWriter{failAt: tc.failAt, err: writeErr}
			harness.deps.stdout = writer

			if exitCode := run(tc.args, harness.deps); exitCode != 1 {
				t.Fatalf("expected exit code 1, got %d", exitCode)
			}
			if !containsString(harness.errorMessages, "Failed to write output") {
				t.Fatalf("expected output error, got %v", harness.errorMessages)
			}
			if writer.calls != tc.failAt {
				t.Fatalf("expected failure on write %d, got %d calls", tc.failAt, writer.calls)
			}
		})
	}
}

func TestRunConfigSetupErrors(t *testing.T) {
	createErr := errors.New("create failed")
	chmodErr := errors.New("chmod failed")
	readErr := errors.New("read failed")
	testCases := []struct {
		name      string
		configure func(*fakeFileSystem)
		expected  string
	}{
		{
			name: "create",
			configure: func(fs *fakeFileSystem) {
				fs.statErr = os.ErrNotExist
				fs.createErr = createErr
			},
			expected: "Failed to create config file",
		},
		{
			name: "chmod",
			configure: func(fs *fakeFileSystem) {
				fs.chmodErr = chmodErr
			},
			expected: "Failed to create config file",
		},
		{
			name: "read",
			configure: func(fs *fakeFileSystem) {
				fs.readErr = readErr
			},
			expected: "Failed to read config file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			harness := newAppTestHarness()
			tc.configure(harness.fs)
			if exitCode := run([]string{"config"}, harness.deps); exitCode != 1 {
				t.Fatalf("expected exit code 1, got %d", exitCode)
			}
			if !containsString(harness.errorMessages, tc.expected) {
				t.Fatalf("expected error %q, got %v", tc.expected, harness.errorMessages)
			}
		})
	}
}

func TestRunPull(t *testing.T) {
	pulledValue := "1"
	pulledConfigs := []config.Config{{Domain: "com.example.app", Key: "enabled", Value: &pulledValue, Type: "boolean"}}
	pullErr := errors.New("pull failed")
	writeErr := errors.New("write failed")

	testCases := []struct {
		name          string
		args          []string
		input         string
		pullError     error
		writeError    error
		exitCode      int
		writeCalls    int
		success       bool
		warning       bool
		output        string
		expectedError string
	}{
		{name: "skip confirmation", args: []string{"pull", "-y"}, exitCode: 0, writeCalls: 1, success: true},
		{name: "confirm yes", args: []string{"pull"}, input: "yes\n", exitCode: 0, writeCalls: 1, success: true, warning: true},
		{name: "confirm no", args: []string{"pull"}, input: "no\n", exitCode: 0, warning: true, output: "Operation cancelled."},
		{name: "confirmation eof", args: []string{"pull"}, exitCode: 1, warning: true, output: "Failed to read input"},
		{name: "pull failure", args: []string{"pull", "-y"}, pullError: pullErr, exitCode: 1, expectedError: "Failed to pull configurations"},
		{name: "write failure", args: []string{"pull", "-y"}, writeError: writeErr, exitCode: 1, writeCalls: 1, expectedError: "Failed to write config file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			harness := newAppTestHarness()
			harness.fs.content = "com.example.app enabled\n"
			harness.fs.writeErr = tc.writeError
			harness.deps.stdin = strings.NewReader(tc.input)
			harness.deps.pull = func([]config.Config) ([]config.Config, error) {
				if tc.pullError != nil {
					return nil, tc.pullError
				}
				return pulledConfigs, nil
			}

			exitCode := run(tc.args, harness.deps)
			if exitCode != tc.exitCode {
				t.Fatalf("expected exit code %d, got %d", tc.exitCode, exitCode)
			}
			if harness.fs.writeCalls != tc.writeCalls {
				t.Fatalf("expected %d writes, got %d", tc.writeCalls, harness.fs.writeCalls)
			}
			if tc.success != (len(harness.successMessages) == 1) {
				t.Fatalf("unexpected success messages: %v", harness.successMessages)
			}
			if tc.warning != (len(harness.warnings) == 1) {
				t.Fatalf("unexpected warnings: %v", harness.warnings)
			}
			if tc.output != "" && !strings.Contains(harness.stdout.String(), tc.output) {
				t.Fatalf("expected output %q, got %q", tc.output, harness.stdout.String())
			}
			if tc.expectedError != "" && !containsString(harness.errorMessages, tc.expectedError) {
				t.Fatalf("expected error %q, got %v", tc.expectedError, harness.errorMessages)
			}
		})
	}
}

func TestRunPush(t *testing.T) {
	pushErr := errors.New("push failed")
	testCases := []struct {
		name          string
		pushError     error
		exitCode      int
		success       bool
		expectedError string
	}{
		{name: "success", exitCode: 0, success: true},
		{name: "failure", pushError: pushErr, exitCode: 1, expectedError: "Failed to push configurations"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			harness := newAppTestHarness()
			harness.fs.content = "com.example.app value \"hello\" string\n"
			pushCalls := 0
			harness.deps.push = func(configs []config.Config) error {
				pushCalls++
				if len(configs) != 1 {
					t.Fatalf("expected one config, got %d", len(configs))
				}
				return tc.pushError
			}

			exitCode := run([]string{"push"}, harness.deps)
			if exitCode != tc.exitCode {
				t.Fatalf("expected exit code %d, got %d", tc.exitCode, exitCode)
			}
			if pushCalls != 1 {
				t.Fatalf("expected one push call, got %d", pushCalls)
			}
			if tc.success != (len(harness.successMessages) == 1) {
				t.Fatalf("unexpected success messages: %v", harness.successMessages)
			}
			if tc.expectedError != "" && !containsString(harness.errorMessages, tc.expectedError) {
				t.Fatalf("expected error %q, got %v", tc.expectedError, harness.errorMessages)
			}
		})
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
