package defaults

import (
	"strings"
	"testing"
)

func TestMapMacOSTypeToInternal(t *testing.T) {
	testCases := []struct {
		macOSType    string
		expectedType string
		expectError  bool
	}{
		{macOSType: "integer", expectedType: "integer"},
		{macOSType: "boolean", expectedType: "boolean"},
		{macOSType: "string", expectedType: "string"},
		{macOSType: "float", expectedType: "float"},
		{macOSType: "real", expectedType: "float"},
		{macOSType: "date", expectedType: "date"},
		{macOSType: "array", expectError: true},
		{macOSType: "dictionary", expectError: true},
		{macOSType: "data", expectError: true},
		{macOSType: "unknown", expectError: true},
		{macOSType: "", expectError: true},
	}

	for _, tc := range testCases {
		result, err := mapMacOSTypeToInternal(tc.macOSType)
		if tc.expectError {
			if err == nil {
				t.Errorf("mapMacOSTypeToInternal(%s) expected an error", tc.macOSType)
			}
			continue
		}
		if err != nil {
			t.Errorf("mapMacOSTypeToInternal(%s) returned error: %v", tc.macOSType, err)
			continue
		}
		if result != tc.expectedType {
			t.Errorf("mapMacOSTypeToInternal(%s) = %s, expected %s", tc.macOSType, result, tc.expectedType)
		}
	}
}

func TestMapInternalTypeToFlag(t *testing.T) {
	testCases := []struct {
		internalType string
		expectedFlag string
	}{
		{"integer", "-int"},
		{"boolean", "-bool"},
		{"string", ""},
		{"float", "-float"},
		{"date", "-date"},
		{"array", ""},
		{"dict", ""},
		{"data", ""},
		{"unknown", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		result := mapInternalTypeToFlag(tc.internalType)
		if result != tc.expectedFlag {
			t.Errorf("mapInternalTypeToFlag(%s) = %s, expected %s", tc.internalType, result, tc.expectedFlag)
		}
	}
}

func TestParseMacOSTypeOutput(t *testing.T) {
	testCases := []struct {
		name         string
		output       string
		expectedType string
		expectError  bool
	}{
		{name: "integer", output: "Type is integer\n", expectedType: "integer"},
		{name: "real", output: "Type is real", expectedType: "float"},
		{name: "malformed", output: "integer", expectError: true},
		{name: "empty", output: "", expectError: true},
		{name: "unknown", output: "Type is future-type", expectError: true},
		{name: "compound", output: "Type is array", expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseMacOSTypeOutput(tc.output)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected an error, got type %q", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if actual != tc.expectedType {
				t.Fatalf("expected type %q, got %q", tc.expectedType, actual)
			}
		})
	}
}

func TestValidateTypeRejectsUnsupportedType(t *testing.T) {
	err := ValidateType("unsupported")
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
	if !strings.Contains(err.Error(), "unsupported defaults type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTypeAcceptsWritableScalarTypes(t *testing.T) {
	for _, valueType := range []string{"", "string", "integer", "boolean", "float", "date"} {
		t.Run(valueType, func(t *testing.T) {
			if err := ValidateType(valueType); err != nil {
				t.Fatalf("expected type %q to be writable, got %v", valueType, err)
			}
		})
	}
}

func TestValidateValueAcceptsAndNormalizesBooleanValues(t *testing.T) {
	testCases := []struct {
		value    string
		expected string
	}{
		{value: "1", expected: "true"},
		{value: "0", expected: "false"},
		{value: "true", expected: "true"},
		{value: "false", expected: "false"},
		{value: "yes", expected: "true"},
		{value: "no", expected: "false"},
		{value: "TRUE", expected: "true"},
		{value: "FALSE", expected: "false"},
		{value: "YES", expected: "true"},
		{value: "NO", expected: "false"},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			if err := ValidateValue(tc.value, "boolean"); err != nil {
				t.Fatalf("expected %q to be valid, got %v", tc.value, err)
			}
			actual, err := normalizeValue(tc.value, "boolean")
			if err != nil {
				t.Fatalf("expected %q to normalize, got %v", tc.value, err)
			}
			if actual != tc.expected {
				t.Fatalf("expected %q to normalize to %q, got %q", tc.value, tc.expected, actual)
			}
		})
	}
}

func TestValidateValueRejectsInvalidBooleanValues(t *testing.T) {
	for _, value := range []string{"", "2", "on", "off", "maybe", " true "} {
		t.Run(value, func(t *testing.T) {
			err := ValidateValue(value, "boolean")
			if err == nil {
				t.Fatal("expected invalid boolean value error")
			}
			if !strings.Contains(err.Error(), "invalid boolean value") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNormalizeValuePreservesNonBooleanValues(t *testing.T) {
	for _, valueType := range []string{"", "string", "integer", "float", "date"} {
		t.Run(valueType, func(t *testing.T) {
			const value = "value with spaces"
			actual, err := normalizeValue(value, valueType)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if actual != value {
				t.Fatalf("expected value to be preserved, got %q", actual)
			}
		})
	}
}
