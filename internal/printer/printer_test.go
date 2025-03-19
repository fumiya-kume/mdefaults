package printer

import (
	"testing"
)

func TestPrintSuccess(t *testing.T) {
	// This test simply verifies that the function doesn't panic
	// We can't easily capture colored output in tests
	PrintSuccess("Test success message")
}

func TestPrintError(t *testing.T) {
	// This test simply verifies that the function doesn't panic
	// We can't easily capture colored output in tests
	PrintError("Test error message")
}
