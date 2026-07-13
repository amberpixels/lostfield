package formatter_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/expectto/be"
	"github.com/expectto/be/be_string"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/analysis"

	"github.com/amberpixels/lostfield/internal/lf/formatter"
)

// converterSrc is a minimal source file with a converter-shaped function,
// used to build a real formatter.FormatContext (Fset positions + on-disk source line).
const converterSrc = `package sample

type User struct {
	ID    int64
	Email string
}

type UserDTO struct {
	ID    int64
	Email string
}

func ConvertUser(u User) UserDTO {
	return UserDTO{ID: u.ID}
}
`

// buildFormatContext writes src to a temp file, parses it and returns a
// formatter.FormatContext for the named function with the given validation payload.
func buildFormatContext(
	t *testing.T,
	validation *formatter.ConverterValidationResult,
) *formatter.FormatContext {
	src, funcName := converterSrc, "ConvertUser"
	t.Helper()

	filename := filepath.Join(t.TempDir(), "converter.go")
	if err := os.WriteFile(filename, []byte(src), 0o600); err != nil {
		t.Fatalf("writing temp source: %v", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing temp source: %v", err)
	}

	var fn *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if decl, ok := n.(*ast.FuncDecl); ok && decl.Name.Name == funcName {
			fn = decl
			return false
		}
		return true
	})
	if fn == nil {
		t.Fatalf("function %q not found in source", funcName)
	}

	return &formatter.FormatContext{
		Filename:   filename,
		Fn:         fn,
		Pass:       &analysis.Pass{Fset: fset},
		Validation: validation,
		Index:      1,
		Total:      1,
	}
}

func TestDefaultFormat(t *testing.T) {
	g := NewWithT(t)

	t.Run("single diagnostic has no numbering prefix", func(t *testing.T) {
		g := NewWithT(t)
		ctx := buildFormatContext(t, &formatter.ConverterValidationResult{
			ConverterType:       "converter",
			MissingInputFields:  []string{"u.Email"},
			MissingOutputFields: []string{"Email"},
		})

		out := formatter.New(formatter.FormatterDefault).Format(ctx)
		g.Expect(out).To(be.Eq("ConvertUser: incomplete converter with missing fields: u.Email, Email"))
	})

	t.Run("multiple diagnostics get [i/N] numbering", func(t *testing.T) {
		g := NewWithT(t)
		ctx := buildFormatContext(t, &formatter.ConverterValidationResult{
			MissingInputFields: []string{"u.Email"},
		})
		ctx.Index, ctx.Total = 2, 3

		out := formatter.New(formatter.FormatterDefault).Format(ctx)
		g.Expect(out).To(be.All(
			be_string.HavingPrefix("[2/3] "),
			be_string.ContainingSubstring("ConvertUser: incomplete converter with missing fields: u.Email"),
		))
	})

	t.Run("no missing fields still names the converter", func(t *testing.T) {
		g := NewWithT(t)
		ctx := buildFormatContext(t, &formatter.ConverterValidationResult{})

		out := formatter.New(formatter.FormatterDefault).Format(ctx)
		g.Expect(out).To(be.Eq("ConvertUser: incomplete converter"))
	})

	// The default format must stay single-line and ANSI-free: it is the format
	// consumed by go vet -json, editors and golangci-lint.
	ctx := buildFormatContext(t, &formatter.ConverterValidationResult{
		MissingInputFields: []string{"u.Email"},
	})
	out := formatter.New(formatter.FormatterDefault).Format(ctx)
	g.Expect(out).NotTo(be.Any(
		be_string.ContainingSubstring("\x1b["),
		be_string.ContainingSubstring("\n"),
	))
}

func TestPrettyFormat(t *testing.T) {
	newValidation := func(inFields, outFields []string) *formatter.ConverterValidationResult {
		return &formatter.ConverterValidationResult{
			ConverterType:       "converter",
			MissingInputFields:  inFields,
			MissingOutputFields: outFields,
		}
	}

	t.Run("plain output with NO_COLOR", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "1")

		ctx := buildFormatContext(t,
			newValidation([]string{"u.Email"}, []string{"Email"}))

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be.All(
			be_string.ContainingSubstring("func ConvertUser(u User) UserDTO {"),
			be_string.ContainingSubstring("^^^^^^^^^^^ detected as converter"),
			be_string.ContainingSubstring("= note: missing fields (2):"),
			be_string.ContainingSubstring("u.Email → ??"),
			be_string.ContainingSubstring("??      → Email"),
		))
		g.Expect(out).NotTo(be_string.ContainingSubstring("\x1b["))
	})

	t.Run("colored output when NO_COLOR is unset", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "")

		ctx := buildFormatContext(t,
			newValidation([]string{"u.Email"}, nil))

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be_string.ContainingSubstring("\x1b["))
	})

	t.Run("truncates to 3 fields per side with a hint", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "1")

		ctx := buildFormatContext(t, newValidation(
			[]string{"u.A", "u.B", "u.C", "u.D", "u.E"},
			[]string{"A", "B", "C", "D"},
		))

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be.All(
			be_string.ContainingSubstring("... and 2 more input fields"),
			be_string.ContainingSubstring("... and 1 more output fields"),
			be_string.ContainingSubstring("hint: re-run with -lostfield.verbose"),
		))
		g.Expect(out).NotTo(be_string.ContainingSubstring("u.D →"))
	})

	t.Run("verbose shows all fields without a hint", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "1")

		ctx := buildFormatContext(t, newValidation(
			[]string{"u.A", "u.B", "u.C", "u.D", "u.E"},
			nil,
		))
		ctx.Verbose = true

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be.All(
			be_string.ContainingSubstring("u.D"),
			be_string.ContainingSubstring("u.E"),
		))
		g.Expect(out).NotTo(be_string.ContainingSubstring("hint: re-run"))
	})

	t.Run("numbering label on multi-diagnostic runs", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "1")

		ctx := buildFormatContext(t,
			newValidation([]string{"u.Email"}, nil))
		ctx.Index, ctx.Total = 2, 5

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be_string.ContainingSubstring("[2/5]"))
	})

	t.Run("survives unreadable source file", func(t *testing.T) {
		g := NewWithT(t)
		t.Setenv("NO_COLOR", "1")

		ctx := buildFormatContext(t,
			newValidation([]string{"u.Email"}, nil))
		ctx.Filename = filepath.Join(t.TempDir(), "does-not-exist.go")

		out := formatter.New(formatter.FormatterPretty).Format(ctx)
		g.Expect(out).To(be_string.ContainingSubstring("<source unavailable>"))
	})
}

func TestShortenLine(t *testing.T) {
	g := NewWithT(t)

	t.Run("short lines pass through", func(t *testing.T) {
		g := NewWithT(t)
		g.Expect(formatter.ShortenLine("short", 120)).To(be.Eq("short"))
	})

	t.Run("long ASCII lines are truncated with ellipsis", func(t *testing.T) {
		g := NewWithT(t)
		out := formatter.ShortenLine(strings.Repeat("x", 200), 120)
		g.Expect(out).To(be_string.HavingSuffix("…"))
		g.Expect(utf8.RuneCountInString(out)).To(be.Eq(120))
	})

	// Truncation must be rune-safe: multibyte characters are never split.
	out := formatter.ShortenLine(strings.Repeat("é", 200), 120)
	g.Expect(utf8.ValidString(out)).To(be.True())
	g.Expect(utf8.RuneCountInString(out)).To(be.Eq(120))
	g.Expect(out).To(be_string.HavingSuffix("…"))
}
