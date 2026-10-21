package golint

import (
	"go/ast"
	"go/token"
	"strings"
)

// Tests may use fake process liveness only on visibly owned temporary roots.
// Index assignments once and memoize root provenance: an alias, reassignment or
// conditional write cannot turn a runtime-selected root into a private one.
func testSweepPrivateRoot(f *ast.File, _ *CheckContext) []ast.Node {
	writes := map[*ast.Object][]ast.Expr{}
	var calls []*ast.CallExpr
	escaped := map[*ast.Object]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range x.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Obj != nil {
					var rhs ast.Expr
					if len(x.Rhs) == len(x.Lhs) {
						rhs = x.Rhs[i]
					}
					writes[id.Obj] = append(writes[id.Obj], rhs)
				}
			}
		case *ast.ValueSpec:
			for i, id := range x.Names {
				if id.Obj != nil {
					var rhs ast.Expr
					if len(x.Values) == len(x.Names) {
						rhs = x.Values[i]
					}
					writes[id.Obj] = append(writes[id.Obj], rhs)
				}
			}
		case *ast.UnaryExpr:
			if x.Op == token.AND {
				if id, ok := x.X.(*ast.Ident); ok && id.Obj != nil {
					escaped[id.Obj] = true
				}
			}
		case *ast.RangeStmt:
			for _, expr := range []ast.Expr{x.Key, x.Value} {
				if id, ok := expr.(*ast.Ident); ok && id.Obj != nil {
					writes[id.Obj] = append(writes[id.Obj], nil)
				}
			}
		case *ast.CallExpr:
			calls = append(calls, x)
		}
		return true
	})
	memo := map[*ast.Object]byte{} //1 visiting,2 unsafe,3 private
	var private func(ast.Expr) bool
	private = func(expr ast.Expr) bool {
		switch x := expr.(type) {
		case *ast.ParenExpr:
			return private(x.X)
		case *ast.Ident:
			if x.Obj == nil || escaped[x.Obj] {
				return false
			}
			if state := memo[x.Obj]; state != 0 {
				return state == 3
			}
			memo[x.Obj] = 1
			candidates := writes[x.Obj]
			safe := len(candidates) > 0
			for _, rhs := range candidates {
				if !private(rhs) {
					safe = false
				}
			}
			if safe {
				memo[x.Obj] = 3
			} else {
				memo[x.Obj] = 2
			}
			return safe
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			receiver, ok := sel.X.(*ast.Ident)
			if !ok {
				return false
			}
			if sel.Sel.Name == "TempDir" && len(x.Args) == 0 && receiver.Obj != nil {
				field, ok := receiver.Obj.Decl.(*ast.Field)
				if !ok {
					return false
				}
				typ := field.Type
				if pointer, ok := typ.(*ast.StarExpr); ok {
					typ = pointer.X
				}
				named, ok := typ.(*ast.SelectorExpr)
				if !ok {
					return false
				}
				pkg, ok := named.X.(*ast.Ident)
				return ok && pkg.Obj == nil && pkg.Name == localName(f, "testing") && (named.Sel.Name == "T" || named.Sel.Name == "B" || named.Sel.Name == "F" || named.Sel.Name == "TB")
			}
			if receiver.Obj != nil || receiver.Name != localName(f, "path/filepath") || sel.Sel.Name != "Join" || len(x.Args) == 0 || !private(x.Args[0]) {
				return false
			}
			for _, arg := range x.Args[1:] {
				part, ok := stringLit(arg)
				if !ok || strings.HasPrefix(part, "/") || strings.ContainsAny(part, "\\:") {
					return false
				}
				for _, seg := range strings.Split(part, "/") {
					if seg == ".." {
						return false
					}
				}
			}
			return true
		}
		return false
	}
	calleeMemo := map[*ast.Object]byte{}
	var sweep func(ast.Expr) bool
	sweep = func(expr ast.Expr) bool {
		if paren, ok := expr.(*ast.ParenExpr); ok {
			return sweep(paren.X)
		}
		id, ok := expr.(*ast.Ident)
		if !ok {
			return false
		}
		if id.Name == "sweepStaleTemp" && (id.Obj == nil || id.Obj.Kind == ast.Fun) {
			return true
		}
		if id.Obj == nil {
			return false
		}
		if state := calleeMemo[id.Obj]; state != 0 {
			return state == 3
		}
		calleeMemo[id.Obj] = 1
		for _, rhs := range writes[id.Obj] {
			if sweep(rhs) {
				calleeMemo[id.Obj] = 3
				return true
			}
		}
		calleeMemo[id.Obj] = 2
		return false
	}
	var out []ast.Node
	for _, call := range calls {
		if len(call.Args) > 0 && sweep(call.Fun) && !private(call.Args[0]) {
			out = append(out, call)
		}
	}
	return out
}
