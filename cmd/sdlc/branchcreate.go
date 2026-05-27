// branchcreate.go — branch-creation helpers shared by `sdlc change-code`
// (and historically `sdlc start`, now removed). Factored out so the
// worktree-or-in-place choice has one source of truth.
package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// commitUntrackedIssueFile commits + pushes one untracked file before
// branch creation, so the new branch starts from a tracked state.
// Push failures are warnings, not fatal — same posture as start.go's
// pre-#39 behavior and the legacy Makefile target.
func commitUntrackedIssueFile(stderr io.Writer, untrackedFile string, r gitRunner) error {
	if untrackedFile == "" {
		return nil
	}
	cinfo(stderr, fmt.Sprintf("Committing %s before branch creation...", untrackedFile))
	if out, err := r.Git("add", untrackedFile); err != nil {
		return fmt.Errorf("git add %s: %v\n%s", untrackedFile, err, out)
	}
	if out, err := r.Git("commit", "-m", "committing issue file before branch creation"); err != nil {
		return fmt.Errorf("git commit: %v\n%s", err, out)
	}
	if out, err := r.Git("push"); err != nil {
		cwarn(stderr, fmt.Sprintf("push failed, continuing with branch creation: %v\n%s", err, out))
	}
	return nil
}

// createWorktreeBranch creates a fresh git worktree on a new branch
// under ../worktree/<repo-dir-name>/<name>/, and writes the worktree
// path to <repo-root>/.goto so the `g` shell alias can cd there.
//
// Returns the worktree path on success.
func createWorktreeBranch(stdout, stderr io.Writer, name string, r gitRunner) (string, error) {
	repoTop, err := gitx.RepoTopLevel()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %v", err)
	}
	repoDir := filepath.Base(repoTop)
	wtRoot := filepath.Join(filepath.Dir(repoTop), "worktree", repoDir)
	wtPath := filepath.Join(wtRoot, name)

	if err := r.MkdirAll(wtRoot); err != nil {
		return "", fmt.Errorf("mkdir %s: %v", wtRoot, err)
	}
	if out, err := r.Git("worktree", "add", "-b", name, wtPath, "HEAD"); err != nil {
		return "", fmt.Errorf("git worktree add: %v\n%s", err, out)
	}
	cok(stderr, fmt.Sprintf("Worktree created at %s on branch %s", wtPath, name))

	gotoPath := filepath.Join(repoTop, ".goto")
	if err := r.WriteFile(gotoPath, []byte(wtPath)); err != nil {
		cwarn(stderr, fmt.Sprintf(".goto write failed: %v", err))
	} else {
		cok(stderr, "Run: g (to cd into worktree)")
	}
	fmt.Fprintln(stdout, wtPath)
	return wtPath, nil
}

// createInPlaceBranch creates a new branch on the current worktree.
// The working tree (including any uncommitted plan edits) carries
// forward to the new branch — that's the whole point of "in-place":
// the operator stays put and starts coding.
//
// Returns the branch name on success (same as input, for symmetry
// with createWorktreeBranch's return-the-location pattern).
func createInPlaceBranch(stdout, stderr io.Writer, name string, r gitRunner) (string, error) {
	if out, err := r.Git("checkout", "-b", name); err != nil {
		return "", fmt.Errorf("git checkout -b %s: %v\n%s", name, err, out)
	}
	cok(stderr, fmt.Sprintf("Branch %s created in place (working tree carried forward)", name))
	fmt.Fprintln(stdout, name)
	return name, nil
}
