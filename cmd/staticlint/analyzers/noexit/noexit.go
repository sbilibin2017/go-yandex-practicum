// Package noexit provides an analyzer that forbids direct calls to os.Exit
// inside the main function of the main package.
//
// This is useful to enforce graceful shutdown and proper error handling
// instead of abrupt termination using os.Exit.
package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is a static analysis Analyzer that reports any direct calls to os.Exit
// inside the main function of the main package.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "forbids direct calls to os.Exit in function main of package main",
	Run:  run,
}

// run implements the analysis logic.
// It inspects the AST of the package and reports diagnostics if a call to os.Exit
// is found inside the main function of the main package.
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function is forbidden")
				}
				return true
			})

			return false
		})
	}
	return nil, nil
}
