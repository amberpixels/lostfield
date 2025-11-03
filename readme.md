# LostField

A Go linter that ensures converter functions use all fields from both input and output structs, preventing "leaky" conversions where fields are accidentally omitted.

## What it does

`lostfield` analyzes converter functions (functions that transform one struct type to another) and reports when fields are missing from either:
- **Input fields** - fields that should be read but aren't
- **Output fields** - fields that should be set but aren't

This helps catch bugs where new fields are added to structs but converter functions aren't updated accordingly.

## Installation

```bash
go install github.com/amberpixels/lostfield/cmd/lostfield@latest
```

## Usage

### Standalone

```bash
# With go vet
go vet -vettool=$(which lostfield) ./...
```

### With golangci-lint

Add to your `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - lostfield

linters-settings:
  lostfield:
    include-methods: false
    allow-getters: true
    exclude-fields: "CreatedAt,UpdatedAt"
```

See [.golangci.example.yml](.golangci.example.yml) for detailed configuration options.

## Configuration

### Command-line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-include-methods` | bool | `false` | Check method receivers in addition to plain functions |
| `-allow-getters` | bool | `true` | Allow Get* methods as substitute for direct field access |
| `-exclude-fields` | string | `""` | Comma-separated regex patterns for field names to ignore |
| `-min-similarity` | float64 | `0.0` | Type name similarity threshold (0.0-1.0, 0=substring matching) |
| `-ignore-tags` | string | `""` | Comma-separated struct tags that mark fields to ignore |
| `-format` | string | `default` | Output format: `default` (standard go vet), `pretty` (Rust-like with colors) |

### Examples

```bash
# Check only plain functions (default)
lostfield ./...

# Also check methods
lostfield -include-methods=true ./...

# Ignore timestamp fields
lostfield -exclude-fields="CreatedAt,UpdatedAt,DeletedAt" ./...

# Use pretty (Rust-like) output format with colors
lostfield -format=pretty ./...

# Strict mode: no getters, check methods
lostfield -include-methods=true -allow-getters=false ./...
```

**Note:** The `-format` flag works with `go vet` directly. When using with **golangci-lint**, the output format is controlled by golangci-lint's `-out-format` flag (json, tab, checkstyle, etc.), and the linter will use the `default` format automatically.

## Example

Given these types:

```go
type User struct {
    ID        int64
    Username  string
    Email     string
    CreatedAt time.Time
}

type UserDTO struct {
    ID       int64
    Username string
    Email    string
}
```

This converter would trigger a warning:

```go
func ConvertUserToDTO(user User) UserDTO {
    return UserDTO{
        ID:       user.ID,
        Username: user.Username,
        // Missing: Email
    }
}
```

**Output (default format):**

```
converters/readmeExample/converter.go:3:6: ConvertUserToDTO: incomplete converter with missing fields: user.Email, user.CreatedAt, Email
```

**Output (pretty format with `-format=pretty`):**

```
   |
12 | func ConvertUserToDTO(user User) UserDTO {
   |      ^^^^^^^^^^^^^^^^ detected as converter
   |
   = note: missing fields:
     user.Email     → ??
     user.CreatedAt → ??
     ??             → Email
```

## Contributing

Contributions are welcome!

## License

[MIT](LICENSE)
