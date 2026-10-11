package activetime

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// #252: the tracker's card commits are read beside HEAD, and a commit
// reachable from both is one event.
func TestLoadWindowCommitsReadsExtraRefsOnce(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")
	run("commit", "-q", "--allow-empty", "-m", "#9 code")
	run("switch", "-q", "--orphan", "tracker")
	run("commit", "-q", "--allow-empty", "-m", "#9: tracker: update card")
	run("switch", "-q", "main")
	orig := resolveCommitWorkspace
	resolveCommitWorkspace = func(r string) (workspace.Identity, error) { return workspace.Identity{Repo: filepath.Base(r)}, nil }
	t.Cleanup(func() { resolveCommitWorkspace = orig })

	head, err := loadWindowCommits(repo, "", "", Scope{})
	if err != nil || len(head) != 1 {
		t.Fatalf("HEAD only: %v %v", head, err)
	}
	both, err := loadWindowCommits(repo, "", "", Scope{}, "refs/heads/tracker", "refs/heads/main")
	if err != nil || len(both) != 2 {
		t.Fatalf("with the tracker ref (and HEAD twice): %+v %v", both, err)
	}
	for _, c := range both {
		if len(c.Issues) != 1 || c.Issues[0] != "9" {
			t.Errorf("commit %q not attributed to #9: %v", c.Subject, c.Issues)
		}
	}
}

// #321: the tracker interleaves every slot's card writes. Off the issue
// branch (unscoped: a resting branch, or main after landing) another issue's
// card write two seconds before a run claimed the whole run, and #317
// measured 0. A tracker-only commit is the measured issue's activity only
// when its lead names that issue, in every scope.
func TestTrackerCommitsOfOtherIssuesDoNotClaim(t *testing.T) {
	repo := gitInit(t)
	branchCommit(t, repo, "2025-12-31T23:00:00+00:00", "base", "base")
	gitAt(t, repo, "", "branch", "-M", "main")
	gitAt(t, repo, "", "switch", "-q", "--orphan", "tracker")
	branchCommit(t, repo, "2026-01-01T00:00:00+00:00", "card9", "#9: tracker: update card")
	branchCommit(t, repo, "2026-01-01T00:09:58+00:00", "card5", "#5: tracker: update card")
	branchCommit(t, repo, "2026-01-01T01:00:00+00:00", "card9", "#9: tracker: done")
	gitAt(t, repo, "", "switch", "-q", "main")
	const since, until = "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z"
	scope := Scope{Issue: "9"} // unscoped: HEAD has not diverged from main

	commits, err := loadWindowCommits(repo, since, until, scope, "refs/heads/tracker")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"00:00 #9: tracker: update card", "01:00 #9: tracker: done"}
	if got := boundaryShape(commits); !reflect.DeepEqual(got, want) {
		t.Errorf("boundaries = %v, want %v: #5's card write is another slot's", got, want)
	}
	// One run 00:10–00:50 (events every 10 min); #5's card is 2 s before it.
	var lines []string
	for _, m := range []string{"10", "20", "30", "40", "50"} {
		lines = append(lines, `{"timestamp":"2026-01-01T00:`+m+`:00Z","type":"user","message":{"content":"x"}}`)
	}
	res, err := Compute(Options{
		Dirs: []string{eventsDir(t, lines...)}, GitRepo: repo, ExtraRefs: []string{"refs/heads/tracker"},
		Scope: scope, SinceISO: since, UntilISO: until, Issues: []string{"5", "9"},
		CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !approx(res.PerIssue["9"], 40) || res.PerIssue["5"] != 0 {
		t.Errorf("per issue = %v, want #9 = 40 min and nothing for #5", res.PerIssue)
	}
}
