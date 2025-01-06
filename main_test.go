package main

import (
	"errors"
	"testing"
)

func TestPull_Success(t *testing.T) {
	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadResult: "1"},
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
	if updatedConfigs[0].Value != "1" {
		t.Errorf("Expected value '1', got %s", updatedConfigs[0].Value)
	}
}

func TestPull_ReadError(t *testing.T) {

	defaults := []DefaultsCommand{
		&MockDefaultsCommand{domain: "com.apple.dock", key: "autohide", ReadError: errors.New("read error")},
	}

	updatedConfigs, err := pullImpl(defaults)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}
