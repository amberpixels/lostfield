package config_test

import (
	"flag"
	"strings"
	"testing"

	"github.com/amberpixels/lostfield/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	// Verify boolean defaults
	if cfg.AllowMethodConverters != true {
		t.Errorf("IncludeMethods: got %v, want true", cfg.AllowMethodConverters)
	}

	if cfg.AllowGetters != true {
		t.Errorf("AllowGetters: got %v, want true", cfg.AllowGetters)
	}

	if cfg.AllowAggregators != true {
		t.Errorf("AllowAggregators: got %v, want true", cfg.AllowAggregators)
	}

	if cfg.Verbose != false {
		t.Errorf("Verbose: got %v, want false", cfg.Verbose)
	}

	if cfg.IncludeGenerated != false {
		t.Errorf("IncludeGenerated: got %v, want false", cfg.IncludeGenerated)
	}

	if cfg.IncludeDeprecated != false {
		t.Errorf("IncludeDeprecated: got %v, want false", cfg.IncludeDeprecated)
	}

	// Verify non-boolean defaults
	if len(cfg.ExcludeFieldPatterns) > 0 {
		t.Errorf("ExcludeFieldPatterns: got %q, want empty string", cfg.ExcludeFieldPatterns)
	}

	if len(cfg.ExcludeConverterPatterns) > 0 {
		t.Errorf("ExcludeConverterPatterns: got %q, want empty string", cfg.ExcludeConverterPatterns)
	}

	// Verify ExcludeFilePatterns has expected defaults
	expectedFilePatterns := []string{"*_test.go", "*.pb.go", "*/vendor/*"}
	if len(cfg.ExcludeFilePatterns) != len(expectedFilePatterns) {
		t.Errorf("ExcludeFilePatterns length: got %d, want %d", len(cfg.ExcludeFilePatterns), len(expectedFilePatterns))
	}
	for i, pattern := range expectedFilePatterns {
		if i >= len(cfg.ExcludeFilePatterns) || cfg.ExcludeFilePatterns[i] != pattern {
			t.Errorf("ExcludeFilePatterns[%d]: got %q, want %q", i, cfg.ExcludeFilePatterns, expectedFilePatterns)
			break
		}
	}

	if cfg.MinTypeNameSimilarity != 0.0 {
		t.Errorf("MinTypeSimilarity: got %v, want 0.0", cfg.MinTypeNameSimilarity)
	}

	if len(cfg.IgnoreFieldTags) > 0 {
		t.Errorf("IgnoreFieldTags: got %q, want empty string", cfg.IgnoreFieldTags)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("DefaultConfig should validate cleanly, got: %v", err)
	}
}

func TestRegisterFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		value     string
		wantErr   bool
		checkFunc func(*testing.T, *config.Config)
	}{
		{
			name:     "include-methods flag",
			flagName: "-include-methods",
			value:    "true",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				if !cfg.AllowMethodConverters {
					t.Errorf("IncludeMethods: got false, want true")
				}
			},
		},
		{
			name:     "allow-getters flag",
			flagName: "-allow-getters",
			value:    "false",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				if cfg.AllowGetters {
					t.Errorf("AllowGetters: got true, want false")
				}
			},
		},
		{
			name:     "allow-aggregators flag",
			flagName: "-allow-aggregators",
			value:    "true",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				if !cfg.AllowAggregators {
					t.Errorf("AllowAggregators: got false, want true")
				}
			},
		},
		{
			name:     "exclude-fields flag",
			flagName: "-exclude-fields",
			value:    "CreatedAt,UpdatedAt",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				want := "CreatedAt,UpdatedAt"
				if strings.Join(cfg.ExcludeFieldPatterns, ",") != want {
					t.Errorf("ExcludeFieldPatterns: got %q, want %q", cfg.ExcludeFieldPatterns, want)
				}
			},
		},
		{
			name:     "exclude-converters flag",
			flagName: "-exclude-converters",
			value:    "Get*,Map*,to*",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				want := "Get*,Map*,to*"
				if strings.Join(cfg.ExcludeConverterPatterns, ",") != want {
					t.Errorf("ExcludeConverterPatterns: got %q, want %q", cfg.ExcludeConverterPatterns, want)
				}
			},
		},
		{
			name:     "exclude-files flag",
			flagName: "-exclude-files",
			value:    "*_test.go,*.pb.go",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				want := "*_test.go,*.pb.go"
				if strings.Join(cfg.ExcludeFilePatterns, ",") != want {
					t.Errorf("ExcludeFilePatterns: got %q, want %q", cfg.ExcludeFilePatterns, want)
				}
			},
		},
		{
			name:     "min-similarity flag",
			flagName: "-min-similarity",
			value:    "0.8",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				want := 0.8
				if cfg.MinTypeNameSimilarity != want {
					t.Errorf("MinTypeSimilarity: got %v, want %v", cfg.MinTypeNameSimilarity, want)
				}
			},
		},
		{
			name:     "ignore-tags flag",
			flagName: "-ignore-tags",
			value:    "json:\"-\",lostfield:\"ignore\"",
			checkFunc: func(t *testing.T, cfg *config.Config) {
				want := "json:\"-\",lostfield:\"ignore\""
				if strings.Join(cfg.IgnoreFieldTags, ",") != want {
					t.Errorf("IgnoreFieldTags: got %q, want %q", cfg.IgnoreFieldTags, want)
				}
			},
		},
		{
			name:     "min-similarity out of range",
			flagName: "-min-similarity",
			value:    "1.5",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			flagName: "-format",
			value:    "fancy",
			wantErr:  true,
		},
		{
			name:     "invalid non-marshallable-fields",
			flagName: "-non-marshallable-fields",
			value:    "sloppy",
			wantErr:  true,
		},
		{
			name:     "invalid field-validation-mode",
			flagName: "-field-validation-mode",
			value:    "union",
			wantErr:  true,
		},
		{
			name:     "invalid fix-mode",
			flagName: "-fix-mode",
			value:    "smrt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			config.RegisterFlags(fs, &cfg)

			// Parse the flag
			args := []string{tt.flagName + "=" + tt.value}
			err := fs.Parse(args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, &cfg)
			}
		})
	}
}
