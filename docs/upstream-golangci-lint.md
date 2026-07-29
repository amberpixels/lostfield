# Upstream golangci-lint submission kit

Goal: get `lostfield` included as a public linter in golangci-lint (and thus
listed on golangci-lint.run). Reference process:
https://golangci-lint.run/docs/contributing/new-linters/

## Prerequisites (must be done first)

- [ ] A tag pushed to GitHub - golangci-lint's `go.mod` must be able to require a
      released version of `github.com/amberpixels/lostfield`.
- [ ] `go` directive at or below golangci-lint's own floor. Theirs is `go 1.25.0`,
      under a comment reserving changes to their maintainers, and a dependency
      declaring more rewrites it silently via `go mod tidy`. lostfield is `1.25.0`.
- [ ] Repo public with LICENSE, README, tests, CI green (all in place).

## Changes to make in a golangci-lint fork

Branch off `main` of github.com/golangci/golangci-lint.

### 1. `pkg/golinters/lostfield/lostfield.go`

```go
package lostfield

import (
	"github.com/amberpixels/lostfield"

	"github.com/golangci/golangci-lint/v2/pkg/config"
	"github.com/golangci/golangci-lint/v2/pkg/goanalysis"
)

func New(settings *config.LostFieldSettings) *goanalysis.Linter {
	cfg := lostfield.DefaultConfig()

	if settings != nil {
		cfg.AllowMethodConverters = settings.IncludeMethods
		cfg.AllowGetters = settings.AllowGetters
		cfg.AllowAggregators = settings.AllowAggregators
		cfg.ExcludeFieldPatterns = settings.ExcludeFields
		cfg.ExcludeConverterPatterns = settings.ExcludeConverters
		cfg.OnlyConverterPatterns = settings.OnlyConverters
		cfg.ExcludeFilePatterns = settings.ExcludeFiles
		cfg.MinTypeNameSimilarity = settings.MinSimilarity
		cfg.IgnoreFieldTags = settings.IgnoreTags
		cfg.IncludeGenerated = settings.IncludeGenerated
		cfg.IncludeDeprecated = settings.IncludeDeprecated
		cfg.IncludePrivateFields = settings.IncludePrivateFields
		cfg.NonMarshallableFieldsHandling = lostfield.NonMarshallableFieldsHandling(settings.NonMarshallableFields)
		cfg.FieldValidationMode = lostfield.FieldValidationMode(settings.FieldValidationMode)
	}

	return goanalysis.
		NewLinterFromAnalyzer(lostfield.NewAnalyzer(cfg)).
		WithLoadMode(goanalysis.LoadModeTypesInfo)
}
```

Notes:
- The import above is the module root, which holds only the analyzer. The
  module-plugin registration lives in `github.com/amberpixels/lostfield/plugin`, so
  upstream never picks up `plugin-module-register` or an `init()`. Importing the root
  package pulls exactly one external module, `golang.org/x/tools`.
- `format`/`verbose`/`fix-mode` are intentionally NOT exposed: output formatting
  belongs to golangci-lint, `verbose` writes to stderr mid-run, and suggested fixes
  flow through `--fix` once the linter is registered `WithAutoFix()`.
- `lostfield.NewAnalyzer` validates the config and reports an invalid enum value
  as a run error - no extra validation needed in the wrapper.

### 2. `pkg/config/linters_settings.go` (settings struct + defaults)

```go
type LostFieldSettings struct {
	IncludeMethods        bool     `mapstructure:"include-methods"`
	AllowGetters          bool     `mapstructure:"allow-getters"`
	AllowAggregators      bool     `mapstructure:"allow-aggregators"`
	ExcludeFields         []string `mapstructure:"exclude-fields"`
	ExcludeConverters     []string `mapstructure:"exclude-converters"`
	OnlyConverters        []string `mapstructure:"only-converters"`
	ExcludeFiles          []string `mapstructure:"exclude-files"`
	MinSimilarity         float64  `mapstructure:"min-similarity"`
	IgnoreTags            []string `mapstructure:"ignore-tags"`
	IncludeGenerated      bool     `mapstructure:"include-generated"`
	IncludeDeprecated     bool     `mapstructure:"include-deprecated"`
	IncludePrivateFields  bool     `mapstructure:"include-private-fields"`
	NonMarshallableFields string   `mapstructure:"non-marshallable-fields"`
	FieldValidationMode   string   `mapstructure:"field-validation-mode"`
}
```

Add matching defaults to the settings defaults var (mirror
`lostfield.DefaultConfig()`: methods/getters/aggregators true,
exclude-files `["*_test.go", "*.pb.go", "*/vendor/*"]`,
non-marshallable-fields `adaptive`, field-validation-mode `strict`).

### 3. Registration in `pkg/lint/lintersdb/builder_linter.go`

```go
linter.NewConfig(lostfield.New(&cfg.Linters.Settings.LostField)).
	WithSince("next_version").
	WithLoadForGoAnalysis().
	WithAutoFix().
	WithURL("https://github.com/amberpixels/lostfield"),
```

Two v2 notes, both checked against today's `builder_linter.go`:

- **No `WithPresets`.** Presets were removed in v2; the method does not exist (zero
  occurrences in that file). A v1-shaped `WithPresets(linter.PresetBugs)` will not compile.
- **`WithAutoFix()` is required** for the suggested fixes to be reachable. It is how v2
  marks a linter as `--fix`-capable (40 uses). Without it the `fixer` package is dead
  weight to golangci-lint users, since `SuggestedFixes` alone is not enough.

### 4. Functional test `pkg/golinters/lostfield/testdata/lostfield.go`

```go
//golangcitest:args -Elostfield
package testdata

type User struct {
	ID       int64
	Username string
	Email    string
}

type UserDTO struct {
	ID       int64
	Username string
	Email    string
}

func ConvertUserToDTO(user User) UserDTO { // want `ConvertUserToDTO: incomplete converter with missing fields: user.Email, Email`
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
	}
}
```

### 5. Reference config

Add the settings block (with defaults) to `.golangci.next.reference.yml` under
`linters-settings`.

### 6. Verify

```bash
go mod tidy   # pulls github.com/amberpixels/lostfield@v0.3.0
go run ./cmd/golangci-lint/ run --no-config --default=none --enable=lostfield \
    ./pkg/golinters/lostfield/testdata/lostfield.go
go test ./pkg/golinters/lostfield/...
```

## PR description draft

> ### Add `lostfield` linter
>
> `lostfield` (https://github.com/amberpixels/lostfield) detects incomplete
> struct converter functions - functions that transform one struct type into
> another but silently drop fields. It reports fields of the input that are
> never read and fields of the output that are never set, catching the common
> bug where a field is added to a model but its converters are not updated.
>
> - Pure `go/analysis` implementation (`unitchecker`-compatible), analyzer
>   exposed via `lostfield.NewAnalyzer(cfg)`.
> - No global state: config is captured per-analyzer, safe for concurrent passes.
> - Supports methods, getters, nested fields, embedded structs, delegating and
>   aggregating converters, deprecated-field handling, struct-tag ignores,
>   regex field exclusion, and a name-similarity threshold to control detection
>   strictness.
> - Suggested fixes (`safe`/`smart`) are emitted as `analysis.SuggestedFix`, so
>   `--fix` works once registered `WithAutoFix()`.
> - Importing it costs one external module, `golang.org/x/tools`. The module-plugin
>   registration lives in a separate `./plugin` package, so nothing else comes along.
> - Tagged release: v0.3.0, Go floor 1.25.0 to match golangci-lint's own. MIT licensed.
>   Tests: analysistest corpus with 21 scenarios plus unit tests (84% coverage), CI
>   running at the floor and at the declared minimum.

## Checklist before opening the PR

- [ ] lostfield `v0.3.0` tag pushed
- [ ] fork branch `feat/add-lostfield-linter` builds and its functional test passes
- [ ] `.golangci.next.reference.yml` updated
- [ ] PR title follows their convention: `Add lostfield linter`
