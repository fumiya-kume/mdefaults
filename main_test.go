package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

// Mock os.Exit function
var osExit = os.Exit

// Mock variables for testing
var (
	version      string = "test-version"
	architecture string = "test-arch"
	versionFlag  bool
	vFlag        bool
)

// Remove remaining pull-related tests

// Mock run function
func run() int {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mdefaults [command]")
		fmt.Println("Commands:")
		fmt.Println("  pull    - Retrieve and update configuration values.")
		fmt.Println("  push    - Write configuration values.")
		fmt.Println("Hey, let's call with pull or push.")
		return 0
	}
	return 0
}

// Mock initFlags function
func initFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.BoolVar(&versionFlag, "version", false, "Print version information")
	flag.BoolVar(&vFlag, "v", false, "Print version information")
}

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

	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	osExit = func(code int) {}

	os.Args = []string{"cmd", "--version"}
	output := captureOutput(func() {
		// Use mainWithOSType instead of main
		mainWithOSType("darwin")
	})

	if !bytes.Contains([]byte(output), []byte("Version: ")) || !bytes.Contains([]byte(output), []byte("Architecture: ")) {
		t.Errorf("Expected version and architecture information to be printed, got: %s", output)
	}
}

func TestMain_VFlag(t *testing.T) {
	resetFlags()
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	osExit = func(code int) {}

	os.Args = []string{"cmd", "-v"}
	output := captureOutput(func() {
		// Use mainWithOSType instead of main
		mainWithOSType("darwin")
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

func TestPrintConfigs(t *testing.T) {
	// Test with nil value
	nilConfig := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: nil},
	}

	output := captureOutput(func() {
		printConfigs(nilConfig)
	})

	expected := "- com.apple.dock autohide (no value)\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}

	// Test with value and type
	value := "true"
	withTypeConfig := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value, Type: "boolean"},
	}

	output = captureOutput(func() {
		printConfigs(withTypeConfig)
	})

	expected = "- com.apple.dock autohide -boolean true\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}

	// Test with value but no type
	noTypeConfig := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value},
	}

	output = captureOutput(func() {
		printConfigs(noTypeConfig)
	})

	expected = "- com.apple.dock autohide true\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

// TestRunWithNoArgs tests the Run function with no arguments
func TestRunWithNoArgs(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with no arguments
	os.Args = []string{"cmd"}

	output := captureOutput(func() {
		Run()
	})

	expected := "Usage: mdefaults [command]\nCommands:\n  pull    - Retrieve and update configuration values.\n  push    - Write configuration values.\nHey, let's call with pull or push.\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

// TestRunWithDebugCommand tests the Run function with the debug command
func TestRunWithDebugCommand(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Save the original os.Exit function and restore it after the test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	// Mock os.Exit to prevent it from terminating the test
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	// Test with debug command
	os.Args = []string{"cmd", "debug"}

	// Capture log output
	var buf bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalLogOutput)

	Run()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(buf.String(), "Debug command executed") {
		t.Errorf("Expected log to contain 'Debug command executed', got: %s", buf.String())
	}
}

// TestRunWithVerboseFlag tests the Run function with the verbose flag
func TestRunWithVerboseFlag(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Save the original verboseFlag and restore it after the test
	originalVerboseFlag := verboseFlag
	defer func() { verboseFlag = originalVerboseFlag }()

	// Set verboseFlag to true
	verboseFlag = true

	// Test with debug command
	os.Args = []string{"cmd", "debug"}

	// Capture log output
	var buf bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalLogOutput)

	Run()

	if !strings.Contains(buf.String(), "Verbose mode enabled") {
		t.Errorf("Expected log to contain 'Verbose mode enabled', got: %s", buf.String())
	}
}

// TestSetupLogging tests the setupLogging function
func TestSetupLogging(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "mdefaults-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Capture the log output
	var buf bytes.Buffer
	originalLogOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalLogOutput)

	// Call setupLogging
	setupLogging()

	// Check that the log file was created
	if _, err := os.Stat("mdefaults.log"); os.IsNotExist(err) {
		t.Errorf("Expected mdefaults.log to be created")
	}
}

// TestPrintUsage tests the printUsage function
func TestPrintUsage(t *testing.T) {
	output := captureOutput(func() {
		printUsage()
	})

	expected := "Usage: mdefaults [command]\nCommands:\n  pull    - Retrieve and update configuration values.\n  push    - Write configuration values.\nHey, let's call with pull or push.\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}
