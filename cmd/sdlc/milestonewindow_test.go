package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// windowRepo builds a temp git repo and chdir's into it, returning a runGit
// helper bound to the repo plus the issues dir. The issue file is committed so
// `git log -- <issuePath>` (previousReviewBoundary's scope) has something to
// follow. Restores cwd on cleanup.
func windowRepo(t *testing.T, issueNum int) (runGit func(args ...string), issuesDir, issuePath string) {
	t.Helper()
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	runGit = func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	runGit("init", "-q", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test")
	runGit("config", "commit.gpgsign", "false")

	// Initial commit so the first #N commit has a parent (branch start = firstSHA^).
	if err := os.WriteFile("README", []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README")
	runGit("commit", "-q", "-m", "init")

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
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
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

// #58: the FIRST milestone has no prior boundary, so it falls back to the branch
// start (parent of the first #N commit) — unchanged from the pre-#58 behavior,
// just no longer keyed to the Mx subject tag.
func TestBoundaryWindowBase_FirstMilestoneBasesOnBranchStart(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	// Only M1 work exists — the M1 close (with its trailer) hasn't happened yet,
	// which is the real state at the moment `milestone-close --milestone M1` runs.
	firstWork := commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")

	if got := previousReviewBoundary(issuePath); got != "" {
		t.Fatalf("previousReviewBoundary = %q, want empty (no prior boundary)", got)
	}

	base := boundaryWindowBase("58", "M1", issuePath)
	wantParent := strings.TrimSpace(captureGit(t, "rev-parse", firstWork+"^"))
	gotResolved := strings.TrimSpace(captureGit(t, "rev-parse", base))
	if gotResolved != wantParent {
		t.Fatalf("boundaryWindowBase(M1) resolved to %q, want branch start (firstWork^ = %q)", gotResolved, wantParent)
	}
}

// #58: a whole-issue close (milestone == "") always spans the branch — it never
// consults the prior boundary, so its window stays branch-start..HEAD even when
// a Review-Verdict trailer exists earlier on the branch.
func TestBoundaryWindowBase_WholeIssueIgnoresPriorBoundary(t *testing.T) {
	runGit, _, issuePath := windowRepo(t, 58)

	firstWork := commitTouchingIssue(t, runGit, issuePath, "m1work", "#58 M1: build the thing", "")
	commitTouchingIssue(t, runGit, issuePath, "m1close", "#58 M1: close",
		"Done.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD")

	base := boundaryWindowBase("58", "", issuePath) // milestone "" → whole-issue
	wantParent := strings.TrimSpace(captureGit(t, "rev-parse", firstWork+"^"))
	gotResolved := strings.TrimSpace(captureGit(t, "rev-parse", base))
	if gotResolved != wantParent {
		t.Fatalf("boundaryWindowBase(whole-issue) = %q (→ %q), want branch start %q", base, gotResolved, wantParent)
	}
}
