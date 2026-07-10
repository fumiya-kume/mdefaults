package push

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
)

type typedWrite struct {
	value     string
	valueType string
}

type fakeDefaultsWriter struct {
	writes          []string
	typedWrites     []typedWrite
	contexts        []context.Context
	writeErr        error
	typedWriteError error
}

func (f *fakeDefaultsWriter) Write(ctx context.Context, value string) error {
	f.contexts = append(f.contexts, ctx)
	f.writes = append(f.writes, value)
	return f.writeErr
}

func (f *fakeDefaultsWriter) WriteWithType(ctx context.Context, value, valueType string) error {
	f.contexts = append(f.contexts, ctx)
	f.typedWrites = append(f.typedWrites, typedWrite{value: value, valueType: valueType})
	return f.typedWriteError
}

func TestPushImplEmptyConfigs(t *testing.T) {
	factoryCalls := 0
	err := pushImpl(context.Background(), nil, func(_, _ string) defaultsWriter {
		factoryCalls++
		return &fakeDefaultsWriter{}
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("expected no writers to be created, got %d", factoryCalls)
	}
}

func TestPushImplRejectsNilValue(t *testing.T) {
	factoryCalls := 0
	err := pushImpl(context.Background(), []config.Config{
		{Domain: "com.example.app", Key: "missing", Value: nil},
	}, func(_, _ string) defaultsWriter {
		factoryCalls++
		return &fakeDefaultsWriter{}
	})

	if err == nil {
		t.Fatal("expected an error for a nil value")
	}
	if !strings.Contains(err.Error(), "value is nil") {
		t.Fatalf("expected nil value context, got %v", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("expected no writer to be created, got %d", factoryCalls)
	}
}

func TestPushImplDoesNotAccessSystemDefaults(t *testing.T) {
	const configCount = 1000
	writer := &fakeDefaultsWriter{}
	configs := make([]config.Config, configCount)
	for i := range configs {
		value := fmt.Sprintf("value%d", i)
		configs[i] = config.Config{
			Domain: fmt.Sprintf("domain%d", i),
			Key:    fmt.Sprintf("key%d", i),
			Value:  &value,
		}
	}

	factoryCalls := 0
	err := pushImpl(context.Background(), configs, func(_, _ string) defaultsWriter {
		factoryCalls++
		return writer
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if factoryCalls != configCount {
		t.Fatalf("expected %d writers to be requested, got %d", configCount, factoryCalls)
	}
	if len(writer.writes) != configCount {
		t.Fatalf("expected %d writes, got %d", configCount, len(writer.writes))
	}
}

func TestPushImplStopsAfterWriteError(t *testing.T) {
	plainErr := errors.New("plain write failed")
	plainWriter := &fakeDefaultsWriter{writeErr: plainErr}
	typedWriter := &fakeDefaultsWriter{}
	plainValue := "plain"
	typedValue := "1"
	typedWriterCreated := false

	err := pushImpl(context.Background(), []config.Config{
		{Domain: "com.example.app", Key: "plain", Value: &plainValue, Type: "string"},
		{Domain: "com.example.app", Key: "typed", Value: &typedValue, Type: "boolean"},
	}, func(_, key string) defaultsWriter {
		if key == "typed" {
			typedWriterCreated = true
			return typedWriter
		}
		return plainWriter
	})

	if !errors.Is(err, plainErr) {
		t.Fatalf("expected plain write error, got %v", err)
	}
	if typedWriterCreated {
		t.Fatal("expected push to stop before creating the typed writer")
	}
}

func TestPushImplValidatesAllConfigsBeforeWriting(t *testing.T) {
	validValue := "valid"
	invalidValue := "invalid"
	writer := &fakeDefaultsWriter{}
	factoryCalls := 0

	err := pushImpl(context.Background(), []config.Config{
		{Domain: "com.example.app", Key: "valid", Value: &validValue, Type: "string"},
		{Domain: "com.example.app", Key: "invalid", Value: &invalidValue, Type: "array"},
	}, func(_, _ string) defaultsWriter {
		factoryCalls++
		return writer
	})

	if err == nil {
		t.Fatal("expected unsupported type error")
	}
	if factoryCalls != 0 || len(writer.writes) != 0 {
		t.Fatalf("expected validation before writes, got %d factories and %d writes", factoryCalls, len(writer.writes))
	}
}

func TestPushImplRejectsInvalidBooleanBeforeWriting(t *testing.T) {
	validValue := "valid"
	invalidBoolean := "maybe"
	factoryCalls := 0

	err := pushImpl(context.Background(), []config.Config{
		{Domain: "com.example.app", Key: "valid", Value: &validValue, Type: "string"},
		{Domain: "com.example.app", Key: "invalid", Value: &invalidBoolean, Type: "boolean"},
	}, func(_, _ string) defaultsWriter {
		factoryCalls++
		return &fakeDefaultsWriter{}
	})

	if err == nil {
		t.Fatal("expected invalid boolean value error")
	}
	if !strings.Contains(err.Error(), "write com.example.app invalid") || !strings.Contains(err.Error(), "invalid boolean value") {
		t.Fatalf("expected contextual boolean value error, got %v", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("expected validation before writer creation, got %d calls", factoryCalls)
	}
}

func TestPushImplRejectsInvalidDomainOrKeyBeforeWriting(t *testing.T) {
	value := "value"
	testCases := []struct {
		name   string
		domain string
		key    string
	}{
		{name: "empty domain", key: "key"},
		{name: "empty key", domain: "com.example.app"},
		{name: "both empty"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factoryCalls := 0
			err := pushImpl(context.Background(), []config.Config{
				{Domain: tc.domain, Key: tc.key, Value: &value, Type: "string"},
			}, func(_, _ string) defaultsWriter {
				factoryCalls++
				return &fakeDefaultsWriter{}
			})

			if err == nil {
				t.Fatal("expected invalid domain or key error")
			}
			if !strings.Contains(err.Error(), "domain and key cannot be empty") {
				t.Fatalf("unexpected error: %v", err)
			}
			if factoryCalls != 0 {
				t.Fatalf("expected validation before writer creation, got %d calls", factoryCalls)
			}
		})
	}
}

func TestPushImplRejectsEveryUnsupportedTypeBeforeWriting(t *testing.T) {
	value := "value"
	unsupportedTypes := []string{"array", "dict", "data", "future-type"}

	for _, valueType := range unsupportedTypes {
		t.Run(valueType, func(t *testing.T) {
			factoryCalls := 0
			err := pushImpl(context.Background(), []config.Config{
				{Domain: "com.example.app", Key: "key", Value: &value, Type: valueType},
			}, func(_, _ string) defaultsWriter {
				factoryCalls++
				return &fakeDefaultsWriter{}
			})

			if err == nil {
				t.Fatalf("expected type %q to be rejected", valueType)
			}
			if !strings.Contains(err.Error(), "write com.example.app key") || !strings.Contains(err.Error(), valueType) {
				t.Fatalf("expected contextual type error, got %v", err)
			}
			if factoryCalls != 0 {
				t.Fatalf("expected preflight rejection, got %d factory calls", factoryCalls)
			}
		})
	}
}

func TestPushImplStopsAfterTypedWriteError(t *testing.T) {
	typedErr := errors.New("typed write failed")
	typedWriter := &fakeDefaultsWriter{typedWriteError: typedErr}
	plainWriter := &fakeDefaultsWriter{}
	typedValue := "true"
	plainValue := "later"
	plainWriterCreated := false

	err := pushImpl(context.Background(), []config.Config{
		{Domain: "com.example.app", Key: "typed", Value: &typedValue, Type: "boolean"},
		{Domain: "com.example.app", Key: "plain", Value: &plainValue, Type: "string"},
	}, func(_, key string) defaultsWriter {
		if key == "typed" {
			return typedWriter
		}
		plainWriterCreated = true
		return plainWriter
	})

	if !errors.Is(err, typedErr) {
		t.Fatalf("expected typed write error, got %v", err)
	}
	if !strings.Contains(err.Error(), "write com.example.app typed") {
		t.Fatalf("expected config context in error, got %v", err)
	}
	if len(typedWriter.typedWrites) != 1 {
		t.Fatalf("expected one typed write attempt, got %d", len(typedWriter.typedWrites))
	}
	if plainWriterCreated || len(plainWriter.writes) != 0 {
		t.Fatal("expected push to stop before the later plain write")
	}
}

func TestPushImplForwardsProvidedContext(t *testing.T) {
	type contextKey string
	const key contextKey = "test-key"
	ctx := context.WithValue(context.Background(), key, "test-value")
	writer := &fakeDefaultsWriter{}
	value := "value"

	err := pushImpl(ctx, []config.Config{
		{Domain: "com.example.app", Key: "key", Value: &value},
	}, func(_, _ string) defaultsWriter {
		return writer
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(writer.contexts) != 1 {
		t.Fatalf("expected one context, got %d", len(writer.contexts))
	}
	if actual := writer.contexts[0].Value(key); actual != "test-value" {
		t.Fatalf("expected forwarded context value, got %v", actual)
	}
}
