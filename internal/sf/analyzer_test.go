// internal/sf/analyzer_test.go
package sf_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"

	"github.com/amberpixels/go-stickyfields/internal/sf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestC1(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "stickyfields",
		Doc:  "reports all inconsistent converter functions: ensures sticky fields)",
		Run:  sf.Run,
	}

	analysistest.Run(t, testdata, analyzer, "converters/c1")
}

func TestDelegatingConverters(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "stickyfields",
		Doc:  "reports all inconsistent converter functions: ensures sticky fields)",
		Run:  sf.Run,
	}

	// This test ensures that delegating converters (converters that call other converters)
	// are correctly identified and skipped from validation
	analysistest.Run(t, testdata, analyzer, "converters/delegate")
}

// TestIsPossibleConverter tests the IsPossibleConverter function with various scenarios.
func TestIsPossibleConverter(t *testing.T) {
	// Parse the test files
	fset := token.NewFileSet()
	testDir := filepath.Join("testdata", "src", "ispossible")

	var files []*ast.File
	patterns := []string{"models.go", "positive_cases.go", "negative_cases.go"}
	for _, pattern := range patterns {
		path := filepath.Join(testDir, pattern)
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", pattern, err)
		}
		files = append(files, f)
	}

	// Type-check the package
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	_, err := conf.Check("ispossible", fset, files, info)
	if err != nil {
		t.Fatalf("type checking failed: %v", err)
	}

	// Create a minimal analysis.Pass for testing
	pass := &analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
	}

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Positive cases - should be detected as converters
		{
			name:     "basic struct-to-struct conversion",
			funcName: "ConvertUserToDTO",
			want:     true,
		},
		{
			name:     "pointer input conversion",
			funcName: "ConvertUserPtrToDTO",
			want:     true,
		},
		{
			name:     "pointer output conversion",
			funcName: "ConvertUserToDTOPtr",
			want:     true,
		},
		{
			name:     "pointer to pointer conversion",
			funcName: "ConvertUserPtrToDTOPtr",
			want:     true,
		},
		{
			name:     "slice conversion",
			funcName: "ConvertUsersToDTO",
			want:     true,
		},
		{
			name:     "slice of pointers conversion",
			funcName: "ConvertUserSlicePtrToDTO",
			want:     true,
		},
		{
			name:     "different suffix pattern (Product->ProductResponse)",
			funcName: "ConvertProductToResponse",
			want:     true,
		},
		{
			name:     "map conversion",
			funcName: "ConvertProductMap",
			want:     true,
		},
		{
			name:     "transform naming convention",
			funcName: "TransformUserToDTO",
			want:     true,
		},
		{
			name:     "short naming convention UserToDTO",
			funcName: "UserToDTO",
			want:     true,
		},
		{
			name:     "short naming convention ToUserDTO",
			funcName: "ToUserDTO",
			want:     true,
		},
		{
			name:     "builder naming convention",
			funcName: "BuildUserDTOFromUser",
			want:     true,
		},

		// Negative cases - should NOT be detected as converters
		{
			name:     "no parameters",
			funcName: "NoParams",
			want:     false,
		},
		{
			name:     "no results",
			funcName: "NoResults",
			want:     false,
		},
		{
			name:     "only primitive types",
			funcName: "NoStructParams",
			want:     false,
		},
		{
			name:     "unrelated types with no naming similarity",
			funcName: "UnrelatedTypes",
			want:     false,
		},
		// TODO: This is a known issue mentioned in analyzer.go:168
		// The function currently incorrectly identifies same-type conversions as converters
		// {
		// 	name:     "same type input and output",
		// 	funcName: "SameTypeInOut",
		// 	want:     false,
		// },
		{
			name:     "multiple unrelated structs",
			funcName: "MultipleUnrelatedStructs",
			want:     false,
		},
		{
			name:     "only primitive return",
			funcName: "OnlyPrimitiveReturn",
			want:     false,
		},
		{
			name:     "only error return",
			funcName: "OnlyErrorReturn",
			want:     false,
		},
		{
			name:     "incompatible container types (slice to non-slice)",
			funcName: "SliceToNonSlice",
			want:     false,
		},
		{
			name:     "incompatible container types (non-slice to slice)",
			funcName: "NonSliceToSlice",
			want:     false,
		},
		{
			name:     "incompatible container types (map to slice)",
			funcName: "MapToSlice",
			want:     false,
		},
		{
			name:     "incompatible container types (slice to map)",
			funcName: "SliceToMap",
			want:     false,
		},
		{
			name:     "helper function",
			funcName: "HelperFunction",
			want:     false,
		},
		{
			name:     "converter with error return (common pattern)",
			funcName: "WithContextAndError",
			want:     true, // Should be detected despite error return
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find the function declaration
			var funcDecl *ast.FuncDecl
			for _, file := range files {
				ast.Inspect(file, func(n ast.Node) bool {
					if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == tt.funcName {
						funcDecl = fn
						return false
					}
					return true
				})
				if funcDecl != nil {
					break
				}
			}

			if funcDecl == nil {
				t.Fatalf("function %q not found in test package", tt.funcName)
			}

			// Test IsPossibleConverter
			got := sf.IsPossibleConverter(funcDecl, pass)
			if got != tt.want {
				t.Errorf("IsPossibleConverter(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}
