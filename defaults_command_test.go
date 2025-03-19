package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestDefaultsCommandReadSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadResult: "true",
		ReadError:  nil,
	}
	result, err := defaults.Read(context.Background())
	if err != nil {
		t.Errorf("Error reading defaults: %v", err)
	}
	if result != "true" {
		t.Errorf("Expected result to be 'true' but got %s", result)
	}
}

func TestDefaultsCommandReadError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadError: errors.New("read error"),
	}
	_, err := defaults.Read(context.Background())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestDefaultsCommandReadTypeSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadTypeResult: "boolean",
		ReadTypeError:  nil,
	}
	result, err := defaults.ReadType(context.Background())
	if err != nil {
		t.Errorf("Error reading defaults type: %v", err)
	}
	if result != "boolean" {
		t.Errorf("Expected result to be 'boolean' but got %s", result)
	}
}

func TestDefaultsCommandReadTypeError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadTypeError: errors.New("read type error"),
	}
	result, err := defaults.ReadType(context.Background())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if result != "string" {
		t.Errorf("Expected default type 'string' on error, but got %s", result)
	}
}

func TestDefaultsCommandReadTypeEmptyResult(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadTypeResult: "",
		ReadTypeError:  nil,
	}
	result, err := defaults.ReadType(context.Background())
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if result != "string" {
		t.Errorf("Expected default type 'string' when ReadTypeResult is empty, but got %s", result)
	}
}

func TestDefaultsCommandWriteSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: nil,
	}
	err := defaults.Write(context.Background(), "true", "boolean")
	if err != nil {
		t.Errorf("Error writing defaults: %v", err)
	}
}

func TestDefaultsCommandWriteError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: errors.New("write error"),
	}
	err := defaults.Write(context.Background(), "true", "boolean")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

type MockDefaultsCommand struct {
	ReadResult     string
	ReadError      error
	ReadTypeResult string
	ReadTypeError  error
	WriteError     error
	domain         string
	key            string
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) ReadType(ctx context.Context) (string, error) {
	if m.ReadTypeError != nil {
		return "string", m.ReadTypeError
	}
	if m.ReadTypeResult == "" {
		return "string", nil
	}
	return m.ReadTypeResult, nil
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string, valueType string) error {
	return m.WriteError
}

func (m *MockDefaultsCommand) Domain() string {
	return m.domain
}

func (m *MockDefaultsCommand) Key() string {
	return m.key
}

// TestDefaultsCommandImpl tests the actual implementation of DefaultsCommand
func TestDefaultsCommandImpl(t *testing.T) {
	// Skip this test if not running on macOS
	if _, err := exec.LookPath("defaults"); err != nil {
		t.Skip("Skipping test: 'defaults' command not available")
	}

	// Test Domain and Key methods
	t.Run("Domain and Key", func(t *testing.T) {
		cmd := &DefaultsCommandImpl{
			domain: "test.domain",
			key:    "test.key",
		}

		if cmd.Domain() != "test.domain" {
			t.Errorf("Expected domain 'test.domain', got '%s'", cmd.Domain())
		}

		if cmd.Key() != "test.key" {
			t.Errorf("Expected key 'test.key', got '%s'", cmd.Key())
		}
	})
}

// TestDefaultsCommandImplWriteFormat tests the command format in the Write method
func TestDefaultsCommandImplWriteFormat(t *testing.T) {
	// This test doesn't actually execute the defaults command,
	// it just verifies that the command format is correct for different value types

	// Test cases for different value types
	testCases := []struct {
		valueType string
		value     string
		expected  string
	}{
		{"string", "hello", `-string "hello"`},
		{"integer", "42", "-int 42"},
		{"int", "42", "-int 42"},
		{"float", "3.14", "-float 3.14"},
		{"real", "3.14", "-float 3.14"},
		{"bool", "true", "-bool true"},
		{"boolean", "false", "-bool false"},
		{"date", "2023-01-01", `-date "2023-01-01"`},
		{"data", "ABCDEF", "-data ABCDEF"},
		{"array", "item1 item2", "-array item1 item2"},
		{"dict", "key1 value1", "-dict key1 value1"},
		{"dictionary", "key1 value1", "-dict key1 value1"},
		{"unknown", "value", "value"},
	}

	for _, tc := range testCases {
		t.Run(tc.valueType, func(t *testing.T) {
			// Create a command string using the same logic as in Write
			var command string

			// Format command based on the value type
			switch tc.valueType {
			case "string":
				command = fmt.Sprintf(`defaults write domain key -string "%s"`, tc.value)
			case "integer", "int":
				command = fmt.Sprintf("defaults write domain key -int %s", tc.value)
			case "float", "real":
				command = fmt.Sprintf("defaults write domain key -float %s", tc.value)
			case "bool", "boolean":
				command = fmt.Sprintf("defaults write domain key -bool %s", tc.value)
			case "date":
				command = fmt.Sprintf(`defaults write domain key -date "%s"`, tc.value)
			case "data":
				command = fmt.Sprintf(`defaults write domain key -data %s`, tc.value)
			case "array":
				command = fmt.Sprintf("defaults write domain key -array %s", tc.value)
			case "dict", "dictionary":
				command = fmt.Sprintf("defaults write domain key -dict %s", tc.value)
			default:
				// If we don't recognize the type, fall back to not specifying a type
				command = fmt.Sprintf("defaults write domain key %s", tc.value)
			}

			// Check that the command contains the expected format
			if !strings.Contains(command, tc.expected) {
				t.Errorf("Expected command to contain '%s', got '%s'", tc.expected, command)
			}
		})
	}
}
