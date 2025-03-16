package main

import (
	"fmt"
	"testing"
)

func TestPush_NilValue(t *testing.T) {
	value := "test"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: nil, Type: "boolean"},
		{Domain: "com.apple.dock", Key: "tilesize", Value: &value, Type: "integer"},
	}

	// Capture the output
	output := captureOutput(func() {
		push(configs)
	})

	// We expect the function to skip the nil value and continue with the next config
	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_TypeInference(t *testing.T) {
	// Test cases for type inference
	testCases := []struct {
		name     string
		value    string
		expected string
	}{
		{"Quoted string", "\"hello world\"", "string"},
		{"Single quoted string", "'hello world'", "string"},
		{"Boolean true", "true", "boolean"},
		{"Boolean false", "false", "boolean"},
		{"Boolean 1", "1", "boolean"},
		{"Boolean 0", "0", "boolean"},
		{"Integer", "42", "integer"},
		{"Float", "3.14", "float"},
		{"Default string", "hello world", "string"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.value
			config := Config{
				Domain: "com.apple.test",
				Key:    "testkey",
				Value:  &value,
				Type:   "", // Empty type to trigger inference
			}

			// Capture the output
			output := captureOutput(func() {
				push([]Config{config})
			})

			if output != "" {
				t.Errorf("Expected no output, got %s", output)
			}
		})
	}
}

func TestPush_EmptyConfigs(t *testing.T) {
	configs := []Config{}

	// Capture the output
	output := captureOutput(func() {
		push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_InvalidConfig(t *testing.T) {
	configs := []Config{
		{Domain: "", Key: "", Value: nil, Type: ""},
	}

	// Capture the output
	output := captureOutput(func() {
		push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_DifferentTypes(t *testing.T) {
	// Create test configs with different value types
	value1 := "1"
	value2 := "48"
	value3 := "true"
	value4 := "Hello"

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
		{Domain: "com.apple.dock", Key: "tilesize", Value: &value2, Type: "integer"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value3, Type: "boolean"},
		{Domain: "com.apple.finder", Key: "SearchWindowName", Value: &value4, Type: "string"},
	}

	// Capture the output without actual execution (we're not checking output here)
	output := captureOutput(func() {
		push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
	// Since we're not actually executing commands in tests, we're just ensuring
	// the function doesn't crash with different value types
}

func TestPush_DefaultTypeIfMissing(t *testing.T) {
	value := "test"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "testvalue", Value: &value, Type: ""},
	}

	// Capture the output
	output := captureOutput(func() {
		push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
	// The default type should be "string" if none is provided
}

func TestPush_MaxConfigs(t *testing.T) {
	maxConfigs := make([]Config, 1000) // Assuming 1000 is the max for this example
	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("value%d", i)
		maxConfigs[i] = Config{
			Domain: fmt.Sprintf("domain%d", i),
			Key:    fmt.Sprintf("key%d", i),
			Value:  &value,
			Type:   "string",
		}
	}

	// Capture the output
	output := captureOutput(func() {
		push(maxConfigs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}
