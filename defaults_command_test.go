package main

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
	err := defaults.Write(context.Background(), "true")
	if err != nil {
		t.Errorf("Error writing defaults: %v", err)
	}
}

func TestDefaultsCommandWriteError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: errors.New("write error"),
	}
	err := defaults.Write(context.Background(), "true")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

type MockDefaultsCommand struct {
	ReadResult string
	ReadError  error
	WriteError error
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string) error {
	return m.WriteError
}
