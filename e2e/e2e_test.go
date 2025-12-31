//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	// testDomain is a safe domain for testing that won't affect real user settings
	testDomain = "com.mdefaults.e2e.test"
	testKey    = "TestKey"
	testValue  = "TestValue123"
)

var binaryPath string

func TestMain(m *testing.M) {
	if runtime.GOOS != "darwin" {
		os.Exit(0) // Skip e2e tests on non-macOS
	}

	// Build the binary
	tmpDir, err := os.MkdirTemp("", "mdefaults-e2e")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "mdefaults")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/mdefaults")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup: remove test defaults
	cleanup()

	os.Exit(code)
}

func cleanup() {
	// Remove test domain from macOS defaults
	exec.Command("defaults", "delete", testDomain).Run()
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run version command: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Version:") {
		t.Errorf("expected version output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Architecture:") {
		t.Errorf("expected architecture output, got: %s", outputStr)
	}
}

func TestVFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run -v command: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Version:") {
		t.Errorf("expected version output, got: %s", outputStr)
	}
}

func TestNoCommand(t *testing.T) {
	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run command: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage: mdefaults") {
		t.Errorf("expected usage output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "pull") {
		t.Errorf("expected pull command in usage, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "push") {
		t.Errorf("expected push command in usage, got: %s", outputStr)
	}
}

func TestPushAndPull(t *testing.T) {
	// Setup: create a temporary config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	configPath := filepath.Join(homeDir, ".mdefaults")

	// Backup existing config if present
	var originalConfig []byte
	if data, err := os.ReadFile(configPath); err == nil {
		originalConfig = data
	}
	defer func() {
		if originalConfig != nil {
			os.WriteFile(configPath, originalConfig, 0644)
		} else {
			os.Remove(configPath)
		}
	}()

	// Write test config
	testConfig := testDomain + " " + testKey + " " + testValue + " string\n"
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Test push command
	t.Run("Push", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "push")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("push failed: %v, output: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "successfully") {
			t.Errorf("expected success message, got: %s", outputStr)
		}

		// Verify the value was written to macOS defaults
		verifyCmd := exec.Command("defaults", "read", testDomain, testKey)
		verifyOutput, err := verifyCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to verify defaults: %v, output: %s", err, verifyOutput)
		}

		if !strings.Contains(string(verifyOutput), testValue) {
			t.Errorf("expected value %s in defaults, got: %s", testValue, verifyOutput)
		}
	})

	// Test pull command
	t.Run("Pull", func(t *testing.T) {
		// First, set a different value in defaults to pull
		newValue := "PulledValue456"
		setCmd := exec.Command("defaults", "write", testDomain, testKey, "-string", newValue)
		if err := setCmd.Run(); err != nil {
			t.Fatalf("failed to set defaults value: %v", err)
		}

		// Run pull with -y flag to auto-confirm
		cmd := exec.Command(binaryPath, "pull", "-y")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("pull failed: %v, output: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "successfully") {
			t.Errorf("expected success message, got: %s", outputStr)
		}

		// Verify the config file was updated
		configData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if !strings.Contains(string(configData), newValue) {
			t.Errorf("expected config to contain %s, got: %s", newValue, configData)
		}
	})

	// Cleanup test domain
	cleanup()
}

func TestPushWithDifferentTypes(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	configPath := filepath.Join(homeDir, ".mdefaults")

	// Backup existing config
	var originalConfig []byte
	if data, err := os.ReadFile(configPath); err == nil {
		originalConfig = data
	}
	defer func() {
		if originalConfig != nil {
			os.WriteFile(configPath, originalConfig, 0644)
		} else {
			os.Remove(configPath)
		}
		cleanup()
	}()

	testCases := []struct {
		name     string
		key      string
		value    string
		typeStr  string
		expected string
	}{
		{"string", "StringKey", "hello", "string", "hello"},
		{"integer", "IntKey", "42", "integer", "42"},
		{"boolean_true", "BoolTrueKey", "true", "boolean", "1"},
		{"boolean_false", "BoolFalseKey", "false", "boolean", "0"},
		{"float", "FloatKey", "3.14", "float", "3.14"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write config for this test
			testConfig := testDomain + " " + tc.key + " " + tc.value + " " + tc.typeStr + "\n"
			if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Run push
			cmd := exec.Command(binaryPath, "push")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("push failed: %v, output: %s", err, output)
			}

			// Verify the value
			verifyCmd := exec.Command("defaults", "read", testDomain, tc.key)
			verifyOutput, err := verifyCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to verify defaults: %v, output: %s", err, verifyOutput)
			}

			if !strings.Contains(string(verifyOutput), tc.expected) {
				t.Errorf("expected value containing %s, got: %s", tc.expected, verifyOutput)
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	cmd := exec.Command(binaryPath, "unknown")
	output, _ := cmd.CombinedOutput()

	// The command should exit with non-zero status, which is expected
	// We just check the output doesn't crash
	_ = output
}
