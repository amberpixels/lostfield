package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/amberpixels/go-stickyfields/internal/config"
	"github.com/amberpixels/go-stickyfields/internal/sf"
)

func main() {
	unitchecker.Main(&analysis.Analyzer{
		Name: config.LinterName,
		Doc:  config.LinterDoc,
		Run:  sf.Run,
	})
}
