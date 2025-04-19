package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	// Test creating a new error
	err := New(InvalidArgument, "test message")

	// Check that the error is of the correct type
	appErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", err)
	}

	// Check that the error code is set correctly
	if appErr.Code != InvalidArgument {
		t.Errorf("Expected error code %d, got %d", InvalidArgument, appErr.Code)
	}

	// Check that the message is set correctly
	if appErr.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", appErr.Message)
	}

	// Check that the wrapped error is nil
	if appErr.Err != nil {
		t.Errorf("Expected nil wrapped error, got %v", appErr.Err)
	}
}

func TestWrap(t *testing.T) {
	// Test wrapping a nil error
	err := Wrap(nil, FileNotFound, "file not found")
	if err != nil {
		t.Errorf("Expected nil error when wrapping nil, got %v", err)
	}

	// Test wrapping a standard error
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, FileNotFound, "file not found")

	// Check that the error is of the correct type
	appErr, ok := wrappedErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", wrappedErr)
	}

	// Check that the error code is set correctly
	if appErr.Code != FileNotFound {
		t.Errorf("Expected error code %d, got %d", FileNotFound, appErr.Code)
	}

	// Check that the message is set correctly
	if appErr.Message != "file not found" {
		t.Errorf("Expected message 'file not found', got '%s'", appErr.Message)
	}

	// Check that the wrapped error is set correctly
	if appErr.Err != originalErr {
		t.Errorf("Expected wrapped error %v, got %v", originalErr, appErr.Err)
	}
}

func TestError_Error(t *testing.T) {
	// Test error message formatting without a wrapped error
	err := New(InvalidArgument, "invalid argument")
	expectedMsg := fmt.Sprintf("[ERROR-%04d] invalid argument", InvalidArgument)
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test error message formatting with a wrapped error
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, FileNotFound, "file not found")
	expectedWrappedMsg := fmt.Sprintf("[ERROR-%04d] file not found: %v", FileNotFound, originalErr)
	if wrappedErr.Error() != expectedWrappedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedWrappedMsg, wrappedErr.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	// Test unwrapping an error
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, FileNotFound, "file not found")

	// Check that the unwrapped error is the original error
	appErr, ok := wrappedErr.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", wrappedErr)
	}

	if appErr.Unwrap() != originalErr {
		t.Errorf("Expected unwrapped error %v, got %v", originalErr, appErr.Unwrap())
	}
}

func TestGetErrorCode(t *testing.T) {
	// Test getting the error code from a nil error
	code := GetErrorCode(nil)
	if code != 0 {
		t.Errorf("Expected error code 0 for nil error, got %d", code)
	}

	// Test getting the error code from an application error
	err := New(InvalidArgument, "invalid argument")
	code = GetErrorCode(err)
	if code != InvalidArgument {
		t.Errorf("Expected error code %d, got %d", InvalidArgument, code)
	}

	// Test getting the error code from a standard error
	stdErr := errors.New("standard error")
	code = GetErrorCode(stdErr)
	if code != Unknown {
		t.Errorf("Expected error code %d for standard error, got %d", Unknown, code)
	}
}

func TestGetErrorMessage(t *testing.T) {
	testCases := []struct {
		code     ErrorCode
		expected string
	}{
		{InvalidArgument, "Invalid argument provided"},
		{FileNotFound, "File not found"},
		{FilePermissionDenied, "Permission denied when accessing file"},
		{FileReadError, "Error reading file"},
		{FileWriteError, "Error writing file"},
		{DefaultsEmptyDomainOrKey, "Domain and key cannot be empty"},
		{DefaultsReadError, "Error reading defaults"},
		{DefaultsWriteError, "Error writing defaults"},
		{DefaultsCommandExecutionError, "Error executing defaults command"},
		{ConfigParseError, "Error parsing configuration"},
		{ConfigWriteError, "Error writing configuration"},
		{Unknown, "Unknown error"},
		{99999, "Unknown error"}, // Test an undefined error code
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ErrorCode_%d", tc.code), func(t *testing.T) {
			message := GetErrorMessage(tc.code)
			if message != tc.expected {
				t.Errorf("Expected message '%s' for error code %d, got '%s'", tc.expected, tc.code, message)
			}
		})
	}
}

func TestErrorIntegration(t *testing.T) {
	// Test the full error chain: create, wrap, get code, get message
	originalErr := errors.New("file does not exist")
	wrappedErr := Wrap(originalErr, FileNotFound, "could not open config file")

	// Check the error message format
	expectedMsg := fmt.Sprintf("[ERROR-%04d] could not open config file: file does not exist", FileNotFound)
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, wrappedErr.Error())
	}

	// Check that the error code can be extracted
	code := GetErrorCode(wrappedErr)
	if code != FileNotFound {
		t.Errorf("Expected error code %d, got %d", FileNotFound, code)
	}

	// Check that the user-friendly message can be retrieved
	friendlyMsg := GetErrorMessage(code)
	if friendlyMsg != "File not found" {
		t.Errorf("Expected friendly message 'File not found', got '%s'", friendlyMsg)
	}

	// Check that the error message contains the expected format with code and message
	if !strings.Contains(wrappedErr.Error(), fmt.Sprintf("[ERROR-%04d]", code)) {
		t.Errorf("Error message '%s' does not contain expected code format", wrappedErr.Error())
	}
}
