package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheck = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "Check os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.ExprStmt:
				if call, ok := x.X.(*ast.CallExpr); ok {
					fn, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return false
					}

					ident, ok := fn.X.(*ast.Ident)
					if !ok {
						return false
					}

					if ident.Name == "os" && fn.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "os.Exit")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
