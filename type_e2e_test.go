package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestValueTypeE2E tests specifically the type support functionality in a controlled environment
func TestValueTypeE2E(t *testing.T) {
	// Check for short mode using environment variable instead of testing.Short()
	if os.Getenv("GO_TEST_SHORT") == "1" {
		t.Skip("Skipping E2E tests in short mode")
	}
	// Skip if not running in CI environment to prevent messing with local settings
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping value type E2E tests when not in CI environment")
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

	// Test different value types
	testCases := []struct {
		name      string
		domain    string
		key       string
		value     string
		valueType string
	}{
		{"Boolean", "com.test.mdefaults.e2e", "boolTest", "true", "boolean"},
		{"Integer", "com.test.mdefaults.e2e", "intTest", "42", "integer"},
		{"Float", "com.test.mdefaults.e2e", "floatTest", "3.14159", "float"},
		{"String", "com.test.mdefaults.e2e", "stringTest", "Hello World", "string"},
	}

	// Test each value type with push and pull operations
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Cleanup previous test values if they exist
			deleteCmd := exec.Command("defaults", "delete", tc.domain, tc.key)
			// Ignore any errors, key might not exist
			if err := deleteCmd.Run(); err != nil {
				log.Printf("Note: Failed to delete %s.%s - may not exist: %v", tc.domain, tc.key, err)
			}

			// Create test config with a single key-value pair
			testConfig := tc.domain + " " + tc.key + " -" + tc.valueType + " " + tc.value
			if err := os.WriteFile(originalConfig, []byte(testConfig), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Build the mdefaults binary if it doesn't exist
			if _, err := os.Stat("./mdefaults"); os.IsNotExist(err) {
				cmd := exec.Command("go", "build", "-o", "mdefaults")
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("Failed to build mdefaults: %v\nOutput: %s", err, output)
				}
			}

			// Execute push command with -y flag
			pushCmd := exec.Command("./mdefaults", "push", "-y")
			if output, err := pushCmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to execute push command: %v\nOutput: %s", err, output)
			}

			// Verify that the value was written with the correct type
			typeCmd := exec.Command("defaults", "read-type", tc.domain, tc.key)
			typeOutput, err := typeCmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to read type: %v\nOutput: %s", err, typeOutput)
			} else {
				typeStr := strings.TrimSpace(string(typeOutput))
				expectedType := "Type is " + tc.valueType
				if !strings.Contains(typeStr, expectedType) {
					t.Errorf("Expected type to be '%s', but got '%s'", expectedType, typeStr)
				}
			}

			// Read value to verify it was set correctly
			readCmd := exec.Command("defaults", "read", tc.domain, tc.key)
			readOutput, err := readCmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to read value: %v\nOutput: %s", err, readOutput)
			} else {
				valueStr := strings.TrimSpace(string(readOutput))
				if tc.valueType == "string" {
					// For string values, defaults may add quotes
					if !strings.Contains(valueStr, tc.value) {
						t.Errorf("Expected value to contain '%s', but got '%s'", tc.value, valueStr)
					}
				} else if tc.valueType == "boolean" {
					// macOS defaults uses 1/0 for boolean values
					expectedBool := "1"
					if tc.value == "false" {
						expectedBool = "0"
					}
					if valueStr != expectedBool {
						t.Errorf("Expected boolean value to be '%s', but got '%s'", expectedBool, valueStr)
					}
				} else if valueStr != tc.value {
					t.Errorf("Expected value to be '%s', but got '%s'", tc.value, valueStr)
				}
			}

			// Remove the config before the pull test
			if err := os.Remove(originalConfig); err != nil {
				log.Printf("Failed to remove config file: %v", err)
			}

			// Create a basic config without type
			basicConfig := tc.domain + " " + tc.key
			if err := os.WriteFile(originalConfig, []byte(basicConfig), 0644); err != nil {
				t.Fatalf("Failed to write basic config: %v", err)
			}

			// Execute pull command with -y flag
			pullCmd := exec.Command("./mdefaults", "pull", "-y")
			if output, err := pullCmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to execute pull command: %v\nOutput: %s", err, output)
			}

			// Read the updated config file
			updatedConfig, err := os.ReadFile(originalConfig)
			if err != nil {
				t.Fatalf("Failed to read updated config: %v", err)
			}

			// Verify that pull saved the correct type
			configStr := string(updatedConfig)
			expectedTypeMarker := "-" + tc.valueType
			if !strings.Contains(configStr, expectedTypeMarker) {
				t.Errorf("Expected config to contain type marker '%s', but got: %s", expectedTypeMarker, configStr)
			}

			// Clean up by deleting the test key
			deleteCleanupCmd := exec.Command("defaults", "delete", tc.domain, tc.key)
			if err := deleteCleanupCmd.Run(); err != nil {
				log.Printf("Warning: Failed to clean up test key %s.%s: %v", tc.domain, tc.key, err)
			}
		})
	}
}
