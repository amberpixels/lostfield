package config_test

import (
	"testing"

	"github.com/amberpixels/lostfield/internal/config"
	"github.com/expectto/be"
	"github.com/expectto/be/be_string"
	. "github.com/onsi/gomega"
)

func TestConfigValidate(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		g := NewWithT(t)
		cfg := config.DefaultConfig()
		g.Expect(cfg.Validate()).To(Succeed())
	})

	t.Run("invalid enum values are rejected with a helpful message", func(t *testing.T) {
		cases := []struct {
			name    string
			mutate  func(*config.Config)
			wantErr string
		}{
			{
				name:    "non-marshallable-fields",
				mutate:  func(c *config.Config) { c.NonMarshallableFieldsHandling = "sloppy" },
				wantErr: `invalid non-marshallable-fields value "sloppy"`,
			},
			{
				name:    "field-validation-mode",
				mutate:  func(c *config.Config) { c.FieldValidationMode = "union" },
				wantErr: `invalid field-validation-mode value "union"`,
			},
			{
				name:    "format",
				mutate:  func(c *config.Config) { c.Format = "fancy" },
				wantErr: `invalid format value "fancy"`,
			},
			{
				name:    "fix-mode",
				mutate:  func(c *config.Config) { c.FixMode = "smrt" },
				wantErr: `invalid fix-mode value "smrt"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				cfg := config.DefaultConfig()
				tc.mutate(&cfg)

				err := cfg.Validate()
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(be.All(
					be_string.ContainingSubstring(tc.wantErr),
					be_string.ContainingSubstring("supported:"),
				))
			})
		}
	})

	t.Run("min-similarity out of range is rejected", func(t *testing.T) {
		g := NewWithT(t)
		cfg := config.DefaultConfig()
		cfg.MinTypeNameSimilarity = 1.5
		g.Expect(cfg.Validate()).To(MatchError(be_string.ContainingSubstring("min-similarity")))

		cfg.MinTypeNameSimilarity = -0.1
		g.Expect(cfg.Validate()).To(MatchError(be_string.ContainingSubstring("min-similarity")))
	})

	t.Run("invalid exclude-fields regex is rejected", func(t *testing.T) {
		g := NewWithT(t)
		cfg := config.DefaultConfig()
		cfg.ExcludeFieldPatterns = []string{"CreatedAt", "[unclosed"}

		err := cfg.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(be_string.ContainingSubstring(`invalid exclude-fields pattern "[unclosed"`))
	})
}
