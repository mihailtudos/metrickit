// Package noexitcheckanalyzer implements a static analyzer that detects usage of os.Exit
// within the main function. This is useful for identifying potentially harmful
// direct process termination in the program's entry point.
//
// The analyzer reports an error when it finds any direct calls to os.Exit within
// the main function. It optimizes the analysis by first checking for the presence
// of the "os" package import before performing detailed AST traversal.
package noexitcheckanalyzer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is an analysis.Analyzer that checks for os.Exit calls within the main function.
// It requires the inspect analyzer to traverse the AST.
//
// # Analyzer Name: exitinmain
//
// # Analyzer Reports:
//   - When os.Exit is called directly within the main function
//
// # Example problematic code:
//
//	package main
//	import "os"
//	func main() {
//	    os.Exit(1) // This will be reported
//	}
//
// # Example valid code:
//
//	package main
//	import "os"
//	func main() {
//	    // Use defer, return, or error handling instead
//	    if err := someFunc(); err != nil {
//	        return
//	    }
//	}
var Analyzer = &analysis.Analyzer{
	Name: "exitinmain",
	Doc:  "checks for os.Exit calls within main function",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

// hasOSImport checks if the given file imports the "os" package.
// It returns true if the file has an import statement for "os",
// false otherwise.
//
// This function is used as an optimization to skip AST traversal
// for files that couldn't possibly contain os.Exit calls.
func hasOSImport(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path.Value == `"os"` {
			return true
		}
	}
	return false
}

// run implements the analysis logic for the exitinmain analyzer.
// It traverses the AST of each file that imports "os" and looks
// for calls to os.Exit within the main function.
//
// The analysis is performed in two steps:
//  1. Check if the file imports the "os" package
//  2. If "os" is imported, inspect the main function for os.Exit calls
//
// Parameters:
//   - pass: The analysis pass context containing the files to analyze
//
// Returns:
//   - interface{}: Always returns nil as this analyzer doesn't produce results for other analyzers
//   - error: Returns an error if the analysis fails, nil otherwise
func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Iterate through each file in the package
	for _, file := range pass.Files {
		// First, check if any of the files have "os" import
		if !hasOSImport(file) {
			continue
		}

		// Only proceed with AST inspection if "os" is imported
		nodeFilter := []ast.Node{
			(*ast.FuncDecl)(nil),
		}

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return
			}

			// Only check the main function
			if funcDecl.Name.Name != "main" {
				return
			}

			// Walk the function body looking for os.Exit calls
			ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
				if callExpr, ok := node.(*ast.CallExpr); ok {
					if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := selExpr.X.(*ast.Ident); ok {
							if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
								pass.Reportf(callExpr.Pos(), "os.Exit called within main function")
							}
						}
					}
				}
				return true
			})
		})
	}

	return nil, nil
}
