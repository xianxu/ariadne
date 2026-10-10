package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestPlanReviewWindow(t *testing.T) {
	mb, start, head := strings.Repeat("1", 40), strings.Repeat("2", 40), strings.Repeat("3", 40)
	hr, s, trailer := strings.Repeat("4", 40), strings.Repeat("5", 40), strings.Repeat("6", 40)
	cases := []struct {
		name     string
		f        windowFacts
		kind     windowKind
		base     string
		human    string
		noteHave string
	}{
		{"whole issue: branch patch", windowFacts{MainBase: mb, Head: head, Reviewed: hr, Rebased: s}, windowBranchPatch, mb, mb, ""},
		{"whole issue on main: branch start", windowFacts{BranchStart: start, Head: head}, windowBranchPatch, start, start, ""},
		{"no anchor at all", windowFacts{Milestone: "M1", Head: head}, windowNone, "", "", ""},
		{"first milestone: branch patch", windowFacts{Milestone: "M1", MainBase: mb, Head: head}, windowBranchPatch, mb, mb, ""},
		{"later milestone: interdiff from S, read as H_r", windowFacts{Milestone: "M2", MainBase: mb, Head: head, Reviewed: hr, PriorBoundary: "M1", Rebased: s}, windowInterdiff, s, hr, ""},
		{"conflict: still the interdiff, labelled", windowFacts{Milestone: "M2", MainBase: mb, Head: head, Reviewed: hr, Rebased: s, Conflicted: []string{"a.go"}}, windowInterdiff, s, hr, "conflict resolutions in a.go"},
		{"unreplayable: branch patch fallback", windowFacts{Milestone: "M2", MainBase: mb, Head: head, Reviewed: hr, RebaseErr: "reviewed head x is not in this repository"}, windowBranchFallback, mb, mb, "not in this repository"},
		{"unusable ledger: named fallback", windowFacts{Milestone: "M2", MainBase: mb, Head: head, LedgerErr: "the boundary ledger is unreadable: x", LegacyTrailerBase: trailer}, windowBranchFallback, mb, mb, "unreadable"},
		{"pre-#304 ledger: trailer boundary", windowFacts{Milestone: "M2", MainBase: mb, Head: head, LegacyTrailerBase: trailer}, windowInterdiff, trailer, trailer, "pre-#304"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := planReviewWindow(c.f)
			if w.Kind != c.kind || w.Base != c.base || w.HumanBase != c.human {
				t.Fatalf("got kind=%d base=%q human=%q, want kind=%d base=%q human=%q", w.Kind, w.Base, w.HumanBase, c.kind, c.base, c.human)
			}
			if c.noteHave != "" && !strings.Contains(w.Note, c.noteHave) {
				t.Fatalf("note %q lacks %q", w.Note, c.noteHave)
			}
		})
	}
}

func TestFormatReviewWindow(t *testing.T) {
	mb, hr, s := "aaaaaaaa11", "bbbbbbbb22", "cccccccc33"
	cases := []struct {
		w    reviewWindow
		want string
	}{
		{reviewWindow{Kind: windowBranchPatch, Base: mb, HumanBase: mb, IssueCommits: 7, Files: 12},
			"branch patch vs main@aaaaaaaa: 7 issue commit(s), 12 file(s)"},
		{reviewWindow{Kind: windowInterdiff, Base: s, HumanBase: hr, MainBase: mb, PriorBoundary: "M1", Files: 3},
			"interdiff since M1 review (bbbbbbbb), rebased onto main@aaaaaaaa: 3 file(s)"},
		{reviewWindow{Kind: windowInterdiff, Base: s, HumanBase: hr, MainBase: mb, PriorBoundary: "M1", Files: 4,
			Note: "includes conflict resolutions in a.go"},
			"interdiff since M1 review (bbbbbbbb), rebased onto main@aaaaaaaa: 4 file(s) — includes conflict resolutions in a.go"},
		{reviewWindow{Kind: windowBranchFallback, Base: mb, HumanBase: mb, IssueCommits: 2, Files: 5, Note: "why"},
			"branch patch vs main@aaaaaaaa (the last review cannot be replayed): 2 issue commit(s), 5 file(s) — why"},
	}
	for _, c := range cases {
		if got := formatReviewWindow(c.w); got != c.want {
			t.Errorf("formatReviewWindow:\n got %q\nwant %q", got, c.want)
		}
	}
}

// The issue's Done-when fixture (#304): issue commits, a finalized M1 review, an
// integration of main carrying foreign changes, then more issue commits. The M2 window
// shows only M2's changes; the whole-issue window shows all of the issue's and none of
// main's — for a merge, a rebase, and a merge that resolves a conflict.
func TestMilestoneWindow_ExcludesIntegratedMain(t *testing.T) {
	for _, mode := range []string{"merge", "rebase", "conflict"} {
		t.Run(mode, func(t *testing.T) {
			runGit, _, issuePath := windowRepo(t, 304)
			plansDir := "workshop/plans"
			write := func(name, body string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			commit := func(msg string) string {
				t.Helper()
				runGit("add", "-A")
				runGit("commit", "-q", "-m", msg)
				return strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
			}
			write("shared.go", "line1\nline2\nline3\n")
			commitTouchingIssue(t, runGit, issuePath, "filed", "#304: file", "")
			runGit("switch", "-q", "-c", "issue-304")
			write("m1.go", "package m1\n")
			if mode == "conflict" {
				write("shared.go", "line1\nISSUE\nline3\n")
			}
			h1 := commit("#304 M1: work")
			// M1's review finalized at h1 (D5): the ledger is the boundary record.
			l := gatestate.Ledger{Gate: boundaryGateKind.Gate, IssueNum: 304, IDPrefix: boundaryGateKind.IDPrefix,
				Rounds: []gatestate.Round{{N: 1, Boundary: "M1", Agent: "claude", Timestamp: "t", Reviewed: h1}}}
			if err := writeBoundaryGateLedger(plansDir, filepath.Base(issuePath), l, "repo"); err != nil {
				t.Fatal(err)
			}
			commit("#304 M1: close")

			runGit("switch", "-q", "main")
			write("foreign.go", "package foreign\n")
			if mode == "conflict" {
				write("shared.go", "line1\nMAIN\nline3\n")
			}
			commit("#999: main moves on")
			runGit("switch", "-q", "issue-304")
			switch mode {
			case "merge":
				runGit("merge", "-q", "--no-edit", "main")
			case "rebase":
				runGit("rebase", "-q", "main")
			case "conflict":
				_ = exec.Command("git", "merge", "-q", "--no-edit", "main").Run() // conflicts by design
				write("shared.go", "line1\nRESOLVED\nline3\n")
				commit("#304: merge main")
			}
			write("m2.go", "package m2\n")
			commit("#304 M2: work")

			names := func(base string) []string {
				t.Helper()
				out := strings.Fields(captureGit(t, "diff", "--name-only", base, "HEAD"))
				sort.Strings(out)
				return out
			}
			m2 := planBoundaryWindow(context.Background(), "304", "M2", issuePath, plansDir)
			if m2.Kind != windowInterdiff {
				t.Fatalf("M2 window kind = %d, want interdiff (%s)", m2.Kind, formatReviewWindow(m2))
			}
			got := strings.Join(names(m2.Base), " ")
			want := "m2.go " + filepath.ToSlash(plansDir) + "/000304-x-close-gate.md"
			if mode == "conflict" {
				want = "m2.go shared.go " + filepath.ToSlash(plansDir) + "/000304-x-close-gate.md"
				if !strings.Contains(m2.Note, "shared.go") {
					t.Errorf("conflict window note %q does not name shared.go", m2.Note)
				}
				if d := captureGit(t, "diff", m2.Base, "HEAD", "--", "shared.go"); !strings.Contains(d, "+RESOLVED") || strings.Contains(d, "foreign") {
					t.Errorf("the conflict interdiff must show the resolution only:\n%s", d)
				}
			}
			if got != want {
				t.Errorf("M2 window files = %q, want %q (%s)", got, want, formatReviewWindow(m2))
			}
			whole := planBoundaryWindow(context.Background(), "304", "", issuePath, plansDir)
			for _, n := range names(whole.Base) {
				if n == "foreign.go" {
					t.Errorf("whole-issue window carries main's foreign.go (%s)", formatReviewWindow(whole))
				}
			}
			if all := strings.Join(names(whole.Base), " "); !strings.Contains(all, "m1.go") || !strings.Contains(all, "m2.go") {
				t.Errorf("whole-issue window %q must carry m1.go and m2.go", all)
			}
		})
	}
}

// #304 D11: one FRESH main. The branch merged a main newer than the stale local
// tracking ref; measuring against that stale ref would put the base behind the merge
// and pull main's foreign change into the branch patch.
func TestReviewWindow_StaleTrackingRefDoesNotWidenTheBranchPatch(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	r.git("switch", "-q", "-c", "issue-304")
	writeRepoFile(t, r.root, "mine.go", "package mine\n")
	r.git("add", "mine.go")
	r.git("commit", "-qm", "#304: mine")

	// Another slot lands foreign work on origin/main.
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", r.origin, other)
	writeRepoFile(t, other, "foreign.go", "package foreign\n")
	testfix.Git(t, other, "add", "foreign.go")
	testfix.Git(t, other, "-c", "user.name=o", "-c", "user.email=o@o", "commit", "-qm", "#999: foreign")
	testfix.Git(t, other, "push", "-q", "origin", "main")

	// This branch merges that newer main by URL, so refs/remotes/origin/main stays stale.
	r.git("fetch", "-q", r.origin, "main:refs/tmp/newmain")
	r.git("merge", "-q", "--no-edit", "refs/tmp/newmain")
	if stale, fresh := r.git("rev-parse", "origin/main"), r.git("rev-parse", "refs/tmp/newmain"); stale == fresh {
		t.Fatal("fixture: the tracking ref must be stale")
	}

	w := planBoundaryWindow(context.Background(), "304", "", "", "")
	out := captureGit(t, "diff", "--name-only", w.Base, "HEAD")
	if strings.Contains(out, "foreign.go") {
		t.Fatalf("a stale tracking ref widened the branch patch to main's foreign.go (%s):\n%s", formatReviewWindow(w), out)
	}
	if !strings.Contains(out, "mine.go") {
		t.Fatalf("the branch patch lost the issue's own change:\n%s", out)
	}
}

// #304 D10: an interdiff's diff base is a synthetic commit; the trailer names the
// reviewed head it stands for, never S.
func TestReviewTrailers_InterdiffNamesReviewedHead(t *testing.T) {
	s, hr, head := strings.Repeat("5", 40), strings.Repeat("4", 40), strings.Repeat("3", 40)
	got := strings.Join(reviewTrailers(reviewResult{Verdict: "SHIP", Base: s[:7], BaseLong: s, WindowBase: hr, Head: head}), "\n")
	if !strings.Contains(got, "Review-Window: "+hr[:8]+".."+head[:8]) || strings.Contains(got, s[:8]) {
		t.Fatalf("trailer must name the reviewed head, not S:\n%s", got)
	}
	plain := strings.Join(reviewTrailers(reviewResult{Verdict: "SHIP", Base: s[:7], BaseLong: s, Head: head}), "\n")
	if !strings.Contains(plain, "Review-Window: "+s[:8]+".."+head[:8]) {
		t.Fatalf("without a window base the trailer keeps the diff base:\n%s", plain)
	}
}

// #304 BR-2: a --no-judge milestone's binary-committed `Review-Verdict: not-run`
// evidence is NOT a review boundary. With no stamped round, the next milestone's
// window must still cover the skipped milestone's work.
func TestMilestoneWindow_NotRunEvidenceIsNotABoundary(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 304)
	commitTouchingIssue(t, runGit, issuePath, "filed", "#304: file", "")
	runGit("switch", "-q", "-c", "issue-304")
	commitTouchingIssue(t, runGit, issuePath, "m1.go", "#304 M1: close", "Review-Verdict: not-run\nReview-Window: a..b\nReview-Reason: --no-judge")
	commitMarkerOnly(t, runGit, "m2.go", "#304 M2: work")

	w := planBoundaryWindow(context.Background(), "304", "M2", issuePath, "workshop/plans")
	if w.Kind != windowBranchPatch {
		t.Fatalf("a not-run trailer became a boundary: %s", formatReviewWindow(w))
	}
	if out := captureGit(t, "diff", "--name-only", w.Base, "HEAD"); !strings.Contains(out, "m1.go") {
		t.Fatalf("the skipped M1's work escaped M2's window:\n%s", out)
	}
	// A finalizing trailer (pre-#304 history) still is one.
	commitTouchingIssue(t, runGit, issuePath, "m2b.go", "#304 M2: close", "Review-Verdict: SHIP\nReview-Window: a..b")
	commitMarkerOnly(t, runGit, "m3.go", "#304 M3: work")
	if w := planBoundaryWindow(context.Background(), "304", "M3", issuePath, "workshop/plans"); w.Kind != windowInterdiff || !strings.Contains(w.Note, "pre-#304") {
		t.Fatalf("a SHIP trailer must still bound a pre-#304 window: %s", formatReviewWindow(w))
	}
}

// #304 BR-3: a criss-cross history has no single branch point. The window must SAY so
// and over-cover, never silently diff from a guessed base under an "interdiff" label.
func TestReviewWindow_CrissCrossIsANamedFallback(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 304)
	commitTouchingIssue(t, runGit, issuePath, "filed", "#304: file", "")
	runGit("switch", "-q", "-c", "issue-304")
	commitMarkerOnly(t, runGit, "b.go", "#304: b")
	runGit("switch", "-q", "main")
	commitMarkerOnly(t, runGit, "a.go", "#999: a")
	a := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	runGit("merge", "-q", "--no-edit", "issue-304~0")
	runGit("switch", "-q", "issue-304")
	runGit("merge", "-q", "--no-edit", a)
	runGit("switch", "-q", "main")
	commitMarkerOnly(t, runGit, "a2.go", "#999: a2")
	runGit("switch", "-q", "issue-304")
	commitMarkerOnly(t, runGit, "b2.go", "#304: b2")
	if bases := strings.Fields(captureGit(t, "merge-base", "--all", "main", "HEAD")); len(bases) < 2 {
		t.Skipf("fixture did not produce a criss-cross (%v)", bases)
	}
	for _, m := range []string{"", "M2"} {
		w := planBoundaryWindow(context.Background(), "304", m, issuePath, "workshop/plans")
		if w.Kind != windowBranchFallback || !strings.Contains(w.Note, "criss-cross") {
			t.Fatalf("milestone %q: want a named criss-cross fallback, got %s", m, formatReviewWindow(w))
		}
	}
}

// #304 BR-5: a hand-edited ledger value is never handed to git. An option-like
// `reviewed:` reads as unusable, and the window over-covers to the branch patch.
func TestMilestoneWindow_RejectsANonSHAReviewedValue(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 304)
	commitTouchingIssue(t, runGit, issuePath, "filed", "#304: file", "")
	runGit("switch", "-q", "-c", "issue-304")
	commitMarkerOnly(t, runGit, "m1.go", "#304 M1: work")
	l := gatestate.Ledger{Gate: boundaryGateKind.Gate, IssueNum: 304, IDPrefix: boundaryGateKind.IDPrefix,
		Rounds: []gatestate.Round{{N: 1, Boundary: "M1", Agent: "claude", Timestamp: "t", Reviewed: "--output=/tmp/x"}}}
	if err := writeBoundaryGateLedger("workshop/plans", filepath.Base(issuePath), l, "repo"); err != nil {
		t.Fatal(err)
	}
	w := planBoundaryWindow(context.Background(), "304", "M2", issuePath, "workshop/plans")
	if w.Kind != windowBranchFallback || !strings.Contains(w.Note, "not a commit id") {
		t.Fatalf("a non-SHA reviewed value must be a NAMED fallback, not a pre-#304 boundary: %s", formatReviewWindow(w))
	}
}

// #304 BR-4: a tracker repository whose main cannot be fetched measures against main
// as last fetched, and the window SAYS so.
func TestReviewWindow_NamesAStaleMainWhenTheFetchFails(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	r.git("switch", "-q", "-c", "issue-304")
	writeRepoFile(t, r.root, "mine.go", "package mine\n")
	r.git("add", "mine.go")
	r.git("commit", "-qm", "#304: mine")
	r.git("remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	w := planBoundaryWindow(context.Background(), "304", "", "", "")
	if !strings.Contains(w.Note, "last fetched") {
		t.Fatalf("a failed fetch must be named in the window: %s", formatReviewWindow(w))
	}
}
