package config

import (
	"flag"
	"strings"
)

// Linter metadata constants.
const (
	LinterName = "lostfield"
	LinterDoc  = "reports all inconsistent converter functions: finds lost fields)"
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

// Config holds all configuration for the analyzer.
type Config struct {
	// AllowMethodConverters enables looking for converters in methods in addition to plain functions.
	// Default: true
	AllowMethodConverters bool

	// AllowGetters allows Get* methods as a substitute for field access
	// Default: true
	AllowGetters bool

	// AllowAggregators enables detection of slice->non-slice converters
	// where the output struct contains a field that holds the converted slice.
	// Default: true
	AllowAggregators bool

	// ExcludeFieldPatterns is a comma-separated list of regex patterns for field names to ignore
	// Default: []
	ExcludeFieldPatterns []string

	// ExcludeConverterPatterns is a comma-separated list of glob patterns for function/method names to exclude from converter detection.
	// Supports wildcards: * matches any sequence of characters, ? matches a single character.
	// Examples: "Get*", "Map*", "to*", "*Helper"
	// Default: []
	ExcludeConverterPatterns []string

	// ExcludeFilePatterns is a comma-separated list of glob patterns for file paths to exclude from analysis.
	// Supports wildcards: * matches any sequence of characters, ? matches a single character.
	// Patterns are matched against the full file path.
	// Examples: "*_test.go", "*.pb.go", "*/vendor/*"
	// Default: ["*_test.go", "*.pb.go", "*/vendor/*"]
	ExcludeFilePatterns []string

	// MinTypeNameSimilarity is the minimum type name similarity ratio (0.0-1.0, 0=substring matching)
	MinTypeNameSimilarity float64

	// IgnoreFieldTags is a comma-separated list of struct tags that mark fields to be ignored.
	// Default: []
	IgnoreFieldTags []string

	// IncludeGenerated includes generated code files in analysis.
	// Default: false
	IncludeGenerated bool

	// IgnoreDeprecated makes linter to ignore lost fields if they are deprecated.
	//
	// Deprecated fields are identified by "Deprecated:" in their documentation comments.
	//
	// Note: (Experimental): this might not work if source of the field is third-party,
	//       or out-of-analysis, e.g. generated file.
	//       So we won't be able to readme the comment and find that the field is deprecated.
	IgnoreDeprecated bool

	// Verbose enables verbose output.
	Verbose bool

	// Format specifies the output format for diagnostics.
	// Supported values: "default" (standard go vet format), "custom" (Rust-like pretty format).
	// Default: "default"
	Format string

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
	NonMarshallableFieldsHandling NonMarshallableFieldsHandling

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
	IncludePrivateFields bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		AllowMethodConverters:         true,
		AllowGetters:                  true,
		AllowAggregators:              true,
		ExcludeFieldPatterns:          []string{},
		ExcludeConverterPatterns:      []string{},
		ExcludeFilePatterns:           []string{"*_test.go", "*.pb.go", "*/vendor/*"},
		MinTypeNameSimilarity:         0.0, // 0 = use strict substring matching.
		IgnoreFieldTags:               []string{},
		IncludeGenerated:              false,
		IgnoreDeprecated:              false,
		Verbose:                       false,          // Quiet output by default
		Format:                        "default",      // Use standard go vet format by default
		NonMarshallableFieldsHandling: HandleAdaptive, // Adapt to what's present in both input and output models by default
		IncludePrivateFields:          false,          // Ignore private fields by default
	}
}

// current holds the active configuration.
var current = DefaultConfig()

// Get returns the current configuration.
func Get() Config {
	return current
}

// SetConfig sets the current configuration (useful for testing).
func SetConfig(cfg Config) {
	current = cfg
}

// splitCommaSeparated splits a comma-separated string into a slice of strings.
// Returns an empty slice if the input string is empty.
func splitCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// RegisterFlags registers configuration flags with the analyzer's FlagSet.
func RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&current.AllowMethodConverters, "include-methods", current.AllowMethodConverters,
		"check method receivers in addition to plain functions")

	fs.BoolVar(&current.AllowGetters, "allow-getters", current.AllowGetters,
		"allow Get* methods as a substitute for direct field access")

	fs.BoolVar(&current.AllowAggregators, "allow-aggregators", current.AllowAggregators,
		"enable detection of slice->non-slice aggregating converters")

	fs.Func(
		"exclude-fields",
		"comma-separated regex patterns for field names to ignore (e.g., 'CreatedAt,UpdatedAt,.*ID')",
		func(s string) error {
			current.ExcludeFieldPatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Func(
		"exclude-converters",
		"comma-separated glob patterns for function/method names to exclude from converter detection (e.g., 'Get*,Map*,to*')",
		func(s string) error {
			current.ExcludeConverterPatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Func(
		"exclude-files",
		"comma-separated glob patterns for file paths to exclude from analysis (e.g., '*_test.go,*.pb.go,*/vendor/*')",
		func(s string) error {
			current.ExcludeFilePatterns = splitCommaSeparated(s)
			return nil
		},
	)

	fs.Float64Var(&current.MinTypeNameSimilarity, "min-similarity", current.MinTypeNameSimilarity,
		"minimum type name similarity ratio (0.0-1.0, 0=substring matching, higher=stricter)")

	fs.Func("ignore-tags", "comma-separated struct tags to ignore fields (e.g., 'lostfield:\"ignore\"')",
		func(s string) error {
			current.IgnoreFieldTags = splitCommaSeparated(s)
			return nil
		})

	fs.BoolVar(&current.Verbose, "verbose", current.Verbose,
		"enable verbose output")

	fs.BoolVar(&current.IncludeGenerated, "include-generated", current.IncludeGenerated,
		"include generated code files in analysis (default: exclude)")

	fs.BoolVar(&current.IgnoreDeprecated, "include-deprecated", current.IgnoreDeprecated,
		"include deprecated fields in validation (default: exclude)")

	fs.StringVar(&current.Format, "format", current.Format,
		"output format for diagnostics (default: standard go vet format, custom: Rust-like pretty format)")

	fs.Func(
		"non-marshallable-fields",
		"how to handle non-marshallable field types (ignore, adaptive, strict)",
		func(s string) error {
			current.NonMarshallableFieldsHandling = NonMarshallableFieldsHandling(s)
			return nil
		},
	)

	fs.BoolVar(&current.IncludePrivateFields, "include-private-fields", current.IncludePrivateFields,
		"validate unexported (private) fields in converters (default: ignore private fields)")
}
