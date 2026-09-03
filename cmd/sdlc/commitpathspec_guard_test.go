package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
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
	assertWiring(t, commitWirings)
}

// assertWiring checks each entry-point → helper edge by parsing the source. One
// implementation for the commit-pathspec, plan-reader and plan-writer guards
// (ARCH-DRY): three copies of an AST walk is how the third one drifts.
func assertWiring(t *testing.T, edges []wiring) {
	t.Helper()
	fset := token.NewFileSet()
	// calls[file:func] = set of function names it calls.
	calls := map[string]map[string]bool{}
	for _, dir := range guardScanDirs {
		pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", dir, err)
		}
		for _, pkg := range pkgs {
			for path, file := range pkg.Files {
				// Directory-qualified: keying on the basename alone would merge
				// two same-named functions in different packages into one call
				// set, and the guard would report an edge satisfied by the wrong
				// one. No basename collides today; this is why it can't start to.
				base := qualifiedFile(dir, path)
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
	}
	for _, w := range edges {
		key := w.file + ":" + w.entry
		called, ok := calls[key]
		if !ok {
			t.Errorf("stale wiring: %s does not exist — update the edge list naming it", key)
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

// guardScanDirs are the packages the source-level guards parse. Hard-coded, and
// TestGuardScopeCoversEveryPackage keeps the list honest by failing when a
// package outside it references a plan-counting regex — a population defined by
// a list is exactly what these guards exist to distrust.
var guardScanDirs = []string{".", "internal/issue"}

// planItemMatchers are the regexes that COUNT plan items. Any function using one
// is by definition a plan-item reader and must take its body from
// PlanItemsBody — the filtered section — or it silently disagrees with the
// others about the same Plan (#211 BR-4).
//
// Deriving the reader set from these, rather than listing readers by hand, is
// the #208 rule: a hand-maintained restatement of the model stops covering the
// sixth member the day someone adds it, and says nothing while it does.
var planItemMatchers = []string{
	"PlanUncheckedRE", "PlanItemRE", "nonEmptyPlanItemRE", "milestonePlanRE", "milestoneLabelRE",
}

// planItemReaderExemptions holds a function that uses one of the counting
// regexes above but must NOT read the filtered body, with its reason.
// Membership is the acknowledgment; the guard also refuses a STALE entry.
//
// TickMilestone, the obvious other candidate, needs no entry: it is a writer
// that builds its pattern inline rather than referencing a named matcher, so the
// derivation never classes it as a reader. An entry added for it was rejected by
// the stale check on exactly that ground.
var planItemReaderExemptions = map[string]string{
	"close.go:milestonesInPlanOrder": "a PURE helper that RECEIVES an already-filtered plan body. " +
		"Exempt here and covered by planItemBodySources instead: extracting the regex into this " +
		"helper moved milestonePlanRE out of its caller, which silently dropped that caller from " +
		"this derivation — so the edge is named explicitly rather than assumed",
}

// planItemBodySources are the functions that OBTAIN a plan body to hand to a
// pure helper. The derivation above finds regex users; it cannot find a caller
// that fetches the body and delegates the counting, and extracting a helper is
// exactly what turns the second kind into the first.
//
// Found by probing my own exemption (#211 close review): I wrote that
// milestonesInPlanOrder was safe because "its caller is itself checked by this
// guard", then reverted that caller and nothing fired. The rationale was false.
var planItemBodySources = []wiring{
	{"close.go", "findMilestonesMissingVerdict", "PlanItemsBody",
		"obtains the plan body for milestonesInPlanOrder; the raw section would " +
			"surface milestones quoted inside a fenced example and demand review evidence for them"},
}

// TestPlanItemBodySourcesUsePlanItemsBody covers the delegating callers.
func TestPlanItemBodySourcesUsePlanItemsBody(t *testing.T) {
	assertWiring(t, planItemBodySources)
}

// planItemWriters are the call sites that REWRITE plan rows. The reader guard
// below has a writer sibling for the same reason it exists at all: reverting
// close.go's tick to the old whole-body ReplaceAll left the suite green, because
// the behavioural test drives issue.TickMilestone directly and never traverses
// computeClose (close review BR-21(b)).
var planItemWriters = []wiring{
	{"close.go", "computeClose", "TickMilestone",
		"the milestone tick must go through the scoped, fence-filtered writer — " +
			"inline rewriting is what ticked quoted rows anywhere in the document"},
}

// TestPlanItemWritersUseTickMilestone is the writer half of the routing guard.
func TestPlanItemWritersUseTickMilestone(t *testing.T) {
	assertWiring(t, planItemWriters)
}

// TestPlanItemReadersUsePlanItemsBody pins the ROUTING, which no behavioral test
// in this tree can reach.
//
// #211's BR-10 found the gap and BR-16 found my false claim about having closed
// it: I reverted the HELPER (removing the filter inside PlanItemsBody) and
// watched a test go red, then reported that as evidence the routing was pinned.
// It was not — reverting the four CALL SITES to PlanSectionBody left the whole
// suite green, because TestPlanItemReadersAgree calls PlanItemsBody directly and
// re-implements close's guard rather than driving it.
//
// That is the same rule #206 landed for the commit-pathspec class: a fix at a
// call site is pinned only by a test entering the production entry point, and
// where the entry point is not in-process drivable (computeClose die()s past the
// test seam) a source-level guard must assert the wiring.
func TestPlanItemReadersUsePlanItemsBody(t *testing.T) {
	fset := token.NewFileSet()
	seen, counters := map[string]bool{}, map[string]string{}
	for _, dir := range guardScanDirs {
		pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", dir, err)
		}
		for _, pkg := range pkgs {
			for path, file := range pkg.Files {
				base := qualifiedFile(dir, path) // see assertWiring on why
				for _, decl := range file.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}
					key := base + ":" + fn.Name.Name
					ast.Inspect(fn, func(n ast.Node) bool {
						// A function referencing a counting regex IS a plan-item
						// reader — derived, not listed.
						if id, ok := n.(*ast.Ident); ok {
							for _, m := range planItemMatchers {
								if id.Name == m {
									counters[key] = m
								}
							}
						}
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}
						name := ""
						switch f := call.Fun.(type) {
						case *ast.Ident:
							name = f.Name
						case *ast.SelectorExpr:
							name = f.Sel.Name
						}
						if name == "PlanItemsBody" {
							seen[key] = true
						}
						return true
					})
				}
			}
		}
	}
	if len(counters) == 0 {
		t.Fatal("no function references any plan-item counting regex — planItemMatchers is stale, " +
			"so this guard covers nothing")
	}
	for key, matcher := range counters {
		if seen[key] {
			continue
		}
		if why, ok := planItemReaderExemptions[key]; ok {
			t.Logf("%s: exempt — %s", key, why)
			continue
		}
		t.Errorf("%s uses %s but never calls PlanItemsBody.\n"+
			"  PlanSectionBody returns the RAW section, so a `- [ ]` inside a quoted\n"+
			"  example counts as open work and this reader silently disagrees with the\n"+
			"  others about the same Plan (#211 BR-4). If this site legitimately needs\n"+
			"  the unfiltered body, add it to planItemReaderExemptions with the reason.",
			key, matcher)
	}
	for key, why := range planItemReaderExemptions {
		if _, ok := counters[key]; !ok {
			t.Errorf("stale exemption %s (%s) — it no longer uses a counting regex; drop it", key, why)
		}
	}
}

// qualifiedFile renders a parsed file's key as "<dir>/<base>" for a nested
// package and plain "<base>" at the package root, so guards can name an edge as
// "close.go:computeClose" or "internal/issue/plan.go:CountPlanItems" without two
// packages' same-named files colliding.
func qualifiedFile(dir, path string) string {
	base := filepath.Base(path)
	if dir == "." {
		return base
	}
	return dir + "/" + base
}

// TestGuardScopeCoversEveryPackage pins the guards' own blind spot.
//
// The derived plan-item guards and assertWiring scan a hard-coded list of
// directories ("." and "internal/issue"). A reader added in a THIRD package is
// invisible to the derivation — it would reference a counting regex, never call
// PlanItemsBody, and the guard would say nothing. That is the same failure the
// guards exist to prevent, one level up: a population defined by a list rather
// than by the tree.
//
// Fully deriving the scope would mean walking every package under cmd/sdlc on
// every run. Instead this asserts the LIST is complete for the property that
// matters: no package outside guardScanDirs references a plan-counting regex.
// When one does, this fails and names it, and the fix is one entry.
func TestGuardScopeCoversEveryPackage(t *testing.T) {
	scanned := map[string]bool{}
	for _, d := range guardScanDirs {
		scanned[filepath.Clean(d)] = true
	}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Clean(filepath.Dir(path))
		if scanned[dir] {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		for _, m := range planItemMatchers {
			if strings.Contains(string(src), m) {
				t.Errorf("%s references %s but lives in %q, which the plan-item guards do not scan "+
					"(guardScanDirs = %v).\n  Add %q to guardScanDirs, or the derivation is blind to "+
					"every reader in that package.", path, m, dir, guardScanDirs, dir)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
