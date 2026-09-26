package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestVerbContextsReachTrackerReads guards #252 BR-23: a verb's context must
// reach every tracker read and card write, so cancellation stops them and the
// command's records scope (one fetch per command) applies. context.Background()
// is allowed only at the root and in the named nil/absent-context fallbacks.
func TestVerbContextsReachTrackerReads(t *testing.T) {
	allowed := map[string]bool{
		"executeCLI":                      true, // the root context
		"commandContext":                  true, // nil fallback for direct callers
		"changeCodeContext":               true,
		"openTrackerAt":                   true,
		"context":                         true, // planningReviewTransaction.context
		"dispatchPlanningReview":          true,
		"runJudge":                        true,
		"runDurableMerge":                 true,
		"dispatchBoundaryReview":          true,
		"withRequiredRepoTransactionLock": true,
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range pkgs["main"].Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
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
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "context" && (sel.Sel.Name == "Background" || sel.Sel.Name == "TODO") && !allowed[fn.Name.Name] {
					t.Errorf("%s: %s uses context.%s; thread the verb's context instead", fset.Position(call.Pos()), fn.Name.Name, sel.Sel.Name)
				}
				return true
			})
		}
	}
}
