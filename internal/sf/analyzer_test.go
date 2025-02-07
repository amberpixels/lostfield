// internal/sf/analyzer_test.go
package sf_test

import (
	"testing"

	"github.com/amberpixels/go-stickyfields/internal/sf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestC1(t *testing.T) {
	testdata := analysistest.TestData()

	analyzer := &analysis.Analyzer{
		Name: "stickyfields",
		Doc:  "reports all inconsistent converter functions: ensures sticky fields)",
		Run:  sf.Run,
	}

	analysistest.Run(t, testdata, analyzer, "converters/c1")
}
