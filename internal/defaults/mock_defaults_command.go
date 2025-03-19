package defaults

import (
	"context"
)

// MockDefaultsCommand is a mock implementation of the DefaultsCommand interface for testing.
type MockDefaultsCommand struct {
	ReadResult string
	ReadError  error
	WriteError error
	DomainVal  string
	KeyVal     string
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string) error {
	return m.WriteError
}

func (m *MockDefaultsCommand) Domain() string {
	return m.DomainVal
}

func (m *MockDefaultsCommand) Key() string {
	return m.KeyVal
}
