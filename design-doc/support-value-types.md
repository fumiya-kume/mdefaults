# Support Value Types in mdefaults

## Problem Statement

Currently, mdefaults stores all macOS default values as strings, but macOS defaults have specific types (integer, boolean, string, float, etc.). When values are stored with incorrect types, it can cause macOS to crash after login. This is a critical issue that needs to be addressed to ensure system stability.

## Solution Overview

Extend the mdefaults system to:
1. Detect and store value types during pull operations using `defaults read-type`
2. Apply values with correct types during push operations using type flags
3. Maintain backward compatibility with existing config files
4. Support all major macOS defaults types

## Technical Design

### Type System

Support the following macOS defaults types:
- `string` - String values (default for backward compatibility)
- `integer` - Integer values (mapped to `-int` flag)
- `boolean` - Boolean values (mapped to `-bool` flag)
- `float` - Float values (mapped to `-float` flag)
- `date` - Date values (mapped to `-date` flag)
- `array` - Array values (mapped to `-array` flag)
- `dict` - Dictionary values (mapped to `-dict` flag)
- `data` - Data values (mapped to `-data` flag)

### Config Structure Changes

Extend the `Config` struct to include type information:
```go
type Config struct {
    Domain string
    Key    string
    Value  *string
    Type   string // New field for storing the macOS defaults type
}
```

### File Format Changes

New config file format (backward compatible):
```
domain key value type
```

Old format continues to work:
```
domain key value
```

### Interface Extensions

Extend `DefaultsCommand` interface:
```go
type DefaultsCommand interface {
    Read(ctx context.Context) (string, error)
    ReadType(ctx context.Context) (string, error) // New method
    Write(ctx context.Context, value string) error
    WriteWithType(ctx context.Context, value string, valueType string) error // New method
    Domain() string
    Key() string
}
```

## Implementation Strategy

1. **TDD Approach**: Write tests first for each component
2. **Backward Compatibility**: Ensure old config files continue to work
3. **Error Handling**: Graceful fallback to string type if type detection fails
4. **Type Mapping**: Map internal types to defaults command flags

## Testing Strategy

- Unit tests for each component with mocks
- Integration tests with different value types
- Backward compatibility tests
- Error handling tests
