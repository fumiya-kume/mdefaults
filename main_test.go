package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"
)

// Mock os.Exit function
var osExit = os.Exit

// Remove remaining pull-related tests

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
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

// Helper function to capture output
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = originalStdout

	return buf.String()
}
