package defaults

import (
	"context"
)

// MockDefaultsCommand is a mock implementation of the DefaultsCommand interface for testing.
type MockDefaultsCommand struct {
	ReadResult   string
	ReadError    error
	TypeResult   string
	TypeError    error
	WriteError   error
	DomainVal    string
	KeyVal       string
	WriteHistory []struct {
		Value     string
		ValueType string
	}
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) ReadType(ctx context.Context) (string, error) {
	return m.TypeResult, m.TypeError
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string, valueType string) error {
	if m.WriteHistory == nil {
		m.WriteHistory = make([]struct {
			Value     string
			ValueType string
		}, 0)
	}

	// Record the write operation for test verification
	m.WriteHistory = append(m.WriteHistory, struct {
		Value     string
		ValueType string
	}{Value: value, ValueType: valueType})

	return m.WriteError
}

func (m *MockDefaultsCommand) Domain() string {
	return m.DomainVal
}

func (m *MockDefaultsCommand) Key() string {
	return m.KeyVal
}
