package main

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestDefaultsCommandAvailability checks if the 'defaults' command is available
// and logs detailed information to help diagnose CI issues.
func TestDefaultsCommandAvailability(t *testing.T) {
	// Skip if not running in CI environment
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping defaults availability test when not in CI environment")
	}

	// Log OS information
	t.Logf("Running on OS: %s", runtime.GOOS)
	
	// Check if we're on macOS
	if runtime.GOOS != "darwin" {
		t.Logf("Not running on macOS, 'defaults' command is not expected to be available")
		return
	}

	// Try to execute 'defaults' command
	cmd := exec.Command("defaults", "help")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("'defaults' command is not available: %v", err)
		t.Logf("Output: %s", string(output))
		
		// Check PATH environment variable
		t.Logf("PATH: %s", os.Getenv("PATH"))
		
		// Try to find the defaults command
		findCmd := exec.Command("which", "defaults")
		findOutput, findErr := findCmd.CombinedOutput()
		if findErr != nil {
			t.Logf("Failed to locate 'defaults' command: %v", findErr)
		} else {
			t.Logf("'defaults' command location: %s", string(findOutput))
		}
	} else {
		t.Logf("'defaults' command is available")
		t.Logf("Output: %s", string(output))
	}
}

// TestSkipE2ETestsOnNonMacOS verifies that E2E tests are properly skipped on non-macOS platforms
func TestSkipE2ETestsOnNonMacOS(t *testing.T) {
	// This test doesn't actually skip itself, it just verifies the logic
	// that should be used in the E2E tests
	
	if runtime.GOOS != "darwin" {
		t.Logf("Not running on macOS, E2E tests should be skipped")
	} else {
		// On macOS, check if defaults command is actually available
		cmd := exec.Command("defaults", "help")
		if err := cmd.Run(); err != nil {
			t.Logf("On macOS but 'defaults' command failed: %v", err)
			t.Logf("E2E tests should be skipped in this environment")
		} else {
			t.Logf("On macOS and 'defaults' command is available, E2E tests can run")
		}
	}
}