package main

import (
	"flag"

	"github.com/amberpixels/go-stickyfields/internal/config"
	"github.com/amberpixels/go-stickyfields/internal/sf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	var flags flag.FlagSet
	config.RegisterFlags(&flags)

	unitchecker.Main(&analysis.Analyzer{
		Name:  config.LinterName,
		Doc:   config.LinterDoc,
		Run:   sf.Run,
		Flags: flags,
	})
}
