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

func TestAggregatingConverter(t *testing.T) {
	// An aggregating converter (slice -> non-slice with a slice field) that drops
	// element fields on both sides is reported with the Categories[].Field notation.
	runAnalysisTest(t, "converters/19-aggregating",
		DiagnosticAssertion{
			FunctionName:  "AggregateDetailsIncomplete",
			FieldsMissing: []string{"detail.Sections", "Categories[].Sections"},
		},
	)
}

func TestAggregatingConvertersEnabled(t *testing.T) {
	// Aggregating converters (slice -> non-slice with valid field mapping)
	// With proper field mapping, should pass validation. Expect 0 diagnostics
	runAnalysisTest(t, "converters/8-slice-to-non-slice")
}

func TestDeprecatedFields(t *testing.T) {
	t.Run("default: deprecated fields are excluded from validation", func(t *testing.T) {
		// Skipping the deprecated OldName field is fine by default;
		// copying it anyway is also fine.
		runAnalysisTest(t, "converters/9-deprecated/clean")
	})

	t.Run("include-deprecated: deprecated fields are validated", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.IncludeDeprecated = true
		runAnalysisTestWithConfig(t, "converters/9-deprecated/dirty", cfg,
			DiagnosticAssertion{
				FunctionName:  "ConvertEventToReplySkippingDeprecated",
				FieldsMissing: []string{"model.OldName", "OldName"},
			},
		)
	})
}

func TestExcludeFields(t *testing.T) {
	t.Run("off: skipped timestamp fields are reported", func(t *testing.T) {
		runAnalysisTest(t, "converters/16-exclude-fields/off",
			DiagnosticAssertion{
				FunctionName:  "ConvertArticle",
				FieldsMissing: []string{"a.CreatedAt", "a.UpdatedAt"},
			},
		)
	})

	t.Run("on: excluded fields (leaf and nested path) are not required", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.ExcludeFieldPatterns = []string{"CreatedAt", "UpdatedAt", `Meta\.Internal`}
		runAnalysisTestWithConfig(t, "converters/16-exclude-fields/on", cfg)
	})
}

func TestIgnoreTags(t *testing.T) {
	t.Run("off: tagged fields are still required", func(t *testing.T) {
		runAnalysisTest(t, "converters/17-ignore-tags/off",
			DiagnosticAssertion{
				FunctionName:  "ConvertProduct",
				FieldsMissing: []string{"p.Secret", "p.Audit", "p.Comments"},
			},
		)
	})

	t.Run("on: tagged fields are skipped", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.IgnoreFieldTags = []string{`lostfield:"ignore"`, "internal", `json:"-"`}
		runAnalysisTestWithConfig(t, "converters/17-ignore-tags/on", cfg)
	})
}

func TestMinSimilarity(t *testing.T) {
	// With min-similarity=0.6:
	// - Message vs MessageNewParams (dice ~0.57) -> callAndMeter is NOT a converter
	// - UserModel vs UserModelDTO (dice > 0.6) -> ConvertUser is still validated
	cfg := config.DefaultConfig()
	cfg.MinTypeNameSimilarity = 0.6
	runAnalysisTestWithConfig(t, "converters/18-min-similarity", cfg,
		DiagnosticAssertion{
			FunctionName:  "ConvertUser",
			FieldsMissing: []string{"u.Name", "Name"},
		},
	)
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

func TestPrivateFields(t *testing.T) {
	t.Run("default mode (private fields ignored)", func(t *testing.T) {
		// In default mode (IncludePrivateFields=false), private fields are skipped
		// Only public fields (ID, Name) are validated
		cfg := config.DefaultConfig()
		cfg.IncludePrivateFields = false
		runAnalysisTestWithConfig(t, "converters/11-private-fields/default", cfg)
	})
}

func TestNestedFields(t *testing.T) {
	t.Run("12-nested-fields:clean", func(t *testing.T) {
		runAnalysisTest(t, "converters/12-nested-fields/clean")
	})

	t.Run("12-nested-fields:dirty", func(t *testing.T) {
		// Converters with missing nested fields - all cases are now fully detected
		// Uses "want" comments in test code to specify exact expected diagnostics
		runAnalysisTest(t, "converters/12-nested-fields/dirty",
			DiagnosticAssertion{
				FunctionName:  "ConvertEventToDTO_FullInlineDeclaration_Missed_Deep_Field",
				FieldsMissing: []string{"event.User.Role.Name", "User.Role.Name"},
			},
			DiagnosticAssertion{
				FunctionName:  "ConvertEventToDTO_FullInlineDeclaration_Missed_Pointer_Field",
				FieldsMissing: []string{"event.User.Group", "User.Group"},
			},
			DiagnosticAssertion{
				FunctionName:  "ConvertEventToDTO_Mixed_Missed_Nested_Field",
				FieldsMissing: []string{"event.Owner.Group", "result.Owner.Group"},
			},
			DiagnosticAssertion{
				FunctionName:  "ConvertEventToDTO_DotNotation_Missed_First_Level",
				FieldsMissing: []string{"event.User", "result.User", "result.Owner.ID", "result.Owner.Name"},
			},
		)
	})
}

func TestSliceInlineMapping(t *testing.T) {
	t.Run("13-slice-inline-mapping:clean", func(t *testing.T) {
		// Slice-to-slice converters with inline composite literal mapping
		// should NOT produce diagnostics when all fields are properly mapped
		runAnalysisTest(t, "converters/13-slice-inline-mapping/clean")
	})

	t.Run("13-slice-inline-mapping:dirty", func(t *testing.T) {
		// Slice-to-slice converter with missing Weight field
		runAnalysisTest(t, "converters/13-slice-inline-mapping/dirty",
			DiagnosticAssertion{
				FunctionName:  "ConvertPetInfosToDTO_MissingWeight",
				FieldsMissing: []string{"rec.Weight", "Weight"},
			},
		)
	})
}

func TestFixSafe(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.FixMode = "safe"
	runAnalysisTestWithConfig(t, "converters/14-fix-safe", cfg,
		DiagnosticAssertion{
			FunctionName: "ConvertUserToDTO_MissingFields",
			FieldsMissing: []string{
				"user.Email",
				"user.Phone",
				"result.Email",
				"result.Phone",
			},
		},
	)
}

func TestFixSafe_HasSuggestedFixes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.FixMode = "safe"

	diagnostics := runRawAnalysisTestWithConfig(t, "converters/14-fix-safe", &cfg)
	for _, diag := range diagnostics {
		if len(diag.SuggestedFixes) == 0 {
			t.Error("expected SuggestedFixes to be populated in safe mode")
		}
		if len(diag.SuggestedFixes) != 1 {
			t.Errorf("expected exactly 1 suggested fix in safe mode, got %d", len(diag.SuggestedFixes))
		}
		for _, fix := range diag.SuggestedFixes {
			if len(fix.TextEdits) == 0 {
				t.Error("expected TextEdits in suggested fix")
			}
		}
	}
}

func TestFixSmart_HasSuggestedFixes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.FixMode = "smart"

	diagnostics := runRawAnalysisTestWithConfig(t, "converters/15-fix-smart", &cfg)
	for _, diag := range diagnostics {
		if len(diag.SuggestedFixes) == 0 {
			t.Error("expected SuggestedFixes to be populated in smart mode")
		}
		// Smart mode should have 2 fixes: smart first, safe second
		if len(diag.SuggestedFixes) != 2 {
			t.Errorf("expected 2 suggested fixes in smart mode, got %d", len(diag.SuggestedFixes))
		}
	}
}

func TestFixDisabled_NoSuggestedFixes(t *testing.T) {
	diagnostics := runRawAnalysisTestWithConfig(t, "converters/14-fix-safe", nil)
	for _, diag := range diagnostics {
		if len(diag.SuggestedFixes) != 0 {
			t.Error("expected no SuggestedFixes when fix mode is disabled")
		}
	}
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
		{
			name:     "same type input and output",
			funcName: "SameTypeInOut",
			want:     false,
		},
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
			defaultCfg := config.DefaultConfig()
			got := lf.IsPossibleConverter(funcDecl, pass, &defaultCfg)
			if got != tt.want {
				t.Errorf("IsPossibleConverter(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}
