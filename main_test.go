package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Mock os.Exit function
var osExit = os.Exit

// Remove remaining pull-related tests

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
