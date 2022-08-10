// Package analyzer search call os.Exit() in main packages main.go and report position
// Implement analysis.Analyzer type interface for multi-check
package analyzer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitanalyzer",
	Doc:  "Don't allow os.Exit in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, value := range pass.Files {
		if value.Name.Name == "main" {
			ast.Inspect(value, func(n ast.Node) bool {
				if expr, ok := n.(*ast.CallExpr); ok {
					if fun, ok := expr.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := fun.X.(*ast.Ident); ok {
							if (ident.Name == "os") && (fun.Sel.Name == "Exit") {
								pass.Reportf(fun.Pos(), "expression has os.Exit call in main package")
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
