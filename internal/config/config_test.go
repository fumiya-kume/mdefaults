package config

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfigFile_Success(t *testing.T) {
	fs := &MockFileSystem{HomeDir: "/mock/home", StatError: nil, CreateErr: nil, ConfigFileContent: "com.apple.dock autohide 1\ncom.apple.finder ShowPathbar true\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	value1 := "1"
	value2 := "true"
	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "string"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2, Type: "string"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Fatalf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, config := range configs {
		if config.Domain != expectedConfigs[i].Domain {
			t.Errorf("Expected domain %s, got %s", expectedConfigs[i].Domain, config.Domain)
		}
		if config.Key != expectedConfigs[i].Key {
			t.Errorf("Expected key %s, got %s", expectedConfigs[i].Key, config.Key)
		}
		if *config.Value != *expectedConfigs[i].Value {
			t.Errorf("Expected value %s, got %s", *expectedConfigs[i].Value, *config.Value)
		}
		if config.Type != expectedConfigs[i].Type {
			t.Errorf("Expected type %s, got %s", expectedConfigs[i].Type, config.Type)
		}
	}
}

func TestReadConfigFile_Error(t *testing.T) {
	fs := &MockFileSystem{StatError: errors.New("read error")}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "read error") {
		t.Errorf("Expected error to contain 'read error', got %v", err)
	}
}

func TestGenerateConfigFileContent(t *testing.T) {
	value1 := "1"
	value2 := "true"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
	}

	expectedContent := "com.apple.dock autohide \"1\" string\ncom.apple.finder ShowPathbar \"true\" string\n"
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestGenerateConfigFileContent_Empty(t *testing.T) {
	configs := []Config{}

	expectedContent := ""
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestWriteConfigFile_Success(t *testing.T) {
	mockFS := &MockFileSystem{
		HomeDir:          "/mock/home",
		WriteFileContent: "",
		WriteFileErr:     nil,
	}
	value1 := "1"
	value2 := "true"

	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2},
	}

	err := WriteConfigFile(mockFS, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := "com.apple.dock autohide \"1\" string\ncom.apple.finder ShowPathbar \"true\" string\n"
	if mockFS.WriteFileContent != expectedContent {
		t.Errorf("Expected WriteFileContent %q, got %q", expectedContent, mockFS.WriteFileContent)
	}
}

func TestWriteConfigFile_Error(t *testing.T) {
	mockFS := &MockFileSystem{
		HomeDir:      "/mock/home",
		WriteFileErr: errors.New("write error"),
	}

	value1 := "1"
	configs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1},
	}

	err := WriteConfigFile(mockFS, configs)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, mockFS.WriteFileErr) {
		t.Errorf("Expected error %v, got %v", mockFS.WriteFileErr, err)
	}
}

func TestReadConfigFile_VariousInputTypes(t *testing.T) {
	// Test with various input types
	configContent := `com.apple.dock autohide 1
com.apple.finder ShowPathbar true
com.example.app floatValue 3.14
com.example.app negativeValue -42
com.example.app zeroValue 0
com.example.app specialChars !@#$%^&*()
com.example.app emptyValue
com.example.app longValue 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890
`
	fs := &MockFileSystem{HomeDir: "/mock/home", ConfigFileContent: configContent}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Expected values
	value1 := "1"
	value2 := "true"
	value3 := "3.14"
	value4 := "-42"
	value5 := "0"
	value6 := "!@#$%^&*()"
	value8 := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

	expectedConfigs := []Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "string"},
		{Domain: "com.apple.finder", Key: "ShowPathbar", Value: &value2, Type: "string"},
		{Domain: "com.example.app", Key: "floatValue", Value: &value3, Type: "string"},
		{Domain: "com.example.app", Key: "negativeValue", Value: &value4, Type: "string"},
		{Domain: "com.example.app", Key: "zeroValue", Value: &value5, Type: "string"},
		{Domain: "com.example.app", Key: "specialChars", Value: &value6, Type: "string"},
		{Domain: "com.example.app", Key: "emptyValue", Value: nil, Type: "string"},
		{Domain: "com.example.app", Key: "longValue", Value: &value8, Type: "string"},
	}

	if len(configs) != len(expectedConfigs) {
		t.Fatalf("Expected %d configs, got %d", len(expectedConfigs), len(configs))
	}

	for i, config := range configs {
		if config.Domain != expectedConfigs[i].Domain {
			t.Errorf("Expected domain %s, got %s", expectedConfigs[i].Domain, config.Domain)
		}
		if config.Key != expectedConfigs[i].Key {
			t.Errorf("Expected key %s, got %s", expectedConfigs[i].Key, config.Key)
		}
		if config.Value == nil || expectedConfigs[i].Value == nil {
			if config.Value != expectedConfigs[i].Value {
				t.Errorf("Expected value %v, got %v", expectedConfigs[i].Value, config.Value)
			}
		} else if *config.Value != *expectedConfigs[i].Value {
			t.Errorf("Expected value %s, got %s", *expectedConfigs[i].Value, *config.Value)
		}
		if config.Type != expectedConfigs[i].Type {
			t.Errorf("Expected type %s, got %s", expectedConfigs[i].Type, config.Type)
		}
	}
}

func TestGenerateConfigFileContent_VariousInputTypes(t *testing.T) {
	// Test with various input types
	value1 := "3.14"
	value2 := "-42"
	value3 := "0"
	value4 := "!@#$%^&*()"
	value5 := ""
	value6 := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

	configs := []Config{
		{Domain: "com.example.app", Key: "floatValue", Value: &value1},
		{Domain: "com.example.app", Key: "negativeValue", Value: &value2},
		{Domain: "com.example.app", Key: "zeroValue", Value: &value3},
		{Domain: "com.example.app", Key: "specialChars", Value: &value4},
		{Domain: "com.example.app", Key: "emptyValue", Value: &value5},
		{Domain: "com.example.app", Key: "longValue", Value: &value6},
	}

	expectedContent := "com.example.app floatValue \"3.14\" string\ncom.example.app negativeValue \"-42\" string\ncom.example.app zeroValue \"0\" string\ncom.example.app specialChars \"!@#$%^&*()\" string\ncom.example.app emptyValue \"\" string\ncom.example.app longValue \"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890\" string\n"
	content := GenerateConfigFileContent(configs)

	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestReadConfigFileWithoutValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.apple.dock autohide   \r\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected one config, got %d", len(configs))
	}
	if configs[0].Value != nil {
		t.Fatalf("expected a nil value, got %q", *configs[0].Value)
	}
}

func TestConfigFileRoundTripPreservesPullTarget(t *testing.T) {
	original := []Config{{Domain: "com.apple.dock", Key: "autohide"}}
	content := GenerateConfigFileContent(original)
	if content != "com.apple.dock autohide\n" {
		t.Fatalf("unexpected generated content: %q", content)
	}

	fs := &MockFileSystem{ConfigFileContent: content}
	actual, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(actual) != 1 || actual[0].Domain != original[0].Domain || actual[0].Key != original[0].Key || actual[0].Value != nil {
		t.Fatalf("pull target mismatch: expected %+v, got %+v", original, actual)
	}
}

func TestReadConfigFilePreservesLegacyTypedEmptyValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app emptyValue  string\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(configs) != 1 || configs[0].Value == nil || *configs[0].Value != "" || configs[0].Type != "string" {
		t.Fatalf("expected a typed empty value, got %+v", configs)
	}
}

func TestReadConfigFileDoesNotTreatUnicodeSeparatorAsLegacyEmptyValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key\u00a0string\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(configs) != 1 || configs[0].Value == nil || *configs[0].Value != "string" {
		t.Fatalf("expected string value, got %+v", configs)
	}
}

func TestReadConfigFileRejectsLegacyEmptyValueWithUnknownType(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key  future-type\n"}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("expected unknown legacy type error")
	}
}

func TestConfigFileRoundTripPreservesValues(t *testing.T) {
	spaced := "hello world"
	multiline := "line one\nline \"two\"\t"
	empty := ""
	date := "2026-07-11 01:23:45 +0000"
	original := []Config{
		{Domain: "com.example.app", Key: "spaced", Value: &spaced, Type: "string"},
		{Domain: "com.example.app", Key: "multiline", Value: &multiline, Type: "string"},
		{Domain: "com.example.app", Key: "empty", Value: &empty, Type: "string"},
		{Domain: "com.example.app", Key: "date", Value: &date, Type: "date"},
	}
	fs := &MockFileSystem{ConfigFileContent: GenerateConfigFileContent(original)}

	actual, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(actual) != len(original) {
		t.Fatalf("expected %d configs, got %d", len(original), len(actual))
	}
	for i := range original {
		if actual[i].Domain != original[i].Domain || actual[i].Key != original[i].Key || actual[i].Type != original[i].Type {
			t.Fatalf("config %d metadata mismatch: expected %+v, got %+v", i, original[i], actual[i])
		}
		if actual[i].Value == nil || *actual[i].Value != *original[i].Value {
			t.Fatalf("config %d value mismatch: expected %q, got %v", i, *original[i].Value, actual[i].Value)
		}
	}
}

func TestReadConfigFilePreservesLegacyTypedValueWithSpaces(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app typed hello world string\n"}

	configs, err := ReadConfigFile(fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected one config, got %d", len(configs))
	}
	if configs[0].Value == nil || *configs[0].Value != "hello world" {
		t.Fatalf("expected %q, got %v", "hello world", configs[0].Value)
	}
	if configs[0].Type != "string" {
		t.Fatalf("expected string type, got %q", configs[0].Type)
	}
}

func TestReadConfigFileRejectsAmbiguousUnquotedValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key hello world\n"}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("expected ambiguous value error")
	}
}

func TestReadConfigFileRejectsMalformedQuotedValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key \"unterminated string\n"}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("expected malformed quoted value error")
	}
}

func TestReadConfigFileRejectsUnknownType(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key \"value\" future-type\n"}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
}

func TestReadConfigFileRequiresSeparatorAfterQuotedValue(t *testing.T) {
	fs := &MockFileSystem{ConfigFileContent: "com.example.app key \"value\"integer\n"}

	_, err := ReadConfigFile(fs)
	if err == nil {
		t.Fatal("expected missing separator error")
	}
}

func TestParseConfigLineRejectsMalformedStructure(t *testing.T) {
	testCases := []struct {
		name        string
		line        string
		errorSubstr string
	}{
		{name: "missing domain", line: "", errorSubstr: "missing domain"},
		{name: "missing key", line: "com.example.app", errorSubstr: "missing key"},
		{name: "extra fields after quoted value", line: `com.example.app key "value" string extra`, errorSubstr: "unexpected fields"},
		{name: "invalid quoted escape", line: `com.example.app key "\xZZ" string`, errorSubstr: "invalid quoted value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfigLine(tc.line)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.errorSubstr) {
				t.Fatalf("expected error containing %q, got %v", tc.errorSubstr, err)
			}
		})
	}
}

func TestParseConfigLineAcceptsQuotedValueWithoutExplicitType(t *testing.T) {
	actual, err := parseConfigLine(`com.example.app key "hello world"`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if actual.Domain != "com.example.app" || actual.Key != "key" || actual.Type != "string" {
		t.Fatalf("unexpected config metadata: %+v", actual)
	}
	if actual.Value == nil || *actual.Value != "hello world" {
		t.Fatalf("expected quoted string value, got %v", actual.Value)
	}
}

func FuzzConfigFileRoundTrip(f *testing.F) {
	seedValues := []string{
		"",
		"plain",
		"hello world",
		"line one\nline \"two\"\t\\",
		"日本語とemoji🚀",
		string([]byte{0, 1, 0x7f, 0xff}),
	}
	for i, value := range seedValues {
		f.Add(value, uint8(i))
	}

	valueTypes := []string{"string", "integer", "boolean", "float", "date", "array", "dict", "data"}
	f.Fuzz(func(t *testing.T, value string, typeIndex uint8) {
		valueType := valueTypes[int(typeIndex)%len(valueTypes)]
		original := Config{
			Domain: "com.example.app",
			Key:    "roundTrip",
			Value:  &value,
			Type:   valueType,
		}
		fs := &MockFileSystem{ConfigFileContent: GenerateConfigFileContent([]Config{original})}

		actual, err := ReadConfigFile(fs)
		if err != nil {
			t.Fatalf("round trip failed for type %q and value %q: %v", valueType, value, err)
		}
		if len(actual) != 1 {
			t.Fatalf("expected one config, got %d", len(actual))
		}
		if actual[0].Domain != original.Domain || actual[0].Key != original.Key || actual[0].Type != original.Type {
			t.Fatalf("metadata mismatch: expected %+v, got %+v", original, actual[0])
		}
		if actual[0].Value == nil || *actual[0].Value != value {
			t.Fatalf("value mismatch: expected %q, got %v", value, actual[0].Value)
		}
	})
}
