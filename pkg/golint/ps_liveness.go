package golint

import (
	"go/ast"
	"path"
	"strings"
	"unicode"
)

// ps state text is a diagnostic snapshot, not a process completion oracle:
// Darwin can show a killed descendant as ?E while it finishes exiting (#567).
// Owned-group tests use procgroup; release tests wait on the resource itself.
func psLivenessQueries(f *ast.File, _ *CheckContext) []ast.Node {
	return psLivenessCalls(f, psLivenessQuery)
}

func psLivenessCalls(f *ast.File, query func(*ast.CallExpr, string) bool) []ast.Node {
	pkg := localName(f, "os/exec")
	if pkg == "" {
		return nil
	}
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok && query(call, pkg) {
			out = append(out, call)
		}
		return true
	})
	return out
}

func psLivenessQuery(call *ast.CallExpr, pkg string) bool {
	name := ""
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok && id.Obj == nil && id.Name == pkg {
			name = fn.Sel.Name
		}
	case *ast.Ident:
		if pkg == "." && fn.Obj == nil {
			name = fn.Name
		}
	}
	offset := 0
	if name == "CommandContext" {
		offset = 1
	} else if name != "Command" {
		return false
	}
	if len(call.Args) <= offset {
		return false
	}
	bin, _ := stringLit(call.Args[offset])
	if path.Base(bin) != "ps" {
		return false
	}
	for _, arg := range call.Args[offset+1:] {
		text, _ := stringLit(arg)
		for _, prefix := range []string{"--format=", "-o", "-O"} {
			if format, ok := strings.CutPrefix(text, prefix); ok {
				text = format
				break
			}
		}
		for _, field := range strings.FieldsFunc(text, func(r rune) bool { return r == ',' || unicode.IsSpace(r) }) {
			key, _, _ := strings.Cut(field, "=")
			if key == "stat" || key == "state" {
				return true
			}
		}
	}
	return false
}
