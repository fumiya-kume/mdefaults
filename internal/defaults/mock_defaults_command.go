package defaults

import (
	"context"
)

// MockDefaultsCommand is a mock implementation of the DefaultsCommand interface for testing.
type MockDefaultsCommand struct {
	ReadResult      string
	ReadError       error
	ReadTypeResult  string
	ReadTypeError   error
	WriteError      error
	WriteTypeError  error
	DomainVal       string
	KeyVal          string
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) ReadType(ctx context.Context) (string, error) {
	if m.ReadTypeResult == "" {
		return "string", m.ReadTypeError
	}
	return m.ReadTypeResult, m.ReadTypeError
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string) error {
	return m.WriteError
}

func (m *MockDefaultsCommand) WriteWithType(ctx context.Context, value string, valueType string) error {
	return m.WriteTypeError
}

func (m *MockDefaultsCommand) Domain() string {
	return m.DomainVal
}

func (m *MockDefaultsCommand) Key() string {
	return m.KeyVal
}
