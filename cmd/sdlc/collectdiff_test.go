package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// TestCollectDiff_Plan_FollowsArchiveRename pins #64: when a close-and-archive
// (issues/ → history/) is already committed in the judged range, the Plan
// judge's collectDiff must hand the agent the HEAD-existing history/ path (with
// the done content), never the deleted issues/ path (whose only readable
// version is the stale base). Uses a throwaway repo + the git() helper.
func TestCollectDiff_Plan_FollowsArchiveRename(t *testing.T) {
	dir := testfix.Repo(t)

	issuesDir, historyDir := "workshop/issues", "workshop/history"
	for _, d := range []string{issuesDir, historyDir} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// BASE: two working issues in issues/, unchecked plans.
	write(filepath.Join(issuesDir, "000099-foo.md"), "---\nid: 99\nstatus: working\n---\n\n## Plan\n- [ ] do the thing\n")
	write(filepath.Join(issuesDir, "000098-bar.md"), "---\nid: 98\nstatus: working\n---\n\n## Plan\n- [ ] other thing\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base: two working issues")
	base := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))

	// HEAD: archive #99 properly (done + ticked, moved to history/); archive #98
	// while STILL incomplete (status flipped but plan unchecked) — the judge must
	// still be able to see #98's unchecked state to fail it (#64 done-when #2).
	git(t, dir, "rm", "-q", filepath.Join(issuesDir, "000099-foo.md"))
	write(filepath.Join(historyDir, "000099-foo.md"), "---\nid: 99\nstatus: done\n---\n\n## Plan\n- [x] do the thing\n\n## Log\n- closed\n")
	git(t, dir, "rm", "-q", filepath.Join(issuesDir, "000098-bar.md"))
	write(filepath.Join(historyDir, "000098-bar.md"), "---\nid: 98\nstatus: done\n---\n\n## Plan\n- [ ] other thing\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "close+archive #99 (done) and #98 (incomplete)")

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	diff, changed, err := collectDiff(judge.Plan, base, "", issuesDir, historyDir)
	if err != nil {
		t.Fatal(err)
	}

	contains := func(s string) bool {
		for _, c := range changed {
			if c == s {
				return true
			}
		}
		return false
	}
	// Core #64 invariant (rename-robust): the agent is handed the HEAD-existing
	// history/ path for each archived issue — including the still-INCOMPLETE #98
	// (so the judge can read it and fail it, #64 done-when #2) — and NEVER the
	// deleted issues/ path (whose only readable version is the stale base).
	for _, want := range []string{
		filepath.Join(historyDir, "000099-foo.md"),
		filepath.Join(historyDir, "000098-bar.md"),
	} {
		if !contains(want) {
			t.Errorf("changedIssues should include archived path %s; got %v", want, changed)
		}
	}
	for _, bad := range []string{
		filepath.Join(issuesDir, "000099-foo.md"),
		filepath.Join(issuesDir, "000098-bar.md"),
	} {
		if contains(bad) {
			t.Errorf("changedIssues must NOT include deleted issues/ path %s; got %v", bad, changed)
		}
	}
	// The diff points at the archived (HEAD) content, not the stale issues/ path —
	// holds whether git records the move as a rename or delete+add.
	if !strings.Contains(diff, historyDir+"/000099-foo.md") {
		t.Errorf("diff should reference the archived history/ path (#64):\n%s", diff)
	}
}
