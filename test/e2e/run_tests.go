package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("Starting mdefaults e2e tests")

	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		fmt.Println("Error: This test script must be run on macOS")
		os.Exit(1)
	}

	// Set up test environment
	testDir, err := os.MkdirTemp("", "mdefaults-e2e-test")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".mdefaults")
	backupDir := filepath.Join(os.Getenv("HOME"), ".mdefaults.backup")

	// Backup existing configuration if it exists
	if _, err := os.Stat(configDir); err == nil {
		fmt.Println("Backing up existing configuration")
		if err := os.Rename(configDir, backupDir); err != nil {
			log.Fatalf("Failed to backup configuration: %v", err)
		}
	}

	// Create test configuration directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// Register cleanup function to run on exit
	defer func() {
		fmt.Println("Cleaning up test environment")
		os.RemoveAll(testDir)
		os.Remove(filepath.Join(configDir, "config"))

		// Restore backup if it exists
		if _, err := os.Stat(backupDir); err == nil {
			os.RemoveAll(configDir)
			os.Rename(backupDir, configDir)
		}

		fmt.Println("Cleanup complete")
	}()

	// Get the directory of the script
	_, scriptPath, _, _ := runtime.Caller(0)
	scriptDir := filepath.Dir(scriptPath)

	// Path to the mdefaults binary
	mdefaultsBin := filepath.Join(scriptDir, "..", "..", "mdefaults")

	// If the binary doesn't exist, try to build it
	if _, err := os.Stat(mdefaultsBin); os.IsNotExist(err) {
		fmt.Println("mdefaults binary not found, building it")
		cmd := exec.Command("go", "build", "-o", "mdefaults", "./cmd/mdefaults")
		cmd.Dir = filepath.Join(scriptDir, "..", "..")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to build mdefaults: %v", err)
		}
		mdefaultsBin = filepath.Join(scriptDir, "..", "..", "mdefaults")
	}

	// Test 1: Create a test configuration file
	fmt.Println("Test 1: Creating test configuration file")
	configContent := "com.apple.dock autohide\ncom.apple.finder ShowPathbar\n"
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(configContent), 0644); err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}

	// Test 2: Run mdefaults pull
	fmt.Println("Test 2: Running mdefaults pull")
	cmd := exec.Command(mdefaultsBin, "pull", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run mdefaults pull: %v", err)
	}

	// Verify the configuration file was updated
	if _, err := os.Stat(filepath.Join(configDir, "config")); os.IsNotExist(err) {
		fmt.Println("Error: Configuration file not found after pull")
		os.Exit(1)
	}

	// Test 3: Modify the configuration file
	fmt.Println("Test 3: Modifying configuration file")
	// Save the original value of autohide
	cmd = exec.Command("defaults", "read", "com.apple.dock", "autohide")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to read dock autohide value: %v", err)
	}
	originalAutohide := strings.TrimSpace(string(output))

	// Toggle the value
	var newAutohide string
	if originalAutohide == "1" {
		newAutohide = "false"
	} else {
		newAutohide = "true"
	}

	// Update the configuration file
	configContent = fmt.Sprintf("com.apple.dock autohide %s\ncom.apple.finder ShowPathbar\n", newAutohide)
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(configContent), 0644); err != nil {
		log.Fatalf("Failed to update config file: %v", err)
	}

	// Test 4: Run mdefaults push
	fmt.Println("Test 4: Running mdefaults push")
	cmd = exec.Command(mdefaultsBin, "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run mdefaults push: %v", err)
	}

	// Verify the changes were applied
	cmd = exec.Command("defaults", "read", "com.apple.dock", "autohide")
	output, err = cmd.Output()
	if err != nil {
		log.Fatalf("Failed to read dock autohide value: %v", err)
	}
	currentAutohide := strings.TrimSpace(string(output))

	expectedValue := newAutohide
	if (newAutohide == "true" && currentAutohide == "1") || (newAutohide == "false" && currentAutohide == "0") {
		fmt.Println("Configuration applied correctly")
	} else {
		fmt.Printf("Error: Configuration not applied correctly\nExpected: %s, Got: %s\n", expectedValue, currentAutohide)
		os.Exit(1)
	}

	// Test 5: Restore original value
	fmt.Println("Test 5: Restoring original value")
	configContent = fmt.Sprintf("com.apple.dock autohide %s\ncom.apple.finder ShowPathbar\n", originalAutohide)
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(configContent), 0644); err != nil {
		log.Fatalf("Failed to update config file: %v", err)
	}

	cmd = exec.Command(mdefaultsBin, "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run mdefaults push: %v", err)
	}

	// Verify the original value was restored
	cmd = exec.Command("defaults", "read", "com.apple.dock", "autohide")
	output, err = cmd.Output()
	if err != nil {
		log.Fatalf("Failed to read dock autohide value: %v", err)
	}
	currentAutohide = strings.TrimSpace(string(output))

	if currentAutohide == originalAutohide {
		fmt.Println("Original value restored correctly")
	} else {
		fmt.Printf("Error: Original value not restored correctly\nExpected: %s, Got: %s\n", originalAutohide, currentAutohide)
		os.Exit(1)
	}

	fmt.Println("All tests passed successfully!")
}
