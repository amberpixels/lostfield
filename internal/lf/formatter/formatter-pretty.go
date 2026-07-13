package formatter

import (
	"bytes"
	"fmt"
	"go/ast"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"golang.org/x/tools/go/analysis"
)

// prettyFormatter implements a Rust-like pretty diagnostic format.
// It is completely independent and creates its own detailed message from raw validation data.
//
// The pretty format is intended for humans reading terminal output; it embeds
// ANSI colors and multi-line source excerpts directly into the diagnostic message.
// Do not use it with machine-readable consumers (go vet -json, editors, golangci-lint) -
// use the default format there.
type prettyFormatter struct {
	// colorize enables ANSI colors. Colors are on by default (output usually goes
	// through `go vet`, which pipes it, so TTY auto-detection would always say no),
	// but the standard NO_COLOR convention (https://no-color.org) is respected.
	colorize bool

	// fileLines caches source file contents per filename for the current formatter
	// lifetime (one Run call), so each file is read at most once per run.
	fileLines map[string][]string
}

func newPrettyFormatter() *prettyFormatter {
	return &prettyFormatter{
		colorize:  os.Getenv("NO_COLOR") == "",
		fileLines: make(map[string][]string),
	}
}

// sprintFunc returns a colorizing sprint function for attrs, or a plain
// pass-through when colors are disabled.
func (c *prettyFormatter) sprintFunc(attrs ...color.Attribute) func(a ...any) string {
	if !c.colorize {
		return fmt.Sprint
	}
	col := color.New(attrs...)
	col.EnableColor() // per-instance: no global color.NoColor mutation
	return col.SprintFunc()
}

// Format produces a Rust-like pretty diagnostic message with colors and formatting.
// It independently formats the message from the raw validation data.
func (c *prettyFormatter) Format(ctx *FormatContext) string {
	// Create the detailed field mapping message from raw validation data
	message := c.formatValidationMessage(ctx.Validation, ctx.Verbose)

	var buf bytes.Buffer
	c.prettyPrint(&buf, ctx.Filename, ctx.Fn, ctx.Pass, message, ctx.Validation.ConverterType, ctx.Index, ctx.Total)
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
		buf.WriteString("\n  hint: re-run with -lostfield.verbose for full list," +
			" or -lostfield.only-converters=\"FuncName\" to target a specific function")
	}

	return buf.String()
}

// sourceLine returns the 1-based line of the given file, reading and caching the
// file contents on first access. Returns "" when the file or line is unavailable.
func (c *prettyFormatter) sourceLine(filename string, line int) string {
	lines, ok := c.fileLines[filename]
	if !ok {
		data, err := os.ReadFile(filename)
		if err != nil {
			c.fileLines[filename] = nil
			return ""
		}
		lines = strings.Split(string(data), "\n")
		c.fileLines[filename] = lines
	}
	if line < 1 || line > len(lines) {
		return ""
	}
	return lines[line-1]
}

// prettyPrint writes a linter message in a Rust-like style to the given writer.
// It extracts the source line from the file (using filename and pos.Line), shortens it to a maximum
// width (120 characters) while preserving the beginning, adjusts the caret position, and prints
// the formatted diagnostic.
func (c *prettyFormatter) prettyPrint(
	w *bytes.Buffer,
	filename string,
	fn *ast.FuncDecl,
	pass *analysis.Pass,
	message string,
	converterType string,
	index, total int,
) {
	pos := pass.Fset.Position(fn.Name.Pos())

	sourceLine := c.sourceLine(filename, pos.Line)
	if sourceLine == "" {
		sourceLine = "<source unavailable>"
	}

	// Shorten the source line (rune-safe).
	const maxWidth = 120
	shortLine := shortenLine(sourceLine, maxWidth)

	// Determine the gutter width (using the line number).
	lineNumStr := strconv.Itoa(pos.Line)
	gutterWidth := len(lineNumStr)

	// pos.Column is a 1-based byte offset in the original line; convert it to a
	// rune-based caret column so multibyte characters don't shift the caret.
	byteCaret := max(min(pos.Column-1, len(sourceLine)), 0)
	newCaret := min(
		utf8.RuneCountInString(sourceLine[:byteCaret]),
		utf8.RuneCountInString(shortLine),
	)

	blue := c.sprintFunc(color.FgBlue)
	red := c.sprintFunc(color.FgRed)
	yellow := c.sprintFunc(color.FgYellow)

	// Print header.
	fnName := fn.Name.Name
	fnNameLen := utf8.RuneCountInString(fnName)

	// Print with extra spacing (4 spaces min) before the function code
	const minSpacing = 4
	fmt.Fprintf(w, "\n%*s %s\n", gutterWidth, "", blue("|"))
	lineNum := blue(fmt.Sprintf("%*d", gutterWidth, pos.Line))
	pipe := blue(" |")
	fmt.Fprintf(w, "%s%s %s %*s%s\n",
		lineNum, pipe, "", minSpacing, "", shortLine)

	// Adjust caret position to account for the min spacing
	caretLine := strings.Repeat(" ", newCaret+minSpacing) + yellow(strings.Repeat("^", fnNameLen))

	// Show converter type on the caret line
	var typeLabel string
	if converterType != "" {
		typeLabel = " detected as " + converterType
	}
	fmt.Fprintf(w, "%*s %s  %s%s\n",
		gutterWidth, "", blue("|"), caretLine, yellow(typeLabel))

	// Add blank line with just the pipe
	fmt.Fprintf(w, "%*s %s\n", gutterWidth, "", blue("|"))

	// Build numbering label for multi-diagnostic runs
	var numLabel string
	if total > 1 {
		numLabel = fmt.Sprintf(" [%d/%d]", index, total)
	}

	// Handle multi-line messages (the note section)
	messageLines := strings.Split(message, "\n")
	for i, line := range messageLines {
		if line != "" {
			if i == 0 && strings.HasPrefix(line, "=") {
				// Color the '=' in blue, numbering in yellow, rest in red
				fmt.Fprintf(w, "%*s %s%s%s\n", gutterWidth, "", blue("="), yellow(numLabel), red(line[1:]))
			} else {
				fmt.Fprintf(w, "%*s %s\n", gutterWidth, "", red(line))
			}
		}
	}
}

// shortenLine shortens a given line to at most maxWidth runes. If the line is
// truncated, an ellipsis ("…") replaces the tail. Truncation is rune-safe:
// multibyte characters are never split.
func shortenLine(line string, maxWidth int) string {
	runes := []rune(line)
	if len(runes) <= maxWidth {
		return line
	}
	return string(runes[:maxWidth-1]) + "…"
}
