package checkers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPanicChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "panic.go")
}

func TestExitChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "exit_outside_main.go")
}

func TestExitInMain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "exit_in_main.go")
}

func TestCleanCode(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "clean.go")
}