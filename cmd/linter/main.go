package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Rusich90/shurl.git/cmd/linter/checkers"
)

func main() {
	singlechecker.Main(checkers.NewAnalyzer())
}
