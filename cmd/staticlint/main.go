package main

import (
	"github.com/sbilibin2017/go-yandex-practicum/cmd/staticlint/analyzers/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// main runs the staticlint tool by aggregating multiple Go analyzers.
//
// It combines analyzers from the staticcheck and simple packages,
// as well as a custom analyzer `noexit`, and executes them using multichecker.
//
// This tool performs static code analysis checks on Go code to identify
// potential issues, enforcing code quality and best practices.
func main() {
	var analyzers []*analysis.Analyzer

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	analyzers = append(analyzers, noexit.Analyzer)

	multichecker.Main(analyzers...)
}
