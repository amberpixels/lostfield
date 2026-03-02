package fixer

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OutputStyle indicates how the output value is constructed in a converter.
type OutputStyle int

const (
	// OutputStyleDotAssignment means fields are set via dot notation (e.g., result.Field = value).
	OutputStyleDotAssignment OutputStyle = iota
	// OutputStyleCompositeLit means the output is built via a composite literal (e.g., &Type{Field: value}).
	OutputStyleCompositeLit
)

// FixContext holds the context needed to generate suggested fixes for a converter.
type FixContext struct {
	InVar         string        // input variable name (e.g., "pbTicket")
	OutVar        string        // output variable name (e.g., "result")
	InFieldVar    string        // field access var (may differ for slice inline mapping)
	InStruct      *types.Struct // input struct type
	OutStruct     *types.Struct // output struct type
	InNamedType   *types.Named  // for getter method lookup
	FnBodyLbrace  token.Pos     // insertion point after opening brace
	FnBodyRbrace  token.Pos     // insertion point before closing brace
	LastReturnPos token.Pos     // position of last return statement (insert output stubs before this)
	OutputStyle   OutputStyle   // composite-lit vs dot-assignment
	CompLitRbrace token.Pos     // for inserting into composite literal (smart fix)
	IsSliceInline bool          // true when inFieldVar is a loop variable
}

// outInsertPos returns the best position to insert output stubs.
// Prefers before the last return statement to avoid unreachable code.
func (fc *FixContext) outInsertPos() token.Pos {
	if fc.LastReturnPos.IsValid() {
		return fc.LastReturnPos
	}
	return fc.FnBodyRbrace
}

// ValidationResult provides the missing fields needed for fix generation.
// This avoids an import cycle with the lf package.
type ValidationResult struct {
	MissingInputFields  []string
	MissingOutputFields []string
}

// GenerateFixes generates suggested fixes for a converter validation result.
// When mode is "smart", smart fix is listed first (applied by -fix), safe fix second.
// When mode is "safe", only safe fix is generated.
func GenerateFixes(
	fixCtx *FixContext,
	validation *ValidationResult,
	mode string,
) []analysis.SuggestedFix {
	if fixCtx == nil || mode == "" {
		return nil
	}

	// Separate missing fields into top-level names (strip var prefix).
	missingIn := topLevelFields(validation.MissingInputFields, fixCtx.InFieldVar)
	missingOut := topLevelFields(validation.MissingOutputFields, fixCtx.OutVar)

	if len(missingIn) == 0 && len(missingOut) == 0 {
		return nil
	}

	var fixes []analysis.SuggestedFix

	if mode == "smart" {
		smartFix := generateSmartFix(fixCtx, missingIn, missingOut)
		if smartFix != nil {
			fixes = append(fixes, *smartFix)
		}
	}

	safeFix := generateSafeFix(fixCtx, missingIn, missingOut)
	if safeFix != nil {
		fixes = append(fixes, *safeFix)
	}

	return fixes
}

// topLevelFields extracts unique top-level field names from qualified field names.
// E.g., ["inVar.Foo", "inVar.Bar.Baz"] becomes ["Foo", "Bar"].
func topLevelFields(qualifiedFields []string, varName string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, qf := range qualifiedFields {
		// Strip "varName." prefix
		field := qf
		if varName != "" {
			field = strings.TrimPrefix(qf, varName+".")
		}
		// Take only the top-level field name
		if idx := strings.Index(field, "."); idx >= 0 {
			field = field[:idx]
		}
		if !seen[field] {
			seen[field] = true
			result = append(result, field)
		}
	}

	return result
}

// generateSafeFix generates a safe fix that suppresses warnings without changing behavior.
// Uses `_ = var.Field` pattern for both input and output missing fields.
func generateSafeFix(
	fixCtx *FixContext,
	missingIn, missingOut []string,
) *analysis.SuggestedFix {
	var edits []analysis.TextEdit

	// For slice inline mapping, emit a TODO comment instead of stubs for input fields
	if fixCtx.IsSliceInline && len(missingIn) > 0 {
		var lines []string
		for _, field := range missingIn {
			lines = append(lines, fmt.Sprintf(
				"\t// TODO(lostfield): handle %s.%s in loop body",
				fixCtx.InFieldVar, field,
			))
		}
		text := "\n" + strings.Join(lines, "\n") + "\n"
		edits = append(edits, analysis.TextEdit{
			Pos:     fixCtx.FnBodyLbrace + 1,
			End:     fixCtx.FnBodyLbrace + 1,
			NewText: []byte(text),
		})
	} else if len(missingIn) > 0 {
		// Insert `_ = inVar.Field` after opening brace
		var lines []string
		for _, field := range missingIn {
			lines = append(lines, fmt.Sprintf("\t_ = %s.%s", fixCtx.InFieldVar, field))
		}
		text := "\n" + strings.Join(lines, "\n") + "\n"
		edits = append(edits, analysis.TextEdit{
			Pos:     fixCtx.FnBodyLbrace + 1,
			End:     fixCtx.FnBodyLbrace + 1,
			NewText: []byte(text),
		})
	}

	// Insert output field stubs before last return (or before closing brace if no return).
	if len(missingOut) > 0 {
		insertPos := fixCtx.outInsertPos()
		if fixCtx.OutVar != "" {
			var lines []string
			for _, field := range missingOut {
				lines = append(lines, fmt.Sprintf("\t_ = %s.%s", fixCtx.OutVar, field))
			}
			text := "\n" + strings.Join(lines, "\n") + "\n"
			edits = append(edits, analysis.TextEdit{
				Pos:     insertPos,
				End:     insertPos,
				NewText: []byte(text),
			})
		} else if fixCtx.CompLitRbrace.IsValid() {
			// No output variable but composite literal exists: insert commented keys.
			var lines []string
			for _, field := range missingOut {
				lines = append(lines, fmt.Sprintf(
					"\t\t// TODO(lostfield): %s is not set", field,
				))
			}
			text := "\n" + strings.Join(lines, "\n") + "\n"
			edits = append(edits, analysis.TextEdit{
				Pos:     fixCtx.CompLitRbrace,
				End:     fixCtx.CompLitRbrace,
				NewText: []byte(text),
			})
		}
	}

	if len(edits) == 0 {
		return nil
	}

	return &analysis.SuggestedFix{
		Message:   "Suppress warnings with _ = var.Field",
		TextEdits: edits,
	}
}

// generateSmartFix generates smart fixes that infer correct field mappings.
func generateSmartFix(
	fixCtx *FixContext,
	missingIn, missingOut []string,
) *analysis.SuggestedFix {
	// Build sets for quick lookup
	missingInSet := make(map[string]bool)
	for _, f := range missingIn {
		missingInSet[f] = true
	}
	missingOutSet := make(map[string]bool)
	for _, f := range missingOut {
		missingOutSet[f] = true
	}

	// Fields in both sides can get smart fixes
	var smartAssignments []string
	var smartCompLitEntries []string
	var safeInStubs []string
	var safeOutStubs []string
	var todoComments []string

	// Build field type maps for the structs
	inFieldTypes := buildFieldTypeMap(fixCtx.InStruct)
	outFieldTypes := buildFieldTypeMap(fixCtx.OutStruct)

	for _, field := range missingIn {
		if missingOutSet[field] {
			// Field is missing on both sides — attempt smart mapping
			inType := inFieldTypes[field]
			outType := outFieldTypes[field]

			assignment := inferAssignment(fixCtx, field, inType, outType)
			if assignment != "" {
				if fixCtx.OutputStyle == OutputStyleCompositeLit {
					smartCompLitEntries = append(smartCompLitEntries, assignment)
				} else {
					smartAssignments = append(smartAssignments, assignment)
				}
			}
			// Input side is handled by the assignment reading from it
			continue
		}
		// Only missing on input side — use safe stub
		if fixCtx.IsSliceInline {
			todoComments = append(todoComments, fmt.Sprintf(
				"\t// TODO(lostfield): handle %s.%s in loop body",
				fixCtx.InFieldVar, field,
			))
		} else {
			safeInStubs = append(safeInStubs, fmt.Sprintf("\t_ = %s.%s", fixCtx.InFieldVar, field))
		}
	}

	for _, field := range missingOut {
		if missingInSet[field] {
			// Already handled above in the smart mapping
			continue
		}
		// Only missing on output side — use safe stub
		if fixCtx.OutVar != "" {
			safeOutStubs = append(safeOutStubs, fmt.Sprintf("\t_ = %s.%s", fixCtx.OutVar, field))
		}
	}

	var edits []analysis.TextEdit

	// Insert safe input stubs + TODO comments after opening brace
	var afterLbrace []string
	afterLbrace = append(afterLbrace, todoComments...)
	afterLbrace = append(afterLbrace, safeInStubs...)
	if len(afterLbrace) > 0 {
		text := "\n" + strings.Join(afterLbrace, "\n") + "\n"
		edits = append(edits, analysis.TextEdit{
			Pos:     fixCtx.FnBodyLbrace + 1,
			End:     fixCtx.FnBodyLbrace + 1,
			NewText: []byte(text),
		})
	}

	// Insert smart assignments and safe output stubs before closing brace
	if fixCtx.OutputStyle == OutputStyleCompositeLit && len(smartCompLitEntries) > 0 {
		// Insert into composite literal
		text := "\n" + strings.Join(smartCompLitEntries, "\n") + "\n"
		edits = append(edits, analysis.TextEdit{
			Pos:     fixCtx.CompLitRbrace,
			End:     fixCtx.CompLitRbrace,
			NewText: []byte(text),
		})
	}

	var beforeReturn []string
	if fixCtx.OutputStyle == OutputStyleDotAssignment {
		beforeReturn = append(beforeReturn, smartAssignments...)
	}
	beforeReturn = append(beforeReturn, safeOutStubs...)
	if len(beforeReturn) > 0 {
		insertPos := fixCtx.outInsertPos()
		text := "\n" + strings.Join(beforeReturn, "\n") + "\n"
		edits = append(edits, analysis.TextEdit{
			Pos:     insertPos,
			End:     insertPos,
			NewText: []byte(text),
		})
	}

	if len(edits) == 0 {
		return nil
	}

	return &analysis.SuggestedFix{
		Message:   "Auto-fix field mappings (smart)",
		TextEdits: edits,
	}
}

// inferAssignment generates the right-hand side for a field assignment.
// Returns the full assignment line or empty string if it cannot be inferred.
func inferAssignment(fixCtx *FixContext, field string, inType, outType types.Type) string {
	inVar := fixCtx.InFieldVar
	outVar := fixCtx.OutVar

	if inType == nil || outType == nil {
		// Can't determine types — fall back to safe with TODO
		if fixCtx.OutputStyle == OutputStyleCompositeLit {
			return fmt.Sprintf("\t\t%s: %s.%s, // TODO(lostfield): verify type", field, inVar, field)
		}
		return fmt.Sprintf("\t%s.%s = %s.%s // TODO(lostfield): verify type", outVar, field, inVar, field)
	}

	// 1. Types identical — direct assignment
	if types.Identical(inType, outType) {
		if fixCtx.OutputStyle == OutputStyleCompositeLit {
			return fmt.Sprintf("\t\t%s: %s.%s,", field, inVar, field)
		}
		return fmt.Sprintf("\t%s.%s = %s.%s", outVar, field, inVar, field)
	}

	// 2. Getter exists
	if fixCtx.InNamedType != nil {
		getterName := "Get" + field
		for i := 0; i < fixCtx.InNamedType.NumMethods(); i++ {
			method := fixCtx.InNamedType.Method(i)
			if method.Name() == getterName {
				sig, ok := method.Type().(*types.Signature)
				if ok && sig.Params().Len() == 0 && sig.Results().Len() >= 1 {
					if fixCtx.OutputStyle == OutputStyleCompositeLit {
						return fmt.Sprintf("\t\t%s: %s.%s(),", field, inVar, getterName)
					}
					return fmt.Sprintf("\t%s.%s = %s.%s()", outVar, field, inVar, getterName)
				}
			}
		}
	}

	// 3. Types convertible
	if types.ConvertibleTo(inType, outType) {
		outTypeName := types.TypeString(outType, nil)
		if fixCtx.OutputStyle == OutputStyleCompositeLit {
			return fmt.Sprintf("\t\t%s: %s(%s.%s),", field, outTypeName, inVar, field)
		}
		return fmt.Sprintf("\t%s.%s = %s(%s.%s)", outVar, field, outTypeName, inVar, field)
	}

	// 4. Incompatible — safe fallback with TODO
	if fixCtx.OutputStyle == OutputStyleCompositeLit {
		return fmt.Sprintf("\t\t// TODO(lostfield): convert %s.%s to %s", inVar, field, field)
	}
	return fmt.Sprintf("\t// TODO(lostfield): convert %s.%s\n\t_ = %s.%s", inVar, field, inVar, field)
}

// buildFieldTypeMap builds a map from field name to field type for a struct.
func buildFieldTypeMap(st *types.Struct) map[string]types.Type {
	m := make(map[string]types.Type)
	if st == nil {
		return m
	}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Exported() {
			m[f.Name()] = f.Type()
		}
	}
	return m
}
