package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"syscall"

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
	return openReviewIsolationContext(context.Background(), stderr, m)
}

func openReviewIsolationContext(ctx context.Context, stderr io.Writer, m judge.ReviewWindowManifest) (*reviewbox.Checkout, reviewbox.Isolation, error) {
	protect, err := reviewProtectPathsContext(ctx, m.RepoRoot)
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	root, roots, note, err := reviewbox.Roots()
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	if note != "" {
		cwarn(stderr, "boundary review: "+note)
	}
	// Reclaim crashed owners' checkouts first, on every root (#468). Best
	// effort: debris that cannot be removed is reported, never a reason to
	// refuse this review.
	for _, r := range roots {
		_, problems := reviewbox.Sweep(r)
		for _, p := range problems {
			cwarn(stderr, "boundary review: "+p.Error())
		}
	}
	checkout, err := reviewbox.CreateContext(ctx, root, m.RepoRoot, m.HeadSHA, reviewOverlays(m))
	if fallback := roots[len(roots)-1]; err != nil && root != fallback && (errors.Is(err, fs.ErrPermission) || errors.Is(err, syscall.EROFS)) {
		// The RAM disk refused the write (an agent sandbox denies it): the
		// normal root, rather than no review (#468). Any other failure (a bad
		// head, git) is the review's, and stands.
		cwarn(stderr, fmt.Sprintf("boundary review: %s refused the review checkout (%v); using %s", root, err, fallback))
		checkout, err = reviewbox.CreateContext(ctx, fallback, m.RepoRoot, m.HeadSHA, reviewOverlays(m))
	}
	if err != nil {
		return nil, reviewbox.Isolation{}, err
	}
	deps, dependencyProtect, err := prepareReviewDependencies(ctx, checkout, m.RepoRoot)
	if err != nil {
		return nil, reviewbox.Isolation{}, errors.Join(err, checkout.Close())
	}
	protect = append(protect, dependencyProtect...)
	fence, reason := reviewbox.Probe()
	iso := reviewbox.Isolation{Fence: fence, Dir: checkout.Dir, Protect: protect, Reason: reason, Dependencies: deps}
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
	root, _, _, err := reviewbox.Roots()
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

// reviewProtectPaths are what the reviewer must not write: every worktree of
// the source repository (the reviewed one, the operator's :0, sibling slots)
// and its Git common dir (refs, config, objects, every worktree's index). Each
// is canonical, because the fence matches the paths the kernel resolves. A
// worktree whose directory is gone (prunable) has nothing to protect.
func reviewProtectPaths(repoRoot string) ([]string, error) {
	return reviewProtectPathsContext(context.Background(), repoRoot)
}

func reviewProtectPathsContext(ctx context.Context, repoRoot string) ([]string, error) {
	common, err := reviewDependencyGit(ctx, repoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, fmt.Errorf("resolve review git common dir: %w", err)
	}
	list, err := reviewDependencyGit(ctx, repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list review worktrees: %w", err)
	}
	paths := []string{repoRoot, strings.TrimSpace(string(common))}
	for _, line := range strings.Split(string(list), "\n") {
		if wt, ok := strings.CutPrefix(line, "worktree "); ok {
			paths = append(paths, wt)
		}
	}
	var out []string
	seen := map[string]bool{}
	for _, p := range paths {
		c, err := reviewbox.Canonical([]string{p})
		if err != nil {
			if p == repoRoot {
				return nil, fmt.Errorf("resolve review repository root: %w", err)
			}
			continue
		}
		if !seen[c[0]] {
			seen[c[0]] = true
			out = append(out, c[0])
		}
	}
	return out, nil
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
