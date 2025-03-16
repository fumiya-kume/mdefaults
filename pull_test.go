package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestPull_Success(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{
			domain:         "com.apple.dock",
			key:            "autohide",
			ReadResult:     "1",
			ReadTypeResult: "boolean",
		},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" {
		t.Errorf("Expected domain 'com.apple.dock', got %s", updatedConfigs[0].Domain)
	}
	if updatedConfigs[0].Key != "autohide" {
		t.Errorf("Expected key 'autohide', got %s", updatedConfigs[0].Key)
	}
	if *updatedConfigs[0].Value != "1" {
		t.Errorf("Expected value '1', got %s", *updatedConfigs[0].Value)
	}
	if updatedConfigs[0].Type != "boolean" {
		t.Errorf("Expected type 'boolean', got %s", updatedConfigs[0].Type)
	}
}

func TestPull_ReadError(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := pullImpl(defaults)
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" {
		t.Errorf("Expected domain 'com.apple.dock', got %s", updatedConfigs[0].Domain)
	}
	if updatedConfigs[0].Key != "autohide" {
		t.Errorf("Expected key 'autohide', got %s", updatedConfigs[0].Key)
	}
	if updatedConfigs[0].Value != nil {
		t.Errorf("Expected nil value, got %v", *updatedConfigs[0].Value)
	}
}

func TestPull_ReadTypeError(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{
			domain:        "com.apple.dock",
			key:           "autohide",
			ReadResult:    "1",
			ReadTypeError: errors.New("read type error"),
		},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	// Should default to string type when read-type fails
	if updatedConfigs[0].Type != "string" {
		t.Errorf("Expected default type 'string', got %s", updatedConfigs[0].Type)
	}
}

func TestPull_MultipleConfigs(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{
			domain:         "com.apple.dock",
			key:            "autohide",
			ReadResult:     "1",
			ReadTypeResult: "boolean",
		},
		&MockDefaultsCommand{
			domain:         "com.apple.finder",
			key:            "ShowPathbar",
			ReadResult:     "true",
			ReadTypeResult: "boolean",
		},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || *updatedConfigs[0].Value != "1" || updatedConfigs[0].Type != "boolean" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
	if updatedConfigs[1].Domain != "com.apple.finder" || updatedConfigs[1].Key != "ShowPathbar" || *updatedConfigs[1].Value != "true" || updatedConfigs[1].Type != "boolean" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[1])
	}
}

func TestPull_DifferentTypes(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1", ReadTypeResult: "boolean"},
		&MockDefaultsCommand{domain: "com.apple.dock", key: "tilesize", ReadResult: "48", ReadTypeResult: "integer"},
		&MockDefaultsCommand{domain: "com.apple.dock", key: "largesize", ReadResult: "64.0", ReadTypeResult: "float"},
		&MockDefaultsCommand{domain: "com.apple.dock", key: "appname", ReadResult: "Finder", ReadTypeResult: "string"},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 4 {
		t.Errorf("Expected 4 configs, got %d", len(updatedConfigs))
	}

	// Check each type was correctly recorded
	expectedTypes := []string{"boolean", "integer", "float", "string"}
	for i, expectedType := range expectedTypes {
		if updatedConfigs[i].Type != expectedType {
			t.Errorf("Config %d: Expected type '%s', got '%s'", i, expectedType, updatedConfigs[i].Type)
		}
	}
}

func TestPull_EmptyConfigs(t *testing.T) {
	defaults := []DefaultsCommand{}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MixedResults(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1", ReadTypeResult: "boolean"},
		&MockDefaultsCommand{domain: "com.apple.finder", key: "ShowPathbar", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := pullImpl(defaults)
	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || *updatedConfigs[0].Value != "1" || updatedConfigs[0].Type != "boolean" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
	if updatedConfigs[1].Domain != "com.apple.finder" || updatedConfigs[1].Key != "ShowPathbar" || updatedConfigs[1].Value != nil {
		t.Errorf("Unexpected config: %+v", updatedConfigs[1])
	}
}

func TestPull_InvalidConfig(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "", key: "", ReadError: errors.New("invalid config")},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MaxConfigs(t *testing.T) {
	maxConfigs := make([]DefaultsCommand, 1000) // Assuming 1000 is the max for this example
	for i := 0; i < 1000; i++ {
		maxConfigs[i] = &MockDefaultsCommand{
			domain:         fmt.Sprintf("domain%d", i),
			key:            fmt.Sprintf("key%d", i),
			ReadResult:     "value",
			ReadTypeResult: "string",
		}
	}

	updatedConfigs, err := pullImpl(maxConfigs)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1000 {
		t.Errorf("Expected 1000 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_ErrorHandling(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadError: errors.New("unexpected error")},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || updatedConfigs[0].Value != nil {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
}
