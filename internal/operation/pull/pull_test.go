package pull

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

type tracingDefaultsCommand struct {
	domain       string
	key          string
	readResult   string
	readType     string
	readErr      error
	readTypeErr  error
	sharedTrace  *[]string
	contextTrace *[]context.Context
}

func (c *tracingDefaultsCommand) Read(ctx context.Context) (string, error) {
	*c.sharedTrace = append(*c.sharedTrace, c.key+":read")
	if c.contextTrace != nil {
		*c.contextTrace = append(*c.contextTrace, ctx)
	}
	return c.readResult, c.readErr
}

func (c *tracingDefaultsCommand) ReadType(ctx context.Context) (string, error) {
	*c.sharedTrace = append(*c.sharedTrace, c.key+":read-type")
	if c.contextTrace != nil {
		*c.contextTrace = append(*c.contextTrace, ctx)
	}
	if c.readType == "" {
		return "string", c.readTypeErr
	}
	return c.readType, c.readTypeErr
}

func (c *tracingDefaultsCommand) Write(context.Context, string) error {
	panic("unexpected write")
}

func (c *tracingDefaultsCommand) WriteWithType(context.Context, string, string) error {
	panic("unexpected typed write")
}

func (c *tracingDefaultsCommand) Domain() string { return c.domain }

func (c *tracingDefaultsCommand) Key() string { return c.key }

func TestPull_Success(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" {
		t.Errorf("Expected domain 'com.apple.dock', got %s", updatedConfigs[0].Domain)
	}
	if updatedConfigs[0].Key != "autohide" {
		t.Errorf("Expected key 'autohide', got %s", updatedConfigs[0].Key)
	}
	if *updatedConfigs[0].Value != "1" {
		t.Errorf("Expected value '1', got %s", *updatedConfigs[0].Value)
	}
}

func TestPull_ReadError(t *testing.T) {
	readErr := errors.New("read error")
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadError: readErr},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}
	if updatedConfigs != nil {
		t.Fatalf("expected no partial configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MultipleConfigs(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.finder", KeyVal: "ShowPathbar", ReadResult: "true"},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(updatedConfigs))
	}
	if updatedConfigs[0].Domain != "com.apple.dock" || updatedConfigs[0].Key != "autohide" || *updatedConfigs[0].Value != "1" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[0])
	}
	if updatedConfigs[1].Domain != "com.apple.finder" || updatedConfigs[1].Key != "ShowPathbar" || *updatedConfigs[1].Value != "true" {
		t.Errorf("Unexpected config: %+v", updatedConfigs[1])
	}
}

func TestPull_EmptyConfigs(t *testing.T) {
	defaultsCmds := []defaults.DefaultsCommand{}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(updatedConfigs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(updatedConfigs))
	}
}

func TestPull_MixedResults(t *testing.T) {
	readErr := errors.New("read error")
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.dock", KeyVal: "autohide", ReadResult: "1"},
		&defaults.MockDefaultsCommand{DomainVal: "com.apple.finder", KeyVal: "ShowPathbar", ReadError: readErr},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}
	if updatedConfigs != nil {
		t.Fatalf("expected no partial configs, got %d", len(updatedConfigs))
	}
}

func TestPull_InvalidConfig(t *testing.T) {
	readErr := errors.New("invalid config")
	defaultsCmds := []defaults.DefaultsCommand{
		&defaults.MockDefaultsCommand{DomainVal: "", KeyVal: "", ReadError: readErr},
	}

	updatedConfigs, err := PullImpl(defaultsCmds)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected invalid config error, got %v", err)
	}
	if updatedConfigs != nil {
		t.Fatalf("expected no partial configs, got %d", len(updatedConfigs))
	}
}

func TestPullImplPreservesValueNewlinesAndRemovesOnlyCommandTerminator(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
	}{
		{name: "no terminator", output: "value", expected: "value"},
		{name: "one terminator", output: "value\n", expected: "value"},
		{name: "value ending in newline", output: "line one\nline two\n\n", expected: "line one\nline two\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commands := []defaults.DefaultsCommand{
				&defaults.MockDefaultsCommand{
					DomainVal:  "com.example.app",
					KeyVal:     "multiline",
					ReadResult: tc.output,
				},
			}

			actual, err := PullImpl(commands)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(actual) != 1 || actual[0].Value == nil || *actual[0].Value != tc.expected {
				t.Fatalf("expected value %q, got %+v", tc.expected, actual)
			}
		})
	}
}

func TestPullImplStopsAfterReadFailure(t *testing.T) {
	readErr := errors.New("read failed")
	trace := []string{}
	commands := []defaults.DefaultsCommand{
		&tracingDefaultsCommand{
			domain:      "com.example.app",
			key:         "first",
			readErr:     readErr,
			sharedTrace: &trace,
		},
		&tracingDefaultsCommand{
			domain:      "com.example.app",
			key:         "second",
			sharedTrace: &trace,
		},
	}

	actual, err := PullImpl(commands)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}
	if actual != nil {
		t.Fatalf("expected no partial configs, got %+v", actual)
	}
	if !reflect.DeepEqual(trace, []string{"first:read"}) {
		t.Fatalf("expected immediate stop after read failure, got trace %v", trace)
	}
	if !strings.Contains(err.Error(), "read com.example.app first") {
		t.Fatalf("expected command context in error, got %v", err)
	}
}

func TestPullImplStopsAfterReadTypeFailure(t *testing.T) {
	typeErr := errors.New("read type failed")
	trace := []string{}
	commands := []defaults.DefaultsCommand{
		&tracingDefaultsCommand{
			domain:      "com.example.app",
			key:         "first",
			readResult:  "value\n",
			readTypeErr: typeErr,
			sharedTrace: &trace,
		},
		&tracingDefaultsCommand{
			domain:      "com.example.app",
			key:         "second",
			sharedTrace: &trace,
		},
	}

	actual, err := PullImpl(commands)
	if !errors.Is(err, typeErr) {
		t.Fatalf("expected type error, got %v", err)
	}
	if actual != nil {
		t.Fatalf("expected no partial configs, got %+v", actual)
	}
	wantTrace := []string{"first:read", "first:read-type"}
	if !reflect.DeepEqual(trace, wantTrace) {
		t.Fatalf("expected stop after type failure, got trace %v", trace)
	}
	if !strings.Contains(err.Error(), "read type com.example.app first") {
		t.Fatalf("expected command context in error, got %v", err)
	}
}

func TestPullImplUsesLiveContextsForReadAndReadType(t *testing.T) {
	trace := []string{}
	contexts := []context.Context{}
	commands := []defaults.DefaultsCommand{
		&tracingDefaultsCommand{
			domain:       "com.example.app",
			key:          "key",
			readResult:   "value\n",
			sharedTrace:  &trace,
			contextTrace: &contexts,
		},
	}

	if _, err := PullImpl(commands); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(contexts) != 2 {
		t.Fatalf("expected two contexts, got %d", len(contexts))
	}
	for i, ctx := range contexts {
		if ctx == nil || ctx.Err() != nil {
			t.Fatalf("expected live context at call %d, got %v", i, ctx)
		}
	}
}
