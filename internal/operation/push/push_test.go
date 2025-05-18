package push

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
)

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

func TestPush_EmptyConfigs(t *testing.T) {
	configs := []config.Config{}

	// Capture the output
	output := captureOutput(func() {
		Push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_DifferentTypes(t *testing.T) {
	// Create a mock implementation for testing
	intValue := "42"
	boolValue := "true"
	floatValue := "3.14"
	stringValue := "hello"

	configs := []config.Config{
		{Domain: "com.example.app", Key: "intSetting", Value: &intValue, Type: "int"},
		{Domain: "com.example.app", Key: "boolSetting", Value: &boolValue, Type: "bool"},
		{Domain: "com.example.app", Key: "floatSetting", Value: &floatValue, Type: "float"},
		{Domain: "com.example.app", Key: "stringSetting", Value: &stringValue, Type: "string"},
	}

	// Set up a mock way to verify the Write calls
	// For real testing, you would inject a mock DefaultsCommand implementation
	// and verify that it's called with the expected types

	// This is a simple output test that just verifies the function doesn't crash
	output := captureOutput(func() {
		Push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_InvalidConfig(t *testing.T) {
	configs := []config.Config{
		{Domain: "", Key: "", Value: nil, Type: "string"},
	}

	// Capture the output
	output := captureOutput(func() {
		Push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_MaxConfigs(t *testing.T) {
	maxConfigs := make([]config.Config, 1000) // Assuming 1000 is the max for this example
	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("value%d", i)
		maxConfigs[i] = config.Config{
			Domain: fmt.Sprintf("domain%d", i),
			Key:    fmt.Sprintf("key%d", i),
			Value:  &value,
			Type:   "string",
		}
	}

	// Capture the output
	output := captureOutput(func() {
		Push(maxConfigs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}
