package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E runs end-to-end tests for the mdefaults tool.
// These tests are designed to be run in a CI environment and won't affect the host system.
func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	// Skip if not running in CI environment to prevent messing with local settings
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping E2E tests when not in CI environment")
	}

	// Setup a test directory with a temporary .mdefaults file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Backup original .mdefaults if it exists
	originalConfig := filepath.Join(homeDir, ".mdefaults")
	backupConfig := filepath.Join(homeDir, ".mdefaults.bak")

	hasOriginalConfig := false
	if _, err := os.Stat(originalConfig); err == nil {
		hasOriginalConfig = true
		if err := os.Rename(originalConfig, backupConfig); err != nil {
			t.Fatalf("Failed to backup original config: %v", err)
		}
		defer func() {
			if err := os.Remove(originalConfig); err != nil {
				log.Printf("Failed to remove test config: %v", err)
			}
			if hasOriginalConfig {
				if err := os.Rename(backupConfig, originalConfig); err != nil {
					log.Printf("Failed to restore original config: %v", err)
				}
			}
		}()
	}

	// Create test config file with test values
	testConfig := `com.apple.homeenergyd Migration24
com.apple.iCal CALPrefLastTruthFileMigrationVersion
com.apple.WindowManager LastHeartbeatDateString.daily`

	if err := os.WriteFile(originalConfig, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Build the mdefaults binary
	cmd := exec.Command("go", "build", "-o", "mdefaults")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build mdefaults: %v\nOutput: %s", err, output)
	}

	// Test the pull command
	t.Run("PullCommand", func(t *testing.T) {
		cmd := exec.Command("./mdefaults", "pull", "-y") // Add -y flag to skip confirmation
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to execute pull command: %v\nOutput: %s", err, output)
		}

		// Read the updated config file
		updatedConfig, err := os.ReadFile(originalConfig)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		// Verify that the config has been updated with types
		updatedStr := string(updatedConfig)
		if !strings.Contains(updatedStr, "-") {
			t.Errorf("Expected config to contain type information, but got: %s", updatedStr)
		}

		// Check that our test keys exist in the updated config
		for _, key := range []string{"Migration24", "CALPrefLastTruthFileMigrationVersion", "LastHeartbeatDateString"} {
			if !strings.Contains(updatedStr, key) {
				t.Errorf("Expected updated config to contain key '%s', but it doesn't", key)
			}
		}
	})

	// Test the push command
	t.Run("PushCommand", func(t *testing.T) {
		// First, modify the config to set predictable test values
		testValues := `com.apple.homeenergyd Migration24 1
com.apple.iCal CALPrefLastTruthFileMigrationVersion 2
com.apple.WindowManager LastHeartbeatDateString.daily "hogehoge"`

		if err := os.WriteFile(originalConfig, []byte(testValues), 0644); err != nil {
			t.Fatalf("Failed to write test values: %v", err)
		}

		// Run push command
		cmd := exec.Command("./mdefaults", "push", "-y") // Add -y flag to skip confirmation if needed
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to execute push command: %v\nOutput: %s", err, output)
		}

		// Verify the values were set correctly using defaults command
		for _, tc := range []struct {
			domain        string
			key           string
			expectedValue string
			expectedType  string
		}{
			{"com.apple.homeenergyd", "Migration24", "1", "boolean"},
			{"com.apple.iCal", "CALPrefLastTruthFileMigrationVersion", "2", "integer"},
			{"com.apple.WindowManager", "LastHeartbeatDateString.daily", "hogehoge", "string"},
		} {
			// Check value
			valueCmd := exec.Command("defaults", "read", tc.domain, tc.key)
			valueOutput, err := valueCmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to read %s.%s value: %v\nOutput: %s", tc.domain, tc.key, err, valueOutput)
				continue
			}

			value := strings.TrimSpace(string(valueOutput))
			if value != tc.expectedValue {
				t.Errorf("Expected %s.%s value to be '%s', but got '%s'", tc.domain, tc.key, tc.expectedValue, value)
			}

			// Check type
			typeCmd := exec.Command("defaults", "read-type", tc.domain, tc.key)
			typeOutput, err := typeCmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to read %s.%s type: %v\nOutput: %s", tc.domain, tc.key, err, typeOutput)
				continue
			}

			typeStr := strings.TrimSpace(string(typeOutput))
			expectedTypeStr := "Type is " + tc.expectedType
			if !strings.Contains(typeStr, expectedTypeStr) {
				t.Errorf("Expected %s.%s type to be '%s', but got '%s'", tc.domain, tc.key, expectedTypeStr, typeStr)
			}
		}

		// Clean up by restoring original defaults values
		for _, item := range []struct {
			domain string
			key    string
		}{
			{"com.apple.homeenergyd", "Migration24"},
			{"com.apple.iCal", "CALPrefLastTruthFileMigrationVersion"},
			{"com.apple.WindowManager", "LastHeartbeatDateString.daily"},
		} {
			cmd := exec.Command("defaults", "delete", item.domain, item.key)
			if err := cmd.Run(); err != nil {
				// Just log errors here, as keys might not exist
				log.Printf("Warning: Failed to delete %s.%s: %v", item.domain, item.key, err)
			}
		}
	})
}
