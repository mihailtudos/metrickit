package main

import (
	"fmt"
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

var standard = []*analysis.Analyzer{
	printf.Analyzer,
	shadow.Analyzer,
	structtag.Analyzer,
	errcheck.Analyzer,
}

var listOfExtraAnalyzers = []string{
	"S1032",
	"S1034",
	"S1016",
	"ST1000",
	"ST1001",
}

// AnalyzersList contains a list of Analyzer that will be applied when building the staticchecker
type AnalyzersList struct {
Checkers []*analysis.Analyzer
}

func main() {
	checks := &AnalyzersList{
		Checkers: make([]*analysis.Analyzer, len(standard)),
	}

	checks.Checkers = append(checks.Checkers, standard...)

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}

	appendOtherAnalyzers(checks, listOfExtraAnalyzers)

	fmt.Println(len(checks.Checkers))
	multichecker.Main(
		checks.Checkers...,
	)
}

func appendOtherAnalyzers(checks *AnalyzersList, others []string) {
	for _, v := range stylecheck.Analyzers {
		if slices.Contains(others, v.Analyzer.Name) {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}

	for _, v := range simple.Analyzers {
		if slices.Contains(others, v.Analyzer.Name) {
			checks.Checkers = append(checks.Checkers, v.Analyzer)
		}
	}
}
