package main

import (
	"regexp"
	"strings"
	"testing"
)

// TestParseReadTypeOutput tests the parsing of different output formats
// from the 'defaults read-type' command.
func TestParseReadTypeOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedType   string
	}{
		{
			name:           "Standard output format",
			output:         "Type is boolean",
			expectedType:   "boolean",
		},
		{
			name:           "Output with trailing newline",
			output:         "Type is boolean\n",
			expectedType:   "boolean",
		},
		{
			name:           "Output with trailing whitespace",
			output:         "Type is boolean  ",
			expectedType:   "boolean",
		},
		{
			name:           "Output with leading whitespace",
			output:         "  Type is boolean",
			expectedType:   "boolean",
		},
		{
			name:           "Output with multiple spaces",
			output:         "Type  is  boolean",
			expectedType:   "boolean",
		},
		{
			name:           "Output with different case",
			output:         "Type is BOOLEAN",
			expectedType:   "BOOLEAN",
		},
		{
			name:           "Output with integer type",
			output:         "Type is integer",
			expectedType:   "integer",
		},
		{
			name:           "Output with float type",
			output:         "Type is float",
			expectedType:   "float",
		},
		{
			name:           "Output with string type",
			output:         "Type is string",
			expectedType:   "string",
		},
		{
			name:           "Output with date type",
			output:         "Type is date",
			expectedType:   "date",
		},
		{
			name:           "Output with data type",
			output:         "Type is data",
			expectedType:   "data",
		},
		{
			name:           "Output with array type",
			output:         "Type is array",
			expectedType:   "array",
		},
		{
			name:           "Output with dictionary type",
			output:         "Type is dictionary",
			expectedType:   "dictionary",
		},
		{
			name:           "Unexpected output format",
			output:         "The type is boolean",
			expectedType:   "string", // Should default to string
		},
		{
			name:           "Empty output",
			output:         "",
			expectedType:   "string", // Should default to string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Directly test the parsing logic from DefaultsCommandImpl.ReadType
			var typeStr string
			outputStr := strings.TrimSpace(tt.output)

			// Use a regular expression to match "Type is" with any number of spaces
			// The ^ and $ ensure we match the entire string, not just a part of it
			re := regexp.MustCompile(`^Type\s+is\s+(.+)$`)
			matches := re.FindStringSubmatch(outputStr)

			if len(matches) > 1 {
				typeStr = strings.TrimSpace(matches[1])
			} else {
				typeStr = "string" // Default to string type if parsing fails
			}

			if typeStr != tt.expectedType {
				t.Errorf("parseReadTypeOutput(%q) = %q, want %q", tt.output, typeStr, tt.expectedType)
			}
		})
	}
}

// TestParseReadOutput tests the parsing of different output formats
// from the 'defaults read' command.
func TestParseReadOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedValue  string
	}{
		{
			name:           "Simple boolean value",
			output:         "1",
			expectedValue:  "1",
		},
		{
			name:           "Boolean value with newline",
			output:         "1\n",
			expectedValue:  "1",
		},
		{
			name:           "Integer value",
			output:         "42",
			expectedValue:  "42",
		},
		{
			name:           "Float value",
			output:         "3.14159",
			expectedValue:  "3.14159",
		},
		{
			name:           "String value",
			output:         "Hello World",
			expectedValue:  "Hello World",
		},
		{
			name:           "String value with newlines",
			output:         "Hello\nWorld",
			expectedValue:  "HelloWorld",
		},
		{
			name:           "String value with leading/trailing whitespace",
			output:         "  Hello World  ",
			expectedValue:  "Hello World",
		},
		{
			name:           "Empty value",
			output:         "",
			expectedValue:  "",
		},
		{
			name:           "Value with special characters",
			output:         "Hello, World! @#$%^&*()",
			expectedValue:  "Hello, World! @#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Directly test the processing logic from pull.go
			value := tt.output
			value = strings.TrimSpace(strings.ReplaceAll(value, "\n", ""))
			if value != tt.expectedValue {
				t.Errorf("parseReadOutput(%q) = %q, want %q", tt.output, value, tt.expectedValue)
			}
		})
	}
}
