package config

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Linter metadata constants.
const (
	LinterName = "lostfield"
	LinterDoc  = "reports all inconsistent converter functions: finds lost fields"
)

// NonMarshallableFieldsHandling specifies how to handle non-marshallable field types
// (functions, channels, etc.) in converter function validation.
type NonMarshallableFieldsHandling string

const (
	// HandleIgnore: Skip all non-marshallable fields (func, chan, etc.) entirely.
	// No validation is performed for these field types in any converter.
	HandleIgnore NonMarshallableFieldsHandling = "ignore"

	// HandleAdaptive: Validate non-marshallable fields ONLY if they exist in both input AND output models.
	// Smart & pragmatic: adapts to the actual model structure.
	// Example: Apple{Callback func()} → ApiApple{} : Callback ignored (missing in output)
	// Example: Apple{Callback func()} → ApiBridge{Callback func()} : Callback validated (present in both).
	HandleAdaptive NonMarshallableFieldsHandling = "adaptive"

	// HandleStrict: Treat non-marshallable fields like normal fields - they MUST be handled.
	// If input has a func or channel field, the converter must read/use it.
	HandleStrict NonMarshallableFieldsHandling = "strict"
)

// FieldValidationMode specifies which fields must be validated in a converter.
type FieldValidationMode string

const (
	// ModeStrict: All fields from both input and output types must be handled.
	// Example: Input{Name, Email, Phone} → Output{Name, Email, Surname}
	// Must validate: Name, Email, Phone (input) and Name, Email, Surname (output).
	ModeStrict FieldValidationMode = "strict"

	// ModeIntersection: Only fields present in both input and output types must be handled.
	// Pragmatic approach: ignores fields that only exist in one side.
	// Example: Input{Name, Email, Phone} → Output{Name, Email, Surname}
	// Must validate only: Name, Email (intersection).
	ModeIntersection FieldValidationMode = "intersection"
)

// Format specifies the output format for diagnostics.
type Format string

const (
	// FormatDefault: standard go vet single-line format.
	FormatDefault Format = "default"

	// FormatPretty: Rust-like multi-line pretty format with source excerpts.
	FormatPretty Format = "pretty"
)

// FixMode controls whether diagnostics carry SuggestedFixes for automatic fixing.
type FixMode string

const (
	// FixModeDisabled: fix generation is disabled; diagnostics are reported without fixes.
	FixModeDisabled FixMode = ""

	// FixModeSafe: generate safe fixes that suppress warnings without changing behavior
	// (`_ = var.Field` stubs and TODO comments).
	FixModeSafe FixMode = "safe"

	// FixModeSmart: infer the correct field mapping (direct assignment, getter call, or
	// type conversion) and fall back to the safe fix for incompatible types.
	// The smart fix is listed first (applied by -fix), safe fix second.
	FixModeSmart FixMode = "smart"
)

// Config holds all configuration for the analyzer.
//
// The json tags are the canonical setting names: they match the CLI flag names and are
// used to decode settings coming from golangci-lint (module plugin or upstream).
type Config struct {
	// AllowMethodConverters enables looking for converters in methods in addition to plain functions.
	// Default: true
	AllowMethodConverters bool `json:"include-methods" mapstructure:"include-methods"`

	// AllowGetters allows Get* methods as a substitute for field access
	// Default: true
	AllowGetters bool `json:"allow-getters" mapstructure:"allow-getters"`

	// AllowAggregators enables detection of slice->non-slice converters
	// where the output struct contains a field that holds the converted slice.
	// Default: true
	AllowAggregators bool `json:"allow-aggregators" mapstructure:"allow-aggregators"`

	// ExcludeFieldPatterns is a list of regex patterns for field names to ignore.
	// Patterns are matched against both the leaf field name (e.g. "CreatedAt") and
	// the full nested path (e.g. "User.Role.CreatedAt").
	// Default: []
	ExcludeFieldPatterns []string `json:"exclude-fields" mapstructure:"exclude-fields"`

	// ExcludeConverterPatterns is a list of glob patterns for function/method names to exclude from converter detection.
	// Supports wildcards: * matches any sequence of characters, ? matches a single character.
	// Examples: "Get*", "Map*", "to*", "*Helper"
	// Default: []
	ExcludeConverterPatterns []string `json:"exclude-converters" mapstructure:"exclude-converters"`

	// OnlyConverterPatterns is a list of glob patterns for function/method names to include.
	// When non-empty, only converters matching at least one pattern are analyzed (inverse of exclude-converters).
	// Supports wildcards: * matches any sequence of characters, ? matches a single character.
	// Examples: "CuratorPurchase", "Convert*"
	// Default: []
	OnlyConverterPatterns []string `json:"only-converters" mapstructure:"only-converters"`

	// ExcludeFilePatterns is a list of glob patterns for file paths to exclude from analysis.
	// Supports wildcards: * matches any sequence of characters, ? matches a single character.
	// Patterns are matched against the full file path.
	// Examples: "*_test.go", "*.pb.go", "*/vendor/*"
	// Default: ["*_test.go", "*.pb.go", "*/vendor/*"]
	ExcludeFilePatterns []string `json:"exclude-files" mapstructure:"exclude-files"`

	// MinTypeNameSimilarity is the minimum type name similarity ratio (0.0-1.0).
	// 0.0 keeps the default behavior: case-insensitive substring matching between
	// input/output type names. Values above 0.0 require the names to be at least
	// that similar (Sørensen–Dice bigram coefficient) — recommended ~0.4-0.5 to
	// reduce false positives on incidentally-similar names.
	MinTypeNameSimilarity float64 `json:"min-similarity" mapstructure:"min-similarity"`

	// IgnoreFieldTags is a list of struct tags that mark fields to be ignored.
	// Each entry is either a bare tag key (e.g. "lostfield" — any value matches)
	// or key:"value" form (e.g. `lostfield:"ignore"` — the tag value must match).
	// Default: []
	IgnoreFieldTags []string `json:"ignore-tags" mapstructure:"ignore-tags"`

	// IncludeGenerated includes generated code files in analysis.
	// Default: false
	IncludeGenerated bool `json:"include-generated" mapstructure:"include-generated"`

	// IncludeDeprecated includes deprecated fields in validation.
	//
	// Deprecated fields are identified by "Deprecated:" in their documentation comments.
	// By default they are excluded from validation (a converter may legitimately skip them).
	//
	// Note (Experimental): detection only works for fields whose source is part of the
	// current analysis pass. Fields from third-party or out-of-analysis packages
	// (e.g. generated files excluded from the run) cannot be identified as deprecated.
	//
	// Default: false (deprecated fields are ignored)
	IncludeDeprecated bool `json:"include-deprecated" mapstructure:"include-deprecated"`

	// Verbose enables verbose output.
	Verbose bool `json:"verbose" mapstructure:"verbose"`

	// Format specifies the output format for diagnostics.
	// Supported values: "default" (standard go vet format), "pretty" (Rust-like pretty format).
	// Default: "default"
	Format Format `json:"format" mapstructure:"format"`

	// NonMarshallableFieldsHandling specifies how to handle non-marshallable field types
	// (functions, channels, etc. - types that cannot be serialized/marshalled).
	// Examples of non-marshallable fields: func(), chan string, func(error) error
	//
	// Behavior:
	//   - "ignore": Skip non-marshallable fields entirely during validation.
	//     Example: Input has `Handler func()`, Output doesn't → both are valid (no error)
	//
	//   - "adaptive" (default): Validate non-marshallable fields ONLY if present in both input AND output.
	//     Example Input:  `Apple { ID string, Callback func() }`
	//     Example Output: `ApiApple { ID string }` (no Callback)
	//     Result: Callback is ignored (not an error) because it's missing from output
	//
	//   - "strict": Treat non-marshallable fields like normal fields - they MUST be handled.
	//     Example: Input has `Callback func()` → error if not read in converter
	//
	// Default: "adaptive"
	NonMarshallableFieldsHandling NonMarshallableFieldsHandling `json:"non-marshallable-fields" mapstructure:"non-marshallable-fields"`

	// IncludePrivateFields enables validation of unexported (private) fields in converters.
	// Private fields are lowercase-starting fields (e.g., `id`, `internalCache`, `mutex`).
	//
	// Behavior:
	//   - false (default): Skip private fields entirely during validation.
	//     Example: Input has `id string` (private), it's not checked in the converter
	//
	//   - true: Treat private fields like public fields - they MUST be handled.
	//     Example: Input has `id string` (private) → error if not read/set in converter
	//
	// Note: In Go, only exported (uppercase) fields can be accessed across packages in converters.
	// This option is mostly useful for converters within the same package.
	//
	// Default: false (private fields ignored)
	IncludePrivateFields bool `json:"include-private-fields" mapstructure:"include-private-fields"`

	// FieldValidationMode specifies which fields must be validated in converters.
	// Determines the scope of field validation for converter functions.
	//
	// Behavior:
	//   - "strict" (default): All fields from both input and output types must be handled.
	//     Example: Input{Name, Email, Phone} → Output{Name, Email, Surname}
	//     Error if: Phone (from input) or Surname (from output) not handled
	//
	//   - "intersection": Only fields present in both input and output must be handled.
	//     Example: Input{Name, Email, Phone} → Output{Name, Email, Surname}
	//     Only Name and Email must be validated (intersection)
	//
	// Default: "strict"
	FieldValidationMode FieldValidationMode `json:"field-validation-mode" mapstructure:"field-validation-mode"`

	// FixMode controls whether diagnostics carry SuggestedFixes for automatic fixing.
	// When combined with the -fix flag (from unitchecker), fixes are applied automatically.
	//
	// Behavior:
	//   - "" (empty, default): Fix generation is disabled. Diagnostics are reported without fixes.
	//   - "safe": Generate safe fixes that suppress warnings without changing behavior.
	//     Uses `_ = var.Field` pattern for both input and output missing fields.
	//   - "smart": Generate smart fixes that infer the correct field mapping.
	//     For fields present in both missing input and output, generates direct assignments,
	//     getter calls, or type conversions. Falls back to safe fix for incompatible types.
	//     When mode is "smart", the smart fix is listed first (applied by -fix), safe fix second.
	//
	// Default: "" (disabled)
	FixMode FixMode `json:"fix-mode" mapstructure:"fix-mode"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		AllowMethodConverters:         true,
		AllowGetters:                  true,
		AllowAggregators:              true,
		ExcludeFieldPatterns:          []string{},
		ExcludeConverterPatterns:      []string{},
		OnlyConverterPatterns:         []string{},
		ExcludeFilePatterns:           []string{"*_test.go", "*.pb.go", "*/vendor/*"},
		MinTypeNameSimilarity:         0.0, // 0 = use substring matching.
		IgnoreFieldTags:               []string{},
		IncludeGenerated:              false,
		IncludeDeprecated:             false,
		Verbose:                       false,           // Quiet output by default
		Format:                        FormatDefault,   // Use standard go vet format by default
		NonMarshallableFieldsHandling: HandleAdaptive,  // Adapt to what's present in both input and output models by default
		IncludePrivateFields:          false,           // Ignore private fields by default
		FieldValidationMode:           ModeStrict,      // Validate all fields by default
		FixMode:                       FixModeDisabled, // Fix generation disabled by default
	}
}

// Validate checks that all enum-like and numeric settings hold supported values.
// It is called after flag parsing (standalone binary) and after settings decoding
// (golangci-lint plugin), so a typo like `fix-mode: smrt` fails loudly instead of
// silently disabling the feature.
func (c *Config) Validate() error {
	switch c.NonMarshallableFieldsHandling {
	case HandleIgnore, HandleAdaptive, HandleStrict:
	default:
		return fmt.Errorf("invalid non-marshallable-fields value %q (supported: ignore, adaptive, strict)",
			c.NonMarshallableFieldsHandling)
	}

	switch c.FieldValidationMode {
	case ModeStrict, ModeIntersection:
	default:
		return fmt.Errorf("invalid field-validation-mode value %q (supported: strict, intersection)",
			c.FieldValidationMode)
	}

	switch c.Format {
	case FormatDefault, FormatPretty:
	default:
		return fmt.Errorf("invalid format value %q (supported: default, pretty)", c.Format)
	}

	switch c.FixMode {
	case FixModeDisabled, FixModeSafe, FixModeSmart:
	default:
		return fmt.Errorf("invalid fix-mode value %q (supported: safe, smart, or empty to disable)", c.FixMode)
	}

	if c.MinTypeNameSimilarity < 0.0 || c.MinTypeNameSimilarity > 1.0 {
		return fmt.Errorf("invalid min-similarity value %v (must be within 0.0-1.0)", c.MinTypeNameSimilarity)
	}

	for _, p := range c.ExcludeFieldPatterns {
		if _, err := regexp.Compile(p); err != nil {
			return fmt.Errorf("invalid exclude-fields pattern %q: %w", p, err)
		}
	}

	return nil
}

// splitCommaSeparated splits a comma-separated string into a slice of strings.
// Returns an empty slice if the input string is empty.
func splitCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// RegisterFlags registers configuration flags with the analyzer's FlagSet,
// binding them to the given Config instance. Enum-like flags are validated
// at parse time, so an unsupported value fails flag parsing instead of being
// silently accepted.
func RegisterFlags(fs *flag.FlagSet, cfg *Config) {
	fs.BoolVar(&cfg.AllowMethodConverters, "include-methods", cfg.AllowMethodConverters,
		"check method receivers in addition to plain functions")

	fs.BoolVar(&cfg.AllowGetters, "allow-getters", cfg.AllowGetters,
		"allow Get* methods as a substitute for direct field access")

	fs.BoolVar(&cfg.AllowAggregators, "allow-aggregators", cfg.AllowAggregators,
		"enable detection of slice->non-slice aggregating converters")

	fs.Func(
		"exclude-fields",
		"comma-separated regex patterns for field names to ignore (e.g., 'CreatedAt,UpdatedAt,.*ID')",
		func(s string) error {
			patterns := splitCommaSeparated(s)
			for _, p := range patterns {
				if _, err := regexp.Compile(p); err != nil {
					return fmt.Errorf("invalid exclude-fields pattern %q: %w", p, err)
				}
			}
			cfg.ExcludeFieldPatterns = patterns
			return nil
		},
	)

	fs.Func(
		"exclude-converters",
		"comma-separated glob patterns for function/method names to exclude from converter detection (e.g., 'Get*,Map*,to*')",
		func(s string) error {
			cfg.ExcludeConverterPatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Func(
		"only-converters",
		"comma-separated glob patterns for function/method names to include (only matching converters are analyzed)",
		func(s string) error {
			cfg.OnlyConverterPatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Func(
		"exclude-files",
		"comma-separated glob patterns for file paths to exclude from analysis (e.g., '*_test.go,*.pb.go,*/vendor/*')",
		func(s string) error {
			cfg.ExcludeFilePatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Func(
		"min-similarity",
		"minimum type name similarity ratio (0.0-1.0, 0=substring matching, higher=stricter)",
		func(s string) error {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("invalid min-similarity value %q: %w", s, err)
			}
			if v < 0.0 || v > 1.0 {
				return fmt.Errorf("invalid min-similarity value %q (must be within 0.0-1.0)", s)
			}
			cfg.MinTypeNameSimilarity = v
			return nil
		},
	)

	fs.Func("ignore-tags", "comma-separated struct tags to ignore fields (e.g., 'lostfield:\"ignore\"')",
		func(s string) error {
			cfg.IgnoreFieldTags = splitCommaSeparated(s)
			return nil
		})

	fs.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose,
		"enable verbose output")

	fs.BoolVar(&cfg.IncludeGenerated, "include-generated", cfg.IncludeGenerated,
		"include generated code files in analysis (default: exclude)")

	fs.BoolVar(&cfg.IncludeDeprecated, "include-deprecated", cfg.IncludeDeprecated,
		"include deprecated fields in validation (default: exclude)")

	fs.Func(
		"format",
		"output format for diagnostics (default: standard go vet format, pretty: Rust-like pretty format)",
		func(s string) error {
			cfg.Format = Format(s)
			switch cfg.Format {
			case FormatDefault, FormatPretty:
				return nil
			default:
				return fmt.Errorf("invalid format value %q (supported: default, pretty)", s)
			}
		},
	)

	fs.Func(
		"non-marshallable-fields",
		"how to handle non-marshallable field types (ignore, adaptive, strict)",
		func(s string) error {
			cfg.NonMarshallableFieldsHandling = NonMarshallableFieldsHandling(s)
			switch cfg.NonMarshallableFieldsHandling {
			case HandleIgnore, HandleAdaptive, HandleStrict:
				return nil
			default:
				return fmt.Errorf("invalid non-marshallable-fields value %q (supported: ignore, adaptive, strict)", s)
			}
		},
	)

	fs.BoolVar(&cfg.IncludePrivateFields, "include-private-fields", cfg.IncludePrivateFields,
		"validate unexported (private) fields in converters (default: ignore private fields)")

	fs.Func(
		"field-validation-mode",
		"field validation mode (strict: all fields from both input and output, intersection: only common fields)",
		func(s string) error {
			cfg.FieldValidationMode = FieldValidationMode(s)
			switch cfg.FieldValidationMode {
			case ModeStrict, ModeIntersection:
				return nil
			default:
				return fmt.Errorf("invalid field-validation-mode value %q (supported: strict, intersection)", s)
			}
		},
	)

	fs.Func(
		"fix-mode",
		"fix mode for automatic fixes (empty=disabled, safe=suppress warnings, smart=infer mappings)",
		func(s string) error {
			cfg.FixMode = FixMode(s)
			switch cfg.FixMode {
			case FixModeDisabled, FixModeSafe, FixModeSmart:
				return nil
			default:
				return fmt.Errorf("invalid fix-mode value %q (supported: safe, smart, or empty to disable)", s)
			}
		},
	)
}
