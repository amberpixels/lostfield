package sf

import (
	"bufio"
	"fmt"
	"go/token"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PrettyPrint writes a linter message in a Rust-like style to the given writer.
// It extracts the source line from the file (using filename and pos.Line), shortens it to a maximum
// width (80 characters) while preserving the significant ranges, adjusts the caret position, and prints
// the formatted diagnostic.
func PrettyPrint(w io.Writer, filename string, pos token.Position, message string) {
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
	lineNumStr := fmt.Sprintf("%d", pos.Line)
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
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	_ = blue
	_ = bold

	// Print header.
	// fmt.Fprintf(w, "--> %s:%d:%d\n", filename, pos.Line, pos.Column)
	fmt.Fprintf(w, "\033]8;;file://%s:%d:%d\033\\%s\033]8;;\033\\\n",
		filename, pos.Line, pos.Column,
		blue(
			fmt.Sprintf("--> %s:%d:%d", filename, pos.Line, pos.Column),
		),
	)

	fmt.Fprintf(w, "%*s |\n", gutterWidth, "")
	fmt.Fprintf(w, "%*d | %s\n", gutterWidth, pos.Line, shortLine)
	caretLine := strings.Repeat(" ", newCaret) + red("^")
	fmt.Fprintf(w, "%*s | %s %s\n", gutterWidth, "", caretLine, red(message))
	fmt.Fprintf(w, "\n")
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
