package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// osExitAnalyzer reports direct calls to os.Exit from main functions in main packages.
var osExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "reports direct calls to os.Exit from func main in package main",
	Run:  runOSExit,
}

func runOSExit(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Exit" {
					return true
				}
				fn, ok := pass.TypesInfo.Uses[selector.Sel].(*types.Func)
				if ok && fn.Pkg() != nil && fn.Pkg().Path() == "os" && fn.Name() == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit is forbidden in func main; return an error and handle it outside main")
				}
				return true
			})
		}
	}
	return nil, nil
}
