package main

import (
	"errors"
	"testing"
)

func TestPull_Success(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1"},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" {
		t.Errorf("Expected domain 'com.apple.dock', got %s", updatedConfigs[0].Domain)
	}
	if updatedConfigs[0].Key != "autohide" {
		t.Errorf("Expected key 'autohide', got %s", updatedConfigs[0].Key)
	}
	if updatedConfigs[0].Value != "1" {
		t.Errorf("Expected value '1', got %s", updatedConfigs[0].Value)
	}
}

func TestPull_ReadError(t *testing.T) {

	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := pullImpl(defaults)
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MultipleConfigs(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1"},
		&MockDefaultsCommand{domain: "com.apple.finder", key: "ShowPathbar", ReadResult: "true"},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || updatedConfigs[0].Value != "1" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
	if updatedConfigs[1].Domain != "com.apple.finder" || updatedConfigs[1].Key != "ShowPathbar" || updatedConfigs[1].Value != "true" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[1])
	}
}

func TestPull_EmptyConfigs(t *testing.T) {
	defaults := []DefaultsCommand{}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MixedResults(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1"},
		&MockDefaultsCommand{domain: "com.apple.finder", key: "ShowPathbar", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := pullImpl(defaults)
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || updatedConfigs[0].Value != "1" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
}

func TestMain_NoCommand_ShowsHelp(t *testing.T) {
	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	exitCode := 0
	osExit = func(code int) { exitCode = code }

	// Simulate calling the program without any subcommands
	os.Args = []string{"mdefaults"}

	// Capture the output
	output := captureOutput(func() {
		run()
	})

	expectedOutput := "Usage: mdefaults [command]\nCommands:\n  pull    - Retrieve and update configuration values.\n  push    - Write configuration values.\nHey, let's call with pull or push.\n"

	if output != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// Helper function to capture output
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = os.Stdout

	return buf.String()
}
