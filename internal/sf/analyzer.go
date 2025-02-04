package sf

import (
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/tools/go/analysis"
)

// Configuration variable for including methods (functions with receivers) in the check.
// Set to false to consider only plain functions.
var includeMethods = false

// Run function used in analysis.Analyzer
func Run(pass *analysis.Pass) (any, error) {
	color.NoColor = false
	for _, file := range pass.Files {
		// Get the filename from the file position.
		filename := pass.Fset.Position(file.Pos()).Filename

		// Skip files that are test files or are in the vendor directory.
		if strings.HasSuffix(filename, "_test.go") || strings.Contains(filepath.ToSlash(filename), "/vendor/") {
			continue
		}

		// Walk the AST and look for function declarations.
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				if !IsPossibleConverter(fn, pass) {
					return true
				}

				var message string
				if !ValidateConverter(fn, pass) {
					message = " function is considered to be a converter and it does leak fields"
				} else {
					return true
					// message = fmt.Sprintf("Function %q is a proper converter", fn.Name.Name)
				}

				// Write the output to stdout.
				PrettyPrint(os.Stdout, filename, fn, pass, message)

			}
			return true
		})
	}
	return nil, nil
}

// ContainerType represents the “container” kind for a candidate type.
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
}

// extractCandidateType checks if the given type qualifies as a candidate for conversion.
// It recognizes a plain struct, a pointer to a struct, a slice/array of such types,
// or a map whose value is such a type. If so, it returns the candidate (with its
// underlying type name and container type) and ok==true. Otherwise, ok==false.
func extractCandidateType(t types.Type) (cand candidate, ok bool) {
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
	return cand, true
}

// IsPossibleConverter checks whether fn (a function declaration)
// qualifies as a potential converter function based on these rules:
//   - At least one input and one output candidate exist.
//   - Candidate is the argument who fits the candidate type (struct or pointer to struct).
//   - For at least one candidate pair (input, output) with the same container type,
//     the names of the candidate types share a common substring (ignoring case).
//
// TODO: it can't be slice -> item or item -> slice
// TODO: it can't be the same type e.g. HandleRewrites(sectionRewrites) (string, SectionRewrite, erro)
func IsPossibleConverter(fn *ast.FuncDecl, pass *analysis.Pass) bool {
	// If we're not including methods and this function has a receiver, skip it.
	if !includeMethods && fn.Recv != nil {
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

	// Look for at least one candidate pair (in, out)
	// with the same container type that share a common substring in their names.
	for _, inCand := range inCandidates {
		lowerIn := strings.ToLower(inCand.name)
		for _, outCand := range outCandidates {
			// for now do not consider slices, map
			if outCand.containerType == ContainerSlice || outCand.containerType == ContainerMap {
				continue
			}

			// if inCand.containerType != outCand.containerType {
			// continue
			// }

			lowerOut := strings.ToLower(outCand.name)
			if strings.Contains(lowerOut, lowerIn) || strings.Contains(lowerIn, lowerOut) {
				return true
			}
		}
	}

	return false
}

// checkAllFieldsUsed returns true if every field in st (a *types.Struct)
// appears in the provided usedFields map.
func checkAllFieldsUsed(st *types.Struct, usedFields UsageLookup, usedMethodsArg ...UsageLookup) bool {
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		// Here we check only exported fields;
		// adjust this as needed.
		if !field.Exported() {
			continue
		}

		if !usedFields.LookUp(field.Name()) {
			// if methods were given, let's allow via getters
			if len(usedMethodsArg) > 0 {
				if usedMethodsArg[0].LookUp("Get" + field.Name()) {
					continue
				}
			}

			return false
		}
	}
	return true
}

// ValidateConverter checks that the converter function fn uses every field
// of the candidate input model (by reading) and every field of the candidate output model (by writing).
//
// For input, we assume the candidate comes from the first parameter and that it has a name.
// For output, we first try to use a named result; if none, we look for a composite literal.
func ValidateConverter(fn *ast.FuncDecl, pass *analysis.Pass) bool {
	// Get the function object and signature.
	obj := pass.TypesInfo.Defs[fn.Name]
	if obj == nil {
		return false
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return false
	}
	if sig.Params().Len() < 1 || sig.Results().Len() < 1 {
		return false
	}

	// Find the candidate input parameter.
	inCand, inVar, okIn := findCandidateParam(fn.Type.Params, sig.Params())
	if !okIn || inVar == "" {
		// If we cannot determine a candidate or its variable name, we cannot check.
		return false
	}

	// Find the candidate output parameter.
	outCand, outVar, okOut := findCandidateParam(fn.Type.Results, sig.Results())
	if !okOut {
		return false
	}

	// Collect field usages for the input candidate variable.
	fieldsUsedModelIn := CollectUsedFields(fn.Body, inVar)
	methodsUsedModelIn := CollectUsedMethods(fn.Body, inVar)

	checkedModelIn := checkAllFieldsUsed(inCand.structType, fieldsUsedModelIn, methodsUsedModelIn)
	if !checkedModelIn {
		return false
	}

	fieldsUsedModelOut := CollectOutputFields(fn, outVar, outCand.name)
	checkedModelOut := checkAllFieldsUsed(outCand.structType, fieldsUsedModelOut)

	return checkedModelIn && checkedModelOut
}

// findCandidateParam searches the appropriate FieldList (for input or output)
// for the first parameter/result that qualifies as a candidate type.
// It returns the candidate info, the variable name (if any) and true on success.
func findCandidateParam(fieldList *ast.FieldList, sigParams *types.Tuple) (cand candidate, varName string, found bool) {
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

// missingFields returns a slice of exported field names in st that are missing from used.
// func missingFields(st *types.Struct, used map[string]bool) []string {
// 	var missing []string
// 	for i := 0; i < st.NumFields(); i++ {
// 		field := st.Field(i)
// 		// Only consider exported fields.
// 		if !field.Exported() {
// 			continue
// 		}
// 		if !used[field.Name()] {
// 			missing = append(missing, field.Name())
// 		}
// 	}
// 	return missing
// }

// // checkConverterFieldUsageUpdated examines the converter function fn and returns two values:
// // - missingOut: a slice of field names that are missing on the output candidate model.
// // - ok: a boolean that is true if no fields are missing.
// func checkConverterFieldUsageUpdated(fn *ast.FuncDecl, pass *analysis.Pass) (missingOut []string, ok bool) {
// 	// Retrieve the function signature.
// 	obj := pass.TypesInfo.Defs[fn.Name]
// 	if obj == nil {
// 		return []string{"no type info"}, false
// 	}
// 	sig, okSig := obj.Type().(*types.Signature)
// 	if !okSig || sig.Params().Len() < 1 || sig.Results().Len() < 1 {
// 		return []string{"invalid signature"}, false
// 	}

// 	// For output, use our candidate-finding helper.
// 	outCand, outVar, okOut := findCandidateParam(fn, fn.Type.Results, sig.Results())
// 	if !okOut {
// 		return []string{"no candidate output parameter"}, false
// 	}

// 	// Collect output field usage.
// 	outputUsed := collectOutputFields(fn, outVar, outCand.name)
// 	// Determine which exported fields of outCand are missing.
// 	missingOut = missingFields(outCand.structType, outputUsed)
// 	ok = len(missingOut) == 0
// 	return missingOut, ok
// }
