// Package plugin registers lostfield with golangci-lint's module plugin system.
//
// It lives apart from the root package on purpose. golangci-lint imports
// github.com/amberpixels/lostfield for the analyzer itself, and if the registration
// sat there, their binary would carry plugin-module-register and self-register
// lostfield in the custom-plugin registry. .custom-gcl.yml takes an `import` field for
// exactly this, so point it here:
//
//	# .custom-gcl.yml
//	version: v2.12.2
//	plugins:
//	  - module: 'github.com/amberpixels/lostfield'
//	    import: 'github.com/amberpixels/lostfield/plugin'
//	    version: latest
//
//	$ golangci-lint custom
//
// and configure it in .golangci.yml under linters.settings.custom.lostfield.
package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/amberpixels/lostfield"
)

//nolint:gochecknoinits // init-time registration is the golangci-lint module plugin contract
func init() {
	register.Plugin("lostfield", newPlugin)
}

// settings is the configuration surface golangci-lint sees. It deliberately omits
// three of Config's keys rather than accepting and then overriding them:
//
//   - format: golangci-lint owns output formatting. The pretty format embeds ANSI and
//     multi-line excerpts into the message, which garbles their formatters.
//   - verbose: makes the analyzer write to os.Stderr mid-run, interleaving with
//     golangci-lint's own output.
//   - fix-mode: suggested fixes already reach --fix through analysis.SuggestedFix.
//
// Naming them here would let a user set them; leaving them out makes it impossible.
type settings struct {
	IncludeMethods        *bool    `json:"include-methods"`
	AllowGetters          *bool    `json:"allow-getters"`
	AllowAggregators      *bool    `json:"allow-aggregators"`
	ExcludeFields         []string `json:"exclude-fields"`
	ExcludeConverters     []string `json:"exclude-converters"`
	OnlyConverters        []string `json:"only-converters"`
	ExcludeFiles          []string `json:"exclude-files"`
	MinSimilarity         *float64 `json:"min-similarity"`
	IgnoreTags            []string `json:"ignore-tags"`
	IncludeGenerated      *bool    `json:"include-generated"`
	IncludeDeprecated     *bool    `json:"include-deprecated"`
	IncludePrivateFields  *bool    `json:"include-private-fields"`
	NonMarshallableFields *string  `json:"non-marshallable-fields"`
	FieldValidationMode   *string  `json:"field-validation-mode"`
}

// plugin adapts the lostfield analyzer to golangci-lint's LinterPlugin contract.
type plugin struct {
	cfg *lostfield.Config
}

// newPlugin builds the plugin from golangci-lint settings
// (linters.settings.custom.lostfield.settings).
//
// Values are applied on top of DefaultConfig, so an omitted key keeps its default
// rather than being zeroed - which is why the scalars are pointers, and why
// register.DecodeSettings is not used: it decodes into a zero value, which would flip
// the true-by-default options off. Unknown keys are rejected so a typo fails the build.
func newPlugin(raw any) (register.LinterPlugin, error) {
	cfg := lostfield.DefaultConfig()

	if raw != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(raw); err != nil {
			return nil, fmt.Errorf("lostfield: encoding settings: %w", err)
		}

		var s settings
		decoder := json.NewDecoder(&buf)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&s); err != nil {
			return nil, fmt.Errorf("lostfield: decoding settings: %w", err)
		}

		s.applyTo(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("lostfield: %w", err)
	}

	return &plugin{cfg: cfg}, nil
}

// applyTo copies onto cfg only the settings that were actually present.
func (s *settings) applyTo(cfg *lostfield.Config) {
	setBool(&cfg.AllowMethodConverters, s.IncludeMethods)
	setBool(&cfg.AllowGetters, s.AllowGetters)
	setBool(&cfg.AllowAggregators, s.AllowAggregators)
	setBool(&cfg.IncludeGenerated, s.IncludeGenerated)
	setBool(&cfg.IncludeDeprecated, s.IncludeDeprecated)
	setBool(&cfg.IncludePrivateFields, s.IncludePrivateFields)

	if s.MinSimilarity != nil {
		cfg.MinTypeNameSimilarity = *s.MinSimilarity
	}
	if s.NonMarshallableFields != nil {
		cfg.NonMarshallableFieldsHandling = lostfield.NonMarshallableFieldsHandling(*s.NonMarshallableFields)
	}
	if s.FieldValidationMode != nil {
		cfg.FieldValidationMode = lostfield.FieldValidationMode(*s.FieldValidationMode)
	}

	setSlice(&cfg.ExcludeFieldPatterns, s.ExcludeFields)
	setSlice(&cfg.ExcludeConverterPatterns, s.ExcludeConverters)
	setSlice(&cfg.OnlyConverterPatterns, s.OnlyConverters)
	setSlice(&cfg.ExcludeFilePatterns, s.ExcludeFiles)
	setSlice(&cfg.IgnoreFieldTags, s.IgnoreTags)
}

func setBool(dst, src *bool) {
	if src != nil {
		*dst = *src
	}
}

func setSlice(dst *[]string, src []string) {
	if src != nil {
		*dst = src
	}
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{lostfield.NewAnalyzer(p.cfg)}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
