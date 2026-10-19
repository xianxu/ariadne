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
		{"problem", "## Problem\nTestMessageSendToDormantSlotReadmits failed\n", own, true},
		{"legacy log label", "## Log\n- Test: `TestCancelledHandoffReportsAChildsCleanExit`\n", launcher, true},
		{"incidental log", "## Log\nThe suite later failed TestCancelledHandoffReportsAChildsCleanExit in another issue.\n", launcher, false},
		{"substring", "## Problem\nTestMessageSendToDormantSlotReadmitsAgain failed\n", own, false},
		{"wrong package", "## Problem\n`github.com/xianxu/pair/cmd/internal/other.TestMessageSendToDormantSlotReadmits` failed\n", own, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixedIssueCovers(nil, tt.body, tt.test, false); got != tt.want {
				t.Fatalf("covers %q = %v, want %v", tt.test, got, tt.want)
			}
		})
	}
	if !fixedIssueCovers([]string{launcher}, "", launcher, false) {
		t.Fatal("exact flake_tests should cover without details")
	}
	if fixedIssueCovers([]string{launcher}, "", launcher+"Extra", false) {
		t.Fatal("flake_tests must match exactly")
	}
	if fixedIssueCovers(nil, "## Problem\nTestCancelledHandoffReportsAChildsCleanExit", launcher, true) {
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
	}{
		{"pre fix", []fixedFlakeIssue{base}, false, true, flakeFixed},
		{"recurrence", []fixedFlakeIssue{base}, true, true, flakeRecurred},
		{"two owners", []fixedFlakeIssue{base, {ID: "000589", Merge: strings.Repeat("b", 40), Tests: []string{name}, DetailKnown: true}}, false, true, flakeUnknown},
		{"missing merge", []fixedFlakeIssue{{ID: "000585", Tests: []string{name}, DetailKnown: true}}, false, true, flakeUnknown},
		{"git uncertain", []fixedFlakeIssue{base}, false, false, flakeUnknown},
		{"unreadable detail", []fixedFlakeIssue{{ID: "000589", Merge: strings.Repeat("b", 40), DetailKnown: false}, base}, false, true, flakeUnknown},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := classifyFixedFlake(name, tt.issues, false, func(_, descendant string) (bool, error) {
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
	issues := []fixedFlakeIssue{{ID: "000585", Merge: merge, Details: "## Problem\nTestX failed\n", DetailKnown: true}}
	contains := func(ancestor, head string) (bool, error) {
		return flakeGitContains(context.Background(), root, ancestor, head)
	}
	if got, _ := classifyFixedFlake(test, issues, false, contains, merge, before); got != flakeFixed {
		t.Fatalf("before merge: %v", got)
	}
	if got, _ := classifyFixedFlake(test, issues, false, contains, merge, merge); got != flakeRecurred {
		t.Fatalf("after merge: %v", got)
	}
	issues[0].Merge = strings.Repeat("f", 40)
	if got, _ := classifyFixedFlake(test, issues, false, contains, merge, before); got != flakeUnknown {
		t.Fatalf("missing merge: %v", got)
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
