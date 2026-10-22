package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/pkg/layergraph"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const reviewDependencyTimeout = 2 * time.Minute

// prepareReviewDependencies observes the pinned declaration, never acquiring or
// refreshing a source. Each child is owned before the next effect can fail.
func prepareReviewDependencies(ctx context.Context, owner *reviewbox.Checkout, caller string) ([]reviewbox.Dependency, []string, error) {
	ctx, cancel := context.WithTimeout(ctx, reviewDependencyTimeout)
	defer cancel()
	rows, err := resolvedDependencyRowsWithAnchors(caller, owner.Dir, 64, reviewDependencyAnchors{ctx})
	if err != nil {
		return nil, nil, err
	}
	// Reuse exactly the suite's conflict rule before cloning anything.
	if _, err := dependencyEnvironment(nil, caller, rows); err != nil {
		return nil, nil, err
	}
	var deps []reviewbox.Dependency
	var protect []string
	seen := map[string]reviewbox.Dependency{}
	kinds := map[string]bool{}
	var total int64
	for _, row := range rows {
		if kinds[row.Kind] {
			continue
		}
		kinds[row.Kind] = true
		source, err := filepath.EvalSymlinks(row.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("review dependency %s: %w; repair the local dependency and rerun the boundary", row.Kind, err)
		}
		source, err = filepath.Abs(source)
		if err != nil {
			return nil, nil, err
		}
		if previous, ok := seen[source]; ok {
			previous.Kind = row.Kind
			deps = append(deps, previous)
			continue
		}
		top, err := reviewDependencyGit(ctx, source, "rev-parse", "--show-toplevel")
		if err != nil {
			return nil, nil, fmt.Errorf("review dependency %s at %s: %w; repair the dependency and rerun the boundary", row.Kind, source, err)
		}
		if filepath.Clean(strings.TrimSpace(string(top))) != source {
			return nil, nil, fmt.Errorf("review dependency %s at %s is not a repository root; repair the dependency and rerun the boundary", row.Kind, source)
		}
		pin, err := reviewDependencyGit(ctx, source, "rev-parse", "--verify", "refs/remotes/origin/main^{commit}")
		if err != nil {
			return nil, nil, fmt.Errorf("review dependency %s: last-fetched origin/main unavailable: %w; refresh the local dependency outside review and rerun the boundary", row.Kind, err)
		}
		head := strings.TrimSpace(string(pin))
		tree, err := reviewDependencyGit(ctx, source, "ls-tree", "-r", "-l", "-z", "--full-tree", head)
		if err != nil {
			return nil, nil, fmt.Errorf("review dependency %s tree: %w", row.Kind, err)
		}
		_, size, err := reviewbox.DependencyTreeBudget(tree)
		if err != nil {
			return nil, nil, fmt.Errorf("review dependency %s: %w", row.Kind, err)
		}
		total += size
		if total > 1<<30 {
			return nil, nil, fmt.Errorf("review dependencies exceed 1 GiB materialized blob budget")
		}
		age := dependencyFetchObservation(ctx, source, head, time.Now())
		paths, err := reviewProtectPathsContext(ctx, source)
		if err != nil {
			return nil, nil, err
		}
		child, err := reviewbox.CreateDependencyContext(ctx, filepath.Dir(owner.Dir), source, head)
		if err != nil {
			return nil, nil, fmt.Errorf("create review dependency %s: %w", row.Kind, err)
		}
		if err := owner.OwnDependency(child); err != nil {
			_ = child.Close()
			return nil, nil, err
		}
		dep := reviewbox.Dependency{Kind: row.Kind, Dir: child.Dir, Head: head, FetchAge: age}
		seen[source] = dep
		deps = append(deps, dep)
		protect = append(protect, paths...)
		protect = append(protect, child.Dir)
	}
	return deps, protect, ctx.Err()
}

type reviewDependencyAnchors struct{ ctx context.Context }

func (a reviewDependencyAnchors) PrimaryCheckout(layer string) (string, error) {
	return workspace.DependencyAnchor(a, layer)
}

func (a reviewDependencyAnchors) GitInDir(dir string, args ...string) ([]byte, error) {
	return reviewDependencyGit(a.ctx, dir, args...)
}

func reviewDependencyGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	out, stderr, err := gitx.RunGitBoundedContext(ctx, dir, []string{"GIT_NO_REPLACE_OBJECTS=1", "GIT_NO_LAZY_FETCH=1"}, 32<<20, append([]string{"--no-replace-objects", "-c", "protocol.allow=never"}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(stderr)))
	}
	return out, nil
}

// An observation's timestamp is not a commit date or a promise of remote
// freshness. Multi-entry/append fetch files cannot establish an entry's age.
func dependencyFetchObservation(ctx context.Context, source, head string, now time.Time) string {
	p, err := reviewDependencyGit(ctx, source, "rev-parse", "--path-format=absolute", "--git-path", "FETCH_HEAD")
	if err != nil {
		return "unknown (FETCH_HEAD unavailable)"
	}
	path := strings.TrimSpace(string(p))
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Size() > 64<<10 {
		return "unknown (FETCH_HEAD unavailable or not a bounded ordinary file)"
	}
	// The shared bounded reader rejects symlinks and non-ordinary files.
	raw, err := layergraph.ReadDeclaration(path, 64<<10)
	if err != nil {
		return "unknown (FETCH_HEAD unreadable)"
	}
	after, err := os.Lstat(path)
	if err != nil || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "unknown (FETCH_HEAD changed during observation)"
	}
	origin, err := reviewDependencyGit(ctx, source, "config", "--get", "remote.origin.url")
	if err != nil {
		return "unknown (origin identity unavailable)"
	}
	return reviewFetchAge(string(raw), head, strings.TrimSpace(string(origin)), after.ModTime(), now)
}

func reviewFetchAge(raw, head, origin string, modified, now time.Time) string {
	line := strings.TrimSuffix(raw, "\n")
	fields := strings.Split(line, "\t")
	if strings.Contains(line, "\n") || len(fields) != 3 || fields[0] != head || origin == "" || modified.After(now) {
		return "unknown (no single matching origin/main fetch observation)"
	}
	remote, ok := strings.CutPrefix(fields[2], "branch 'main' of ")
	if !ok || strings.TrimSuffix(remote, ".git") != strings.TrimSuffix(origin, ".git") {
		return "unknown (FETCH_HEAD does not identify this origin/main)"
	}
	return fmt.Sprintf("%s (matching FETCH_HEAD file observed %s; remote freshness unverified)", now.Sub(modified).Round(time.Second), modified.UTC().Format(time.RFC3339))
}
