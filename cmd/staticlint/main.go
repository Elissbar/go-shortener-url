package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/tomarrell/wrapcheck/wrapcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		stylecheck.Analyzers[0].Analyzer,
		// публичные анализаторы
		ineffassign.Analyzer,
		wrapcheck.Analyzer,
		// кастомный анализатор
		ExitCheck,
	)

	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	multichecker.Main(
		analyzers...,
	)
}
