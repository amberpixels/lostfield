package config_test

import (
	"flag"
	"testing"

	"github.com/amberpixels/go-stickyfields/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	// Verify defaults match expected values
	if cfg.IncludeMethods != false {
		t.Errorf("IncludeMethods: got %v, want false", cfg.IncludeMethods)
	}

	if cfg.AllowGetters != true {
		t.Errorf("AllowGetters: got %v, want true", cfg.AllowGetters)
	}

	if cfg.ExcludeFieldPatterns != "" {
		t.Errorf("ExcludeFieldPatterns: got %q, want empty string", cfg.ExcludeFieldPatterns)
	}

	if cfg.MinTypeSimilarity != 0.0 {
		t.Errorf("MinTypeSimilarity: got %v, want 0.0", cfg.MinTypeSimilarity)
	}

	if cfg.IgnoreFieldTags != "" {
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
				if !cfg.IncludeMethods {
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
			name:     "exclude-fields flag",
			flagName: "-exclude-fields",
			value:    "CreatedAt,UpdatedAt",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				want := "CreatedAt,UpdatedAt"
				if cfg.ExcludeFieldPatterns != want {
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
				if cfg.MinTypeSimilarity != want {
					t.Errorf("MinTypeSimilarity: got %v, want %v", cfg.MinTypeSimilarity, want)
				}
			},
		},
		{
			name:     "ignore-tags flag",
			flagName: "-ignore-tags",
			value:    "json:\"-\",stickyfields:\"ignore\"",
			checkFunc: func(t *testing.T) {
				cfg := config.Get()
				want := "json:\"-\",stickyfields:\"ignore\""
				if cfg.IgnoreFieldTags != want {
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

	// Verify it's a valid Config struct
	_ = cfg.IncludeMethods
	_ = cfg.AllowGetters
	_ = cfg.ExcludeFieldPatterns
	_ = cfg.MinTypeSimilarity
	_ = cfg.IgnoreFieldTags
}
