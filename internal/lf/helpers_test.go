package lf_test

import (
	"testing"

	"github.com/amberpixels/lostfield/internal/lf"
)

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
