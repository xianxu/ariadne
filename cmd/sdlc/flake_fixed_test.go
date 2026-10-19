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

func TestFixedIssueCoverageUsesOwnershipNotIncidentalLogs(t *testing.T) {
	const launcher = "github.com/xianxu/pair/cmd/internal/launcher.TestCancelledHandoffReportsAChildsCleanExit"
	const own = "github.com/xianxu/pair/cmd/internal/couchcmd.TestMessageSendToDormantSlotReadmits"
	tests := []struct {
		name, body, test string
		want             bool
	}{
		{"problem", "## Problem\nTestMessageSendToDormantSlotReadmits failed in cmd/internal/couchcmd/message_service_test.go\n", own, false},
		{"legacy log label", "## Log\n- Test: `TestCancelledHandoffReportsAChildsCleanExit`, `cmd/internal/launcher/osruntime_launch_test.go`\n", launcher, false},
		{"structured log", "## Log\n- Test: `" + own + "`\n", own, true},
		{"title full name", "# " + own + "\n", own, false},
		{"prose full name", "## Spec\n`" + own + "` failed\n", own, false},
		{"plan full name", "## Plan\n- [ ] Fix `" + own + "`\n", own, false},
		{"done full name", "## Done when\n- `" + own + "` passes\n", own, false},
		{"log mention full name", "## Log\nLater suite mentioned `" + own + "`\n", own, false},
		{"log line with commentary", "## Log\n- Test: `" + own + "`, fixed with another test\n", own, false},
		{"fenced log line", "## Log\n```text\n- Test: " + own + "\n```\n", own, false},
		{"incidental log", "## Log\nThe suite later failed TestCancelledHandoffReportsAChildsCleanExit in another issue.\n", launcher, false},
		{"bare wrong package", "## Problem\nTestMessageSendToDormantSlotReadmits failed in cmd/internal/other/message_service_test.go\n", own, false},
		{"same basename in another package", "## Problem\nTestMessageSendToDormantSlotReadmits failed in pkg/couchcmd/message_service_test.go\n", own, false},
		{"same basename in dedicated log line", "## Log\n- Test: TestMessageSendToDormantSlotReadmits in pkg/couchcmd/message_service_test.go\n", own, false},
		{"nested subpackage path", "## Problem\nTestMessageSendToDormantSlotReadmits failed in cmd/internal/couchcmd/subpkg/message_service_test.go\n", own, false},
		{"bare no package", "## Problem\nTestMessageSendToDormantSlotReadmits failed\n", own, false},
		{"substring", "## Problem\nTestMessageSendToDormantSlotReadmitsAgain failed in cmd/internal/couchcmd\n", own, false},
		{"wrong package", "## Problem\n`github.com/xianxu/pair/cmd/internal/other.TestMessageSendToDormantSlotReadmits` failed\n", own, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixedIssueCovers(nil, tt.body, tt.test); got != tt.want {
				t.Fatalf("covers %q = %v, want %v", tt.test, got, tt.want)
			}
		})
	}
	if !fixedIssueCovers([]string{launcher}, "", launcher) {
		t.Fatal("exact flake_tests should cover without details")
	}
	if fixedIssueCovers([]string{launcher}, "", launcher+"Extra") {
		t.Fatal("flake_tests must match exactly")
	}
	if fixedIssueCovers(nil, "## Problem\nTestCancelledHandoffReportsAChildsCleanExit", launcher) {
		t.Fatal("a bare test name with package collision is ambiguous")
	}
}

func TestFixedFlakeDecisionIsConservative(t *testing.T) {
	const name = "github.com/xianxu/pair/pkg.TestX"
	base := fixedFlakeIssue{ID: "000585", Merge: strings.Repeat("a", 40), Tests: []string{name}, DetailKnown: true}
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
		{"two owners", []fixedFlakeIssue{base, {ID: "000589", Merge: strings.Repeat("b", 40), Tests: []string{name}, DetailKnown: true}}, false, true, flakeUnknown, "multiple terminal issues"},
		{"missing merge", []fixedFlakeIssue{{ID: "000585", Tests: []string{name}, DetailKnown: true}}, false, true, flakeUnknown, "no landing merge"},
		{"git uncertain", []fixedFlakeIssue{base}, false, false, flakeUnknown, "Git ancestry"},
		{"unreadable detail", []fixedFlakeIssue{{ID: "000589", Merge: strings.Repeat("b", 40), DetailKnown: false}, base}, false, true, flakeUnknown, "details unreadable"},
		{"unrelated legacy card with no landing", []fixedFlakeIssue{{ID: "000171", DetailKnown: false}, base}, false, true, flakeFixed, ""},
		{"legacy card explicitly claims test", []fixedFlakeIssue{{ID: "000171", Tests: []string{name}, DetailKnown: false}, base}, false, true, flakeUnknown, "details unreadable"},
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
	issues := []fixedFlakeIssue{{ID: "000585", Merge: merge, Details: "## Problem\ngithub.com/xianxu/pair/pkg.TestX failed\n", DetailKnown: true}}
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
	w := &flakeWorld{main: merge, fixed: []fixedFlakeIssue{{ID: "000448", Merge: merge, Details: "## Problem\n" + name + " failed\n", DetailKnown: true}}}
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
		t.Fatalf("prose-only issue hid the pre-fix row:\n%s", out.String())
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
	w.fixed = append(w.fixed, fixedFlakeIssue{ID: "000449", Merge: merge, Tests: []string{name}, DetailKnown: true})
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
			w := &flakeWorld{main: "main", fixed: []fixedFlakeIssue{{ID: "000448", Merge: "merge", Tests: []string{name}, DetailKnown: true}}}
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

func TestFlakeListReadsArchivedDetailsFromPublishedMain(t *testing.T) {
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
	newTrackerRepo(t, map[string]string{cardPath: string(raw)}, map[string]string{archived: detail})
	view, err := loadFlakeListView(context.Background(), "workshop/issues")
	if err != nil || view.FixedErr != nil {
		t.Fatalf("view %v, fixed evidence %v", err, view.FixedErr)
	}
	if len(view.Fixed) != 1 || !view.Fixed[0].DetailKnown || view.Fixed[0].Merge != landed || !strings.Contains(view.Fixed[0].Details, "# Seeded fixed") {
		t.Fatalf("archived details not joined from main: %+v", view.Fixed)
	}
}
