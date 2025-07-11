package defaults

import (
	"context"
	"testing"
)

func TestMapMacOSTypeToInternal(t *testing.T) {
	testCases := []struct {
		macOSType    string
		expectedType string
	}{
		{"integer", "integer"},
		{"boolean", "boolean"},
		{"string", "string"},
		{"float", "float"},
		{"real", "float"},
		{"date", "date"},
		{"array", "array"},
		{"dictionary", "dict"},
		{"data", "data"},
		{"unknown", "string"},
		{"", "string"},
	}

	for _, tc := range testCases {
		result := mapMacOSTypeToInternal(tc.macOSType)
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
		{"array", "-array"},
		{"dict", "-dict"},
		{"data", "-data"},
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

func TestMockDefaultsCommandReadType(t *testing.T) {
	mock := &MockDefaultsCommand{
		DomainVal:      "com.apple.dock",
		KeyVal:         "autohide",
		ReadTypeResult: "boolean",
	}

	result, err := mock.ReadType(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "boolean" {
		t.Errorf("Expected 'boolean', got %s", result)
	}
}

func TestMockDefaultsCommandWriteWithType(t *testing.T) {
	mock := &MockDefaultsCommand{
		DomainVal: "com.apple.dock",
		KeyVal:    "autohide",
	}

	err := mock.WriteWithType(context.Background(), "1", "boolean")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
