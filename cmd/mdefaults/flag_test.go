package main

import (
	"flag"
	"os"
	"testing"
)

func TestInitFlags(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset flags before test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset flag variables
	versionFlag = false
	vFlag = false
	verboseFlag = false
	yesFlag = false

	// Initialize flags
	initFlags()

	// Test cases
	testCases := []struct {
		name     string
		args     []string
		flagVar  *bool
		expected bool
	}{
		{"version flag", []string{"cmd", "-version"}, &versionFlag, true},
		{"v flag", []string{"cmd", "-v"}, &vFlag, true},
		{"verbose flag", []string{"cmd", "-verbose"}, &verboseFlag, true},
		{"y flag", []string{"cmd", "-y"}, &yesFlag, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags before each test case
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)

			// Reset flag variables
			versionFlag = false
			vFlag = false
			verboseFlag = false
			yesFlag = false

			// Initialize flags
			initFlags()

			// Set args and parse
			os.Args = tc.args
			flag.Parse()

			// Check if flag is set correctly
			if *tc.flagVar != tc.expected {
				t.Errorf("Expected %v to be %v, got %v", tc.name, tc.expected, *tc.flagVar)
			}
		})
	}
}

func TestFlagDefaults(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset flags before test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset flag variables
	versionFlag = false
	vFlag = false
	verboseFlag = false
	yesFlag = false

	// Initialize flags
	initFlags()

	// Set args without any flags
	os.Args = []string{"cmd"}
	flag.Parse()

	// Check default values
	if versionFlag != false {
		t.Errorf("Expected versionFlag default to be false, got %v", versionFlag)
	}
	if vFlag != false {
		t.Errorf("Expected vFlag default to be false, got %v", vFlag)
	}
	if verboseFlag != false {
		t.Errorf("Expected verboseFlag default to be false, got %v", verboseFlag)
	}
	if yesFlag != false {
		t.Errorf("Expected yesFlag default to be false, got %v", yesFlag)
	}
}
