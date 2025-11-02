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

	if cfg.IgnoreDeprecated != false {
		t.Errorf("IncludeDeprecated: got %v, want false", cfg.IgnoreDeprecated)
	}

	// Verify non-boolean defaults
	if len(cfg.ExcludeFieldPatterns) > 0 {
		t.Errorf("ExcludeFieldPatterns: got %q, want empty string", cfg.ExcludeFieldPatterns)
	}

	if cfg.MinTypeNameSimilarity != 0.0 {
		t.Errorf("MinTypeSimilarity: got %v, want 0.0", cfg.MinTypeNameSimilarity)
	}

	if len(cfg.IgnoreFieldTags) > 0 {
		t.Errorf("IgnoreFieldTags: got %q, want empty string", cfg.IgnoreFieldTags)
	}
}

func TestRegisterFlags(t *testing.T) {
	// Create a new FlagSet for testing
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	config.RegisterFlags(fs)

	// Test that flags are registered
	tests := []struct {
		name      string
		flagName  string
		value     string
		wantErr   bool
		checkFunc func(*testing.T)
	}{
		{
			name:     "include-methods flag",
			flagName: "-include-methods",
			value:    "true",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				if !cfg.AllowMethodConverters {
					t.Errorf("IncludeMethods: got false, want true")
				}
			},
		},
		{
			name:     "allow-getters flag",
			flagName: "-allow-getters",
			value:    "false",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				if cfg.AllowGetters {
					t.Errorf("AllowGetters: got true, want false")
				}
			},
		},
		{
			name:     "allow-aggregators flag",
			flagName: "-allow-aggregators",
			value:    "true",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				if !cfg.AllowAggregators {
					t.Errorf("AllowAggregators: got false, want true")
				}
			},
		},
		{
			name:     "exclude-fields flag",
			flagName: "-exclude-fields",
			value:    "CreatedAt,UpdatedAt",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				want := "CreatedAt,UpdatedAt"
				if strings.Join(cfg.ExcludeFieldPatterns, ",") != want {
					t.Errorf("ExcludeFieldPatterns: got %q, want %q", cfg.ExcludeFieldPatterns, want)
				}
			},
		},
		{
			name:     "min-similarity flag",
			flagName: "-min-similarity",
			value:    "0.8",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
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
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				want := "json:\"-\",lostfield:\"ignore\""
				if strings.Join(cfg.IgnoreFieldTags, ",") != want {
					t.Errorf("IgnoreFieldTags: got %q, want %q", cfg.IgnoreFieldTags, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags and config for each test
			fs = flag.NewFlagSet("test", flag.ContinueOnError)
			config.RegisterFlags(fs)

			// Parse the flag
			args := []string{tt.flagName + "=" + tt.value}
			err := fs.Parse(args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

func TestConfigGet(t *testing.T) {
	// Get should return the current configuration
	cfg := config.Get()

	// Verify it's a valid Config struct with all fields accessible
	_ = cfg.AllowMethodConverters
	_ = cfg.AllowGetters
	_ = cfg.AllowAggregators
	_ = cfg.Verbose
	_ = cfg.IncludeGenerated
	_ = cfg.IgnoreDeprecated
	_ = cfg.ExcludeFieldPatterns
	_ = cfg.MinTypeNameSimilarity
	_ = cfg.IgnoreFieldTags
}
