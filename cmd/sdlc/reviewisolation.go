package main

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

// openReviewIsolation makes the disposable checkout one boundary-review round
// runs in (#204) and the isolation that confines the reviewer to it. The
// caller owns the returned checkout and must Close it once the reviewer has
// exited. When no OS write fence can be applied (Seatbelt cannot nest inside
// an agent's own sandbox), the round degrades to checkout-only isolation with
// a loud warning; the round and its sidecar record which one it got.
func openReviewIsolation(stderr io.Writer, m judge.ReviewWindowManifest) (*reviewbox.Checkout, reviewbox.Isolation, error) {
	protect, err := reviewProtectPaths(m.RepoRoot)
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	root, err := reviewbox.Root()
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	checkout, err := reviewbox.Create(root, m.RepoRoot, m.HeadSHA, reviewOverlays(m))
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	fence, reason := reviewbox.Probe()
	iso := reviewbox.Isolation{Fence: fence, Dir: checkout.Dir, Protect: protect, Reason: reason}
	if iso.Level() == reviewbox.LevelCheckout {
		cwarn(stderr, fmt.Sprintf("boundary review: write fence unavailable (%s); the reviewer is isolated to a disposable checkout only (isolation: checkout)", reason))
	}
	return checkout, iso, nil
}

// plannedReviewIsolation is the isolation a dispatch would get, for --dry-run:
// the probed fence around a placeholder checkout path. It creates nothing.
func plannedReviewIsolation(m judge.ReviewWindowManifest) (reviewbox.Isolation, error) {
	protect, err := reviewProtectPaths(m.RepoRoot)
	if err != nil {
		return reviewbox.Isolation{}, err
	}
	root, err := reviewbox.Root()
	if err != nil {
		return reviewbox.Isolation{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return reviewbox.Isolation{}, err
	}
	fence, reason := reviewbox.Probe()
	return reviewbox.Isolation{Fence: fence, Dir: filepath.Join(root, "<review-checkout>"), Protect: protect, Reason: reason}, nil
}

// reviewProtectPaths are what the reviewer must not write: the source
// checkout and its Git common dir (refs, config, objects, every worktree's
// index). For a linked worktree the common dir lies outside the checkout.
func reviewProtectPaths(repoRoot string) ([]string, error) {
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return nil, fmt.Errorf("resolve review git common dir: %w", err)
	}
	return []string{repoRoot, strings.TrimSpace(string(out))}, nil
}

// reviewOverlays are the live issue and plan files, copied into the checkout
// so the reviewer reads the text it read before #204 (they may be
// uncommitted at close). An absolute path is read where it lies.
func reviewOverlays(m judge.ReviewWindowManifest) []string {
	var out []string
	for _, p := range []string{m.IssueFile, m.PlanFile} {
		if p != "" && !filepath.IsAbs(p) {
			out = append(out, p)
		}
	}
	return out
}
