package golint

import (
	"go/ast"
	"path"
	"strings"
)

// This tripwire owns the live-owner resume cleanup implementation, not proof
// GC or pre-link setup. One literal callback may contain only one raw effect:
// allowing two would silently lose the authority check between them (#463).
// It is intentionally syntactic, not interprocedural: new cleanup helpers in
// these owners must retain a resume cleanup/retire/release/remove/unlink name.
func resumeCleanupEffects(f *ast.File, context *CheckContext) []ast.Node {
	filename := path.Clean(context.Path)
	owned := false
	for _, owner := range []string{"cmd/sdlc/localcheck_resume.go", "cmd/sdlc/localcheck_resume_environment.go"} {
		owned = owned || filename == owner || strings.HasSuffix(filename, "/"+owner)
	}
	runner := filename == "cmd/sdlc/localcheck_run.go" || strings.HasSuffix(filename, "/cmd/sdlc/localcheck_run.go")
	if !owned && !runner {
		return nil
	}
	type cleanupScope struct{ body, bindings *ast.BlockStmt }
	var scopes []cleanupScope
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if owned && resumeCleanupFunction(fn.Name.Name) {
			scopes = append(scopes, cleanupScope{fn.Body, fn.Body})
		}
		if runner {
			// The strict branch of the deferred scratch cleanup owns resume
			// evidence. Ordinary runner cleanup and pre-launch protection do not.
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				deferred, ok := n.(*ast.DeferStmt)
				if !ok {
					return true
				}
				callback, ok := deferred.Call.Fun.(*ast.FuncLit)
				if !ok {
					return true
				}
				ast.Inspect(callback.Body, func(n ast.Node) bool {
					branch, ok := n.(*ast.IfStmt)
					if !ok {
						return true
					}
					cond, ok := branch.Cond.(*ast.SelectorExpr)
					if ok && cond.Sel.Name == "StrictGroupCleanup" {
						scopes = append(scopes, cleanupScope{branch.Body, fn.Body})
					}
					return true
				})
				return false
			})
		}
	}
	var out []ast.Node
	for _, scope := range scopes {
		// Resolve local function aliases conservatively: any assignment from a
		// raw effect makes every call through that variable an effect as well.
		writes := map[*ast.Object][]ast.Expr{}
		ast.Inspect(scope.bindings, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				if len(x.Lhs) == len(x.Rhs) {
					for i, lhs := range x.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && id.Obj != nil {
							writes[id.Obj] = append(writes[id.Obj], x.Rhs[i])
						}
					}
				}
			case *ast.ValueSpec:
				if len(x.Names) == len(x.Values) {
					for i, id := range x.Names {
						writes[id.Obj] = append(writes[id.Obj], x.Values[i])
					}
				}
			}
			return true
		})
		var effect func(ast.Expr, []ast.Expr, map[*ast.Object]bool) bool
		effect = func(expr ast.Expr, args []ast.Expr, seen map[*ast.Object]bool) bool {
			switch x := expr.(type) {
			case *ast.ParenExpr:
				return effect(x.X, args, seen)
			case *ast.Ident:
				if x.Name == "sync" || x.Name == "syncResumeDirectory" && (x.Obj == nil || x.Obj.Kind == ast.Fun) {
					return true
				}
				if x.Name == "resumeEnvironmentGit" && (x.Obj == nil || x.Obj.Kind == ast.Fun) {
					return resumeWorktreeMutation(args)
				}
				if x.Obj != nil && !seen[x.Obj] {
					seen[x.Obj] = true
					for _, rhs := range writes[x.Obj] {
						if effect(rhs, args, seen) {
							return true
						}
					}
				}
			case *ast.SelectorExpr:
				switch x.Sel.Name {
				case "removeAll", "remove", "sync", "Cleanup":
					return true // injected operation receivers have no static type
				case "Remove", "RemoveAll":
					id, ok := x.X.(*ast.Ident)
					return ok && id.Obj == nil && id.Name == localName(f, "os")
				case "Git", "CleanupGit":
					return resumeWorktreeMutation(args)
				}
			}
			return false
		}
		isEffect := func(call *ast.CallExpr) bool {
			return effect(call.Fun, call.Args, map[*ast.Object]bool{})
		}
		allowed := map[*ast.CallExpr]bool{}
		ast.Inspect(scope.body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || id.Name != "resumeCleanupEffect" || id.Obj != nil && id.Obj.Kind != ast.Fun {
				return true
			}
			var callback *ast.FuncLit
			if len(call.Args) == 3 {
				callback, _ = call.Args[2].(*ast.FuncLit)
			}
			if callback == nil {
				out = append(out, call)
				return true
			}
			var effects []*ast.CallExpr
			deferred := map[*ast.CallExpr]bool{}
			ast.Inspect(callback.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.DeferStmt:
					deferred[x.Call] = true
				case *ast.GoStmt:
					deferred[x.Call] = true
				}
				if _, nested := n.(*ast.FuncLit); nested {
					return false // deferred/asynchronous effects lack the gate
				}
				if inner, ok := n.(*ast.CallExpr); ok && isEffect(inner) {
					effects = append(effects, inner)
				}
				return true
			})
			if len(effects) == 1 && !deferred[effects[0]] {
				allowed[effects[0]] = true
			}
			return true
		})
		ast.Inspect(scope.body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok && isEffect(call) && !allowed[call] {
				out = append(out, call)
			}
			return true
		})
	}
	return out
}

func resumeCleanupFunction(name string) bool {
	name = strings.ToLower(name)
	if !strings.Contains(name, "resume") {
		return false
	}
	for _, operation := range []string{"cleanup", "retire", "release", "remove", "unlink"} {
		if strings.Contains(name, operation) {
			return true
		}
	}
	return strings.HasPrefix(name, "confirmresume") || name == "finishresumescratch"
}

func resumeWorktreeMutation(args []ast.Expr) bool {
	for i, arg := range args {
		if command, ok := stringLit(arg); ok && command == "worktree" {
			if i+1 == len(args) {
				return true
			}
			subcommand, known := stringLit(args[i+1])
			return !known || subcommand != "list"
		}
	}
	return false
}
