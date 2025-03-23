package pull

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func TestPull_Success(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
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
}

func TestPull_ReadError(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := PullImpl(defaultsCmds)
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MultipleConfigs(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.finder", KeyVal: "ShowPathbar", ReadResult: "true"},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || *updatedConfigs[0].Value != "1" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
	if updatedConfigs[1].Domain != "com.apple.finder" || updatedConfigs[1].Key != "ShowPathbar" || *updatedConfigs[1].Value != "true" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[1])
	}
}

func TestPull_EmptyConfigs(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MixedResults(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.finder", KeyVal: "ShowPathbar", ReadError: errors.New("read error")},
	}

	updatedConfigs, _ := PullImpl(defaultsCmds)
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || *updatedConfigs[0].Value != "1" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
}

func TestPull_InvalidConfig(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "", KeyVal: "", ReadError: errors.New("invalid config")},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MaxConfigs(t *testing.T) {
	maxConfigs := make([]defaults.DefaultsCommand, 1000) // Assuming 1000 is the max for this example
	for i := 0; i < 1000; i++ {
		maxConfigs[i] = &defaults.MockDefaultsCommand{DomainVal: fmt.Sprintf("domain%d", i), KeyVal: fmt.Sprintf("key%d", i), ReadResult: "value"}
	}

	updatedConfigs, err := PullImpl(maxConfigs)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1000 {
		t.Errorf("Expected 1000 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_ErrorHandling(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadError: errors.New("unexpected error")},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}
