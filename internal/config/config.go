package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Config represents a configuration entry with domain, key, value, and type.
type Config struct {
	Domain string
	Key    string
	Value  *string
	Type   string
}

// ConfigFilePath is the default path for the configuration file.
var ConfigFilePath = filepath.Join(os.Getenv("HOME"), ".mdefaults")

// FileSystemReader is a minimal interface for file system operations needed by config package
type FileSystemReader interface {
	ReadFile(name string) (string, error)
	WriteFile(name string, content string) error
}

// ReadConfigFile reads the configuration file and returns a slice of Config.
func ReadConfigFile(fs FileSystemReader) ([]Config, error) {
	content, err := fs.ReadFile(ConfigFilePath)
	if err != nil {
		return nil, err
	}
	configs := []Config{}
	for lineNumber, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		config, err := parseConfigLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse config line %d: %w", lineNumber+1, err)
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func parseConfigLine(line string) (Config, error) {
	domain, remainder, ok := takeField(line)
	if !ok {
		return Config{}, fmt.Errorf("missing domain")
	}
	key, valuePart, ok := takeField(remainder)
	if !ok {
		return Config{}, fmt.Errorf("missing key for domain %q", domain)
	}

	parsed := Config{Domain: domain, Key: key, Type: "string"}
	if valuePart == "" {
		return parsed, nil
	}

	trimmedValuePart := strings.TrimSpace(valuePart)
	if trimmedValuePart == "" {
		return parsed, nil
	}

	if strings.HasPrefix(valuePart, "  ") {
		if !isKnownConfigType(trimmedValuePart) {
			return Config{}, fmt.Errorf("ambiguous legacy empty value with unknown type %q", trimmedValuePart)
		}
		emptyValue := ""
		parsed.Value = &emptyValue
		parsed.Type = trimmedValuePart
		return parsed, nil
	}

	if strings.HasPrefix(trimmedValuePart, "\"") {
		value, typePart, err := takeQuotedValue(trimmedValuePart)
		if err != nil {
			return Config{}, err
		}
		if typePart != "" && strings.TrimLeftFunc(typePart, unicode.IsSpace) == typePart {
			return Config{}, fmt.Errorf("quoted value and type must be separated by whitespace")
		}
		parsed.Value = &value

		typeFields := strings.Fields(typePart)
		if len(typeFields) > 1 {
			return Config{}, fmt.Errorf("unexpected fields after quoted value")
		}
		if len(typeFields) == 1 {
			if !isKnownConfigType(typeFields[0]) {
				return Config{}, fmt.Errorf("unknown config type %q", typeFields[0])
			}
			parsed.Type = typeFields[0]
		}
		return parsed, nil
	}

	fields := strings.Fields(trimmedValuePart)
	valueFields := fields
	if len(fields) > 1 {
		if !isKnownConfigType(fields[len(fields)-1]) {
			return Config{}, fmt.Errorf("ambiguous unquoted value; quote values containing spaces")
		}
		parsed.Type = fields[len(fields)-1]
		valueFields = fields[:len(fields)-1]
	}
	value := strings.Join(valueFields, " ")
	parsed.Value = &value
	return parsed, nil
}

func takeField(input string) (field, remainder string, ok bool) {
	input = strings.TrimLeftFunc(input, unicode.IsSpace)
	if input == "" {
		return "", "", false
	}
	separator := strings.IndexFunc(input, unicode.IsSpace)
	if separator < 0 {
		return input, "", true
	}
	return input[:separator], input[separator:], true
}

func takeQuotedValue(input string) (value, remainder string, err error) {
	escaped := false
	for i := 1; i < len(input); i++ {
		switch {
		case escaped:
			escaped = false
		case input[i] == '\\':
			escaped = true
		case input[i] == '"':
			value, err := strconv.Unquote(input[:i+1])
			if err != nil {
				return "", "", fmt.Errorf("invalid quoted value: %w", err)
			}
			return value, input[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("unterminated quoted value")
}

func isKnownConfigType(valueType string) bool {
	switch valueType {
	case "string", "integer", "boolean", "float", "date", "array", "dict", "data":
		return true
	default:
		return false
	}
}

// GenerateConfigFileContent generates the content for the configuration file from a slice of Config.
func GenerateConfigFileContent(configs []Config) string {
	var content strings.Builder
	for _, config := range configs {
		if config.Value == nil {
			content.WriteString(config.Domain)
			content.WriteByte(' ')
			content.WriteString(config.Key)
			content.WriteByte('\n')
			continue
		}
		configType := config.Type
		if configType == "" {
			configType = "string"
		}
		content.WriteString(config.Domain)
		content.WriteByte(' ')
		content.WriteString(config.Key)
		content.WriteByte(' ')
		content.WriteString(strconv.Quote(*config.Value))
		content.WriteByte(' ')
		content.WriteString(configType)
		content.WriteByte('\n')
	}
	return content.String()
}

// WriteConfigFile writes the configs to the configuration file.
func WriteConfigFile(fs FileSystemReader, configs []Config) error {
	content := GenerateConfigFileContent(configs)
	return fs.WriteFile(ConfigFilePath, content)
}
