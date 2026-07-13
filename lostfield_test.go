package lostfield_test

import (
	"path/filepath"
	"testing"

	"github.com/amberpixels/lostfield"
	"github.com/expectto/be"
	"github.com/expectto/be/be_string"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestNewAnalyzerSmoke runs the public analyzer over the shared testdata corpus,
// proving the root package is usable exactly the way golangci-lint consumes it.
func TestNewAnalyzerSmoke(t *testing.T) {
	testdata, err := filepath.Abs(filepath.Join("internal", "lf", "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata dir: %v", err)
	}

	analysistest.Run(t, testdata, lostfield.NewAnalyzer(nil), "converters/1-readme-example")
}

func TestNewAnalyzerWithConfig(t *testing.T) {
	testdata, err := filepath.Abs(filepath.Join("internal", "lf", "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata dir: %v", err)
	}

	cfg := lostfield.DefaultConfig()
	cfg.ExcludeFieldPatterns = []string{"CreatedAt", "UpdatedAt", `Meta\.Internal`}
	analysistest.Run(t, testdata, lostfield.NewAnalyzer(cfg), "converters/16-exclude-fields/on")
}

// TestNewAnalyzerInvalidConfig verifies that a config error surfaces as the
// analyzer's Run error instead of being silently ignored.
func TestNewAnalyzerInvalidConfig(t *testing.T) {
	g := NewWithT(t)

	cfg := lostfield.DefaultConfig()
	cfg.FixMode = "smrt"

	analyzer := lostfield.NewAnalyzer(cfg)
	_, err := analyzer.Run(nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(be.All(
		be_string.ContainingSubstring(`invalid fix-mode value "smrt"`),
		be_string.ContainingSubstring("supported:"),
	))
}
