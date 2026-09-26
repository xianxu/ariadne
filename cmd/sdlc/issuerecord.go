// issuerecord.go — the composed issue reader every consumer uses (#252): card
// fields from the tracker, detail fields from this checkout.
package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// loadIssueRecords composes the tracker's cards with the details in issuesDir,
// for the repository that contains issuesDir (not the process's current one). A
// directory outside any repository, or a repository whose publication remote has
// no tracker, reads its details as the whole record (pre-migration). Gates that
// authorize a write pass tracker.Fresh; read-only views pass PreferFresh and
// label a stale result.
func loadIssueRecords(ctx context.Context, issuesDir string, mode tracker.FetchMode) (tracker.Records, error) {
	abs, err := filepath.Abs(issuesDir)
	if err != nil {
		return tracker.Records{}, err
	}
	root := ""
	for dir := abs; ; dir = filepath.Dir(dir) {
		if out, err := gitx.RunGit("-C", dir, "rev-parse", "--show-toplevel"); err == nil {
			root = strings.TrimSpace(string(out))
			break
		}
		if filepath.Dir(dir) == dir {
			break // not inside a repository (or the directory does not exist yet)
		}
	}
	if root == "" {
		return tracker.LoadRecords(commandContext(ctx), nil, abs, mode)
	}
	return loadIssueRecordsAt(ctx, root, abs, mode)
}

// loadIssueRecordsAt reads the repository rooted at root (issuesDir relative to it).
func loadIssueRecordsAt(ctx context.Context, root, issuesDir string, mode tracker.FetchMode) (tracker.Records, error) {
	if !filepath.IsAbs(issuesDir) {
		issuesDir = filepath.Join(root, issuesDir)
	}
	repo, err := recordsRepository(ctx, root)
	if err != nil {
		return tracker.Records{}, err
	}
	return tracker.LoadRecords(commandContext(ctx), repo, issuesDir, mode)
}

// recordsRepository opens the tracker through the checkout's publication target.
// Without a target the repository can have no tracker — unless one was fetched
// already, which means the configuration broke: that is an error, never a quiet
// fallback to stale mirrors.
func recordsRepository(ctx context.Context, root string) (*tracker.Repository, error) {
	env, err := openTrackerAt(ctx, root)
	if err == nil {
		return env.repo, nil
	}
	fetched, ferr := gitx.RunGit("-C", root, "for-each-ref", "--format=%(refname)", "refs/remotes/*/issue-tracker")
	if ferr == nil && len(fetched) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf("this repository has a fetched issue tracker but its publication target is unusable: %w", err)
}
