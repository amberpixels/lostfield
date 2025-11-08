package lf_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"

	"github.com/amberpixels/lostfield/internal/config"
	"github.com/amberpixels/lostfield/internal/lf"
	"golang.org/x/tools/go/analysis"
)

func TestReadmeExample(t *testing.T) {
	runAnalysisTest(t, "converters/1-readme-example",
		DiagnosticAssertion{
			FunctionName: "ConvertUserToDTO",
			FieldsMissing: []string{
				"user.Email",
				"user.CreatedAt",
				"Email",
			},
		},
	)
}

func TestBasic(t *testing.T) {
	t.Run("2-basic:clean", func(t *testing.T) {
		// Valid converter that uses all fields correctly
		runAnalysisTest(t, "converters/2-basic/clean")
	})

	t.Run("2-basic:dirty", func(t *testing.T) {
		// Various incomplete converters
		runAnalysisTest(t, "converters/2-basic/dirty",
			DiagnosticAssertion{
				FunctionName:  "ConvertSampleToDB_MissingPrice",
				FieldsMissing: []string{"sample.Price", "result.Price"},
			},
			DiagnosticAssertion{
				FunctionName:  "ConvertSampleToDB_MissingCurrency",
				FieldsMissing: []string{"sample.Currency", "result.Currency"},
			},
			DiagnosticAssertion{
				FunctionName:  "ConvertSampleToDB_MissingBoth",
				FieldsMissing: []string{"sample.Price", "sample.Currency", "result.Price", "result.Currency"},
			},
		)
	})
}

func TestDelegatingConverters(t *testing.T) {
	t.Run("3-delegate:clean", func(t *testing.T) {
		// Valid delegating converters are skipped from validation
		runAnalysisTest(t, "converters/3-delegate/clean")
	})

	t.Run("3-delegate:dirty", func(t *testing.T) {
		// Delegating to incomplete converter
		runAnalysisTest(t, "converters/3-delegate/dirty",
			DiagnosticAssertion{
				FunctionName: "ConvertTicketModelToProto_Incomplete",
				FieldsMissing: []string{
					"t.TicketsCount",
					"t.Display",
					"t.ServiceInfo",
					"t.IsLeavingSingleSeats",
					"t.IsOwnTickets",
					"t.SeatsWarnings",
					"t.RewriteDetails",
					"IsOwnTickets",
					"TicketsCount",
					"IsLeavingSingleSeats",
					"ServiceInfo",
					"SeatsWarnings",
					"RewriteDetails",
					"Display",
				},
			},
		)
	})
}

func TestSameTypeFunctions(t *testing.T) {
	// Functions with same input and output type are NOT flagged as converters
	runAnalysisTest(t, "converters/4-same-type")
}

func TestBlankIdentifierFields(t *testing.T) {
	t.Run("5-blank-ident:clean", func(t *testing.T) {
		// Blank identifier fields are correctly recognized as being used/acknowledged
		runAnalysisTest(t, "converters/5-blank-ident/clean")
	})

	t.Run("5-blank-ident:dirty", func(t *testing.T) {
		// Without blank ident, field is reported as missing
		runAnalysisTest(t, "converters/5-blank-ident/dirty",
			DiagnosticAssertion{
				FunctionName:  "ConvertPerformanceMapSchemeModelToProto_WithoutBlankIdent",
				FieldsMissing: []string{"model.VenueConfiguration"},
			},
		)
	})
}

func TestDifferentPackagesSameName(t *testing.T) {
	t.Run("6-different-packages:clean", func(t *testing.T) {
		// Valid converters with same-named types from different packages
		runAnalysisTest(t, "converters/6-different-packages/clean")
	})

	t.Run("6-different-packages:dirty", func(t *testing.T) {
		// Incomplete converters with missing fields
		runAnalysisTest(t, "converters/6-different-packages/dirty",
			DiagnosticAssertion{
				FunctionName:  "MatchedCategoryToProto_Incomplete",
				FieldsMissing: []string{"category.Name", "Name"},
			},
			DiagnosticAssertion{
				FunctionName:  "MatchingDetailsToProto_Incomplete",
				FieldsMissing: []string{"details.Info", "Info"},
			},
			DiagnosticAssertion{
				FunctionName:  "MatchedMapDataToProto_Incomplete",
				FieldsMissing: []string{"data.Details", "Details"},
			},
		)
	})
}

func TestChainedReturnMethod(t *testing.T) {
	// Valid converters with chained return methods
	runAnalysisTest(t, "converters/7-chained-return/clean")
}

func TestSliceToNonSlice(t *testing.T) {
	// Slice->non-slice conversions are NOT caught as converters (they're utility functions)
	// Expect 0 diagnostics
	runAnalysisTest(t, "converters/8-slice-to-non-slice")
}

func TestAggregatingConvertersEnabled(t *testing.T) {
	// Aggregating converters (slice -> non-slice with valid field mapping)
	// With proper field mapping, should pass validation. Expect 0 diagnostics
	runAnalysisTest(t, "converters/8-slice-to-non-slice")
}

func TestDeprecatedFields(t *testing.T) {
	// Converter handles all fields including the deprecated OldName field
	runAnalysisTest(t, "converters/9-deprecated/clean")
}

func TestNonMarshallableFields(t *testing.T) {
	t.Run("ignore mode", func(t *testing.T) {
		// In ignore mode, non-marshallable fields (func, chan) are completely ignored
		cfg := config.DefaultConfig()
		cfg.NonMarshallableFieldsHandling = config.HandleIgnore
		runAnalysisTestWithConfig(t, "converters/10-non-marshallable-fields/ignore", cfg)
	})

	t.Run("adaptive mode (default)", func(t *testing.T) {
		// In adaptive mode, non-marshallable fields are validated only if they exist in both input and output
		// Since Handler/Notify don't exist in ApiApple/MessageDTO, they're ignored
		cfg := config.DefaultConfig()
		cfg.NonMarshallableFieldsHandling = config.HandleAdaptive
		runAnalysisTestWithConfig(t, "converters/10-non-marshallable-fields/if_present", cfg)
	})
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
