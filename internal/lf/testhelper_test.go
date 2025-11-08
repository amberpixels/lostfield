package lf_test

import (
	"strings"
	"testing"

	"github.com/amberpixels/lostfield/internal/lf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// DiagnosticAssertion represents an expected diagnostic with its properties.
type DiagnosticAssertion struct {
	FunctionName  string   // Name of the converter function
	FieldsMissing []string // Exact list of missing fields - must match exactly (no more, no less)
	// If FieldsMissing is empty, the converter is expected to be valid (no errors)
}

// runAnalysisTest runs the lostfield analyzer on a test package and validates against assertions.
// It reads the results from analysistest.Run() and asserts the diagnostics match expectations.
// Usage: runAnalysisTest(t, "converters/myPackage", assertion1, assertion2, ...)
func runAnalysisTest(t *testing.T, pkgPath string, assertions ...DiagnosticAssertion) {
	t.Helper()

	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// analysistest.Run() returns []*Result which has Diagnostics field
	results := analysistest.Run(t, testdata, analyzer, pkgPath)

	// Collect all diagnostics from all results
	var diagnostics []*analysis.Diagnostic
	for _, result := range results {
		for i := range result.Diagnostics {
			diagnostics = append(diagnostics, &result.Diagnostics[i])
		}
	}

	// Check count matches expectations
	if len(diagnostics) != len(assertions) {
		PrintDiagnostics(t, diagnostics)
		t.Fatalf("expected %d diagnostics, got %d", len(assertions), len(diagnostics))
	}

	// Check each diagnostic against its assertion
	for i, assertion := range assertions {
		msg := diagnostics[i].Message

		// Verify function name is in the message
		if !strings.Contains(msg, assertion.FunctionName) {
			t.Errorf("diagnostic %d: message should contain function name %q\n  Full message: %s",
				i, assertion.FunctionName, msg)
			continue
		}

		// Extract and validate fields
		reportedFields := extractMissingFields(msg)

		if !fieldsMatch(reportedFields, assertion.FieldsMissing) {
			t.Errorf("diagnostic %d: fields mismatch\n  Expected: %v\n  Got: %v\n  Message: %s",
				i, assertion.FieldsMissing, reportedFields, msg)
		}
	}
}

// extractMissingFields extracts the list of missing fields from a diagnostic message.
// Expected format: "FunctionName: incomplete converter with missing fields: field1, field2, field3".
func extractMissingFields(msg string) []string {
	// Look for "missing fields: " in the message
	marker := "missing fields: "
	idx := strings.Index(msg, marker)
	if idx == -1 {
		// No missing fields section found, converter is valid
		return []string{}
	}

	// Extract everything after "missing fields: "
	fieldsStr := msg[idx+len(marker):]

	// Split by comma and trim whitespace
	var fields []string
	for _, field := range strings.Split(fieldsStr, ",") {
		trimmed := strings.TrimSpace(field)
		if trimmed != "" {
			fields = append(fields, trimmed)
		}
	}

	return fields
}

// fieldsMatch checks if two field lists are exactly equal (same fields, same order).
func fieldsMatch(got, expected []string) bool {
	if len(got) != len(expected) {
		return false
	}

	for i := range got {
		if got[i] != expected[i] {
			return false
		}
	}

	return true
}

// PrintDiagnostics prints all diagnostics for debugging purposes.
func PrintDiagnostics(t *testing.T, diagnostics []*analysis.Diagnostic) {
	t.Helper()
	t.Logf("Total diagnostics: %d", len(diagnostics))
	for i, diag := range diagnostics {
		t.Logf("  [%d] %s", i, diag.Message)
	}
}
