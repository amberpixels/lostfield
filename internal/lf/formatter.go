package lf

import (
	"bufio"
	"fmt"
	"go/ast"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/tools/go/analysis"
)

// PrettyPrint writes a linter message in a Rust-like style to the given writer.
// It extracts the source line from the file (using filename and pos.Line), shortens it to a maximum
// width (80 characters) while preserving the significant ranges, adjusts the caret position, and prints
// the formatted diagnostic.
// TODO: make a struct-base method (so we do not send `pass` via arg, etc)
func PrettyPrint(
	w io.Writer,
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
	// fmt.Fprintf(w, "\033]8;;file://%s:%d:%d\033\\%s\033]8;;\033\\\n",
	// 	filename, pos.Line, pos.Column,
	// 	blue(
	// 		fmt.Sprintf("--> %s:%d:%d", filename, pos.Line, pos.Column),
	// 	),
	// )

	fnName := fn.Name.Name
	fnNameLen := len(fnName)

	// Print with extra spacing (4 spaces min) before the function code
	const minSpacing = 4
	fmt.Fprintf(w, "\n%*s %s\n", gutterWidth, "", blue("|"))
	lineNum := blue(fmt.Sprintf("%*d", gutterWidth, pos.Line))
	pipe := blue(" |")
	fmt.Fprintf(w, "%s%s %s %*s%s\n", lineNum, pipe, "", minSpacing, "", shortLine)

	// Adjust caret position to account for the min spacing
	caretLine := strings.Repeat(" ", newCaret+minSpacing) + yellow(strings.Repeat("^", fnNameLen))

	// Show converter type on the caret line
	var typeLabel string
	if converterType != "" {
		typeLabel = " detected as " + converterType
	}
	fmt.Fprintf(w, "%*s %s  %s%s\n", gutterWidth, "", blue("|"), caretLine, yellow(typeLabel))

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
	// fmt.Fprintf(w, "\n%s: aborting due to previous error\n", bold("error"))
}

// shortenLine shortens a given line to at most maxWidth characters while preserving
// If any portion is omitted, ellipses ("...") are inserted accordingly.
func shortenLine(line string, maxWidth int) string {
	if len(line) <= maxWidth {
		return line
	}

	// If the significant block is longer than maxWidth, return its first maxWidth characters.
	return line[:maxWidth-1] + "…"
}
