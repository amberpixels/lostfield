package config

import "flag"

// Linter metadata constants.
const (
	LinterName = "stickyfields"
	LinterDoc  = "reports all inconsistent converter functions: ensures sticky fields)"
)

// Config holds all configuration for the analyzer.
type Config struct {
	// IncludeMethods enables checking method receivers in addition to plain functions
	IncludeMethods bool

	// AllowGetters allows Get* methods as a substitute for field access
	AllowGetters bool

	// AllowAggregatorsConverters enables detection of slice->non-slice converters
	// where the output struct contains a field that holds the converted slice
	AllowAggregatorsConverters bool

	// ExcludeFieldPatterns is a comma-separated list of regex patterns for field names to ignore
	ExcludeFieldPatterns string

	// MinTypeSimilarity is the minimum type name similarity ratio (0.0-1.0, 0=substring matching)
	MinTypeSimilarity float64

	// IgnoreFieldTags is a comma-separated list of struct tags that mark fields to ignore
	IgnoreFieldTags string

	// Verbose enables verbose output
	Verbose bool

	// ExcludeGenerated excludes generated code files from analysis
	// Detects files with "DO NOT EDIT" or similar markers
	ExcludeGenerated bool

	// ExcludeDeprecated excludes deprecated fields from validation.
	// Deprecated fields are identified by "Deprecated:" in their documentation comments.
	// This is useful when handling proto-generated code or legacy APIs where deprecated
	// fields are intentionally not converted.
	// Note: For protobuf-generated files to be recognized as deprecated, ensure the .pb.go
	// files are included in the analysis (they should be if you run "go vet ./..." in the
	// directory containing them).
	ExcludeDeprecated bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		IncludeMethods:             false,
		AllowGetters:               true,
		AllowAggregatorsConverters: false,
		ExcludeFieldPatterns:       "",
		MinTypeSimilarity:          0.0, // 0 = use substring matching (current behavior)
		IgnoreFieldTags:            "",
		Verbose:                    false,
		ExcludeGenerated:           true,
		ExcludeDeprecated:          false,
	}
}

// current holds the active configuration.
var current = DefaultConfig()

// Get returns the current configuration.
func Get() Config {
	return current
}

// RegisterFlags registers configuration flags with the analyzer's FlagSet.
func RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&current.IncludeMethods, "include-methods", current.IncludeMethods,
		"check method receivers in addition to plain functions")

	fs.BoolVar(&current.AllowGetters, "allow-getters", current.AllowGetters,
		"allow Get* methods as a substitute for direct field access")

	fs.BoolVar(&current.AllowAggregatorsConverters, "allow-aggregators", current.AllowAggregatorsConverters,
		"enable detection of slice->non-slice aggregating converters")

	fs.StringVar(&current.ExcludeFieldPatterns, "exclude-fields", current.ExcludeFieldPatterns,
		"comma-separated regex patterns for field names to ignore (e.g., 'CreatedAt,UpdatedAt,.*ID')")

	fs.Float64Var(&current.MinTypeSimilarity, "min-similarity", current.MinTypeSimilarity,
		"minimum type name similarity ratio (0.0-1.0, 0=substring matching, higher=stricter)")

	fs.StringVar(&current.IgnoreFieldTags, "ignore-tags", current.IgnoreFieldTags,
		"comma-separated struct tags to ignore fields (e.g., 'stickyfields:\"ignore\"')")

	fs.BoolVar(&current.Verbose, "verbose", current.Verbose,
		"enable verbose output")

	fs.BoolVar(&current.ExcludeGenerated, "exclude-generated", current.ExcludeGenerated,
		"exclude generated code files (detected via DO NOT EDIT markers)")

	fs.BoolVar(&current.ExcludeDeprecated, "exclude-deprecated", current.ExcludeDeprecated,
		"exclude deprecated fields from validation (marked with Deprecated: comments)")
}
