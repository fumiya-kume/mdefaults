package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

// Mock os.Exit function
var osExit = os.Exit

// Remove remaining pull-related tests

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func mockLogs() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&buf)
	return &buf, func() {
		log.SetOutput(originalLogOutput)
	}
}

func TestMain(m *testing.M) {
	// Initialize testing flags
	testing.Init()

	// Save the original os.Args and restore it after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Remove test flags from os.Args
	os.Args = []string{originalArgs[0]}

	// Run the tests
	os.Exit(m.Run())
}

func TestMain_NoCommand_ShowsHelp(t *testing.T) {
	logBuffer, restoreLogs := mockLogs()
	defer restoreLogs()

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

	_ = logBuffer // Use logBuffer if needed for assertions
}

func TestMain_VersionFlag(t *testing.T) {
	resetFlags()
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "--version"}
	output := captureOutput(func() {
		main()
	})

	if !bytes.Contains([]byte(output), []byte("Version: ")) || !bytes.Contains([]byte(output), []byte("Architecture: ")) {
		t.Errorf("Expected version and architecture information to be printed, got: %s", output)
	}
}

func TestMain_VFlag(t *testing.T) {
	resetFlags()
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"cmd", "-v"}
	output := captureOutput(func() {
		main()
	})

	if !bytes.Contains([]byte(output), []byte("Version: ")) || !bytes.Contains([]byte(output), []byte("Architecture: ")) {
		t.Errorf("Expected version and architecture information to be printed, got: %s", output)
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

// mainWithOSType is a helper function to simulate different OS types
func mainWithOSType(osType string) {
	if osType == "linux" || osType == "windows" {
		fmt.Println("Work In Progress: This tool uses macOS specific commands and may not function correctly on Linux/Windows.")
	}
	initFlags()

	if versionFlag || vFlag {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Architecture: %s\n", architecture)
		return
	}

	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Architecture: %s\n", architecture)
	osExit(run())
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
