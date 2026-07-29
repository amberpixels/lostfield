package plugin_test

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	. "github.com/onsi/gomega"

	"github.com/amberpixels/lostfield"
	"github.com/amberpixels/lostfield/plugin"
)

// cfgOf builds the plugin from raw settings and returns the resulting config.
func cfgOf(t *testing.T, raw any) *lostfield.Config {
	t.Helper()

	g := NewWithT(t)
	p, err := plugin.NewPlugin(raw)
	g.Expect(err).NotTo(HaveOccurred())

	return plugin.ConfigOf(p)
}

func TestNewPluginNilSettingsKeepsDefaults(t *testing.T) {
	g := NewWithT(t)
	g.Expect(cfgOf(t, nil)).To(Equal(lostfield.DefaultConfig()))
}

// Omitted keys must keep their defaults rather than being zeroed. The true-by-default
// options are what would silently flip if settings were decoded into a zero Config,
// so they are what this guards.
func TestNewPluginOmittedKeysKeepDefaults(t *testing.T) {
	g := NewWithT(t)
	cfg := cfgOf(t, map[string]any{"min-similarity": 0.6})

	g.Expect(cfg.MinTypeNameSimilarity).To(Equal(0.6))
	g.Expect(cfg.AllowMethodConverters).To(BeTrue())
	g.Expect(cfg.AllowGetters).To(BeTrue())
	g.Expect(cfg.AllowAggregators).To(BeTrue())
	g.Expect(cfg.FieldValidationMode).To(Equal(lostfield.ModeStrict))
	g.Expect(cfg.NonMarshallableFieldsHandling).To(Equal(lostfield.HandleAdaptive))
	g.Expect(cfg.ExcludeFilePatterns).To(Equal([]string{"*_test.go", "*.pb.go", "*/vendor/*"}))
}

func TestNewPluginAppliesSettings(t *testing.T) {
	g := NewWithT(t)
	cfg := cfgOf(t, map[string]any{
		"include-methods":         false,
		"allow-getters":           false,
		"exclude-fields":          []string{"^ID$", "CreatedAt"},
		"exclude-files":           []string{"*_test.go"},
		"ignore-tags":             []string{`lostfield:"ignore"`},
		"include-private-fields":  true,
		"non-marshallable-fields": "strict",
		"field-validation-mode":   "intersection",
	})

	g.Expect(cfg.AllowMethodConverters).To(BeFalse())
	g.Expect(cfg.AllowGetters).To(BeFalse())
	g.Expect(cfg.ExcludeFieldPatterns).To(Equal([]string{"^ID$", "CreatedAt"}))
	g.Expect(cfg.ExcludeFilePatterns).To(Equal([]string{"*_test.go"}))
	g.Expect(cfg.IgnoreFieldTags).To(Equal([]string{`lostfield:"ignore"`}))
	g.Expect(cfg.IncludePrivateFields).To(BeTrue())
	g.Expect(cfg.NonMarshallableFieldsHandling).To(Equal(lostfield.HandleStrict))
	g.Expect(cfg.FieldValidationMode).To(Equal(lostfield.ModeIntersection))
}

// format, verbose and fix-mode are not part of the plugin's settings surface: they
// belong to golangci-lint, to stderr, and to --fix respectively. Being absent from the
// struct means DisallowUnknownFields rejects them, so they cannot be set at all.
func TestNewPluginRejectsKeysGolangciLintOwns(t *testing.T) {
	for _, key := range []string{"verbose", "format", "fix-mode"} {
		t.Run(key, func(t *testing.T) {
			g := NewWithT(t)
			_, err := plugin.NewPlugin(map[string]any{key: "whatever"})
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNewPluginRejectsUnknownKey(t *testing.T) {
	g := NewWithT(t)
	_, err := plugin.NewPlugin(map[string]any{"min-similarty": 0.6}) // typo
	g.Expect(err).To(HaveOccurred())
}

func TestNewPluginRejectsInvalidValue(t *testing.T) {
	g := NewWithT(t)
	_, err := plugin.NewPlugin(map[string]any{"field-validation-mode": "nonsense"})
	g.Expect(err).To(HaveOccurred())
}

// The package's init() is the whole contract with golangci-lint: `golangci-lint custom`
// blank-imports this package and then looks the name up. Importing it here exercises
// that same path, so a move or a rename cannot silently unregister it.
func TestRegisteredUnderItsName(t *testing.T) {
	g := NewWithT(t)

	newFn, err := register.GetPlugin("lostfield")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(newFn).NotTo(BeNil())

	p, err := newFn(nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(plugin.ConfigOf(p)).To(Equal(lostfield.DefaultConfig()))
}

func TestBuildAnalyzers(t *testing.T) {
	g := NewWithT(t)

	p, err := plugin.NewPlugin(nil)
	g.Expect(err).NotTo(HaveOccurred())

	analyzers, err := p.BuildAnalyzers()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(analyzers).To(HaveLen(1))
	g.Expect(analyzers[0].Name).To(Equal("lostfield"))

	g.Expect(p.GetLoadMode()).To(Equal("typesinfo"))
}
