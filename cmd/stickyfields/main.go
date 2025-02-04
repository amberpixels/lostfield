package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/amberpixels/go-stickyfields/internal/sf"
)

func main() {
	analyzer := &analysis.Analyzer{
		Name: "go-stickyfields",
		Doc:  "reports all inconsistent converter functions: ensures sticky fields)",
		Run:  sf.Run,
	}

	unitchecker.Main(analyzer)
}
