# Support Value Types in mdefaults

## Problem Statement

Currently, mdefaults stores all macOS default values as strings, but macOS defaults have specific types (integer, boolean, string, float, etc.). When values are stored with incorrect types, it can cause macOS to crash after login. This is a critical issue that needs to be addressed to ensure system stability.

## Solution Overview

Extend the mdefaults system to:

1. Detect and store value types during pull operations using `defaults read-type`
2. Apply values with correct types during push operations using type flags
3. Maintain backward compatibility with existing config files
4. Safely support scalar macOS defaults types
5. Abort pull and push operations when a value cannot be read or written safely

## Technical Design

### Type System

Support the following scalar macOS defaults types:

- `string` - String values (default for backward compatibility)
- `integer` - Integer values (mapped to `-int` flag)
- `boolean` - Boolean values (mapped to `-bool` flag)
- `float` - Float values (mapped to `-float` flag)
- `date` - Date values (mapped to `-date` flag)

Boolean values read back by macOS as `1` or `0` are normalized to `true` or
`false` before invoking `defaults write -bool`. The aliases `yes` and `no` are
accepted as input; other Boolean spellings are rejected before any write.

`array`, `dictionary`, and `data` values are rejected during pull and push. The
text emitted by `defaults read` for these types cannot be safely replayed as a
single `defaults write` argument, so failing explicitly avoids corrupting them.

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

New config file format:

```text
domain key "escaped value" type
```

Values use Go-compatible quoting so spaces, quotes, tabs, and newlines survive a
write/read round trip.

Old scalar formats and pull targets continue to work:

```text
domain key value
domain key
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
3. **Error Handling**: Abort the pull if value or type detection fails; defaulting an unknown type to string can corrupt macOS preferences
4. **Type Mapping**: Map supported internal types to defaults command flags and reject unsupported types before writing anything

## Testing Strategy

- Unit tests for each component with mocks
- Integration tests with different value types
- Backward compatibility tests
- Quoted-value round-trip tests
- Error handling tests
- Tests use injected command writers so they never modify the host's macOS defaults
