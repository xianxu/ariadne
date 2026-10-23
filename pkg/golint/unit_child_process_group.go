package golint

import (
	"go/ast"
	"path"
	"strings"
)

// The inner flake unit adapter inherits the managed helper's process group.
// Creating another group or nesting a managed runner makes descendants escape
// outer-owner cleanup after wrapper death (ariadne#556). Keep this lexical
// backstop scoped to the production adapter family, including future splits;
// the outer localcheck runner legitimately creates and terminates groups.
func unitChildProcessGroups(f *ast.File, context *CheckContext) []ast.Node {
	filename := path.Clean(context.Path)
	dir := path.Dir(filename)
	base := path.Base(filename)
	if (dir != "cmd/sdlc" && !strings.HasSuffix(dir, "/cmd/sdlc")) ||
		!strings.HasPrefix(base, "flake_run") || !strings.HasSuffix(base, ".go") || strings.HasSuffix(base, "_test.go") {
		return nil
	}
	groupPkg := localName(f, "github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup")
	sysPkg := localName(f, "syscall")
	groupOperation := func(name string) bool { return name == "Configure" || name == "Terminate" }
	sysOperation := func(name string) bool {
		return name == "Setpgid" || name == "Setsid" || name == "Kill"
	}
	var out []ast.Node
	selectorNames := map[*ast.Ident]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			selectorNames[x.Sel] = true
			// Match references, not just calls: assigning a forbidden function
			// to an alias must not hide it from the guard.
			bad := x.Sel.Name == "SysProcAttr"
			if pkg, ok := x.X.(*ast.Ident); ok && pkg.Obj == nil {
				bad = bad || groupPkg != "" && pkg.Name == groupPkg && groupOperation(x.Sel.Name) ||
					sysPkg != "" && pkg.Name == sysPkg && sysOperation(x.Sel.Name)
			}
			if bad {
				out = append(out, x)
			}
		case *ast.Ident:
			if selectorNames[x] {
				return true
			}
			bad := x.Name == "SysProcAttr"
			switch x.Name {
			case "runManagedCommandBatch", "runUnderSlot", "runUnderSlotBatch", "runUnderResource", "runUnderResourceBatch",
				"configureRunnerCleanup", "configureRunnerCleanupWithin", "acquireSuiteSlot":
				bad = true
			}
			if x.Obj == nil {
				bad = bad || groupPkg == "." && groupOperation(x.Name) || sysPkg == "." && sysOperation(x.Name)
			}
			if bad {
				out = append(out, x)
			}
		}
		return true
	})
	return out
}
