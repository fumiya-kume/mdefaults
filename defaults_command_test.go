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

func TestDefaultsCommandReadTypeSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadTypeResult: "boolean",
		ReadTypeError:  nil,
	}
	result, err := defaults.ReadType(context.Background())
	if err != nil {
		t.Errorf("Error reading defaults type: %v", err)
	}
	if result != "boolean" {
		t.Errorf("Expected result to be 'boolean' but got %s", result)
	}
}

func TestDefaultsCommandReadTypeError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		ReadTypeError: errors.New("read type error"),
	}
	result, err := defaults.ReadType(context.Background())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if result != "string" {
		t.Errorf("Expected default type 'string' on error, but got %s", result)
	}
}

func TestDefaultsCommandWriteSuccess(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: nil,
	}
	err := defaults.Write(context.Background(), "true", "boolean")
	if err != nil {
		t.Errorf("Error writing defaults: %v", err)
	}
}

func TestDefaultsCommandWriteError(t *testing.T) {
	defaults := &MockDefaultsCommand{
		WriteError: errors.New("write error"),
	}
	err := defaults.Write(context.Background(), "true", "boolean")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

type MockDefaultsCommand struct {
	ReadResult     string
	ReadError      error
	ReadTypeResult string
	ReadTypeError  error
	WriteError     error
	domain         string
	key            string
}

func (m *MockDefaultsCommand) Read(ctx context.Context) (string, error) {
	return m.ReadResult, m.ReadError
}

func (m *MockDefaultsCommand) ReadType(ctx context.Context) (string, error) {
	if m.ReadTypeError != nil {
		return "string", m.ReadTypeError
	}
	if m.ReadTypeResult == "" {
		return "string", nil
	}
	return m.ReadTypeResult, nil
}

func (m *MockDefaultsCommand) Write(ctx context.Context, value string, valueType string) error {
	return m.WriteError
}

func (m *MockDefaultsCommand) Domain() string {
	return m.domain
}

func (m *MockDefaultsCommand) Key() string {
	return m.key
}
