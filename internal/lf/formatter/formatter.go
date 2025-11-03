package formatter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	FormatterDefault = "default"
	FormatterPretty  = "pretty"
)

// ConverterValidationResult holds the raw validation results that formatters need.
// This allows formatters to be completely independent from each other.
type ConverterValidationResult struct {
	Valid               bool
	ConverterType       string
	MissingInputFields  []string
	MissingOutputFields []string
}

// FormatContext holds the context needed to format a diagnostic message.
// It contains raw data that each formatter can independently format as needed.
type FormatContext struct {
	Filename   string
	Fn         *ast.FuncDecl
	Pass       *analysis.Pass
	Validation *ConverterValidationResult
}

// Formatter interface defines how to format diagnostic messages.
type Formatter interface {
	Format(ctx *FormatContext) string
}

// New returns a formatter based on the format name.
// Supported formats: "default" (standard go vet), "pretty" (Rust-like).
// Returns default formatter for unknown format names.
func New(format string) Formatter {
	switch format {
	case FormatterPretty:
		return &prettyFormatter{}
	case FormatterDefault:
		fallthrough
	default:
		return &defaultFormatter{}
	}
}
