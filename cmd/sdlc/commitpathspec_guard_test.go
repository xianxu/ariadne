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
// whole-tree, with the reason. Membership is the acknowledgment — but it is NOT
// the licence: an entry only takes effect for a function that demonstrably
// stages the whole tree (`git add -A`), which the scan verifies. The first cut
// of this guard keyed the exemption on the function alone, which excused every
// commit in `runPush` on the strength of its unrelated `commit -a` — including
// the archive commit that was one of the seven sites this guard exists for.
type wholeIndexCommitExemption struct {
	file, fn, why string
}

// wholeIndexCommits is the complete allowlist. `runPush`'s `commit -a` needs no
// entry: an argv carrying `-a` says whole-tree in the argv itself, which is the
// rule below rather than an exception to it.
var wholeIndexCommits = []wholeIndexCommitExemption{
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
// The rule it encodes is not "every commit needs a pathspec" but the honest one:
// A COMMIT MUST BE AS NARROW AS ITS ADD. So a commit argv is accepted when it
//
//  1. carries a `--` pathspec separator (the narrow case), or
//  2. carries `-a` (whole-tree, stated in the argv itself), or
//  3. sits in a function that stages `git add -A` AND is listed in
//     wholeIndexCommits with its reason — both halves required, so an
//     allowlist entry can neither go stale nor widen to cover a sibling
//     commit in the same function.
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
				stagesWholeTree := functionStagesWholeTree(fn)
				ast.Inspect(fn, func(n ast.Node) bool {
					lits, isGitArgv := commitArgvLiterals(n)
					if !isGitArgv || !hasLiteral(lits, "commit") {
						return true
					}
					key := base + ":" + fn.Name.Name
					if hasLiteral(lits, "--") || hasLiteral(lits, "-a") {
						return true
					}
					if why, ok := exempt[key]; ok && stagesWholeTree {
						used[key] = true
						t.Logf("%s: whole-index commit allowed — %s", key, why)
						return true
					}
					t.Errorf("%s (%s): git commit argv %v has no `--` pathspec separator.\n"+
						"  A bare commit records the WHOLE INDEX, so a peer agent's staged work is\n"+
						"  swept into it (#206). A commit must be as narrow as its add — commit the\n"+
						"  same paths you staged:\n"+
						"      append([]string{\"commit\", \"-m\", msg, \"--\"}, pathspec...)\n"+
						"  Deliberately whole-tree? Say so in the argv with `-a`, or pair it with\n"+
						"  `git add -A` in the same function and add %s to wholeIndexCommits.",
						key, fset.Position(n.Pos()), lits, key)
					return true
				})
			}
		}
	}
	for _, e := range wholeIndexCommits {
		if !used[e.file+":"+e.fn] {
			t.Errorf("stale exemption %s:%s (%s) — no whole-index commit found there "+
				"(or the function no longer stages `git add -A`); drop the entry",
				e.file, e.fn, e.why)
		}
	}
}

// functionStagesWholeTree reports whether fn builds a `git add -A` argv, which
// is what makes a whole-index commit inside it legitimate rather than a bug.
// This is the half of the exemption the allowlist cannot assert for itself.
func functionStagesWholeTree(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn, func(n ast.Node) bool {
		lits, isGitArgv := commitArgvLiterals(n)
		if isGitArgv && hasLiteral(lits, "add") && hasLiteral(lits, "-A") {
			found = true
		}
		return !found
	})
	return found
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
		case "Git", "GitInDir", "RunGit", "Capture":
			// gitx.RunGit / gitx.Capture are the package's other live git seams
			// (~12 read sites); an inline gitx.RunGit("commit", …) would
			// otherwise escape this guard entirely.
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

// wiring is one "this entry point must call this helper" edge.
type wiring struct {
	file, entry, helper, why string
}

// commitWirings are the edges no behavioral test in this package can reach.
//
// The rule three review rounds converged on: a fix at a CALL SITE is pinned only
// by a test entering through the production entry point — a real-git test that
// hand-builds the argv proves the helper and mocks the wiring. Where the entry
// point is not in-process drivable, that leaves the call site unpinned, and
// deleting it keeps the suite green. runChangeCode is the standing example
// (#191: exitWithCode bypasses the die seam), and runPush / runMerge / runMigrate
// each die() through gates a unit test cannot satisfy.
//
// So the wiring is asserted at the source instead. This is weaker than driving
// the verb — it proves the call exists, not that it runs — but it is strictly
// stronger than nothing, and it fails the moment someone deletes the line.
var commitWirings = []wiring{
	{"changecode.go", "runChangeCode", "syncIssue",
		"Spec piece 3 (#206): change-code lands + publishes the design at the end of planning"},
	{"push.go", "runPush", "archiveCommitArgs",
		"the archive commit must be as narrow as archiveAddArgs staged"},
	{"push.go", "recoverInterruptedArchive", "archiveCommitArgs",
		"same commit, resumed after an interrupted archive"},
	{"merge.go", "runMerge", "archiveCommitArgs",
		"the archive commit in the MAIN worktree, where merge's dirty check never looked"},
	{"migrate.go", "runMigrate", "migrateCommitArgs",
		"both migrate commits and both --no-commit hints share one argv builder"},
	{"issue.go", "runIssueNew", "syncIssuesToMain",
		"the #82 M1 reservation broadcast, with #206's local-commit fallback"},
	{"issue.go", "runIssueSync", "syncIssuesToMain",
		"the verb is a thin exposure of the shared dispatch, not a second sync path"},
	{"startplan.go", "runStartPlan", "syncPointer",
		"the mid-planning durability trigger's only delivery point (#206 BR-4)"},
}

// TestVerbsWireTheirCommitHelpers asserts every edge in commitWirings, and that
// none has gone stale. Deleting `syncIssue(stderr, f, issuePath)` from
// runChangeCode used to leave the whole package green; it now fails here.
func TestVerbsWireTheirCommitHelpers(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse cmd/sdlc: %v", err)
	}
	// calls[file:func] = set of function names it calls.
	calls := map[string]map[string]bool{}
	for _, pkg := range pkgs {
		for path, file := range pkg.Files {
			base := filepath.Base(path)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				key := base + ":" + fn.Name.Name
				calls[key] = map[string]bool{}
				ast.Inspect(fn, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					switch f := call.Fun.(type) {
					case *ast.Ident:
						calls[key][f.Name] = true
					case *ast.SelectorExpr:
						calls[key][f.Sel.Name] = true
					}
					return true
				})
			}
		}
	}
	for _, w := range commitWirings {
		key := w.file + ":" + w.entry
		called, ok := calls[key]
		if !ok {
			t.Errorf("stale wiring: %s does not exist — update commitWirings", key)
			continue
		}
		if !called[w.helper] {
			t.Errorf("%s no longer calls %s.\n  %s\n"+
				"  This edge has no behavioral coverage (the entry point is not in-process\n"+
				"  drivable), so deleting the call would otherwise leave the suite green.",
				key, w.helper, w.why)
		}
	}
}
