package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/amberpixels/lostfield"
)

func main() {
	unitchecker.Main(lostfield.NewAnalyzerWithFlags())
}
