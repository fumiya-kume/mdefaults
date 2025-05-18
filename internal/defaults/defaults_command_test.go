package defaults

import (
	"context"
	"errors"
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

func TestDefaultsCommandWriteSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: nil,
	}
	err := defaults.Write(context.Background(), "true", "bool")
	if err != nil {
		t.Errorf("Error writing defaults: %v", err)
	}
}

func TestDefaultsCommandWriteError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: errors.New("write error"),
	}
	err := defaults.Write(context.Background(), "true", "bool")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// Test ReadType functionality
func TestDefaultsCommandReadType(t *testing.T) {
	testCases := []struct {
		name        string
		typeResult  string
		expectError bool
		typeError   error
	}{
		{"Boolean type", "bool", false, nil},
		{"Integer type", "int", false, nil},
		{"Float type", "float", false, nil},
		{"String type", "string", false, nil},
		{"Error case", "", true, errors.New("read type error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := &MockDefaultsCommand{
				TypeResult: tc.typeResult,
				TypeError:  tc.typeError,
			}
			result, err := defaults.ReadType(context.Background())
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Error reading type: %v", err)
			}
			if !tc.expectError && result != tc.typeResult {
				t.Errorf("Expected type result to be '%s' but got '%s'", tc.typeResult, result)
			}
		})
	}
}

// Additional test cases with different input/output values

func TestDefaultsCommandReadDifferentValues(t *testing.T) {
	testCases := []struct {
		name        string
		readResult  string
		expectError bool
	}{
		{"Boolean false", "false", false},
		{"Integer value", "42", false},
		{"Decimal value", "3.14", false},
		{"String value", "hello world", false},
		{"Empty string", "", false},
		{"Special characters", "!@#$%^&*()", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := &MockDefaultsCommand{
				ReadResult: tc.readResult,
				ReadError:  nil,
			}
			result, err := defaults.Read(context.Background())
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Error reading defaults: %v", err)
			}
			if result != tc.readResult {
				t.Errorf("Expected result to be '%s' but got '%s'", tc.readResult, result)
			}
		})
	}
}

func TestDefaultsCommandWriteDifferentValues(t *testing.T) {
	testCases := []struct {
		name        string
		writeValue  string
		writeType   string
		expectError bool
	}{
		{"Boolean true", "true", "bool", false},
		{"Boolean false", "false", "bool", false},
		{"Integer value", "42", "int", false},
		{"Decimal value", "3.14", "float", false},
		{"String value", "hello world", "string", false},
		{"Empty string", "", "string", false},
		{"Special characters", "!@#$%^&*()", "string", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := &MockDefaultsCommand{
				WriteError: nil,
			}
			err := defaults.Write(context.Background(), tc.writeValue, tc.writeType)
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Error writing defaults: %v", err)
			}
		})
	}
}

func TestDefaultsCommandDomainAndKey(t *testing.T) {
	testCases := []struct {
		name      string
		domain    string
		key       string
		expectErr bool
	}{
		{"Standard domain and key", "com.example.app", "setting", false},
		{"Empty domain", "", "setting", false},
		{"Empty key", "com.example.app", "", false},
		{"Both empty", "", "", false},
		{"Domain with special chars", "com.example-app.test", "setting", false},
		{"Key with special chars", "com.example.app", "setting-name", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := &MockDefaultsCommand{
				DomainVal: tc.domain,
				KeyVal:    tc.key,
			}

			domain := defaults.Domain()
			if domain != tc.domain {
				t.Errorf("Expected domain to be '%s' but got '%s'", tc.domain, domain)
			}

			key := defaults.Key()
			if key != tc.key {
				t.Errorf("Expected key to be '%s' but got '%s'", tc.key, key)
			}
		})
	}
}

func TestDefaultsCommandWithContextCancellation(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadResult: "test",
		ReadError:  nil,
	}

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The mock doesn't actually respect context cancellation,
	// but this test demonstrates how we would test it if it did
	result, err := defaults.Read(ctx)
	if err != nil {
		t.Errorf("Error reading defaults: %v", err)
	}
	if result != "test" {
		t.Errorf("Expected result to be 'test' but got '%s'", result)
	}
}

func TestDefaultsCommandImplReadEmptyDomainOrKey(t *testing.T) {
	testCases := []struct {
		name   string
		domain string
		key    string
	}{
		{"Empty domain", "", "setting"},
		{"Empty key", "com.example.app", ""},
		{"Both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := NewDefaultsCommandImpl(tc.domain, tc.key)
			_, err := defaults.Read(context.Background())
			if err == nil {
				t.Errorf("Expected error for empty domain or key, got nil")
			}
		})
	}
}

func TestDefaultsCommandImplWriteEmptyDomainOrKey(t *testing.T) {
	testCases := []struct {
		name   string
		domain string
		key    string
	}{
		{"Empty domain", "", "setting"},
		{"Empty key", "com.example.app", ""},
		{"Both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaults := NewDefaultsCommandImpl(tc.domain, tc.key)
			err := defaults.Write(context.Background(), "test", "string")
			if err == nil {
				t.Errorf("Expected error for empty domain or key, got nil")
			}
		})
	}
}
