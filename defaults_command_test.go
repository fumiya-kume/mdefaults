package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestDefaultsCommandReadSuccess(t *testing.T) {
	defaults := &MockDefaultsCommandImpl{
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
	defaults := &MockDefaultsCommandImpl{
		ReadError: errors.New("read error"),
	}
	_, err := defaults.Read(context.Background())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestDefaultsCommandWriteSuccess(t *testing.T) {
	defaults := &MockDefaultsCommandImpl{
		WriteError: nil,
	}
	err := defaults.Write(context.Background(), "true")
	if err != nil {
		t.Errorf("Error writing defaults: %v", err)
	}
}

func TestDefaultsCommandWriteError(t *testing.T) {
	defaults := &MockDefaultsCommandImpl{
		WriteError: errors.New("write error"),
	}
	err := defaults.Write(context.Background(), "true")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

type MockDefaultsCommandImpl struct {
	ReadResult string
	ReadError  error
	WriteError error
	Exec       *exec.Cmd
}

func (m *MockDefaultsCommandImpl) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommandImpl) Write(ctx context.Context, value string) error {
	return m.WriteError
}
