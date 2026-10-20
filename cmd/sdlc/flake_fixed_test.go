package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// #479: a fixed issue covers a failing leaf by the gate's rule, flakes.Covers,
// nearest name first as Registry.Cover; two issues at that name stay ambiguous.
func TestFixedIssueCoverageFollowsTheGateRule(t *testing.T) {
	const x = "github.com/xianxu/pair/pkg.TestX"
	issue := func(id string, tests ...string) fixedFlakeIssue { return fixedFlakeIssue{ID: id, Tests: tests} }
	for _, c := range []struct {
		name   string
		issues []fixedFlakeIssue
		leaf   string
		want   []string
	}{
		{"exact top-level", []fixedFlakeIssue{issue("1", x)}, x, []string{"1"}},
		{"name prefix is not a test", []fixedFlakeIssue{issue("1", x)}, x + "Extra", nil},
		{"no card ownership", []fixedFlakeIssue{issue("1")}, x, nil},
		{"exact subtest", []fixedFlakeIssue{issue("1", x+"/sub")}, x + "/sub", []string{"1"}},
		{"subtest covers its own subtests", []fixedFlakeIssue{issue("1", x+"/sub")}, x + "/sub/deeper", []string{"1"}},
		{"subtest never covers a sibling", []fixedFlakeIssue{issue("1", x+"/sub")}, x + "/sib", nil},
		{"subtest never covers its parent", []fixedFlakeIssue{issue("1", x+"/sub")}, x, nil},
		{"top-level covers its subtests", []fixedFlakeIssue{issue("1", x)}, x + "/sub", []string{"1"}},
		{"nearest name wins", []fixedFlakeIssue{issue("1", x), issue("2", x+"/sub")}, x + "/sub", []string{"2"}},
		{"nearest name wins in either order", []fixedFlakeIssue{issue("2", x+"/sub"), issue("1", x)}, x + "/sub/deeper", []string{"2"}},
		{"same name on two issues is ambiguous", []fixedFlakeIssue{issue("1", x+"/sub"), issue("2", x+"/sub")}, x + "/sub", []string{"1", "2"}},
		{"one issue naming both counts once", []fixedFlakeIssue{issue("1", x, x+"/sub")}, x + "/sub", []string{"1"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, i := range fixedCandidates(c.issues, c.leaf) {
				got = append(got, i.ID)
			}
			if !slices.Equal(got, c.want) {
				t.Fatalf("candidates %v, want %v", got, c.want)
			}
		})
	}
}

// #479: flake list judges a fixed issue's subtest like an open one's: the
// logged leaf it names is hidden as fixed, a sibling's stays listed by name.
func TestFlakeListJudgesFixedSubtestsFromTheLog(t *testing.T) {
	for name, c := range map[string]struct {
		logged string
		listed bool
	}{
		"the fixed subtest failed": {flakeSub, false},
		"a sibling failed":         {flakeT1 + "/plain/other.test.sh", true},
	} {
		t.Run(name, func(t *testing.T) {
			w := &flakeWorld{
				attempts:    []SuiteAttempt{attemptAt("f1", "abc", 1, 1, flakeT1), attemptAt("p1", "abc", 2, 0)},
				logSubtests: map[string][]string{"f1": {flakeT1 + "/plain", c.logged}},
				main:        "main",
				fixed:       []fixedFlakeIssue{{ID: "000465", Merge: "merge", Tests: []string{flakeSub}}},
				contains: func(_ context.Context, _, head string) (bool, error) {
					return head == "main", nil
				},
			}
			var out bytes.Buffer
			if err := runFlakeListMode(context.Background(), &out, w.deps(), true); err != nil {
				t.Fatal(err)
			}
			section := strings.SplitN(strings.SplitN(out.String(), "same-head fail→pass, not registered", 2)[1], "same-head flakes fixed", 2)[0]
			if got := strings.Contains(section, c.logged); got != c.listed {
				t.Fatalf("listed as unregistered %v, want %v:\n%s", got, c.listed, out.String())
			}
			if fixed := strings.Contains(out.String(), "  "+flakeSub+"  attempt f1") && strings.Contains(out.String(), "fixed by #000465"); fixed == c.listed {
				t.Fatalf("fixed row %v, want %v:\n%s", fixed, !c.listed, out.String())
			}
		})
	}
}

// #479: flakes.Covers is the one coverage predicate. A parallel *Covers
// matcher in package main is how a fixed issue once matched by exact name.
func TestNoParallelFlakeCoverageMatcher(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pkgs {
		for file, f := range p.Files {
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasSuffix(fn.Name.Name, "Covers") {
					t.Errorf("%s: %s re-implements flake coverage; call flakes.Covers", file, fn.Name.Name)
				}
			}
		}
	}
}

func TestFixedFlakeDecisionIsConservative(t *testing.T) {
	const name = "github.com/xianxu/pair/pkg.TestX"
	base := fixedFlakeIssue{ID: "000585", Merge: strings.Repeat("a", 40), Tests: []string{name}}
	for _, tt := range []struct {
		name   string
		issues []fixedFlakeIssue
		inHead bool
		gitOK  bool
		want   fixedFlakeState
		why    string
	}{
		{"pre fix", []fixedFlakeIssue{base}, false, true, flakeFixed, ""},
		{"recurrence", []fixedFlakeIssue{base}, true, true, flakeRecurred, ""},
		{"two owners", []fixedFlakeIssue{base, {ID: "000589", Merge: strings.Repeat("b", 40), Tests: []string{name}}}, false, true, flakeUnknown, "multiple terminal issues"},
		{"missing merge", []fixedFlakeIssue{{ID: "000585", Tests: []string{name}}}, false, true, flakeUnknown, "no landing merge"},
		{"git uncertain", []fixedFlakeIssue{base}, false, false, flakeUnknown, "Git ancestry"},
		{"unrelated legacy card with no landing", []fixedFlakeIssue{{ID: "000171"}, base}, false, true, flakeFixed, ""},
		{"legacy card explicitly claims test", []fixedFlakeIssue{{ID: "000171", Tests: []string{name}}, base}, false, true, flakeUnknown, "multiple terminal issues"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, _, reason := classifyFixedFlake(name, tt.issues, func(_, descendant string) (bool, error) {
				if !tt.gitOK {
					return false, errFixedGitUnknown
				}
				if descendant == "main" {
					return true, nil
				}
				return tt.inHead, nil
			}, "main", "head")
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if tt.why != "" && !strings.Contains(reason, tt.why) {
				t.Fatalf("reason %q, want %q", reason, tt.why)
			}
		})
	}
}

func TestFixedFlakeAgainstRealGitMergeHistory(t *testing.T) {
	root := t.TempDir()
	git(t, "", "init", "-q", "-b", "main", root)
	git(t, root, "config", "user.name", "Test")
	git(t, root, "config", "user.email", "test@example.com")
	file := filepath.Join(root, "fixture")
	if err := os.WriteFile(file, []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", "fixture")
	git(t, root, "commit", "-qm", "base")
	before := git(t, root, "rev-parse", "HEAD")
	git(t, root, "switch", "-qc", "fix")
	if err := os.WriteFile(file, []byte("fix"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "commit", "-qam", "fix")
	git(t, root, "switch", "-q", "main")
	if err := os.WriteFile(file, []byte("main"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "commit", "-qam", "main moves")
	git(t, root, "merge", "--no-ff", "-qm", "land fix", "-X", "ours", "fix")
	merge := git(t, root, "rev-parse", "HEAD")
	const test = "github.com/xianxu/pair/pkg.TestX"
	issues := []fixedFlakeIssue{{ID: "000585", Merge: merge}}
	contains := func(ancestor, head string) (bool, error) {
		return flakeGitContains(context.Background(), root, ancestor, head)
	}
	if got, _, _ := classifyFixedFlake(test, issues, contains, merge, before); got != flakeUnknown {
		t.Fatalf("prose-only owner hid pre-merge row: %v", got)
	}
	issues[0].Tests = []string{test}
	if got, _, _ := classifyFixedFlake(test, issues, contains, merge, before); got != flakeFixed {
		t.Fatalf("before merge: %v", got)
	}
	if got, _, _ := classifyFixedFlake(test, issues, contains, merge, merge); got != flakeRecurred {
		t.Fatalf("after merge: %v", got)
	}
	issues[0].Merge = strings.Repeat("f", 40)
	if got, _, _ := classifyFixedFlake(test, issues, contains, merge, before); got != flakeUnknown {
		t.Fatalf("missing merge: %v", got)
	}
}

func TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit(t *testing.T) {
	root := t.TempDir()
	git(t, "", "init", "-q", "-b", "main", root)
	git(t, root, "config", "user.name", "Test")
	git(t, root, "config", "user.email", "test@example.com")
	file := filepath.Join(root, "fixture")
	if err := os.WriteFile(file, []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", "fixture")
	git(t, root, "commit", "-qm", "base")
	before := git(t, root, "rev-parse", "HEAD")
	git(t, root, "switch", "-qc", "fix")
	if err := os.WriteFile(file, []byte("fix"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "commit", "-qam", "fix")
	git(t, root, "switch", "-q", "main")
	if err := os.WriteFile(file, []byte("main"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "commit", "-qam", "main moves")
	git(t, root, "merge", "--no-ff", "-qm", "land fix", "-X", "ours", "fix")
	merge := git(t, root, "rev-parse", "HEAD")
	const name = "github.com/xianxu/ariadne/cmd/sdlc.TestOther"
	w := &flakeWorld{main: merge, fixed: []fixedFlakeIssue{{ID: "000448", Merge: merge}}}
	w.contains = func(ctx context.Context, ancestor, head string) (bool, error) {
		return flakeGitContains(ctx, root, ancestor, head)
	}
	w.attempts = []SuiteAttempt{
		attemptAt("before", before, 1, 1, name), attemptAt("before-pass", before, 2, 0),
		attemptAt("after", merge, 3, 1, name), attemptAt("after-pass", merge, 4, 0),
	}
	var out bytes.Buffer
	if err := runFlakeListMode(context.Background(), &out, w.deps(), false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "attempt before on") {
		t.Fatalf("card without flake_tests hid the pre-fix row:\n%s", out.String())
	}
	w.fixed[0].Tests = []string{name}
	out.Reset()
	if err := runFlakeListMode(context.Background(), &out, w.deps(), false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "attempt before on") || !strings.Contains(out.String(), "attempt after on") || !strings.Contains(out.String(), "recurred after #000448") {
		t.Fatalf("default list failed fixed/recurrence distinction:\n%s", out.String())
	}
	out.Reset()
	if err := runFlakeListMode(context.Background(), &out, w.deps(), true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "attempt before on") || !strings.Contains(out.String(), "fixed by #000448 (merge "+merge[:8]+")") {
		t.Fatalf("verbose list omitted fixed row:\n%s", out.String())
	}
	w.fixed = append(w.fixed, fixedFlakeIssue{ID: "000449", Merge: merge, Tests: []string{name}})
	out.Reset()
	if err := runFlakeListMode(context.Background(), &out, w.deps(), false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "attempt before on") || !strings.Contains(out.String(), "coverage unknown: multiple terminal issues") {
		t.Fatalf("ambiguous coverage was hidden or unexplained:\n%s", out.String())
	}
}

func TestFlakeListCachesAncestryAcrossRepeatedAttempts(t *testing.T) {
	for _, n := range []int{4, 8} {
		t.Run(string(rune('0'+n)), func(t *testing.T) {
			const name = "github.com/xianxu/ariadne/cmd/sdlc.TestOther"
			w := &flakeWorld{main: "main", fixed: []fixedFlakeIssue{{ID: "000448", Merge: "merge", Tests: []string{name}}}}
			for i := 1; i <= n; i++ {
				w.attempts = append(w.attempts, attemptAt(sixDigits(i), "head", i, 1, name))
			}
			w.attempts = append(w.attempts, attemptAt("pass", "head", n+1, 0))
			calls := 0
			w.contains = func(_ context.Context, _, head string) (bool, error) { calls++; return head == "main", nil }
			var out bytes.Buffer
			if err := runFlakeListMode(context.Background(), &out, w.deps(), true); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Fatalf("%d attempts caused %d Git calls, want 2", n, calls)
			}
			if strings.Count(out.String(), "fixed by #000448") != n {
				t.Fatalf("lost fixed entries:\n%s", out.String())
			}
		})
	}
}

func TestFlakeListIgnoresPublishedLogOwnership(t *testing.T) {
	const name = "github.com/xianxu/pair/cmd/internal/launcher.TestCancelledHandoffReportsAChildsCleanExit"
	for _, tt := range []struct{ label, log string }{
		{"plain", "- Test: " + name},
		{"html", "<pre>\n- Test: " + name + "\n</pre>"},
		{"unmatched backtick", "`\n- Test: " + name},
	} {
		t.Run(tt.label, func(t *testing.T) {
			cardPath, card, detailPath, detail := seededIssue(t, "000585", "fixed")
			card = strings.Replace(card, "status: open", "status: done\nactual_hours: 1", 1)
			landed := strings.Repeat("a", 40)
			raw, err := issue.SetCardCompletion([]byte(card), issue.Completion{
				Token: "closed-585", Repository: "fixture", ReviewedHEAD: landed, EvidenceCommit: landed, LandedCommit: landed,
			})
			if err != nil {
				t.Fatal(err)
			}
			archived := strings.Replace(detailPath, "workshop/issues/", "workshop/history/issues/", 1)
			newTrackerRepo(t, map[string]string{cardPath: string(raw)}, map[string]string{archived: detail + "\n## Log\n" + tt.log + "\n"})
			view, err := loadFlakeListView(context.Background(), "workshop/issues")
			if err != nil || view.FixedErr != nil {
				t.Fatalf("view %v, fixed evidence %v", err, view.FixedErr)
			}
			if len(view.Fixed) != 1 || view.Fixed[0].Merge != landed || len(view.Fixed[0].Tests) != 0 {
				t.Fatalf("Log text became ownership: %+v", view.Fixed)
			}
			got, _, _ := classifyFixedFlake(name, view.Fixed, func(_, _ string) (bool, error) {
				t.Fatal("no card owner should bypass Git ancestry")
				return false, nil
			}, view.Main, "before")
			if got != flakeUnknown {
				t.Fatalf("Log text hid the failure: %v", got)
			}
		})
	}
}
