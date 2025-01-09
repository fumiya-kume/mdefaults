package main

import (
	"fmt"
	"testing"
)

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
		{Domain: "", Key: "", Value: ""},
	}

	// Capture the output
	output := captureOutput(func() {
		push(configs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}

func TestPush_MaxConfigs(t *testing.T) {
	maxConfigs := make([]Config, 1000) // Assuming 1000 is the max for this example
	for i := 0; i < 1000; i++ {
		maxConfigs[i] = Config{Domain: fmt.Sprintf("domain%d", i), Key: fmt.Sprintf("key%d", i), Value: "value"}
	}

	// Capture the output
	output := captureOutput(func() {
		push(maxConfigs)
	})

	if output != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}
