package pull

import (
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func TestPullImplWithTypes(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{
			DomainVal:      "com.apple.dock",
			KeyVal:         "autohide",
			ReadResult:     "1",
			ReadTypeResult: "boolean",
		},
		&defaults.MockDefaultsCommand{
			DomainVal:      "com.apple.trackpad",
			KeyVal:         "ClickThreshold",
			ReadResult:     "2",
			ReadTypeResult: "integer",
		},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}

	expectedConfigs := []config.Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: stringPtr("1"), Type: "boolean"},
		{Domain: "com.apple.trackpad", Key: "ClickThreshold", Value: stringPtr("2"), Type: "integer"},
	}

	for i, expected := range expectedConfigs {
		if i >= len(updatedConfigs) {
			t.Errorf("Missing config at index %d", i)
			continue
		}
		actual := updatedConfigs[i]
		if actual.Domain != expected.Domain || actual.Key != expected.Key || 
		   *actual.Value != *expected.Value || actual.Type != expected.Type {
			t.Errorf("Config mismatch at index %d: expected %+v, got %+v", i, expected, actual)
		}
	}
}

func TestPullImplWithTypeDetectionFailure(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{
			DomainVal:     "com.apple.dock",
			KeyVal:        "autohide",
			ReadResult:    "1",
			ReadTypeError: &MockError{},
		},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}

	if updatedConfigs[0].Type != "string" {
		t.Errorf("Expected type 'string' when type detection fails, got %s", updatedConfigs[0].Type)
	}
}

type MockError struct{}

func (e *MockError) Error() string {
	return "mock error"
}

func stringPtr(s string) *string {
	return &s
}
