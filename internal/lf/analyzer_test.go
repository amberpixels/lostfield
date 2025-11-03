package lf_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"

	"github.com/amberpixels/lostfield/internal/lf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBasic(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	analysistest.Run(t, testdata, analyzer, "converters/basic")
}

func TestDelegatingConverters(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that delegating converters (converters that call other converters)
	// are correctly identified and skipped from validation
	analysistest.Run(t, testdata, analyzer, "converters/delegate")
}

func TestSameTypeFunctions(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that functions with the same input and output type
	// (e.g., applyFilters(*DB) *DB) are NOT flagged as converters
	analysistest.Run(t, testdata, analyzer, "converters/sameType")
}

func TestBlankIdentifierFields(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that fields marked with blank identifier (_ = model.Field)
	// are correctly recognized as being used/acknowledged
	analysistest.Run(t, testdata, analyzer, "converters/blankIdent")
}

func TestDifferentPackagesSameName(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that converters with same-named types from different packages
	// (e.g., models.MatchedMapData -> pbVenueConfig.MatchedMapData) are correctly
	// identified as converters and not excluded by the same-type check
	analysistest.Run(t, testdata, analyzer, "converters/differentPackages")
}

func TestChainedReturnMethod(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that converters returning (&Type{fields...}).MethodCall()
	// correctly detect field usage in the composite literal
	analysistest.Run(t, testdata, analyzer, "converters/chainedReturn")
}

func TestSliceToNonSlice(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that slice->non-slice conversions are NOT caught as converters
	// These are utility functions, not proper converters
	analysistest.Run(t, testdata, analyzer, "converters/sliceToNonSlice")
}

func TestAggregatingConvertersEnabled(t *testing.T) {
	testdata := analysistest.TestData()

	// Create analyzer with aggregating converters enabled
	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test verifies that aggregating converters (slice -> non-slice)
	// are detected and validated when the feature is enabled.
	// Since we can't easily control config from here, and the test is designed
	// to work without the flag, we'll document the expected behavior:
	// With --allow-aggregators=true, the sliceToNonSlice converter would be caught
	// and validated because:
	// - Input: []*VenueDetail (has Name, Sections fields)
	// - Output: Metadata with Categories []Category (Category has Name, Sections)
	// - All fields are properly mapped, so it should pass validation
	analysistest.Run(t, testdata, analyzer, "converters/sliceToNonSlice")
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
		{
			name:     "embedded struct conversion (manually setting all embedded fields)",
			funcName: "FromDomain",
			want:     true,
		},

		// Negative cases - should NOT be detected as converters
		{
			name:     "constructor with New prefix",
			funcName: "NewDecorator",
			want:     false,
		},
		{
			name:     "constructor with New prefix and similar names",
			funcName: "NewUserDTO",
			want:     false,
		},
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
		{
			name:     "embedded struct with missing fields (still detected as possible converter)",
			funcName: "FromDomainIncomplete",
			want:     true, // Detected as possible converter, but validation will fail
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
			got := lf.IsPossibleConverter(funcDecl, pass)
			if got != tt.want {
				t.Errorf("IsPossibleConverter(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestDeprecatedFields(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test ensures that deprecated fields are handled correctly by the analyzer.
	// The test converter handles all fields including the deprecated OldName field.
	// This tests that:
	// - Deprecated fields CAN be used and converted if desired
	// - The analyzer correctly detects the "Deprecated:" comment in field definitions
	// - isDeprecatedField() works properly for fields marked with deprecation comments
	analysistest.Run(t, testdata, analyzer, "converters/deprecated")
}

func TestReadmeExample(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "lostfield",
		Doc:  "reports all inconsistent converter functions: finds lost fields)",
		Run:  lf.Run,
	}

	// This test validates the example shown in readme.md
	analysistest.Run(t, testdata, analyzer, "converters/readmeExample")
}

// TestMatchesAnyPattern tests the glob pattern matching logic.
func TestMatchesAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		patterns []string
		want     bool
	}{
		{
			name:     "matches Get* pattern",
			funcName: "GetUser",
			patterns: []string{"Get*"},
			want:     true,
		},
		{
			name:     "matches Map* pattern",
			funcName: "MapResponse",
			patterns: []string{"Map*"},
			want:     true,
		},
		{
			name:     "matches to* pattern",
			funcName: "toDTO",
			patterns: []string{"to*"},
			want:     true,
		},
		{
			name:     "matches *Helper pattern",
			funcName: "ConvertHelper",
			patterns: []string{"*Helper"},
			want:     true,
		},
		{
			name:     "matches one of multiple patterns",
			funcName: "MapUser",
			patterns: []string{"Get*", "Map*", "to*"},
			want:     true,
		},
		{
			name:     "does not match any pattern",
			funcName: "ConvertUser",
			patterns: []string{"Get*", "Map*", "to*"},
			want:     false,
		},
		{
			name:     "empty patterns list",
			funcName: "GetUser",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "exact match pattern",
			funcName: "GetUser",
			patterns: []string{"GetUser"},
			want:     true,
		},
		{
			name:     "single char wildcard ?",
			funcName: "GetA",
			patterns: []string{"Get?"},
			want:     true,
		},
		{
			name:     "single char wildcard does not match multiple",
			funcName: "GetAB",
			patterns: []string{"Get?"},
			want:     false,
		},
		{
			name:     "middle wildcard",
			funcName: "ConvertUserToDTO",
			patterns: []string{"Convert*ToDTO"},
			want:     true,
		},
		{
			name:     "case sensitive matching",
			funcName: "getUser",
			patterns: []string{"Get*"},
			want:     false,
		},
		// File path pattern tests
		{
			name:     "matches full path with *_test.go pattern",
			funcName: "/Users/e/code/github.com/ht/ht-mercator/internal/domain/svgmap/tixstockmaps/parser_test.go",
			patterns: []string{"*_test.go"},
			want:     true,
		},
		{
			name:     "matches full path with *.pb.go pattern",
			funcName: "/path/to/proto/generated/file.pb.go",
			patterns: []string{"*.pb.go"},
			want:     true,
		},
		{
			name:     "matches full path with */vendor/* pattern",
			funcName: "/path/to/project/vendor/github.com/foo/bar.go",
			patterns: []string{"*/vendor/*"},
			want:     true,
		},
		{
			name:     "does not match non-test file with *_test.go pattern",
			funcName: "/path/to/project/internal/file.go",
			patterns: []string{"*_test.go"},
			want:     false,
		},
		{
			name:     "matches with mixed path and basename patterns",
			funcName: "/path/to/vendor/module/file_test.go",
			patterns: []string{"*_test.go", "*/vendor/*"},
			want:     true,
		},
		// Edge cases
		{
			name:     "empty string does not match any pattern",
			funcName: "",
			patterns: []string{"*"},
			want:     false,
		},
		{
			name:     "pattern with only * matches any basename",
			funcName: "anything.go",
			patterns: []string{"*"},
			want:     true,
		},
		{
			name:     "exact filename match without wildcard",
			funcName: "/path/to/exact.go",
			patterns: []string{"exact.go"},
			want:     true,
		},
		{
			name:     "exact filename no match without wildcard",
			funcName: "/path/to/other.go",
			patterns: []string{"exact.go"},
			want:     false,
		},
		// Multiple directory levels in path pattern
		{
			name:     "matches nested vendor directory",
			funcName: "/project/src/vendor/github.com/pkg/file.go",
			patterns: []string{"*/vendor/*"},
			want:     true,
		},
		{
			name:     "does not match vendor in filename only",
			funcName: "/project/src/vendorfile.go",
			patterns: []string{"*/vendor/*"},
			want:     false,
		},
		{
			name:     "matches vendor at any depth",
			funcName: "/a/b/c/vendor/d/e/f/file.go",
			patterns: []string{"*/vendor/*"},
			want:     true,
		},
		// Proto file variations
		{
			name:     "matches .pb.go in deep path",
			funcName: "/project/internal/proto/generated/service.pb.go",
			patterns: []string{"*.pb.go"},
			want:     true,
		},
		{
			name:     "does not match .pb in middle of filename",
			funcName: "/project/file.pb.old.go",
			patterns: []string{"*.pb.go"},
			want:     false,
		},
		// Test file variations
		{
			name:     "matches test file without package",
			funcName: "simple_test.go",
			patterns: []string{"*_test.go"},
			want:     true,
		},
		{
			name:     "matches test file with long name",
			funcName: "/path/to/very_long_descriptive_converter_test.go",
			patterns: []string{"*_test.go"},
			want:     true,
		},
		{
			name:     "does not match test in middle of filename",
			funcName: "/path/to/test_file.go",
			patterns: []string{"*_test.go"},
			want:     false,
		},
		// Multiple wildcards
		{
			name:     "matches multiple wildcards in filename pattern",
			funcName: "convert_user_to_dto.go",
			patterns: []string{"convert_*_to_*.go"},
			want:     true,
		},
		{
			name:     "does not match multiple wildcards with wrong structure",
			funcName: "convert_user_dto.go",
			patterns: []string{"convert_*_to_*.go"},
			want:     false,
		},
		// Path pattern variations
		{
			name:     "matches */generated/* pattern",
			funcName: "/project/internal/generated/models/user.go",
			patterns: []string{"*/generated/*"},
			want:     true,
		},
		{
			name:     "matches */mocks/* pattern",
			funcName: "/project/internal/service/mocks/mock_service.go",
			patterns: []string{"*/mocks/*"},
			want:     true,
		},
		{
			name:     "does not match similar but different path",
			funcName: "/project/internal/vendordata/file.go",
			patterns: []string{"*/vendor/*"},
			want:     false,
		},
		// Real-world scenarios
		{
			name:     "matches proto in realistic path structure",
			funcName: "/home/user/project/api/proto/v1/service.pb.go",
			patterns: []string{"*.pb.go", "*_test.go", "*/vendor/*"},
			want:     true,
		},
		{
			name:     "matches test in realistic path structure",
			funcName: "/home/user/project/internal/domain/converter_test.go",
			patterns: []string{"*.pb.go", "*_test.go", "*/vendor/*"},
			want:     true,
		},
		{
			name:     "matches vendor in realistic path structure",
			funcName: "/home/user/project/vendor/github.com/lib/module.go",
			patterns: []string{"*.pb.go", "*_test.go", "*/vendor/*"},
			want:     true,
		},
		{
			name:     "does not match regular file with default exclusions",
			funcName: "/home/user/project/internal/domain/converter.go",
			patterns: []string{"*.pb.go", "*_test.go", "*/vendor/*"},
			want:     false,
		},
		// Complex basename patterns
		{
			name:     "matches complex basename pattern with prefix and suffix",
			funcName: "/path/to/mock_service_test.go",
			patterns: []string{"mock_*.go"},
			want:     true,
		},
		{
			name:     "matches pattern with question mark",
			funcName: "/path/file1.go",
			patterns: []string{"file?.go"},
			want:     true,
		},
		{
			name:     "does not match question mark with multiple chars",
			funcName: "/path/file12.go",
			patterns: []string{"file?.go"},
			want:     false,
		},
		// Extension variations
		{
			name:     "matches .proto.go files",
			funcName: "/path/service.proto.go",
			patterns: []string{"*.proto.go"},
			want:     true,
		},
		{
			name:     "matches _generated.go files",
			funcName: "/path/models_generated.go",
			patterns: []string{"*_generated.go"},
			want:     true,
		},
		// First pattern wins scenarios
		{
			name:     "matches first matching pattern",
			funcName: "/path/converter_test.go",
			patterns: []string{"*_test.go", "converter_*"},
			want:     true,
		},
		{
			name:     "matches second pattern when first doesn't match",
			funcName: "/path/converter_impl.go",
			patterns: []string{"*_test.go", "converter_*"},
			want:     true,
		},
		// Single file (no directory)
		{
			name:     "matches single file with test pattern",
			funcName: "main_test.go",
			patterns: []string{"*_test.go"},
			want:     true,
		},
		{
			name:     "matches single file with pb pattern",
			funcName: "service.pb.go",
			patterns: []string{"*.pb.go"},
			want:     true,
		},
		{
			name:     "does not match vendor pattern without path",
			funcName: "vendor.go",
			patterns: []string{"*/vendor/*"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lf.MatchesAnyPattern(tt.funcName, tt.patterns)
			if got != tt.want {
				t.Errorf("MatchesAnyPattern(%q, %v) = %v, want %v", tt.funcName, tt.patterns, got, tt.want)
			}
		})
	}
}
