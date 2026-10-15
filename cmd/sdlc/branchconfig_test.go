package main

import (
	"bytes"
	"context"
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
