package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

var errInitTrackerAlreadyInitialized = errors.New("tracker already initialized")

// Initialization is presence-only: it never interprets, imports or merges issue
// records. Migration's changed-file policy cannot prove a new repository empty.
func verifyInitializationEmpty(env *migrateEnv, main string) error {
	d := vocab.Issue().Discovery()
	roots := issue.IDDirs(d.Home, d.Archive)
	found, err := initializationEvidence(env, main, roots)
	if err != nil {
		if errors.Is(err, errInitTrackerAlreadyInitialized) {
			return fmt.Errorf("%w; use sdlc issue migrate", err)
		}
		return fmt.Errorf("cannot verify empty issue history; use sdlc issue migrate after resolving the observation: %w", err)
	}
	if found != "" {
		return fmt.Errorf("this repo has issue history (%s); use sdlc issue migrate", found)
	}
	return nil
}

// initializationIssuePath is the single path rule for snapshots, indexes and
// filesystem entries. Plans, assets and side bundles nested below a root are
// not top-level issue documents; the flat legacy root remains a first-class root.
func initializationIssuePath(p string, roots []string) bool {
	_, _, ok := issue.IssueInventoryPath(p, roots)
	return ok
}

func initializationEvidence(env *migrateEnv, main string, roots []string) (string, error) {
	observation, err := observeIssueInventory(env, main)
	if err != nil {
		return "", err
	}
	for _, ref := range observation.Tips {
		if ref.Name == "refs/heads/"+vocab.Issue().Discovery().Tracker {
			return "", fmt.Errorf("%w: pre-existing local reserved branch %s", errInitTrackerAlreadyInitialized, ref.Name)
		}
	}
	tips := observation.Tips
	worktrees := observation.Worktrees
	seen := map[string]bool{}
	for _, tip := range tips {
		if seen[tip.OID] {
			continue
		}
		seen[tip.OID] = true
		found, err := initializationTipEvidence(env, tip.OID, roots)
		if err != nil {
			return "", fmt.Errorf("%s: %w", tip.Name, err)
		}
		if found != "" {
			if tip.Name == gitx.RemoteTrackingRef(env.target.Remote, vocab.Issue().Discovery().Tracker) {
				return "", fmt.Errorf("%w: %s contains issue history at %s", errInitTrackerAlreadyInitialized, tip.Name, found)
			}
			return tip.Name + ": " + found, nil
		}
	}
	for _, wt := range worktrees {
		if wt.Bare || wt.Prunable != nil {
			continue
		}
		args := append([]string{"-C", wt.Path, "--literal-pathspecs", "ls-files", "-z", "--cached", "--"}, roots...)
		raw, err := inventoryGit(env, args...)
		if err != nil {
			return "", err
		}
		if len(raw) > 0 && raw[len(raw)-1] != 0 {
			return "", errors.New("truncated worktree index inventory")
		}
		for _, p := range strings.Split(string(raw), "\x00") {
			if initializationIssuePath(p, roots) {
				return "worktree index " + wt.Path + ": " + p, nil
			}
		}
		found, err := initializationFilesystemEvidence(env, wt.Path, roots)
		if err != nil {
			return "", fmt.Errorf("worktree %s: %w", wt.Path, err)
		}
		if found != "" {
			return "worktree " + wt.Path + ": " + found, nil
		}
	}
	return "", nil
}

// The shared direct-child metadata reader avoids all issue body reads. Pins
// cover unrelated and inherited histories: no difference rule erases presence.
func initializationTipEvidence(env *migrateEnv, tip string, roots []string) (string, error) {
	files, err := issueSnapshotFiles(env, tip, roots)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		return files[0].Path, nil
	}
	return "", nil
}

func initializationFilesystemEvidence(env *migrateEnv, worktree string, roots []string) (string, error) {
	files, err := issueFilesystemFiles(env, worktree, roots)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		return files[0], nil
	}
	return "", nil
}
