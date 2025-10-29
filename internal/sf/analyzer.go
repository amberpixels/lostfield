package sf

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/amberpixels/go-stickyfields/internal/config"
	"golang.org/x/tools/go/analysis"
)

// Run function used in analysis.Analyzer.
func Run(pass *analysis.Pass) (any, error) {
	warningsTotal := 0
	filesTotal := 0
	filesWarned := 0

	for _, file := range pass.Files {
		// Get the filename from the file position.
		filename := pass.Fset.Position(file.Pos()).Filename

		// Skip files that are test files or are in the vendor directory.
		if strings.HasSuffix(filename, "_test.go") || strings.Contains(filepath.ToSlash(filename), "/vendor/") {
			continue
		}

		filesTotal++

		// Walk the AST and look for function declarations.
		var fileContainsWarnings bool
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				if !IsPossibleConverter(fn, pass) {
					return true
				}

				validationResult, err := ValidateConverter(fn, pass)
				if err != nil {
					// fmt.Println("--> Validation error, ignoring ", fn.Name.Name)
					return true
				}

				if validationResult.Valid {
					return true
				}

				message := formatConverterValidationMessage(validationResult)

				var buf bytes.Buffer
				PrettyPrint(&buf, filename, fn, pass, message)

				// Now report the diagnostic using pass.Report.
				pass.Report(analysis.Diagnostic{
					Pos:     fn.Name.Pos(),
					Message: buf.String(),
				})

				warningsTotal++
				fileContainsWarnings = true
			}
			return true
		})
		if fileContainsWarnings {
			filesWarned++
		}
	}

	// At the end of processing all files, print the total number of warnings.
	// Only print if verbose mode is enabled.
	cfg := config.Get()
	if cfg.Verbose {
		if warningsTotal > 0 {
			fmt.Fprintf(
				os.Stdout,
				"\nFiles total analyzed: %d. Warnings: %d caught in %d files\n",
				filesTotal,
				warningsTotal,
				filesWarned,
			)
		} else {
			fmt.Fprintf(os.Stdout, "\nFiles total analyzed: %d. Warnings: 0\n", filesTotal)
		}
	}

	return nil, nil //nolint: nilnil // fix later
}

// formatConverterValidationMessage creates a human-readable two-column format for missing fields.
// Example output:
//
//	converter function is leaking fields:
//	 data.CalculatedAt → ??
//	 data.CreatedAt    → ??
//	 ??                → output.NewField
//
// When there are many fields missing, it simplifies to:
//
//	converter function is leaking fields:
//	 ??  → all output fields
func formatConverterValidationMessage(result *ConverterValidationResult) string {
	var buf strings.Builder
	buf.WriteString("converter function is leaking fields:\n")

	if len(result.MissingInputFields) == 0 && len(result.MissingOutputFields) == 0 {
		return buf.String()
	}

	// Use consistent indentation for the table (aligned with 'func' keyword position)
	// The format is: line_number | code, so we need to indent to align with the code
	const indent = "      "

	// Simplification: if no input fields are missing but many output fields are, just say "all output fields"
	// This makes the output more readable when there are dozens of missing output fields
	if len(result.MissingInputFields) == 0 && len(result.MissingOutputFields) > 5 {
		buf.WriteString(indent + "??  → all output fields\n")
		return buf.String()
	}

	// Similarly, if no output fields are missing but many input fields are, say "all input fields"
	if len(result.MissingOutputFields) == 0 && len(result.MissingInputFields) > 5 {
		buf.WriteString(indent + "all input fields → ??\n")
		return buf.String()
	}

	// Calculate the maximum length for alignment of the arrow
	// We need to find the longest field name to align all arrows
	maxLen := 0
	for _, field := range result.MissingInputFields {
		if len(field) > maxLen {
			maxLen = len(field)
		}
	}
	for _, field := range result.MissingOutputFields {
		if len(field) > maxLen {
			maxLen = len(field)
		}
	}
	// Account for the leading space and ensure minimum width
	if maxLen < 1 {
		maxLen = 1
	}

	// Add input fields (missing in output mapping)
	for _, field := range result.MissingInputFields {
		padding := strings.Repeat(" ", maxLen-len(field)+1)
		buf.WriteString(indent + field + padding + "→ ??\n")
	}

	// Add output fields (missing in input mapping)
	for _, field := range result.MissingOutputFields {
		padding := strings.Repeat(" ", maxLen-len("??")+1)
		buf.WriteString(indent + "??" + padding + "→ " + field + "\n")
	}

	return buf.String()
}

// ContainerType represents the "container" kind for a candidate type.
type ContainerType string

const (
	ContainerNone    ContainerType = "none"    // plain struct
	ContainerPointer ContainerType = "pointer" // pointer to struct
	ContainerSlice   ContainerType = "slice"   // slice or array
	ContainerMap     ContainerType = "map"     // map (using its value type)
)

// candidate holds the underlying candidate type's name and its container type.
type candidate struct {
	name          string
	containerType ContainerType
	structType    *types.Struct
	fullType      types.Type // Full type info for accurate comparisons
}

// extractCandidateType checks if the given type qualifies as a candidate for conversion.
// It recognizes a plain struct, a pointer to a struct, a slice/array of such types,
// or a map whose value is such a type. If so, it returns the candidate (with its
// underlying type name and container type) and ok==true. Otherwise, ok==false.
func extractCandidateType(t types.Type) (candidate, bool) {
	var cand candidate
	// First, check for containers.
	switch tt := t.(type) {
	case *types.Slice, *types.Array:
		cand.containerType = ContainerSlice
		var elem types.Type
		switch x := tt.(type) {
		case *types.Slice:
			elem = x.Elem()
		case *types.Array:
			elem = x.Elem()
		}
		t = elem
	case *types.Map:
		cand.containerType = ContainerMap
		t = tt.Elem()
	default:
		cand.containerType = ContainerNone
	}

	// If the type is a pointer and not already a container, mark it as pointer.
	if ptr, okPtr := t.(*types.Pointer); okPtr {
		if cand.containerType == ContainerNone {
			cand.containerType = ContainerPointer
		}
		t = ptr.Elem()
	}

	// We expect a named type whose underlying type is a struct.
	named, okNamed := t.(*types.Named)
	if !okNamed {
		return candidate{}, false
	}
	st, okStruct := named.Underlying().(*types.Struct)
	if !okStruct {
		return candidate{}, false
	}
	cand.name = named.Obj().Name()
	cand.structType = st
	cand.fullType = named // Store the full type for accurate comparison
	return cand, true
}

// IsPossibleConverter checks whether fn (a function declaration)
// qualifies as a potential converter function based on these rules:
//   - At least one input and one output candidate exist.
//   - Candidate is the argument who fits the candidate type (struct or pointer to struct).
//   - For at least one candidate pair (input, output) with the same container type,
//     the names of the candidate types share a common substring (ignoring case).
//
// TODO: it can't be the same type e.g. HandleRewrites(sectionRewrites) (string, SectionRewrite, erro)
func IsPossibleConverter(fn *ast.FuncDecl, pass *analysis.Pass) bool {
	cfg := config.Get()

	// If we're not including methods and this function has a receiver, skip it.
	if !cfg.IncludeMethods && fn.Recv != nil {
		return false
	}

	obj := pass.TypesInfo.Defs[fn.Name]
	if obj == nil {
		return false
	}

	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return false
	}

	// No arguments: nothing was converted
	if sig.Params().Len() == 0 || sig.Results().Len() == 0 {
		return false
	}

	// Gather candidate types from input parameters.
	var inCandidates []candidate
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		if cand, ok := extractCandidateType(param.Type()); ok {
			inCandidates = append(inCandidates, cand)
		}
	}
	if len(inCandidates) == 0 {
		return false
	}

	// Gather candidate types from output parameters.
	var outCandidates []candidate
	for i := 0; i < sig.Results().Len(); i++ {
		res := sig.Results().At(i)

		if cand, ok := extractCandidateType(res.Type()); ok {
			outCandidates = append(outCandidates, cand)
		}
	}
	if len(outCandidates) == 0 {
		return false
	}

	// Look for at least one candidate pair (in, out) where:
	// - The container types are compatible:
	//    - if the input candidate is a slice or map, then the output candidate must be of the same container type.
	//    - otherwise, if the input candidate is a plain struct or pointer to struct, the output candidate
	//      must also be a plain struct or pointer (i.e. not a slice or map).
	// - The candidate names are different (no same-type conversions like DB -> DB)
	// - And the candidate names share a common substring (ignoring case).
	for _, inCand := range inCandidates {
		lowerIn := strings.ToLower(inCand.name)
		for _, outCand := range outCandidates {
			// Exclude same-type conversions (e.g., *DB -> *DB is not a converter)
			// Compare full types, not just names, to avoid false positives like models.MatchedMapData -> pbVenueConfig.MatchedMapData
			if inCand.fullType != nil && outCand.fullType != nil && types.Identical(inCand.fullType, outCand.fullType) {
				continue
			}

			// Check container type compatibility.
			if inCand.containerType == ContainerSlice || inCand.containerType == ContainerMap {
				if inCand.containerType != outCand.containerType {
					continue // e.g. slice -> non-slice is not allowed.
				}
			} else {
				// inCand is ContainerNone or ContainerPointer.
				// Allow output to be either a plain struct or a pointer.
				if outCand.containerType != ContainerNone && outCand.containerType != ContainerPointer {
					continue
				}
			}

			lowerOut := strings.ToLower(outCand.name)
			if strings.Contains(lowerOut, lowerIn) || strings.Contains(lowerIn, lowerOut) {
				return true
			}
		}
	}

	return false
}

// collectMissingFields is similar to checkAllFieldsUsed but returns a slice of missing field names.
func collectMissingFields(st *types.Struct, usedFields UsageLookup, usedMethodsArg ...UsageLookup) []string {
	cfg := config.Get()
	var missing []string
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		// adjust this as needed.
		if !field.Exported() {
			continue
		}

		if !usedFields.LookUp(field.Name()) {
			// if methods were given, let's allow via getters (if config allows)
			// If a getter method exists (for input candidate) then allow it.
			if cfg.AllowGetters && len(usedMethodsArg) > 0 && usedMethodsArg[0].LookUp("Get"+field.Name()) {
				continue
			}
			missing = append(missing, field.Name())
		}
	}
	return missing
}

// ConverterValidationResult holds the details of a converter function validation.
type ConverterValidationResult struct {
	// Valid is true if every exported field (or getter methods) in
	// both input/output models were used.
	Valid bool
	// MissingInputFields contains the names of exported fields in the input candidate
	// that were not used.
	MissingInputFields []string
	// MissingOutputFields contains the names of exported fields in the output candidate
	// that were not used.
	MissingOutputFields []string
}

func NewOKConverterValidationResult() *ConverterValidationResult {
	return &ConverterValidationResult{Valid: true}
}
func NewFailedConverterValidationResult(in, out []string) *ConverterValidationResult {
	return &ConverterValidationResult{MissingInputFields: in, MissingOutputFields: out}
}

// ValidateConverter checks that the converter function fn uses every field
// of the candidate input model (by reading) and every field of the candidate output model (by writing).
//
// For input, we assume the candidate comes from the first parameter and that it has a name.
// For output, we first try to use a named result; if none, we look for a composite literal.
func ValidateConverter(fn *ast.FuncDecl, pass *analysis.Pass) (*ConverterValidationResult, error) {
	// Retrieve the function object and signature.
	obj := pass.TypesInfo.Defs[fn.Name]
	if obj == nil {
		return nil, fmt.Errorf("cannot get type info for function %q", fn.Name.Name)
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return nil, fmt.Errorf("function %q does not have a valid signature", fn.Name.Name)
	}
	if sig.Params().Len() < 1 || sig.Results().Len() < 1 {
		return nil, fmt.Errorf(
			"function %q must have at least one parameter and one result",
			fn.Name.Name,
		)
	}

	// Find the candidate input parameter.
	inCand, inVar, okIn := findCandidateParam(fn.Type.Params, sig.Params())
	if !okIn || inVar == "" {
		return nil, fmt.Errorf(
			"cannot determine candidate input parameter for function %q",
			fn.Name.Name,
		)
	}

	// Determine the candidate output parameter.
	outCand, outVar, okOut := findCandidateParam(fn.Type.Results, sig.Results())
	if !okOut {
		return nil, fmt.Errorf(
			"cannot determine candidate output parameter for function %q",
			fn.Name.Name,
		)
	}

	// Check if this is a delegating converter (e.g., converts a slice by calling another converter on each element)
	if isDelegatingConverter(fn, inCand, outCand, inVar) {
		// For delegating converters, skip validation since the actual field mapping
		// is delegated to the inner converter function which will be linted separately
		return NewOKConverterValidationResult(), nil
	}

	// Collect field usages for the input candidate variable.
	fieldsUsedModelIn := CollectUsedFields(fn.Body, inVar)
	methodsUsedModelIn := CollectUsedMethods(fn.Body, inVar)
	missingIn := collectMissingFields(inCand.structType, fieldsUsedModelIn, methodsUsedModelIn)
	for i, m := range missingIn {
		missingIn[i] = inVar + "." + m
	}

	// Collect field usages for the output candidate.
	fieldsUsedModelOut := CollectOutputFields(fn, outVar, outCand.name)
	missingOut := collectMissingFields(outCand.structType, fieldsUsedModelOut)
	if outVar != "" {
		for i, m := range missingOut {
			missingOut[i] = outVar + "." + m
		}
	}

	if len(missingIn) == 0 && len(missingOut) == 0 {
		return NewOKConverterValidationResult(), nil
	}

	return NewFailedConverterValidationResult(missingIn, missingOut), nil
}

// findCandidateParam searches the appropriate FieldList (for input or output)
// for the first parameter/result that qualifies as a candidate type.
// It returns the candidate info, the variable name (if any) and true on success.
func findCandidateParam(fieldList *ast.FieldList, sigParams *types.Tuple) (candidate, string, bool) {
	if fieldList == nil {
		return candidate{}, "", false
	}
	// Keep a running count to match the order of parameters/results in sigParams.
	paramIndex := 0
	for _, field := range fieldList.List {
		// A field may declare several names (e.g. "a, b int").
		names := field.Names
		// If no names are present (for results), we still count the parameter.
		n := 1
		if len(names) > 0 {
			n = len(names)
		}
		// For each declared parameter in this field:
		for i := 0; i < n; i++ {
			// Get the type from the signature.
			if paramIndex >= sigParams.Len() {
				break
			}
			paramVar := sigParams.At(paramIndex)
			if c, ok := extractCandidateType(paramVar.Type()); ok {
				// If the AST field has names, use the first one (or the one corresponding to our index).
				if len(names) > 0 {
					return c, names[i].Name, true
				}
				// Otherwise, return with an empty variable name.
				return c, "", true
			}
			paramIndex++
		}
		paramIndex += (n - 1)
	}
	return candidate{}, "", false
}

// isDelegatingConverter checks if a function is a delegating converter:
// - Input parameter is a slice of structs
// - Output parameter is a slice of structs
// - Function loops through input slice and calls another function on each element
// - Results are appended to output (filtering is allowed)
//
// Returns true if this pattern is detected (and validation should be skipped).
func isDelegatingConverter(
	fn *ast.FuncDecl,
	inCand candidate,
	outCand candidate,
	inVar string,
) bool {
	// Check if both input and output are slices
	if inCand.containerType != ContainerSlice || outCand.containerType != ContainerSlice {
		return false
	}

	// Look for a range loop over the input variable
	var foundLoop bool
	var loopVar string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		rangeStmt, ok := n.(*ast.RangeStmt)
		if !ok {
			return true
		}

		// Check if looping over the input variable
		ident, ok := rangeStmt.X.(*ast.Ident)
		if !ok || ident.Name != inVar {
			return true
		}

		// Extract loop variable (e.g., "t" in "for _, t := range tickets")
		if rangeStmt.Value != nil {
			if idVal, ok := rangeStmt.Value.(*ast.Ident); ok {
				loopVar = idVal.Name
			}
		}
		foundLoop = true
		return false
	})

	if !foundLoop || loopVar == "" {
		return false
	}

	// Look for function calls with the loop variable as argument
	// and either append operations OR indexed assignments that look like delegating
	var foundDelegation bool
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.CallExpr:
			// Check for append calls
			if ident, ok := stmt.Fun.(*ast.Ident); ok && ident.Name == "append" {
				// append should have at least 2 args: slice and value
				if len(stmt.Args) >= 2 {
					foundDelegation = true
					return false
				}
			}

		case *ast.AssignStmt:
			// Check for indexed assignments like: protos[i] = ConvertFunc(...)
			if stmt.Tok == token.ASSIGN && len(stmt.Lhs) > 0 && len(stmt.Rhs) > 0 {
				// Check if LHS is an index expression (e.g., protos[i])
				if _, ok := stmt.Lhs[0].(*ast.IndexExpr); ok {
					// If RHS is a function call with the loop variable, it's delegation
					if _, ok := stmt.Rhs[0].(*ast.CallExpr); ok {
						foundDelegation = true
						return false
					}
				}
			}
		}

		return true
	})

	return foundDelegation
}
