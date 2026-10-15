package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func assertIssueHasNoUpstream(t *testing.T, r *trackerRepo, branch string) {
	t.Helper()
	for _, key := range []string{"remote", "merge"} {
		if gitSucceeds(r.root, "config", "--get", "branch."+branch+"."+key) {
			t.Errorf("issue branch has branch.%s.%s = %s", branch, key, r.git("config", "--get", "branch."+branch+"."+key))
		}
	}
	if got := r.git("rev-parse", "main@{upstream}"); got != r.git("rev-parse", "origin/main") {
		t.Errorf("main upstream changed: %s", got)
	}
}

func TestIssuePublicationDoesNotSetUpstream(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	r.git("config", "push.autoSetupRemote", "true")
	const branch = "000374-publish"
	r.git("switch", "--no-track", "-c", branch)
	if err := leasedBranchPush(boundaryEnv(t).git, "origin", branch); err != nil {
		t.Fatal(err)
	}
	head := r.git("rev-parse", "HEAD")
	if got := remoteTip(t, r, branch); got != head {
		t.Fatalf("published %s, want %s", got, head)
	}
	if got := r.git("rev-parse", "refs/remotes/origin/"+branch); got != head {
		t.Fatalf("tracking ref %s, want %s", got, head)
	}
	assertIssueHasNoUpstream(t, r, branch)
}

func TestIssueCreationDoesNotInheritUpstream(t *testing.T) {
	for _, mode := range []string{"always", "inherit"} {
		for _, worktree := range []bool{false, true} {
			t.Run(mode+"/"+map[bool]string{false: "in-place", true: "worktree"}[worktree], func(t *testing.T) {
				r := newTrackerRepo(t, nil, nil)
				r.git("config", "branch.autoSetupMerge", mode)
				const branch = "000374-create"
				var err error
				if worktree {
					_, err = createWorktreeBranch(context.Background(), io.Discard, io.Discard, branch, execGitRunner{}, "workshop/issues")
				} else {
					_, err = createInPlaceBranch(context.Background(), io.Discard, io.Discard, branch, execGitRunner{}, "workshop/issues")
				}
				if err != nil {
					t.Fatal(err)
				}
				if got := r.git("rev-parse", "refs/heads/"+branch); got != r.git("rev-parse", "main") {
					t.Fatal("new branch has wrong tip")
				}
				assertIssueHasNoUpstream(t, r, branch)
			})
		}
	}
}

func TestIssueReopenDoesNotSetUpstream(t *testing.T) {
	r, paths, kept := abandonedHere(t)
	r.git("config", "branch.autoSetupMerge", "always")
	r.git("config", "push.autoSetupRemote", "true")
	if out, err := reopen9(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	assertReopened(t, r, paths["000009"], kept)
	assertIssueHasNoUpstream(t, r, s09Branch)
}

func TestLegacyPRDoesNotSetIssueUpstream(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	r.git("config", "push.autoSetupRemote", "true")
	const branch = "000374-legacy"
	wt, err := createWorktreeBranch(context.Background(), io.Discard, io.Discard, branch, execGitRunner{}, "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(wt); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(r.root)
	oldGH, oldDetect, oldRunner := ghClient, detectRepo, prRunner
	ghClient = &recordingGH{}
	detectRepo = func() (string, error) { return "test/repo", nil }
	prRunner = execGitRunner{}
	t.Cleanup(func() { ghClient, detectRepo, prRunner = oldGH, oldDetect, oldRunner })
	if target, err := resolveLandingTarget(prRunner); err != nil || target != nil {
		t.Fatalf("not legacy feature workspace: %v %v", target, err)
	}
	if err := runPR(io.Discard, io.Discard, &prFlags{IssuesDir: "workshop/issues"}); err != nil {
		t.Fatal(err)
	}
	if got := remoteTip(t, r, branch); got != strings.TrimSpace(testfix.Capture(t, wt, "rev-parse", "HEAD")) {
		t.Fatalf("legacy PR published wrong head: %s", got)
	}
	assertIssueHasNoUpstream(t, r, branch)
}
