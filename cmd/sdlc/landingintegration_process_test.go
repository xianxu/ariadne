//go:build unix

package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Retain the fixture's identity substitution while all Git subprocesses use
// the production context/process-group boundary, including inherited Git calls.
type landingCancellableTransport struct{ gitRunner }

func (r landingCancellableTransport) GitInDir(root string, args ...string) ([]byte, error) {
	if len(args) >= 3 && args[0] == "remote" && args[1] == "get-url" {
		return r.gitRunner.GitInDir(root, "config", "--get-all", "remote."+args[len(args)-1]+".url")
	}
	return r.gitRunner.GitInDir(root, args...)
}

func TestLandingCancelledPushCanResume(t *testing.T) {
	root, _, gh := integrationFixture(t)
	hookPath := filepath.Join(gh.remote, "hooks", "post-receive")
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\nsleep 60 & printf '%s' \"$!\" > \"$SDLC_GIT_CHILD_PID\"; wait\n"), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	original := mergeRunner
	mergeRunner = landingCancellableTransport{boundedLandingGit(ctx, execGitRunner{})}
	defer func() { mergeRunner = original }()
	f := landingFlags()
	f.Context = ctx
	done := make(chan error, 1)
	go func() { done <- runMerge(io.Discard, io.Discard, f) }()
	awaitLandingGitChild(t, pidfile)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled push: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("landing cancellation exceeded bound")
	}
	head := procedureHead(t, root)
	if got := git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch); got != head {
		t.Fatalf("uncertain published head %s want %s", got, head)
	}
	if gh.merges != 0 {
		t.Fatal("cancelled landing merged PR")
	}
	if err := os.Remove(hookPath); err != nil {
		t.Fatal(err)
	}
	mergeRunner = original
	// The remote already changed, but GitHub has not observed it. Re-entry must
	// reconcile the integration evidence without attempting a duplicate merge.
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 || gh.pr.HeadOID != head {
		t.Fatalf("resume merged=%d head=%s want %s", gh.merges, gh.pr.HeadOID, head)
	}
}
