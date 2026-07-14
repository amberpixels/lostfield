<p align="center">
  <img src="logo.svg" alt="LostField" width="560">
</p>

<div align="center">

### No field left behind.

A Go linter that catches incomplete struct converters.

[![Go Reference](https://pkg.go.dev/badge/github.com/amberpixels/lostfield.svg)](https://pkg.go.dev/github.com/amberpixels/lostfield)
[![CI](https://github.com/amberpixels/lostfield/actions/workflows/ci.yml/badge.svg)](https://github.com/amberpixels/lostfield/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/amberpixels/lostfield)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

</div>

---

`lostfield` analyzes converter functions (functions that transform one struct
type into another) and reports fields that leak out of the conversion:

- **Input fields** that should be read but aren't
- **Output fields** that should be set but aren't

It catches the classic bug: a field is added to a model, but the converters
are never updated - and the API silently ships incomplete data.

```go
func ConvertUserToDTO(user User) UserDTO {
    return UserDTO{
        ID:       user.ID,
        Username: user.Username,
        // Missing: Email  <- lostfield reports this
    }
}
```

> [!NOTE]
> lostfield is in **early development** (pre-1.0). It's already used in real
> projects, but flags and defaults may still change before v1.0. Questions,
> ideas, and feedback are very welcome - see [Feedback](#feedback).

## Contents

- [Install](#install)
- [Usage](#usage)
  - [Standalone (go vet)](#standalone-go-vet)
  - [With golangci-lint (module plugin)](#with-golangci-lint-module-plugin)
  - [Programmatic](#programmatic)
- [Where it fits](#where-it-fits)
- [Configuration](#configuration)
  - [Command-line flags](#command-line-flags)
  - [How converter detection works](#how-converter-detection-works)
  - [Deprecated fields](#deprecated-fields)
  - [Examples](#examples)
- [Output](#output)
- [Requirements](#requirements)
- [AI disclosure](#ai-disclosure)
- [Feedback](#feedback)
- [License](#license)

## Install

```bash
go install github.com/amberpixels/lostfield/cmd/lostfield@latest
```

Or pin it per-project with the Go 1.24+ tool directive:

```bash
go get -tool github.com/amberpixels/lostfield/cmd/lostfield@latest
```

## Usage

### Standalone (go vet)

```bash
go vet -vettool=$(which lostfield) ./...

# Typical scoped run: check only your DTO/mapping layer, ignore DB-managed fields
go vet -vettool=$(which lostfield) \
    -lostfield.exclude-fields='^ID$,CreatedAt,UpdatedAt,DeletedAt' \
    ./internal/dtos/...

# With the tool directive, resolve the pinned binary via `go tool -n`
go vet -vettool=$(go tool -n lostfield) ./internal/dtos/...
```

> [!TIP]
> `go vet` caches results per package+flags. Environment-only changes (such as
> `NO_COLOR`) don't invalidate the cache; touch a file or change a flag to
> force a re-run.

### With golangci-lint (module plugin)

`lostfield` supports golangci-lint's [module plugin system](https://golangci-lint.run/plugins/module-plugins/):
you build a custom golangci-lint binary once, then use lostfield like any other linter.

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

See [.golangci.example.yml](.golangci.example.yml) for all settings with
documentation. Unknown setting keys fail the build (typo protection), and
omitted keys keep their defaults.

Upstream inclusion in golangci-lint (no custom binary needed) is planned.

### Programmatic

The analyzer is importable for custom vet tools and multicheckers:

```go
import "github.com/amberpixels/lostfield"

cfg := lostfield.DefaultConfig()
cfg.ExcludeFieldPatterns = []string{"CreatedAt", "UpdatedAt"}
analyzer := lostfield.NewAnalyzer(cfg)
```

## Where it fits

lostfield shines in codebases where mapping code is hand-written and
concentrated in one layer - for example a `dtos/` package converting between
domain models and wire shapes, as in the [DCBA](https://github.com/amberpixels/dcba)
layering convention. Scope the linter to that layer and every
"added a field, forgot the converter" bug is caught at lint time:

```bash
go vet -vettool=$(go tool -n lostfield) \
    -lostfield.exclude-fields='^ID$,CreatedAt,UpdatedAt,DeletedAt' \
    -lostfield.ignore-tags='lostfield:"ignore"' \
    ./internal/dtos/...
```

DB-managed fields (primary keys, timestamps) are excluded by name; foreign-key
fields set by the ORM can be tagged `lostfield:"ignore"` at the model instead.

## Configuration

### Command-line flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-include-methods` | bool | `true` | Check method receivers in addition to plain functions |
| `-allow-getters` | bool | `true` | Allow Get* methods as substitute for direct field access |
| `-allow-aggregators` | bool | `true` | Enable detection of slice-to-non-slice aggregating converters |
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

Invalid values for enum-like flags (`-format`, `-fix-mode`,
`-non-marshallable-fields`, `-field-validation-mode`), out-of-range
`-min-similarity`, and non-compiling `-exclude-fields` regexes are rejected at
startup rather than silently ignored.

### How converter detection works

A function is considered a converter when it takes a struct (or
pointer/slice/map of structs) and returns a different struct whose type name is
*similar* to the input's:

- **`min-similarity: 0` (default)**: names match when one contains the other,
  case-insensitively (`User` -> `UserDTO`). Shared fragments shorter than
  3 characters are ignored.
- **`min-similarity: > 0`**: names match when their Sørensen-Dice bigram
  similarity reaches the threshold. **Recommended: `0.6`** - it keeps genuine
  pairs like `UserModel` -> `UserModelDTO` while dropping incidental pairs like
  `Message` -> `MessageNewParams` (an API params struct, not a conversion).

Constructors (functions starting with `New`) are never treated as converters.
Use `-exclude-converters`/`-only-converters` for name-based control.

### Deprecated fields

Fields whose doc comment contains `Deprecated:` are excluded from validation by
default - a converter may legitimately stop copying them. Pass
`-include-deprecated` to validate them like normal fields.

Limitation: deprecation is only detectable when the struct is defined in the
package under analysis; doc comments of imported types are not available to a
`go/analysis` pass.

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

## Output

**Default format** (single-line, machine-friendly - used by `go vet -json`,
editors, and golangci-lint):

```
converters/readmeExample/converter.go:3:6: ConvertUserToDTO: incomplete converter with missing fields: user.Email, user.CreatedAt, Email
```

**Pretty format** (`-format=pretty`, human-only; respects [`NO_COLOR`](https://no-color.org)):

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

## Requirements

- Go 1.26+

## AI disclosure

lostfield's code is written with heavy AI assistance - and that's by design.
But the AI is a tool here, not the author of record:

- **Every architectural decision is made by a human.** The detection heuristics,
  the config surface, the trade-offs - those are deliberate human choices, not
  whatever a model happened to produce.
- **Every line of code is read and reviewed by a human before it's pushed.**
  Nothing lands in this repository unread.
- **The code is written AI-first.** It's deliberately optimized to be easy for
  AI to read, grep, update, and extend - not primarily for human ergonomics.
  Clear, greppable names and consistent structure win over cleverness.

Responsibility for the code is human. 🤖🤝🧑

## Feedback

lostfield is a solo, opinionated project - but if you stumbled upon it and have
ideas, questions, or bug reports, an
[issue](https://github.com/amberpixels/lostfield/issues) is always welcome :)

## License

[MIT](LICENSE) © [amberpixels](https://amberpixels.io)
