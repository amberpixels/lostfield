package main

import (
	"flag"

	"github.com/amberpixels/lostfield/internal/config"
	"github.com/amberpixels/lostfield/internal/lf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	var flags flag.FlagSet
	config.RegisterFlags(&flags)

	unitchecker.Main(&analysis.Analyzer{
		Name:  config.LinterName,
		Doc:   config.LinterDoc,
		Run:   lf.Run,
		Flags: flags,
	})
}
