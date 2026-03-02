package formatter

import (
	"fmt"
	"strings"
)

// defaultFormatter implements the standard go vet output format.
// This format is minimal and follows Go conventions for linter output.
// It independently creates a concise message from the raw validation data.
type defaultFormatter struct{}

// Format produces a standard go vet format diagnostic message.
// Creates a concise, single-line message from raw validation data.
// Example output: "ToPM: incomplete converter with missing fields: Categories, Sections, URLValidated, Email".
func (d *defaultFormatter) Format(ctx *FormatContext) string {
	fnName := ctx.Fn.Name.Name
	validation := ctx.Validation

	// Collect all missing fields (both input and output)
	missingFields := append(
		[]string(nil),
		validation.MissingInputFields...,
	)
	missingFields = append(missingFields, validation.MissingOutputFields...)

	// Build numbering prefix if total > 1
	var prefix string
	if ctx.Total > 1 {
		prefix = fmt.Sprintf("[%d/%d] ", ctx.Index, ctx.Total)
	}

	// Build a standard go vet-style message
	if len(missingFields) == 0 {
		return fmt.Sprintf("%s%s: incomplete converter", prefix, fnName)
	}

	return fmt.Sprintf("%s%s: incomplete converter with missing fields: %s",
		prefix, fnName, strings.Join(missingFields, ", "))
}
