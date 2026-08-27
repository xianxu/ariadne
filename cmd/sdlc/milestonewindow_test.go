package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// windowRepo builds a temp git repo and chdir's into it, returning a runGit
// helper bound to the repo plus the issues dir. The issue file is committed so
// `git log -- <issuePath>` (previousReviewBoundary's scope) has something to
// follow. Restores cwd on cleanup.
func windowRepo(t *testing.T, issueNum int) (runGit func(args ...string), issuesDir, issuePath string) {
	t.Helper()
	// testfix.Repo seeds the initial commit so the first #N commit has a parent
	// (branch start = firstSHA^) and chdir's in; runGit keeps the callers'
	// bound-closure signature over the shared runner.
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	runGit = func(args ...string) { t.Helper(); testfix.Git(t, dir, args...) }

	issuesDir = "workshop/issues"
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issuePath = filepath.Join(issuesDir, fmt.Sprintf("%06d-x.md", issueNum))
	return runGit, issuesDir, issuePath
}

// commitTouchingIssue writes a marker file + a line to the issue file and commits
// with the given subject/body, so the commit is in `git log -- <issuePath>`.
func commitTouchingIssue(t *testing.T, runGit func(...string), issuePath, marker, subject, body string) string {
	t.Helper()
	if err := os.WriteFile(marker, []byte(marker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Append to the issue file so the commit touches it.
	f, err := os.OpenFile(issuePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("- " + subject + "\n")
	_ = f.Close()
	runGit("add", ".")
	if body == "" {
		runGit("commit", "-q", "-m", subject)
	} else {
		runGit("commit", "-q", "-m", subject, "-m", body)
	}
	return strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
}

// commitMarkerOnly commits a marker file WITHOUT touching the issue file —
// models an inter-milestone side-quest that references #N in its subject but
// doesn't edit the issue file.
func commitMarkerOnly(t *testing.T, runGit func(...string), marker, subject string) string {
	t.Helper()
	if err := os.WriteFile(marker, []byte(marker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit("commit", "-q", "-m", subject)
	return strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
}

func captureGit(t *testing.T, args ...string) string {
	t.Helper()
	return testfix.Capture(t, "", args...)
}

// #58: a milestone's review window must base on the PREVIOUS review boundary
// (the prior milestone close carrying a Review-Verdict: trailer), not on the
// first `#N Mx` commit — so an inter-milestone `#N`-but-not-`Mx` side-quest
// landed between M1's close and M2's first commit falls inside M2's window
// rather than slipping the net (the c52f482 shape from #57).
func TestBoundaryWindowBase_MilestoneBasesOnPriorBoundary(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	// [M1 work] → [M1 close, carries Review-Verdict] → [#58 side-quest, no Mx] → [M2 work]
	commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")
	m1Close := commitTouchingIssue(t, runGit, issuePath, "m1close", "#58 M1: close",
		"Milestone done.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD")
	sideQuest := commitMarkerOnly(t, runGit, "sidequest", "#58: dev-aliases side-quest (no milestone)")
	commitTouchingIssue(t, runGit, issuePath, "m2work", "#58 M2: build the next thing", "")

	// The previous boundary is the M1 close commit.
	if got := previousReviewBoundary(issuePath); got != m1Close {
		t.Fatalf("previousReviewBoundary = %q, want M1-close %q", got, m1Close)
	}

	// boundaryWindowBase for M2 bases on that boundary (not the first #58 M2 commit).
	base := boundaryWindowBase("58", "M2", issuePath)
	if base != m1Close {
		t.Fatalf("boundaryWindowBase(M2) = %q, want prior boundary %q", base, m1Close)
	}

	// The load-bearing assertion: the side-quest commit is INSIDE base..HEAD.
	revs := captureGit(t, "rev-list", base+"..HEAD")
	if !strings.Contains(revs, sideQuest) {
		t.Errorf("inter-milestone side-quest %s not in M2 window %s..HEAD:\n%s", sideQuest, base, revs)
	}
}

// #162: the FIRST milestone on a feature branch starts at the actual branch
// point, not the parent of the first #N commit. An issue can be filed on main
// long before its implementation branch; using its first commit would pull
// unrelated main history into the review and can make the prompt enormous.
func TestBoundaryWindowBase_FirstMilestoneBasesOnFeatureBranchPoint(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	fileIssue := commitTouchingIssue(t, runGit, issuePath, "filed", "#58: file issue", "")
	other := commitMarkerOnly(t, runGit, "other", "#99: unrelated feature")
	branchPoint := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	runGit("switch", "-c", "feature-58")
	impl := commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")

	if got := previousReviewBoundary(issuePath); got != "" {
		t.Fatalf("previousReviewBoundary = %q, want empty (no prior boundary)", got)
	}

	base := boundaryWindowBase("58", "M1", issuePath)
	gotResolved := strings.TrimSpace(captureGit(t, "rev-parse", base))
	if gotResolved != branchPoint {
		t.Fatalf("boundaryWindowBase(M1) resolved to %q, want branch point %q", gotResolved, branchPoint)
	}

	revs := captureGit(t, "rev-list", base+"..HEAD")
	if !strings.Contains(revs, impl) {
		t.Errorf("implementation commit %s missing from M1 window %s..HEAD:\n%s", impl, base, revs)
	}
	for name, sha := range map[string]string{"file-issue": fileIssue, "unrelated": other} {
		if strings.Contains(revs, sha) {
			t.Errorf("%s commit %s must not be in M1 window %s..HEAD:\n%s", name, sha, base, revs)
		}
	}
}

// #58: if a prior milestone close exists but its commit never carried a
// Review-Verdict: trailer (e.g. the operator forgot to paste it), the boundary
// lookup finds nothing and the window falls back to the branch start — so the
// next milestone OVER-covers (re-reviews the prior slice) rather than
// under-covers. Over-cover is the safe direction; this pins it.
func TestBoundaryWindowBase_MissingPriorTrailerFallsBackToBranchStart(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	firstWork := commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")
	// M1 close WITHOUT a Review-Verdict: trailer in the body.
	commitTouchingIssue(t, runGit, issuePath, "m1close", "#58 M1: close", "Milestone done, trailer forgotten.")
	commitTouchingIssue(t, runGit, issuePath, "m2work", "#58 M2: build the next thing", "")

	if got := previousReviewBoundary(issuePath); got != "" {
		t.Fatalf("previousReviewBoundary = %q, want empty (no trailer on prior close)", got)
	}

	base := boundaryWindowBase("58", "M2", issuePath)
	wantParent := strings.TrimSpace(captureGit(t, "rev-parse", firstWork+"^"))
	gotResolved := strings.TrimSpace(captureGit(t, "rev-parse", base))
	if gotResolved != wantParent {
		t.Fatalf("boundaryWindowBase(M2) = %q (→ %q), want branch start %q (over-cover fallback)", base, gotResolved, wantParent)
	}
}

// #58/#77: a whole-issue close (milestone == "") never consults the prior
// boundary. On `main` (no feature branch — merge-base == HEAD), it falls back to
// the issue's branch start, so the window stays branch-start..HEAD even when a
// Review-Verdict trailer exists earlier. This pins the #77 on-main fallback (the
// direct-on-main `sdlc push` flow); the feature-branch case is the test below.
func TestBoundaryWindowBase_WholeIssueIgnoresPriorBoundary(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	firstWork := commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")
	commitTouchingIssue(t, runGit, issuePath, "m1close", "#58 M1: close",
		"Done.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD")

	base := boundaryWindowBase("58", "", issuePath) // milestone "" → whole-issue, on main
	wantParent := strings.TrimSpace(captureGit(t, "rev-parse", firstWork+"^"))
	gotResolved := strings.TrimSpace(captureGit(t, "rev-parse", base))
	if gotResolved != wantParent {
		t.Fatalf("boundaryWindowBase(whole-issue) = %q (→ %q), want branch start %q", base, gotResolved, wantParent)
	}
}

// #77: on a FEATURE BRANCH, a whole-issue close windows on the branch point
// (merge-base with main), NOT the first `#N` commit — so an issue filed early
// (its "#N: file issue" commit) followed by unrelated work merged onto main
// before the branch does NOT pull that unrelated history into the end-of-issue
// review window. This is the over-capture #58's own review surfaced (147 commits
// across ~12 unrelated issues).
func TestBoundaryWindowBase_WholeIssueBasesOnBranchPoint(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 77)

	// Filed early, on main: the "#77: file issue" commit.
	fileIssue := commitTouchingIssue(t, runGit, issuePath, "filed", "#77: file issue", "")
	// Unrelated work from other issues lands on main afterward.
	other1 := commitMarkerOnly(t, runGit, "other1", "#99: unrelated feature")
	other2 := commitMarkerOnly(t, runGit, "other2", "#100: another unrelated fix")
	branchPoint := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD")) // == other2

	// The issue is actually implemented later, on a feature branch off main.
	runGit("switch", "-c", "feature-77")
	impl := commitTouchingIssue(t, runGit, issuePath, "impl", "#77: implement the thing", "")

	base := boundaryWindowBase("77", "", issuePath)
	if gotBase := strings.TrimSpace(captureGit(t, "rev-parse", base)); gotBase != branchPoint {
		t.Fatalf("whole-issue base = %q (→ %q), want branch point (merge-base) %q", base, gotBase, branchPoint)
	}

	// The window base..HEAD includes ONLY the branch's own work...
	revs := captureGit(t, "rev-list", base+"..HEAD")
	if !strings.Contains(revs, impl) {
		t.Errorf("implementation commit %s missing from window %s..HEAD:\n%s", impl, base, revs)
	}
	// ...and excludes the filed-early commit + the unrelated merged history.
	for name, sha := range map[string]string{"file-issue": fileIssue, "unrelated #99": other1, "unrelated #100": other2} {
		if strings.Contains(revs, sha) {
			t.Errorf("%s commit %s must NOT be in whole-issue window %s..HEAD (over-capture):\n%s", name, sha, base, revs)
		}
	}
}

// #194: the window head must be the CONCRETE SHA the review will read, not the
// literal string "HEAD". Everything downstream — the diff, the prompt, the
// Review-Window trailer, the sidecar, and the finalize check — spends this value,
// and the finalize check cannot classify a mid-review delta against a floating ref.
func TestResolveReviewWindow_HeadIsConcreteSHA(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 194)
	commitTouchingIssue(t, runGit, issuePath, "work", "#194: build the thing", "")

	_, _, head := resolveReviewWindow("194", "", issuePath)
	if head == "HEAD" {
		t.Fatal(`resolveReviewWindow head is the literal "HEAD"; it must resolve to a SHA`)
	}
	if want := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD")); head != want {
		t.Fatalf("head = %q, want rev-parse HEAD = %q", head, want)
	}
}

// #194: abbrevSHA must NOT resolve a symbolic ref. shortSHA shells out to
// `git rev-parse --short`, so shortSHA("HEAD") returns the ambient repo's HEAD —
// which in a review window would print a commit the review never read.
func TestAbbrevSHA_DoesNotResolveSymbolicRefs(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"HEAD", "HEAD"}, // the fallback: degraded, not wrong
		{"main", "main"}, // any symbolic ref survives intact
		{"e456565e922af72711492f918c92efea8adbf9bf", "e456565e"}, // full SHA truncates
		{"abc1234", "abc1234"}, // already short
	} {
		if got := abbrevSHA(tc.in); got != tc.want {
			t.Errorf("abbrevSHA(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
