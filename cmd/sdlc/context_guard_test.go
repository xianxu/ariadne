package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerbContextsReachTrackerReads guards #252 BR-23: a verb's context must
// reach every tracker read and card write, so cancellation stops them and the
// command's records scope (one fetch per command) applies. context.Background()
// is allowed only at the root and in the named nil/absent-context fallbacks —
// in package main and in every internal package, whose exported entry points
// take the caller's context rather than minting their own.
func TestVerbContextsReachTrackerReads(t *testing.T) {
	allowed := map[string]bool{
		"main.executeCLI":                      true, // the root context
		"main.commandContext":                  true, // nil fallback for direct callers
		"main.changeCodeContext":               true,
		"main.openTrackerAt":                   true,
		"main.context":                         true, // planningReviewTransaction.context
		"main.dispatchPlanningReview":          true,
		"main.runJudge":                        true,
		"main.runDurableMerge":                 true,
		"main.dispatchBoundaryReview":          true,
		"main.withRequiredRepoTransactionLock": true,
		"gitx.runGitIn":                        true, // the context-free legacy runner
		"gitx.operationContext":                true, // nil fallback of a context-free TrunkFile
	}
	dirs := []string{"."}
	internal, err := filepath.Glob("internal/*")
	if err != nil {
		t.Fatal(err)
	}
	dirs = append(dirs, internal...)
	fset := token.NewFileSet()
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
			return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatal(err)
		}
		for name, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || fn.Body == nil {
						continue
					}
					qualified := name + "." + fn.Name.Name
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}
						sel, ok := call.Fun.(*ast.SelectorExpr)
						if !ok {
							return true
						}
						if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "context" && (sel.Sel.Name == "Background" || sel.Sel.Name == "TODO") && !allowed[qualified] {
							t.Errorf("%s: %s uses context.%s; thread the caller's context instead", fset.Position(call.Pos()), qualified, sel.Sel.Name)
						}
						return true
					})
				}
			}
		}
	}
}
