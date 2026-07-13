package main

import (
	"github.com/amberpixels/lostfield"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(lostfield.NewAnalyzerWithFlags())
}
