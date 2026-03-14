package checkers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPanicChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "panic")
}

func TestExitChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "exit_outside_main")
}

func TestExitInMain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "exit_in_main")
}

func TestCleanCode(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "clean")
}
