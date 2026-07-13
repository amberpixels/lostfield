// Package lostfield exposes the lostfield analyzer: a go/analysis linter that
// ensures converter functions use all fields from both input and output structs,
// preventing "leaky" conversions where fields are accidentally omitted.
//
// Standalone usage (go vet):
//
//	go vet -vettool=$(which lostfield) ./...
//
// Programmatic usage (e.g. golangci-lint integration, custom multicheckers):
//
//	cfg := lostfield.DefaultConfig()
//	cfg.ExcludeFieldPatterns = []string{"CreatedAt", "UpdatedAt"}
//	analyzer := lostfield.NewAnalyzer(cfg)
package lostfield

import (
	"flag"

	"github.com/amberpixels/lostfield/internal/config"
	"github.com/amberpixels/lostfield/internal/lf"
	"golang.org/x/tools/go/analysis"
)

// Config holds all configuration for the lostfield analyzer.
// See the field documentation for defaults and semantics.
type Config = config.Config

// Enum types for Config fields.
type (
	// NonMarshallableFieldsHandling specifies how to handle non-marshallable field types.
	NonMarshallableFieldsHandling = config.NonMarshallableFieldsHandling
	// FieldValidationMode specifies which fields must be validated in a converter.
	FieldValidationMode = config.FieldValidationMode
	// Format specifies the output format for diagnostics.
	Format = config.Format
	// FixMode controls whether diagnostics carry SuggestedFixes.
	FixMode = config.FixMode
)

// Re-exported enum values, so importers never need the internal package.
const (
	HandleIgnore   = config.HandleIgnore
	HandleAdaptive = config.HandleAdaptive
	HandleStrict   = config.HandleStrict

	ModeStrict       = config.ModeStrict
	ModeIntersection = config.ModeIntersection

	FormatDefault = config.FormatDefault
	FormatPretty  = config.FormatPretty

	FixModeDisabled = config.FixModeDisabled
	FixModeSafe     = config.FixModeSafe
	FixModeSmart    = config.FixModeSmart
)

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	cfg := config.DefaultConfig()
	return &cfg
}

// NewAnalyzer builds a lostfield analysis.Analyzer bound to the given configuration.
// Passing nil uses DefaultConfig(). The returned analyzer holds no global state, so
// multiple analyzers with different configs can coexist in one process.
//
// The configuration is validated; an invalid enum value makes the analyzer fail
// at run time with the validation error (analysis.Analyzer construction cannot
// return an error).
func NewAnalyzer(cfg *Config) *analysis.Analyzer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if err := cfg.Validate(); err != nil {
		return &analysis.Analyzer{
			Name: config.LinterName,
			Doc:  config.LinterDoc,
			Run: func(*analysis.Pass) (any, error) {
				return nil, err
			},
		}
	}
	return lf.NewAnalyzer(cfg)
}

// NewAnalyzerWithFlags builds the analyzer with its config bound to CLI flags,
// as used by the standalone `lostfield` binary (unitchecker). Flag values are
// validated at parse time.
func NewAnalyzerWithFlags() *analysis.Analyzer {
	cfg := DefaultConfig()
	analyzer := lf.NewAnalyzer(cfg)

	var fs flag.FlagSet
	config.RegisterFlags(&fs, cfg)
	analyzer.Flags = fs

	return analyzer
}
