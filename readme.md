# LostField

A Go linter that ensures converter functions use all fields from both input and output structs, preventing "leaky" conversions where fields are accidentally omitted.

## What it does

`lostfield` analyzes converter functions (functions that transform one struct type to another) and reports when fields are missing from either:
- **Input fields** - fields that should be read but aren't
- **Output fields** - fields that should be set but aren't

This helps catch bugs where new fields are added to structs but converter functions aren't updated accordingly.

It shines in codebases where mapping code is hand-written and concentrated in one layer (e.g. a `dtos/` package converting between domain models and wire shapes): scope the linter to that layer and every "added a field, forgot the converter" bug is caught at lint time.

## Installation

```bash
go install github.com/amberpixels/lostfield/cmd/lostfield@latest
```

## Usage

### Standalone (go vet)

```bash
go vet -vettool=$(which lostfield) ./...

# Typical scoped run: check only your DTO/mapping layer, ignore DB-managed fields
go vet -vettool=$(which lostfield) \
    -lostfield.exclude-fields='^ID$,CreatedAt,UpdatedAt,DeletedAt' \
    ./internal/dtos/...
```

> Note: `go vet` caches results per package+flags. Environment-only changes
> (such as `NO_COLOR`) don't invalidate the cache; touch a file or change a
> flag to force a re-run.

### With golangci-lint (module plugin)

`lostfield` supports golangci-lint's [module plugin system](https://golangci-lint.run/plugins/module-plugins/): you build a custom golangci-lint binary once, then use lostfield like any other linter.

1. Add a `.custom-gcl.yml` next to your `.golangci.yml` (see [.custom-gcl.example.yml](.custom-gcl.example.yml)):

```yaml
version: v2.12.2
plugins:
  - module: 'github.com/amberpixels/lostfield'
    version: latest
```

2. Build the custom binary:

```bash
golangci-lint custom
```

3. Enable and configure lostfield in `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - lostfield
  settings:
    custom:
      lostfield:
        type: "module"
        description: "Detects incomplete struct converter functions"
        settings:
          min-similarity: 0.6
          exclude-fields: ["^ID$", "CreatedAt", "UpdatedAt", "DeletedAt"]
```

See [.golangci.example.yml](.golangci.example.yml) for all settings with documentation. Unknown setting keys fail the build (typo protection), and omitted keys keep their defaults.

Upstream inclusion in golangci-lint (no custom binary needed) is planned.

### Programmatic (custom vet tools, multicheckers)

The analyzer is importable:

```go
import "github.com/amberpixels/lostfield"

cfg := lostfield.DefaultConfig()
cfg.ExcludeFieldPatterns = []string{"CreatedAt", "UpdatedAt"}
analyzer := lostfield.NewAnalyzer(cfg)
```

## Configuration

### Command-line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-include-methods` | bool | `true` | Check method receivers in addition to plain functions |
| `-allow-getters` | bool | `true` | Allow Get* methods as substitute for direct field access |
| `-allow-aggregators` | bool | `true` | Enable detection of slice→non-slice aggregating converters |
| `-exclude-fields` | string | `""` | Comma-separated regex patterns for field names to ignore (matched against leaf name and full nested path) |
| `-exclude-converters` | string | `""` | Comma-separated glob patterns for function/method names to exclude (e.g., `Get*,Map*`) |
| `-only-converters` | string | `""` | Comma-separated glob patterns for function/method names to include (only matching converters are analyzed) |
| `-exclude-files` | string | `"*_test.go,*.pb.go,*/vendor/*"` | Comma-separated glob patterns for file paths to exclude |
| `-min-similarity` | float64 | `0.0` | Minimum type-name similarity (0.0-1.0). `0` = substring matching; recommended `0.6` to reduce false positives |
| `-ignore-tags` | string | `""` | Comma-separated struct tags marking fields to ignore (bare key or `key:"value"`) |
| `-include-generated` | bool | `false` | Include generated code files in analysis |
| `-include-deprecated` | bool | `false` | Validate deprecated fields too (by default converters may skip them) |
| `-include-private-fields` | bool | `false` | Validate unexported (private) fields in converters |
| `-non-marshallable-fields` | string | `"adaptive"` | How to handle non-marshallable field types: `ignore`, `adaptive`, `strict` |
| `-field-validation-mode` | string | `"strict"` | Field validation mode: `strict` (all fields) or `intersection` (only common fields) |
| `-fix-mode` | string | `""` | Suggested fixes: `safe` (suppressing stubs) or `smart` (inferred mappings); apply with go vet's `-fix` |
| `-format` | string | `"default"` | Output format: `default` (standard go vet), `pretty` (Rust-like, human-only) |
| `-verbose` | bool | `false` | Verbose output (with `-format=pretty`, shows all fields instead of truncating) |

Invalid values for enum-like flags (`-format`, `-fix-mode`, `-non-marshallable-fields`, `-field-validation-mode`), out-of-range `-min-similarity`, and non-compiling `-exclude-fields` regexes are rejected at startup rather than silently ignored.

### How converter detection works

A function is considered a converter when it takes a struct (or pointer/slice/map of structs) and returns a different struct whose type name is *similar* to the input's:

- **`min-similarity: 0` (default)**: names match when one contains the other, case-insensitively (`User` → `UserDTO`). Shared fragments shorter than 3 characters are ignored.
- **`min-similarity: > 0`**: names match when their Sørensen–Dice bigram similarity reaches the threshold. **Recommended: `0.6`** — it keeps genuine pairs like `UserModel` → `UserModelDTO` while dropping incidental pairs like `Message` → `MessageNewParams` (an API params struct, not a conversion).

Constructors (functions starting with `New`) are never treated as converters. Use `-exclude-converters`/`-only-converters` for name-based control.

### Deprecated fields

Fields whose doc comment contains `Deprecated:` are excluded from validation by default — a converter may legitimately stop copying them. Pass `-include-deprecated` to validate them like normal fields.

Limitation: deprecation is only detectable when the struct is defined in the package under analysis; doc comments of imported types are not available to a `go/analysis` pass.

### Examples

```bash
# Check only plain functions
lostfield -include-methods=false ./...

# Ignore timestamp fields
lostfield -exclude-fields="CreatedAt,UpdatedAt,DeletedAt" ./...

# Stricter converter detection (recommended for whole-tree runs)
lostfield -min-similarity=0.6 ./...

# Skip fields tagged `lostfield:"ignore"`
lostfield -ignore-tags='lostfield:"ignore"' ./...

# Use pretty (Rust-like) output format with colors
lostfield -format=pretty ./...

# Strict mode: no getters, check methods
lostfield -include-methods=true -allow-getters=false ./...

# Only validate fields common to both input and output
lostfield -field-validation-mode=intersection ./...

# Exclude specific converter functions
lostfield -exclude-converters="Get*,Map*,to*" ./...

# Target a specific converter function
lostfield -only-converters="CuratorPurchase" -format=pretty -verbose ./...

# Generate suggested fixes and apply them
go vet -vettool=$(which lostfield) -lostfield.fix-mode=smart -fix ./...
```

**Note:** The `-format=pretty` output embeds colors and multi-line source excerpts and is meant for humans; machine consumers (`go vet -json`, editors, golangci-lint) should use the `default` format (golangci-lint enforces this automatically). Pretty output respects [`NO_COLOR`](https://no-color.org).

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

```bash
just test   # run tests
just lint   # run golangci-lint
just run    # run lostfield against a target path
```

## License

[MIT](LICENSE)
