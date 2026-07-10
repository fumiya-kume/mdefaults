package push

import (
	"context"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
)

func TestPushImplUsesTypedWrites(t *testing.T) {
	writers := map[string]*fakeDefaultsWriter{
		"autohide":       {},
		"ClickThreshold": {},
		"scale":          {},
		"timestamp":      {},
	}
	booleanValue := "1"
	integerValue := "2"
	floatValue := "1.25"
	dateValue := "2026-07-11 01:23:45 +0000"
	configs := []config.Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &booleanValue, Type: "boolean"},
		{Domain: "com.apple.trackpad", Key: "ClickThreshold", Value: &integerValue, Type: "integer"},
		{Domain: "com.example.app", Key: "scale", Value: &floatValue, Type: "float"},
		{Domain: "com.example.app", Key: "timestamp", Value: &dateValue, Type: "date"},
	}

	err := pushImpl(context.Background(), configs, func(_, key string) defaultsWriter {
		return writers[key]
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := writers["autohide"].typedWrites; len(got) != 1 || got[0] != (typedWrite{value: "1", valueType: "boolean"}) {
		t.Fatalf("unexpected boolean writes: %+v", got)
	}
	if got := writers["ClickThreshold"].typedWrites; len(got) != 1 || got[0] != (typedWrite{value: "2", valueType: "integer"}) {
		t.Fatalf("unexpected integer writes: %+v", got)
	}
	if got := writers["scale"].typedWrites; len(got) != 1 || got[0] != (typedWrite{value: floatValue, valueType: "float"}) {
		t.Fatalf("unexpected float writes: %+v", got)
	}
	if got := writers["timestamp"].typedWrites; len(got) != 1 || got[0] != (typedWrite{value: dateValue, valueType: "date"}) {
		t.Fatalf("unexpected date writes: %+v", got)
	}
}

func TestPushImplUsesPlainWriteForStringAndEmptyType(t *testing.T) {
	writers := map[string]*fakeDefaultsWriter{
		"explicit": {},
		"legacy":   {},
	}
	explicitValue := "explicit value"
	legacyValue := "legacy value"
	configs := []config.Config{
		{Domain: "com.example.app", Key: "explicit", Value: &explicitValue, Type: "string"},
		{Domain: "com.example.app", Key: "legacy", Value: &legacyValue},
	}

	err := pushImpl(context.Background(), configs, func(_, key string) defaultsWriter {
		return writers[key]
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := writers["explicit"].writes; len(got) != 1 || got[0] != explicitValue {
		t.Fatalf("unexpected explicit string writes: %v", got)
	}
	if got := writers["legacy"].writes; len(got) != 1 || got[0] != legacyValue {
		t.Fatalf("unexpected legacy string writes: %v", got)
	}
}
