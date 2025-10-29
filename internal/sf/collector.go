package sf

import (
	"go/ast"
	"go/token"
	"strings"
)

// CollectingType means which type of node we can "collect" via usageCollector.
type CollectingType int

const (
	RecordUnknown CollectingType = iota
	RecordMethods
	RecordFields
)

// UsageLookup is a map storing name of fields/methods that were used.
type UsageLookup map[string]struct{}

func (ul UsageLookup) LookUp(v string) bool {
	_, ok := ul[v]
	return ok
}

// UsageCollector is a generic AST visitor that collects selector usage for a given variable.
// rType(RecordingType) stands for the type of things we record: fields or methods.
type UsageCollector struct {
	used        UsageLookup
	varName     string
	parentStack []ast.Node
	nodesType   CollectingType
}

func NewUsageCollector(varName string, rType CollectingType) *UsageCollector {
	return &UsageCollector{
		used:        make(UsageLookup),
		varName:     varName,
		parentStack: make([]ast.Node, 0),
		nodesType:   rType,
	}
}

// Visit collects used items (of nodeType) in a given container node.
func (v *UsageCollector) Visit(container ast.Node) ast.Visitor {
	// When node is nil, we're returning from a branch: pop the last parent.
	if container == nil {
		if len(v.parentStack) > 0 {
			v.parentStack = v.parentStack[:len(v.parentStack)-1]
		}
		return nil
	}

	// Push the current node onto the parent stack.
	v.parentStack = append(v.parentStack, container)

	// Check for a selector expression.
	sel, ok := container.(*ast.SelectorExpr)
	if !ok {
		return v
	}

	// Check that the expression's X is an identifier matching varName.
	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != v.varName {
		return v
	}

	// Determine whether this selector is used as part of a call expression.
	var isMethodCall bool
	if len(v.parentStack) >= 2 {
		// The parent is the second-to-last element.
		if call, ok := v.parentStack[len(v.parentStack)-2].(*ast.CallExpr); ok && call.Fun == sel {
			isMethodCall = true
		}
	}

	// Decide based on the mode.
	if v.nodesType == RecordMethods && isMethodCall || !isMethodCall {
		v.used[sel.Sel.Name] = struct{}{}
	}

	// Also check if this selector is used in a blank identifier assignment (e.g., _ = model.Field)
	// This is a valid way to acknowledge that a field exists/is used
	if len(v.parentStack) >= 2 {
		if assign, ok := v.parentStack[len(v.parentStack)-2].(*ast.AssignStmt); ok {
			// Check if LHS is a blank identifier
			if len(assign.Lhs) > 0 {
				if ident, ok := assign.Lhs[0].(*ast.Ident); ok && ident.Name == "_" {
					// This is an assignment to blank identifier, mark field as used
					v.used[sel.Sel.Name] = struct{}{}
				}
			}
		}
	}

	return v
}

func (v *UsageCollector) reset() {
	v.parentStack = make([]ast.Node, 0)
	v.used = make(UsageLookup)
}

func (v *UsageCollector) Walk(container ast.Node) UsageLookup {
	v.reset()
	ast.Walk(v, container)
	return v.used
}

// CollectUsedFields walks the AST rooted at n and returns a set (UsageLookup)
// of field names that are directly accessed on varName (ignoring any method calls).
func CollectUsedFields(n ast.Node, varName string) UsageLookup {
	return NewUsageCollector(varName, RecordFields).Walk(n)
}

// CollectUsedMethods walks the AST rooted at n and returns a set (UsageLookup)
// of method names that are called on varName.
func CollectUsedMethods(n ast.Node, varName string) UsageLookup {
	return NewUsageCollector(varName, RecordMethods).Walk(n)
}

// CollectCompositeLitKeys scans for composite literals of type matching the given candidate,
// and returns a set of keys (field names) that appear in the literal.
// (We assume keys are simple identifiers; more complex cases can be added as needed.)
func CollectCompositeLitKeys(n ast.Node, candidateName string) UsageLookup {
	ul := make(UsageLookup)
	ast.Inspect(n, func(node ast.Node) bool {
		cl, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if ident, ok := cl.Type.(*ast.Ident); ok && ident.Name == candidateName {
			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if keyIdent, ok := kv.Key.(*ast.Ident); ok {
					ul[keyIdent.Name] = struct{}{}
				}
			}
		}
		return true
	})

	return ul
}

// CollectOutputFields inspects fn.Body looking for all field names that are "set"
// on the output value of the converter. It does two things:
//
//	(a) If outVar is non-empty, it collects direct field accesses on that variable.
//	(b) It also scans assignment and return statements to find composite literals
//	    (or their address-of forms) whose type (or underlying type) matches candidateName,
//	    then collects the keys (i.e. field names) provided in the literal.
//
// CollectOutputFields inspects fn.Body and returns a set of field names that are used in
// constructing the output value of the converter. It does two things:
//
//	(a) If outVar is non-empty or can be determined from a local declaration, it collects direct
//	    field accesses on that variable (e.g. out.ID = ...).
//	(b) It scans assignment and return statements for composite literals that initialize a value
//	    of type candidateName (e.g. out = &Category{ Type: ... }).
func CollectOutputFields(fn *ast.FuncDecl, outVar, candidateName string) UsageLookup {
	ul := make(UsageLookup)

	// If no output variable was provided (e.g. unnamed result), try to find a local candidate.
	if outVar == "" {
		outVar = findLocalCandidateVariable(fn, candidateName)
	}

	// (a) If we have an output variable, collect direct field accesses.
	if outVar != "" {
		for k := range CollectUsedFields(fn.Body, outVar) {
			ul[k] = struct{}{}
		}
	}

	// (b) Scan the function body for composite literals in assignments and return statements.
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, expr := range stmt.Rhs {
				extractKeysFromExpr(expr, candidateName, ul)
			}
		case *ast.ReturnStmt:
			for _, expr := range stmt.Results {
				extractKeysFromExpr(expr, candidateName, ul)
			}
		}
		return true
	})

	return ul
}

// extractKeysFromExpr examines expr and, if it is or contains a composite literal
// that initializes a value of type candidateName, it extracts any key names and adds them to keys.
func extractKeysFromExpr(expr ast.Expr, candidateName string, keys UsageLookup) {
	var cl *ast.CompositeLit

	switch x := expr.(type) {
	case *ast.CompositeLit:
		cl = x
	case *ast.UnaryExpr:
		// Handle cases like: out = &Category{...}
		if x.Op == token.AND {
			if lit, ok := x.X.(*ast.CompositeLit); ok {
				cl = lit
			}
		}
	case *ast.CallExpr:
		// Handle cases like: return (&Category{...}).MethodCall()
		// The composite literal is in the receiver of the call
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			// The receiver might be wrapped in parentheses: (&Type{...}).Method()
			recv := sel.X

			// Unwrap ParenExpr if present
			if paren, ok := recv.(*ast.ParenExpr); ok {
				recv = paren.X
			}

			// Now we should have a UnaryExpr with &
			if unaryRecv, ok := recv.(*ast.UnaryExpr); ok && unaryRecv.Op == token.AND {
				if lit, ok := unaryRecv.X.(*ast.CompositeLit); ok {
					cl = lit
				}
			}
		}
	}

	if cl == nil {
		return
	}

	// Determine the type name of the composite literal.
	var typeName string
	switch t := cl.Type.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.SelectorExpr:
		// For types like models.Category, use the selector's identifier.
		typeName = t.Sel.Name
	}

	// Compare candidate names (optionally case-insensitively).
	if !strings.EqualFold(typeName, candidateName) {
		return
	}

	// Extract keys from key-value pairs.
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if keyIdent, ok := kv.Key.(*ast.Ident); ok {
			keys[keyIdent.Name] = struct{}{}
		}
	}
}

// findLocalCandidateVariable scans the function body for a short variable declaration
// that assigns a composite literal (or its address) of type candidateName. If found, it returns
// the variable name (e.g. "out"). Otherwise, it returns the empty string.
func findLocalCandidateVariable(fn *ast.FuncDecl, candidateName string) string {
	var varName string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		decl, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		// Look for a short variable declaration (:=)
		if decl.Tok != token.DEFINE {
			return true
		}
		// Iterate over each LHS/RHS pair.
		for i, lhs := range decl.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if !ok {
				continue
			}
			// Ensure there is a corresponding RHS expression.
			if i >= len(decl.Rhs) {
				continue
			}
			expr := decl.Rhs[i]
			var cl *ast.CompositeLit
			switch x := expr.(type) {
			case *ast.CompositeLit:
				cl = x
			case *ast.UnaryExpr:
				// Handle cases like: out := &Category{ ... }
				if x.Op == token.AND {
					if lit, ok := x.X.(*ast.CompositeLit); ok {
						cl = lit
					}
				}
			}
			if cl == nil {
				continue
			}
			// Determine the type name of the composite literal.
			var typeName string
			switch t := cl.Type.(type) {
			case *ast.Ident:
				typeName = t.Name
			case *ast.SelectorExpr:
				// For types like models.Category, use the selector's identifier.
				typeName = t.Sel.Name
			}
			// Compare candidate names case-insensitively.
			if strings.EqualFold(typeName, candidateName) {
				varName = ident.Name
				return false // stop searching
			}
		}
		return true
	})
	return varName
}
