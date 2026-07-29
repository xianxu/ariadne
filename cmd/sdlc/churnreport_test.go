package main

import (
	"math"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Three commits that rewrite the same file land 30 final lines from 90 inserted —
// rework 3.0. This is the pair#127 shape the metric exists to catch: the final diff
// alone reads as a modest 30-line change and says nothing about the two rewrites.
//
// base is captured BEFORE the first of the three commits, so all 90 insertions fall
// inside the window. (Passing the first commit's own SHA would leave only 60 in-window
// and quietly halve the ratio — the arithmetic here is the contract, not the helper.)
func TestChurnForWindowCountsRework(t *testing.T) {
	// closeRepo (closereview_test.go) builds + chdir's into a temp git repo with an
	// issue file already committed; it is the established helper for git-touching tests
	// in this package. Reuse it rather than standing up a second repo fixture (ARCH-DRY).
	closeRepo(t, 187)
	base := headSHA(t)
	commitFile(t, "cmd/x.go", strings.Repeat("a\n", 30), "#187: v1")
	commitFile(t, "cmd/x.go", strings.Repeat("b\n", 30), "#187: v2")
	commitFile(t, "cmd/x.go", strings.Repeat("c\n", 30), "#187: v3")

	r, err := churnForWindow(base)
	if err != nil {
		t.Fatal(err)
	}
	if r.Final.CodeProd != 30 {
		t.Errorf("CodeProd = %d, want 30", r.Final.CodeProd)
	}
	if r.FinalTotal != 30 {
		t.Errorf("FinalTotal = %d, want 30", r.FinalTotal)
	}
	if math.Abs(r.Rework-3.0) > 0.05 {
		t.Errorf("Rework = %.2f, want ~3.0", r.Rework)
	}
}

// The four buckets must come from the real diff, not just the unit-tested classifier —
// this is the only place the numstat→FileStat→bucket chain runs end to end over git.
func TestChurnForWindowSplitsBuckets(t *testing.T) {
	closeRepo(t, 187)
	base := headSHA(t)
	commitFile(t, "cmd/x.go", strings.Repeat("a\n", 10), "#187: prod")
	commitFile(t, "cmd/x_test.go", strings.Repeat("b\n", 20), "#187: test")
	commitFile(t, "atlas/index.md", strings.Repeat("c\n", 3), "#187: atlas")
	commitFile(t, "workshop/plans/p.md", strings.Repeat("d\n", 7), "#187: workshop")

	r, err := churnForWindow(base)
	if err != nil {
		t.Fatal(err)
	}
	if r.Final.CodeProd != 10 || r.Final.CodeTest != 20 || r.Final.Atlas != 3 || r.Final.Workshop != 7 {
		t.Errorf("buckets = %+v", r.Final)
	}
	// Nothing was rewritten, so every line written survived.
	if math.Abs(r.Rework-1.0) > 0.05 {
		t.Errorf("Rework = %.2f, want ~1.0 with no rewrites", r.Rework)
	}
}

// An unresolvable base (a docs-only window with no #N commit — boundaryWindowBase
// returns "") must degrade to a zero report, never break the close.
func TestChurnForWindowEmptyBase(t *testing.T) {
	r, err := churnForWindow("")
	if err != nil {
		t.Fatalf("empty base must not error: %v", err)
	}
	if r.FinalTotal != 0 {
		t.Errorf("want a zero report, got %+v", r)
	}
}

// A BAD base — as opposed to an absent one — must ERROR rather than report zeroes. This
// is the whole reason churnForWindow drives gitx.RunGit instead of gitx.Capture: Capture
// flattens any git failure to "", so a bogus SHA would print `churn: prod 0 / test 0 /
// …` indistinguishably from a genuinely empty window, in the one number introduced to
// answer "which gates earn their cost". The caller warns and zeroes; it can only do that
// if it is told.
func TestChurnForWindowBadBaseErrors(t *testing.T) {
	closeRepo(t, 187)
	if _, err := churnForWindow("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"); err == nil {
		t.Error("a bad base SHA must error, not degrade to a silent zero report")
	}
}

// headSHA returns the current HEAD commit SHA of the repo the test has chdir'd into.
// Capture returns raw stdout, so the trailing newline must come off before the SHA is
// handed to git as a revision.
func headSHA(t *testing.T) string {
	t.Helper()
	return strings.TrimSpace(testfix.Capture(t, ".", "rev-parse", "HEAD"))
}

// commitFile writes path (creating parents, via the package's existing mkArtifact —
// ARCH-DRY), stages it, and commits with subject. Returns the new commit's SHA.
func commitFile(t *testing.T, path, content, subject string) string {
	t.Helper()
	mkArtifact(t, path, content)
	testfix.Git(t, ".", "add", path)
	testfix.Git(t, ".", "commit", "-q", "-m", subject)
	return headSHA(t)
}
