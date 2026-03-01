package formatter

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/tools/go/analysis"
)

// prettyFormatter implements a Rust-like pretty diagnostic format.
// It is completely independent and creates its own detailed message from raw validation data.
type prettyFormatter struct{}

// Format produces a Rust-like pretty diagnostic message with colors and formatting.
// It independently formats the message from the raw validation data.
func (c *prettyFormatter) Format(ctx *FormatContext) string {
	// Create the detailed field mapping message from raw validation data
	message := c.formatValidationMessage(ctx.Validation, ctx.Verbose)

	var buf bytes.Buffer
	c.prettyPrint(&buf, ctx.Filename, ctx.Fn, ctx.Pass, message, ctx.Validation.ConverterType)
	return buf.String()
}

// maxFieldsPerSide is the maximum number of fields to display per side (input/output)
// before truncating with a "... and N more" message.
const maxFieldsPerSide = 3

// formatValidationMessage creates a detailed field mapping message from raw validation data.
// This is the core message that the pretty printer will display.
// When verbose is false, fields are truncated to maxFieldsPerSide per side with a hint.
func (c *prettyFormatter) formatValidationMessage(validation *ConverterValidationResult, verbose bool) string {
	var buf strings.Builder

	totalMissing := len(validation.MissingInputFields) + len(validation.MissingOutputFields)
	if totalMissing == 0 {
		buf.WriteString("= note: missing fields:\n")
		return buf.String()
	}

	fmt.Fprintf(&buf, "= note: missing fields (%d):\n", totalMissing)

	inFields := validation.MissingInputFields
	outFields := validation.MissingOutputFields

	// Determine how many fields to show per side
	inShow := len(inFields)
	outShow := len(outFields)
	if !verbose {
		inShow = min(len(inFields), maxFieldsPerSide)
		outShow = min(len(outFields), maxFieldsPerSide)
	}

	// Calculate the maximum length for alignment of the arrow (only from visible fields)
	maxLen := 2 // minimum width for "??"
	for _, field := range inFields[:inShow] {
		if len(field) > maxLen {
			maxLen = len(field)
		}
	}

	// Add input fields (missing in output mapping)
	for _, field := range inFields[:inShow] {
		padding := strings.Repeat(" ", maxLen-len(field)+1)
		buf.WriteString("\n  " + field + padding + "→ ??")
	}
	inRemaining := len(inFields) - inShow
	if inRemaining > 0 {
		fmt.Fprintf(&buf, "\n  ... and %d more input fields", inRemaining)
	}

	// Add output fields (missing in input mapping)
	for _, field := range outFields[:outShow] {
		padding := strings.Repeat(" ", maxLen-len("??")+1)
		buf.WriteString("\n  " + "??" + padding + "→ " + field)
	}
	outRemaining := len(outFields) - outShow
	if outRemaining > 0 {
		fmt.Fprintf(&buf, "\n  ... and %d more output fields", outRemaining)
	}

	// Add re-run hint when any fields were truncated
	if inRemaining > 0 || outRemaining > 0 {
		buf.WriteString("\n  hint: re-run with -lostfield.verbose for full list, or -lostfield.only-converters=\"FuncName\" to target a specific function")
	}

	return buf.String()
}

// prettyPrint writes a linter message in a Rust-like style to the given writer.
// It extracts the source line from the file (using filename and pos.Line), shortens it to a maximum
// width (120 characters) while preserving the significant ranges, adjusts the caret position, and prints
// the formatted diagnostic.
func (c *prettyFormatter) prettyPrint(
	w *bytes.Buffer,
	filename string,
	fn *ast.FuncDecl,
	pass *analysis.Pass,
	message string,
	converterType string,
) {
	pos := pass.Fset.Position(fn.Name.Pos())

	// Open the file.
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(w, "error opening file %q: %v\n", filename, err)
		return
	}
	defer file.Close()

	// Read the file line-by-line until we get to the desired line.
	var sourceLine string
	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == pos.Line {
			sourceLine = scanner.Text()
			break
		}
		currentLine++
	}
	if sourceLine == "" {
		sourceLine = "<source unavailable>"
	}

	// Shorten the source line.
	maxWidth := 120
	shortLine := shortenLine(sourceLine, maxWidth)

	// Determine the gutter width (using the line number).
	lineNumStr := strconv.Itoa(pos.Line)
	gutterWidth := len(lineNumStr)

	// Adjust the caret: we know pos.Column is in the original line.
	// Estimate how many characters were trimmed from the beginning.
	origCaret := pos.Column - 1 // 0-indexed in the original line.
	trimmed := 0
	if strings.HasPrefix(shortLine, "…") {
		// Find the index in the original line where the shortened part begins.
		idx := strings.Index(sourceLine, shortLine[3:])
		if idx >= 0 {
			trimmed = idx
		}
	}
	newCaret := origCaret - trimmed
	if strings.HasPrefix(shortLine, "…") {
		newCaret += 1 // account for the ellipsis.
	}
	if newCaret < 0 {
		newCaret = 0
	}
	if newCaret > len(shortLine) {
		newCaret = len(shortLine)
	}

	// Prepare colored output.
	// Force color output even when output is not a TTY (when captured by go vet).
	// Note: We set the global NoColor flag to ensure colors are rendered.
	//nolint:reassign // Set global flag to force color output in go vet
	color.NoColor = false

	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	_ = bold

	// Print header.
	fnName := fn.Name.Name
	fnNameLen := len(fnName)

	// Print with extra spacing (4 spaces min) before the function code
	const minSpacing = 4
	fmt.Fprintf(w, "\n%*s %s\n", gutterWidth, "", blue("|"))
	lineNum := blue(fmt.Sprintf("%*d", gutterWidth, pos.Line))
	pipe := blue(" |")
	//nolint:gosec // CLI output, not web
	fmt.Fprintf(w, "%s%s %s %*s%s\n",
		lineNum, pipe, "", minSpacing, "", shortLine)

	// Adjust caret position to account for the min spacing
	caretLine := strings.Repeat(" ", newCaret+minSpacing) + yellow(strings.Repeat("^", fnNameLen))

	// Show converter type on the caret line
	var typeLabel string
	if converterType != "" {
		typeLabel = " detected as " + converterType
	}
	//nolint:gosec // CLI output, not web
	fmt.Fprintf(w, "%*s %s  %s%s\n",
		gutterWidth, "", blue("|"), caretLine, yellow(typeLabel))

	// Add blank line with just the pipe
	fmt.Fprintf(w, "%*s %s\n", gutterWidth, "", blue("|"))

	// Handle multi-line messages (the note section)
	messageLines := strings.Split(message, "\n")
	for i, line := range messageLines {
		if line != "" {
			if i == 0 && strings.HasPrefix(line, "=") {
				// Color the '=' in blue, rest in red
				eqIndex := strings.Index(line, "=")
				if eqIndex >= 0 {
					fmt.Fprintf(w, "%*s %s%s\n", gutterWidth, "", blue("="), red(line[1:]))
				} else {
					fmt.Fprintf(w, "%*s %s\n", gutterWidth, "", red(line))
				}
			} else {
				fmt.Fprintf(w, "%*s %s\n", gutterWidth, "", red(line))
			}
		}
	}
}

// shortenLine shortens a given line to at most maxWidth characters while preserving
// significant ranges. If any portion is omitted, ellipses ("…") are inserted accordingly.
func shortenLine(line string, maxWidth int) string {
	if len(line) <= maxWidth {
		return line
	}

	// If the line is longer than maxWidth, return its first maxWidth characters.
	return line[:maxWidth-1] + "…"
}
