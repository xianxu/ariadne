package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// wholeIndexCommitExemption names a function whose `git commit` is DELIBERATELY
// whole-tree, with the reason. Membership is the acknowledgment: a commit that
// records everything is a real thing to want, but it has to be said out loud.
type wholeIndexCommitExemption struct {
	file, fn, why string
}

// wholeIndexCommits is the complete allowlist. Both entries pair their commit
// with a whole-tree ADD, which is exactly what makes them legitimate — the rule
// below is "a commit must be as narrow as its add", not "every commit needs a
// pathspec".
var wholeIndexCommits = []wholeIndexCommitExemption{
	{"push.go", "runPush", "`commit -a`: the ship verb auto-commits every tracked change before publishing"},
	{"propagatebase.go", "commitConsumption", "paired with `git add -A`: consuming a base-layer change is a whole-tree event"},
}

// TestGitCommitsCarryTheirPathspec is #206's CLASS guard, and the reason it is a
// source scan rather than another behavioral test.
//
// The bug — narrow the `git add`, then commit the whole index, sweeping a peer
// agent's staged work into a commit that misdescribes it — was found at one site
// and turned out to live at seven. Each round of review found another one, and
// each was fixed as an instance. Behavioral tests pin the sites they drive;
// nothing stopped the EIGHTH site from being written tomorrow. This does: a new
// `git commit` argv built after a narrowed add fails here at once, in the same
// shape as this tree's other drift guards (TestRepoLockCommandMetadata,
// TestForceAckMatchesGateCatalog).
//
// The rule: every git-commit argv in cmd/sdlc must carry a `--` pathspec
// separator, unless its function is in wholeIndexCommits above with a reason.
func TestGitCommitsCarryTheirPathspec(t *testing.T) {
	exempt := map[string]string{}
	for _, e := range wholeIndexCommits {
		exempt[e.file+":"+e.fn] = e.why
	}
	used := map[string]bool{}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse cmd/sdlc: %v", err)
	}
	for _, pkg := range pkgs {
		for path, file := range pkg.Files {
			base := filepath.Base(path)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				ast.Inspect(fn, func(n ast.Node) bool {
					lits, isGitArgv := commitArgvLiterals(n)
					if !isGitArgv || !hasLiteral(lits, "commit") {
						return true
					}
					key := base + ":" + fn.Name.Name
					if why, ok := exempt[key]; ok {
						used[key] = true
						t.Logf("%s: whole-index commit allowed — %s", key, why)
						return true
					}
					if !hasLiteral(lits, "--") {
						t.Errorf("%s (%s): git commit argv %v has no `--` pathspec separator.\n"+
							"  A bare commit records the WHOLE INDEX, so a peer agent's staged work is\n"+
							"  swept into it (#206). Commit the same paths the add staged:\n"+
							"      append([]string{\"commit\", \"-m\", msg, \"--\"}, pathspec...)\n"+
							"  If the commit is deliberately whole-tree, add %s to wholeIndexCommits\n"+
							"  with the reason.",
							key, fset.Position(n.Pos()), lits, key)
					}
					return true
				})
			}
		}
	}
	for _, e := range wholeIndexCommits {
		if !used[e.file+":"+e.fn] {
			t.Errorf("stale exemption %s:%s (%s) — no whole-index commit found there; drop the entry",
				e.file, e.fn, e.why)
		}
	}
}

// commitArgvLiterals returns the string literals of n when n builds a git argv:
// a *.Git / *.GitInDir / gitInDir call, an exec.Command("git", …), or a []string
// composite literal (the `append([]string{"commit", …, "--"}, paths...)` shape).
// Anything else reports false, so a stray "commit" string elsewhere is not a hit.
func commitArgvLiterals(n ast.Node) ([]string, bool) {
	switch e := n.(type) {
	case *ast.CallExpr:
		if !isGitArgvCall(e) {
			return nil, false
		}
		return stringLiterals(e.Args), true
	case *ast.CompositeLit:
		arr, ok := e.Type.(*ast.ArrayType)
		if !ok {
			return nil, false
		}
		if id, ok := arr.Elt.(*ast.Ident); !ok || id.Name != "string" {
			return nil, false
		}
		return stringLiterals(e.Elts), true
	}
	return nil, false
}

func isGitArgvCall(e *ast.CallExpr) bool {
	switch fn := e.Fun.(type) {
	case *ast.Ident:
		return fn.Name == "gitInDir"
	case *ast.SelectorExpr:
		switch fn.Sel.Name {
		case "Git", "GitInDir":
			return true
		case "Command": // exec.Command("git", …)
			lits := stringLiterals(e.Args)
			return len(lits) > 0 && lits[0] == "git"
		}
	}
	return false
}

func stringLiterals(exprs []ast.Expr) []string {
	var out []string
	for _, a := range exprs {
		bl, ok := a.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			continue
		}
		if s, err := strconv.Unquote(bl.Value); err == nil {
			out = append(out, s)
		}
	}
	return out
}

func hasLiteral(lits []string, want string) bool {
	for _, l := range lits {
		if l == want {
			return true
		}
	}
	return false
}
