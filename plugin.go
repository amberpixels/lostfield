package lostfield

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

// init registers lostfield with golangci-lint's module plugin system, so a
// custom golangci-lint binary can be built with:
//
//	# .custom-gcl.yml
//	version: v2.12.2
//	plugins:
//	  - module: 'github.com/amberpixels/lostfield'
//	    version: latest
//
//	$ golangci-lint custom
//
// and configured in .golangci.yml under linters.settings.custom.lostfield.
//
//nolint:gochecknoinits // init-time registration is the golangci-lint module plugin contract
func init() {
	register.Plugin("lostfield", newPlugin)
}

// plugin adapts the lostfield analyzer to golangci-lint's LinterPlugin contract.
type plugin struct {
	cfg *Config
}

// newPlugin builds the plugin from golangci-lint settings
// (linters.settings.custom.lostfield.settings).
//
// Settings are decoded ON TOP of DefaultConfig - omitted keys keep their default
// values (register.DecodeSettings is not used because it decodes into a zero
// Config, which would flip the true-by-default options off). Unknown keys are
// rejected so config typos fail the build instead of being silently ignored.
func newPlugin(settings any) (register.LinterPlugin, error) {
	cfg := DefaultConfig()

	if settings != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(settings); err != nil {
			return nil, fmt.Errorf("lostfield: encoding settings: %w", err)
		}

		decoder := json.NewDecoder(&buf)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("lostfield: decoding settings: %w", err)
		}
	}

	// The pretty format embeds ANSI colors and multi-line source excerpts into
	// diagnostic messages; inside golangci-lint that garbles the output formats.
	// Force the plain single-line format regardless of settings.
	cfg.Format = FormatDefault

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("lostfield: %w", err)
	}

	return &plugin{cfg: cfg}, nil
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{NewAnalyzer(p.cfg)}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
