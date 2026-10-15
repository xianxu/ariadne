package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLandingDefersBlockedConfigCleanup(t *testing.T) {
	roots, remote, gh := landingFixture(t, 1)
	root := roots[1]
	common := git(t, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	lock := filepath.Join(common, "config.lock")
	if err := os.WriteFile(lock, []byte("another config writer\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	if err := runMerge(io.Discard, &stderr, landingFlags()); err != nil {
		t.Fatalf("landed issue must not fail on config lock: %v\n%s", err, &stderr)
	}
	if !strings.Contains(stderr.String(), "deferred") || !strings.Contains(stderr.String(), "config") {
		t.Fatalf("missing deferred cleanup diagnostic: %s", &stderr)
	}
	if got := git(t, root, "for-each-ref", "--format=%(objectname)", "refs/heads/"+landingTestBranch); got != "" {
		t.Fatal("issue ref retained", got)
	}
	archived := git(t, remote, "rev-parse", "main")
	f := landingFlags()
	f.Branch = landingTestBranch
	if err := runMerge(io.Discard, &stderr, f); err != nil {
		t.Fatal("persistent denial must remain nonblocking", err)
	}
	if gh.merges != 1 || git(t, remote, "rev-parse", "main") != archived {
		t.Fatal("retry repeated landing or archive")
	}
	if got := git(t, root, "config", "--get", "branch."+landingTestBranch+".remote"); got != "upstream" {
		t.Fatal("pending metadata not retained", got)
	}
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	if err := runMerge(io.Discard, &stderr, f); err != nil {
		t.Fatal("writable retry", err)
	}
	if out, err := landingOptional(mergeRunner, root, "config", "--get-regexp", `^branch\.`+landingTestBranch+`\.`); err != nil || out != "" {
		t.Fatal("metadata not cleaned", out, err)
	}
}

func TestReconcileRetriesObsoleteBranchConfig(t *testing.T) {
	r, _, _ := externalMergeFixture(t, "closed")
	branch := r.git("branch", "--show-current")
	// This models tracking metadata left by a pre-feature or manual push.
	r.git("config", "branch."+branch+".remote", "origin")
	r.git("config", "branch."+branch+".merge", "refs/heads/"+branch)
	r.git("config", "branch.unrelated.remote", "keep-me")
	var out, stderr bytes.Buffer
	reconcile := func() {
		t.Helper()
		if err := runRecoveryReconcile(context.Background(), &out, &stderr, 330); err != nil {
			t.Fatalf("reconcile: %v\n%s", err, &stderr)
		}
	}
	reconcile()
	if got := r.git("config", "branch."+branch+".remote"); got != "origin" {
		t.Fatal("live branch metadata removed", got)
	}
	r.git("switch", "-q", "main")
	head := r.git("rev-parse", "refs/heads/"+branch)
	r.git("update-ref", "-d", "refs/heads/"+branch, head)
	lock := filepath.Join(r.root, ".git", "config.lock")
	if err := os.WriteFile(lock, []byte("blocked\n"), 0600); err != nil {
		t.Fatal(err)
	}
	reconcile()
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	// A recreated branch is live even though its card is still done.
	r.git("update-ref", "refs/heads/"+branch, head)
	reconcile()
	if got := r.git("config", "branch."+branch+".remote"); got != "origin" {
		t.Fatal("recreated branch metadata removed", got)
	}
	r.git("update-ref", "-d", "refs/heads/"+branch, head)
	reconcile()
	if out, err := landingOptional(execGitRunner{}, r.root, "config", "--get-regexp", `^branch\.`+branch+`\.`); err != nil || out != "" {
		t.Fatal("next reconcile did not remove pending metadata", out, err)
	}
	if got := r.git("config", "branch.unrelated.remote"); got != "keep-me" {
		t.Fatal("unrelated config changed", got)
	}
}

func TestObsoleteConfigGuards(t *testing.T) {
	for _, mode := range []string{"occupied", "ref-present", "ref-probe-fails", "worktree-probe-fails", "manual-recreation"} {
		t.Run(mode, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			const branch = "000374-guard"
			git(t, root, "config", "branch."+branch+".remote", "origin")
			before := git(t, root, "config", "--local", "--list")
			tip := git(t, root, "rev-parse", "HEAD")
			if mode == "ref-present" {
				git(t, root, "update-ref", "refs/heads/"+branch, tip)
			}
			if mode == "occupied" {
				git(t, root, "symbolic-ref", "HEAD", "refs/heads/"+branch)
			}
			reached := false
			query := func(args ...string) (string, error) {
				if mode == "ref-probe-fails" && args[0] == "rev-parse" || mode == "worktree-probe-fails" && args[0] == "worktree" {
					reached = true
					return "", errors.New("probe failed")
				}
				if strings.Contains(strings.Join(args, " "), "--remove-section") {
					if mode != "manual-recreation" {
						t.Error("destructive config write despite protected observation")
					}
					reached = true
					git(t, root, "update-ref", "refs/heads/"+branch, tip)
				}
				return landingGit(execGitRunner{}, root, args...)
			}
			err := removeObsoleteBranchConfig(query, branch)
			if strings.HasSuffix(mode, "fails") && (err == nil || !reached) {
				t.Fatal("failed probe not preserved", err)
			}
			if mode == "occupied" && err != nil {
				t.Fatal(err)
			}
			if mode != "manual-recreation" && git(t, root, "config", "--local", "--list") != before {
				t.Fatal("config changed")
			}
			if mode == "manual-recreation" {
				if err != nil || !reached || git(t, root, "rev-parse", "refs/heads/"+branch) != tip {
					t.Fatal("late manual ref/commit lost", err)
				}
			}
		})
	}
}
