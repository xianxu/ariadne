package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestFixedIssueCoverageUsesCardOnly(t *testing.T) {
	const launcher = "github.com/xianxu/pair/cmd/internal/launcher.TestCancelledHandoffReportsAChildsCleanExit"
	if !fixedIssueCovers([]string{launcher}, launcher) {
		t.Fatal("exact flake_tests should cover without details")
	}
	if fixedIssueCovers([]string{launcher}, launcher+"Extra") {
		t.Fatal("flake_tests must match exactly")
	}
	if fixedIssueCovers(nil, launcher) {
		t.Fatal("no card ownership must stay visible")
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
