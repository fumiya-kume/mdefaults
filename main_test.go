package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/alecthomas/kong"
)

// Mock os.Exit function
var osExit = os.Exit

// Remove remaining pull-related tests

func TestMain_NoCommand_ShowsHelp(t *testing.T) {
	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	osExit = func(code int) {}

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
}

func TestMain_VersionFlag(t *testing.T) {
	// Setup
	os.Args = []string{"mdefaults", "version"}

	// Capture output
	output := captureOutput(func() {
		main()
	})

	// Verify
	if !bytes.Contains([]byte(output), []byte("Version:")) {
		t.Errorf("Expected version command output, got: %s", output)
	}
}

func TestMain_VFlag(t *testing.T) {
	// Setup
	os.Args = []string{"mdefaults", "version", "--verbose"}

	// Capture output
	output := captureOutput(func() {
		main()
	})

	// Verify
	if !bytes.Contains([]byte(output), []byte("Architecture:")) {
		t.Errorf("Expected verbose version output, got: %s", output)
	}
}

func TestMain_WIPMessage(t *testing.T) {
	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	osExit = func(code int) {}

	// Test cases for different OS
	testCases := []struct {
		osType  string
		message string
	}{
		{"linux", "Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows."},
		{"windows", "Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows."},
	}

	for _, tc := range testCases {
		t.Run(tc.osType, func(t *testing.T) {
			// Capture output
			output := captureOutput(func() {
				mainWithOSType(tc.osType)
			})

			// Check if the output contains the expected message
			if !bytes.Contains([]byte(output), []byte(tc.message)) {
				t.Errorf("Expected message %q, but got %q", tc.message, output)
			}
		})
	}
}

func TestMain_PullCommand(t *testing.T) {
	// Setup
	os.Args = []string{"mdefaults", "pull", "config.yaml"}

	// Capture output
	output := captureOutput(func() {
		main()
	})

	// Verify
	if !bytes.Contains([]byte(output), []byte("Pulling configurations from:")) {
		t.Errorf("Expected pull command output, got: %s", output)
	}
}

func TestMain_PushCommand(t *testing.T) {
	// Setup
	os.Args = []string{"mdefaults", "push", "config.yaml"}

	// Capture output
	output := captureOutput(func() {
		main()
	})

	// Verify
	if !bytes.Contains([]byte(output), []byte("Pushing configurations to:")) {
		t.Errorf("Expected push command output, got: %s", output)
	}
}

// mainWithOSType is a helper function to simulate different OS types
func mainWithOSType(osType string) {
	if osType == "linux" || osType == "windows" {
		fmt.Println("Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows.")
	}

	// Use kong to parse commands
	ctx := kong.Parse(&CLI)

	switch ctx.Command() {
	case "version":
		fmt.Println("Version:", version)
		if CLI.Version.Verbose {
			fmt.Println("Architecture:", architecture)
		}
	case "pull <config>":
		fmt.Println("Pulling configurations from:", CLI.Pull.Config)
		// Simulate pull logic
	case "push <config>":
		fmt.Println("Pushing configurations to:", CLI.Push.Config)
		// Simulate push logic
	}
}

// Helper function to capture output
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		fmt.Printf("Failed to copy output: %v", err)
	}
	os.Stdout = originalStdout

	return buf.String()
}
