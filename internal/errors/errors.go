package errors

import (
	"fmt"
)

// ErrorCode represents a specific error type in the application
type ErrorCode int

const (
	// Unknown represents an unknown error
	Unknown ErrorCode = iota
	// InvalidArgument represents an error due to invalid arguments
	InvalidArgument

	// FileNotFound represents an error when a file is not found
	FileNotFound
	// FilePermissionDenied represents an error when file access is denied
	FilePermissionDenied
	// FileReadError represents an error when reading a file
	FileReadError
	// FileWriteError represents an error when writing to a file
	FileWriteError

	// DefaultsEmptyDomainOrKey represents an error when domain or key is empty
	DefaultsEmptyDomainOrKey
	// DefaultsReadError represents an error when reading defaults
	DefaultsReadError
	// DefaultsWriteError represents an error when writing defaults
	DefaultsWriteError
	// DefaultsCommandExecutionError represents an error when executing defaults command
	DefaultsCommandExecutionError

	// ConfigParseError represents an error when parsing configuration
	ConfigParseError
	// ConfigWriteError represents an error when writing configuration
	ConfigWriteError
)

// Error represents an application error with a code and message
type Error struct {
	Code    ErrorCode
	Message string
	Err     error // Original error, if any
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[ERROR-%04d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[ERROR-%04d] %s", e.Code, e.Message)
}

// Unwrap returns the original error
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error with the given code and message
func New(code ErrorCode, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a code and message
func Wrap(err error, code ErrorCode, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return 0
	}

	var appErr *Error
	if e, ok := err.(*Error); ok {
		appErr = e
	} else {
		// If it's not our error type, return Unknown
		return Unknown
	}

	return appErr.Code
}

// GetErrorMessage returns a user-friendly message for an error code
func GetErrorMessage(code ErrorCode) string {
	switch code {
	case InvalidArgument:
		return "Invalid argument provided"
	case FileNotFound:
		return "File not found"
	case FilePermissionDenied:
		return "Permission denied when accessing file"
	case FileReadError:
		return "Error reading file"
	case FileWriteError:
		return "Error writing file"
	case DefaultsEmptyDomainOrKey:
		return "Domain and key cannot be empty"
	case DefaultsReadError:
		return "Error reading defaults"
	case DefaultsWriteError:
		return "Error writing defaults"
	case DefaultsCommandExecutionError:
		return "Error executing defaults command"
	case ConfigParseError:
		return "Error parsing configuration"
	case ConfigWriteError:
		return "Error writing configuration"
	default:
		return "Unknown error"
	}
}
