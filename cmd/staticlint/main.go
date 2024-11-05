/*
Package main provides a static analysis tool that combines multiple analyzers
to perform code checks on Go source files. It leverages existing analyzers
from various sources, including the Go tools and third-party libraries.
*/
package main

import (
	"github.com/mihailtudos/metrickit/analyzer/noexitcheckanalyzer"
	"slices"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// standard is a slice of the default analyzers to be used for static checks.
var standard = []*analysis.Analyzer{
	printf.Analyzer,
	shadow.Analyzer,
	structtag.Analyzer,
	errcheck.Analyzer,
	noexitcheckanalyzer.Analyzer,
}

// listOfExtraAnalyzers holds the names of additional analyzers that can be included.
var listOfExtraAnalyzers = []string{
	"S1032",
	"S1034",
	"S1016",
	"ST1000",
	"ST1001",
}

// AnalyzersList contains a list of analyzers that will be applied when building the static checker.
type AnalyzersList struct {
	Checkers []*analysis.Analyzer // List of analyzers to be executed.
}

// main is the entry point for the application. It initializes the list of analyzers
// and runs the multichecker with the combined analyzers.
func main() {
	checks := &AnalyzersList{
		Checkers: make([]*analysis.Analyzer, 0), // Start with an empty slice
	}

	// Append standard analyzers to the checks.
	checks.Checkers = append(checks.Checkers, standard...)

	// Append analyzers from staticcheck that have names starting with "SA".
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}

	// Append additional analyzers based on the predefined list.
	appendOtherAnalyzers(checks, listOfExtraAnalyzers)

	// Run the multichecker with the combined list of analyzers.
	multichecker.Main(
		checks.Checkers...,
	)
}

// appendOtherAnalyzers adds analyzers from stylecheck and simple that match
// the names in the provided list to the AnalyzersList.
func appendOtherAnalyzers(checks *AnalyzersList, others []string) {
	// Check and append analyzers from stylecheck.
	for _, v := range stylecheck.Analyzers {
		if slices.Contains(others, v.Analyzer.Name) {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}

	// Check and append analyzers from simple.
	for _, v := range simple.Analyzers {
		if slices.Contains(others, v.Analyzer.Name) {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}
}
