package defaults

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type runnerCall struct {
	ctx  context.Context
	name string
	args []string
}

type stubRunner struct {
	output  []byte
	err     error
	handler commandRunner
	calls   []runnerCall
}

func (s *stubRunner) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	s.calls = append(s.calls, runnerCall{
		ctx:  ctx,
		name: name,
		args: slices.Clone(args),
	})
	if s.handler != nil {
		return s.handler(ctx, name, args...)
	}
	return s.output, s.err
}

func newCommandForTest(runner *stubRunner) *DefaultsCommandImpl {
	return newDefaultsCommandImpl("com.example.app", "setting", runner.run)
}

func requireSingleCall(t *testing.T, runner *stubRunner, wantArgs ...string) runnerCall {
	t.Helper()
	if len(runner.calls) != 1 {
		t.Fatalf("expected one runner call, got %d", len(runner.calls))
	}
	call := runner.calls[0]
	if call.name != "defaults" {
		t.Fatalf("expected executable %q, got %q", "defaults", call.name)
	}
	if !slices.Equal(call.args, wantArgs) {
		t.Fatalf("expected arguments %q, got %q", wantArgs, call.args)
	}
	return call
}

func TestNewDefaultsCommandImpl(t *testing.T) {
	command := NewDefaultsCommandImpl("com.example.app", "setting")

	if command.Domain() != "com.example.app" {
		t.Fatalf("unexpected domain: %q", command.Domain())
	}
	if command.Key() != "setting" {
		t.Fatalf("unexpected key: %q", command.Key())
	}
	if command.run == nil {
		t.Fatal("expected the production command runner to be configured")
	}
}

func TestRunCommandReturnsExecutionError(t *testing.T) {
	missingExecutable := filepath.Join(t.TempDir(), "missing-command")

	if _, err := runCommand(context.Background(), missingExecutable); err == nil {
		t.Fatal("expected an execution error")
	}
}

func TestDefaultsCommandImplReadPreservesOutputAndArguments(t *testing.T) {
	runner := &stubRunner{output: []byte("value with spaces\n")}
	command := newCommandForTest(runner)

	actual, err := command.Read(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actual != "value with spaces\n" {
		t.Fatalf("expected output to be preserved, got %q", actual)
	}
	requireSingleCall(t, runner, "read", "com.example.app", "setting")
}

func TestDefaultsCommandImplReadTypeMapsScalarTypesAndArguments(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
	}{
		{name: "integer", output: "Type is integer\n", expected: "integer"},
		{name: "boolean", output: "Type is boolean\n", expected: "boolean"},
		{name: "string", output: "Type is string\n", expected: "string"},
		{name: "float", output: "Type is float\n", expected: "float"},
		{name: "real", output: "Type is real\n", expected: "float"},
		{name: "date", output: "Type is date\n", expected: "date"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &stubRunner{output: []byte(tc.output)}
			command := newCommandForTest(runner)

			actual, err := command.ReadType(context.Background())
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if actual != tc.expected {
				t.Fatalf("expected type %q, got %q", tc.expected, actual)
			}
			requireSingleCall(t, runner, "read-type", "com.example.app", "setting")
		})
	}
}

func TestDefaultsCommandImplWritePreservesValueAsOneArgument(t *testing.T) {
	values := []string{"", "plain", "value with spaces", "line one\nline two"}

	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			runner := &stubRunner{}
			command := newCommandForTest(runner)

			if err := command.Write(context.Background(), value); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			requireSingleCall(t, runner, "write", "com.example.app", "setting", value)
		})
	}
}

func TestDefaultsCommandImplWriteWithTypeUsesExactScalarArguments(t *testing.T) {
	testCases := []struct {
		name      string
		value     string
		valueType string
		expected  []string
	}{
		{
			name:      "legacy empty type",
			value:     "value with spaces",
			valueType: "",
			expected:  []string{"write", "com.example.app", "setting", "value with spaces"},
		},
		{
			name:      "string",
			value:     "value with spaces",
			valueType: "string",
			expected:  []string{"write", "com.example.app", "setting", "value with spaces"},
		},
		{
			name:      "integer",
			value:     "42",
			valueType: "integer",
			expected:  []string{"write", "com.example.app", "setting", "-int", "42"},
		},
		{
			name:      "boolean true",
			value:     "true",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "true"},
		},
		{
			name:      "boolean false",
			value:     "false",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "false"},
		},
		{
			name:      "boolean numeric true",
			value:     "1",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "true"},
		},
		{
			name:      "boolean numeric false",
			value:     "0",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "false"},
		},
		{
			name:      "boolean yes",
			value:     "yes",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "true"},
		},
		{
			name:      "boolean no",
			value:     "no",
			valueType: "boolean",
			expected:  []string{"write", "com.example.app", "setting", "-bool", "false"},
		},
		{
			name:      "float",
			value:     "3.14",
			valueType: "float",
			expected:  []string{"write", "com.example.app", "setting", "-float", "3.14"},
		},
		{
			name:      "date",
			value:     "2026-07-11 01:23:45 +0000",
			valueType: "date",
			expected:  []string{"write", "com.example.app", "setting", "-date", "2026-07-11 01:23:45 +0000"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &stubRunner{}
			command := newCommandForTest(runner)

			if err := command.WriteWithType(context.Background(), tc.value, tc.valueType); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			requireSingleCall(t, runner, tc.expected...)
		})
	}
}

func TestDefaultsCommandImplPropagatesRunnerErrors(t *testing.T) {
	runnerError := errors.New("runner failed")
	testCases := []struct {
		name     string
		invoke   func(*DefaultsCommandImpl) error
		expected []string
	}{
		{
			name: "read",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.Read(context.Background())
				return err
			},
			expected: []string{"read", "com.example.app", "setting"},
		},
		{
			name: "read type",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.ReadType(context.Background())
				return err
			},
			expected: []string{"read-type", "com.example.app", "setting"},
		},
		{
			name: "write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.Write(context.Background(), "value")
			},
			expected: []string{"write", "com.example.app", "setting", "value"},
		},
		{
			name: "typed write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.WriteWithType(context.Background(), "42", "integer")
			},
			expected: []string{"write", "com.example.app", "setting", "-int", "42"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &stubRunner{err: runnerError}
			command := newCommandForTest(runner)

			if err := tc.invoke(command); !errors.Is(err, runnerError) {
				t.Fatalf("expected runner error, got %v", err)
			}
			requireSingleCall(t, runner, tc.expected...)
		})
	}
}

func TestDefaultsCommandImplReadTypeRejectsMalformedOutput(t *testing.T) {
	runner := &stubRunner{output: []byte("integer\n")}
	command := newCommandForTest(runner)

	if _, err := command.ReadType(context.Background()); err == nil {
		t.Fatal("expected malformed type output to fail")
	}
	requireSingleCall(t, runner, "read-type", "com.example.app", "setting")
}

func TestDefaultsCommandImplRejectsInvalidDomainOrKeyBeforeRunning(t *testing.T) {
	invalidCommands := []struct {
		name   string
		domain string
		key    string
	}{
		{name: "empty domain", domain: "", key: "setting"},
		{name: "empty key", domain: "com.example.app", key: ""},
		{name: "both empty", domain: "", key: ""},
	}
	operations := []struct {
		name   string
		invoke func(*DefaultsCommandImpl) error
	}{
		{
			name: "read",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.Read(context.Background())
				return err
			},
		},
		{
			name: "read type",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.ReadType(context.Background())
				return err
			},
		},
		{
			name: "write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.Write(context.Background(), "value")
			},
		},
		{
			name: "typed write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.WriteWithType(context.Background(), "42", "integer")
			},
		},
	}

	for _, invalid := range invalidCommands {
		for _, operation := range operations {
			t.Run(invalid.name+"/"+operation.name, func(t *testing.T) {
				runner := &stubRunner{}
				command := newDefaultsCommandImpl(invalid.domain, invalid.key, runner.run)

				if err := operation.invoke(command); err == nil {
					t.Fatal("expected validation error")
				}
				if len(runner.calls) != 0 {
					t.Fatalf("expected validation before execution, got %d calls", len(runner.calls))
				}
			})
		}
	}
}

func TestDefaultsCommandImplRejectsUnsupportedTypesBeforeRunning(t *testing.T) {
	for _, valueType := range []string{"array", "dict", "data", "future-type"} {
		t.Run(valueType, func(t *testing.T) {
			runner := &stubRunner{}
			command := newCommandForTest(runner)

			if err := command.WriteWithType(context.Background(), "value", valueType); err == nil {
				t.Fatal("expected unsupported type error")
			}
			if len(runner.calls) != 0 {
				t.Fatalf("expected validation before execution, got %d calls", len(runner.calls))
			}
		})
	}
}

func TestDefaultsCommandImplRejectsInvalidBooleanValuesBeforeRunning(t *testing.T) {
	for _, value := range []string{"", "2", "on", "off", "maybe", " true "} {
		t.Run(value, func(t *testing.T) {
			runner := &stubRunner{}
			command := newCommandForTest(runner)

			err := command.WriteWithType(context.Background(), value, "boolean")
			if err == nil {
				t.Fatal("expected invalid boolean value error")
			}
			if !strings.Contains(err.Error(), "invalid boolean value") {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(runner.calls) != 0 {
				t.Fatalf("expected validation before execution, got %d calls", len(runner.calls))
			}
		})
	}
}

func TestDefaultsCommandImplPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	testCases := []struct {
		name   string
		invoke func(*DefaultsCommandImpl) error
	}{
		{
			name: "read",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.Read(ctx)
				return err
			},
		},
		{
			name: "read type",
			invoke: func(command *DefaultsCommandImpl) error {
				_, err := command.ReadType(ctx)
				return err
			},
		},
		{
			name: "write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.Write(ctx, "value")
			},
		},
		{
			name: "typed write",
			invoke: func(command *DefaultsCommandImpl) error {
				return command.WriteWithType(ctx, "42", "integer")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &stubRunner{
				handler: func(received context.Context, _ string, _ ...string) ([]byte, error) {
					return nil, received.Err()
				},
			}
			command := newCommandForTest(runner)

			if err := tc.invoke(command); !errors.Is(err, context.Canceled) {
				t.Fatalf("expected context cancellation, got %v", err)
			}
			if len(runner.calls) != 1 || runner.calls[0].ctx != ctx {
				t.Fatal("expected the caller context to be passed to the runner")
			}
		})
	}
}

func TestMockDefaultsCommandContract(t *testing.T) {
	readErr := errors.New("read failed")
	readTypeErr := errors.New("read type failed")
	writeErr := errors.New("write failed")
	writeTypeErr := errors.New("typed write failed")
	mock := &MockDefaultsCommand{
		ReadResult:     "value",
		ReadError:      readErr,
		ReadTypeResult: "integer",
		ReadTypeError:  readTypeErr,
		WriteError:     writeErr,
		WriteTypeError: writeTypeErr,
		DomainVal:      "com.example.app",
		KeyVal:         "setting",
	}

	if result, err := mock.Read(context.Background()); result != "value" || !errors.Is(err, readErr) {
		t.Fatalf("unexpected read result %q and error %v", result, err)
	}
	if result, err := mock.ReadType(context.Background()); result != "integer" || !errors.Is(err, readTypeErr) {
		t.Fatalf("unexpected read-type result %q and error %v", result, err)
	}
	if err := mock.Write(context.Background(), "value"); !errors.Is(err, writeErr) {
		t.Fatalf("unexpected write error: %v", err)
	}
	if err := mock.WriteWithType(context.Background(), "42", "integer"); !errors.Is(err, writeTypeErr) {
		t.Fatalf("unexpected typed-write error: %v", err)
	}
	if mock.Domain() != "com.example.app" || mock.Key() != "setting" {
		t.Fatalf("unexpected identity: %q %q", mock.Domain(), mock.Key())
	}

	mock.ReadTypeResult = ""
	mock.ReadTypeError = nil
	if result, err := mock.ReadType(context.Background()); result != "string" || err != nil {
		t.Fatalf("expected fallback string type, got %q and %v", result, err)
	}
}
