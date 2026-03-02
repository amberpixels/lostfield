package fixer_test

import (
	"strings"
	"testing"

	"github.com/amberpixels/lostfield/internal/lf/fixer"
)

func TestGenerateFixes_NilContext(t *testing.T) {
	fixes := fixer.GenerateFixes(nil, &fixer.ValidationResult{
		MissingInputFields: []string{"x.Foo"},
	}, "safe")
	if fixes != nil {
		t.Error("expected nil fixes for nil context")
	}
}

func TestGenerateFixes_EmptyMode(t *testing.T) {
	fixes := fixer.GenerateFixes(&fixer.FixContext{}, &fixer.ValidationResult{
		MissingInputFields: []string{"x.Foo"},
	}, "")
	if fixes != nil {
		t.Error("expected nil fixes for empty mode")
	}
}

func TestGenerateFixes_SafeMode(t *testing.T) {
	ctx := &fixer.FixContext{
		InVar:        "sample",
		OutVar:       "result",
		InFieldVar:   "sample",
		FnBodyLbrace: 100,
		FnBodyRbrace: 200,
		OutputStyle:  fixer.OutputStyleDotAssignment,
	}
	validation := &fixer.ValidationResult{
		MissingInputFields:  []string{"sample.Price"},
		MissingOutputFields: []string{"result.Currency"},
	}

	fixes := fixer.GenerateFixes(ctx, validation, "safe")
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	if fixes[0].Message != "Suppress warnings with _ = var.Field" {
		t.Errorf("unexpected fix message: %s", fixes[0].Message)
	}
	if len(fixes[0].TextEdits) != 2 {
		t.Errorf("expected 2 text edits, got %d", len(fixes[0].TextEdits))
	}
}

func TestGenerateFixes_SmartMode(t *testing.T) {
	ctx := &fixer.FixContext{
		InVar:        "sample",
		OutVar:       "result",
		InFieldVar:   "sample",
		FnBodyLbrace: 100,
		FnBodyRbrace: 200,
		OutputStyle:  fixer.OutputStyleDotAssignment,
	}
	validation := &fixer.ValidationResult{
		MissingInputFields:  []string{"sample.Price"},
		MissingOutputFields: []string{"result.Price"},
	}

	fixes := fixer.GenerateFixes(ctx, validation, "smart")
	// Should have smart fix first, safe fix second
	if len(fixes) != 2 {
		t.Fatalf("expected 2 fixes, got %d", len(fixes))
	}
	if fixes[0].Message != "Auto-fix field mappings (smart)" {
		t.Errorf("expected smart fix first, got: %s", fixes[0].Message)
	}
	if fixes[1].Message != "Suppress warnings with _ = var.Field" {
		t.Errorf("expected safe fix second, got: %s", fixes[1].Message)
	}
}

func TestGenerateFixes_SliceInline(t *testing.T) {
	ctx := &fixer.FixContext{
		InVar:         "items",
		OutVar:        "result",
		InFieldVar:    "item",
		FnBodyLbrace:  100,
		FnBodyRbrace:  200,
		OutputStyle:   fixer.OutputStyleDotAssignment,
		IsSliceInline: true,
	}
	validation := &fixer.ValidationResult{
		MissingInputFields: []string{"item.Weight"},
	}

	fixes := fixer.GenerateFixes(ctx, validation, "safe")
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	// Should contain TODO comment instead of _ = stub
	text := string(fixes[0].TextEdits[0].NewText)
	if text == "" {
		t.Error("expected non-empty text edit")
	}
	if !strings.Contains(text, "TODO(lostfield)") {
		t.Errorf("expected TODO comment for slice inline, got: %s", text)
	}
}
