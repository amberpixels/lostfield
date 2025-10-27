package config

import "flag"

// Linter metadata constants
const (
	LinterName = "stickyfields"
	LinterDoc  = "reports all inconsistent converter functions: ensures sticky fields)"
)

// Config holds all configuration for the analyzer
type Config struct {
	// IncludeMethods enables checking method receivers in addition to plain functions
	IncludeMethods bool

	// AllowGetters allows Get* methods as a substitute for field access
	AllowGetters bool

	// ExcludeFieldPatterns is a comma-separated list of regex patterns for field names to ignore
	ExcludeFieldPatterns string

	// MinTypeSimilarity is the minimum type name similarity ratio (0.0-1.0, 0=substring matching)
	MinTypeSimilarity float64

	// IgnoreFieldTags is a comma-separated list of struct tags that mark fields to ignore
	IgnoreFieldTags string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		IncludeMethods:       false,
		AllowGetters:         true,
		ExcludeFieldPatterns: "",
		MinTypeSimilarity:    0.0, // 0 = use substring matching (current behavior)
		IgnoreFieldTags:      "",
	}
}

// current holds the active configuration
var current = DefaultConfig()

// Get returns the current configuration
func Get() Config {
	return current
}

// RegisterFlags registers configuration flags with the analyzer's FlagSet
func RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&current.IncludeMethods, "include-methods", current.IncludeMethods,
		"check method receivers in addition to plain functions")

	fs.BoolVar(&current.AllowGetters, "allow-getters", current.AllowGetters,
		"allow Get* methods as a substitute for direct field access")

	fs.StringVar(&current.ExcludeFieldPatterns, "exclude-fields", current.ExcludeFieldPatterns,
		"comma-separated regex patterns for field names to ignore (e.g., 'CreatedAt,UpdatedAt,.*ID')")

	fs.Float64Var(&current.MinTypeSimilarity, "min-similarity", current.MinTypeSimilarity,
		"minimum type name similarity ratio (0.0-1.0, 0=substring matching, higher=stricter)")

	fs.StringVar(&current.IgnoreFieldTags, "ignore-tags", current.IgnoreFieldTags,
		"comma-separated struct tags to ignore fields (e.g., 'stickyfields:\"ignore\"')")
}
