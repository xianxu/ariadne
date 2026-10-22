package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// Initialization is presence-only: it never interprets, imports or merges issue
// records. Migration's changed-file policy cannot prove a new repository empty.
func verifyInitializationEmpty(env *migrateEnv, main string) error {
	d := vocab.Issue().Discovery()
	roots := issue.IDDirs(d.Home, d.Archive)
	found, err := initializationEvidence(env, main, roots)
	if err != nil {
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
	if _, _, ok := issue.ParseFilename(path.Base(p)); !ok {
		return false
	}
	for _, root := range roots {
		if path.Dir(p) == root {
			return true
		}
	}
	return false
}

func initializationEvidence(env *migrateEnv, main string, roots []string) (string, error) {
	refs, err := migrationBranchRefs(env, main)
	if err != nil {
		return "", err
	}
	raw, err := env.gitRaw(nil, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", err
	}
	worktrees, err := gitx.ParseWorktrees(raw)
	if err != nil {
		return "", err
	}
	tips := append([]migrationBranchRef{{Name: "main", OID: main}}, refs...)
	for _, wt := range worktrees {
		if wt.Bare || wt.Prunable != nil {
			continue
		}
		tips = append(tips, migrationBranchRef{Name: "worktree " + wt.Path, OID: wt.HEAD})
	}
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
			return tip.Name + ": " + found, nil
		}
	}
	for _, wt := range worktrees {
		if wt.Bare || wt.Prunable != nil {
			continue
		}
		args := append([]string{"-C", wt.Path, "--literal-pathspecs", "ls-files", "-z", "--cached", "--"}, roots...)
		raw, err := env.gitRaw(nil, args...)
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
		found, err := initializationFilesystemEvidence(wt.Path, roots)
		if err != nil {
			return "", fmt.Errorf("worktree %s: %w", wt.Path, err)
		}
		if found != "" {
			return "worktree " + wt.Path + ": " + found, nil
		}
	}
	return "", nil
}

// Only metadata is selected. Returning false avoids reading any issue body and
// keeps process count independent of file count. Pins also cover unrelated or
// inherited histories: no merge-base/difference rule can erase present evidence.
func initializationTipEvidence(env *migrateEnv, tip string, roots []string) (string, error) {
	view := env.main.ViewOf(tip)
	kinds := map[string]string{}
	for _, root := range roots {
		present := true
		parts := strings.Split(root, "/")
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			kind, known := kinds[prefix]
			if !known {
				exists, err := view.Exists(prefix)
				if err != nil {
					return "", err
				}
				if exists {
					kind, err = env.git("cat-file", "-t", tip+":"+prefix)
					if err != nil {
						return "", err
					}
				}
				kinds[prefix] = kind
			}
			if kind == "" {
				present = false
				break
			}
			if kind != "tree" {
				return "", fmt.Errorf("issue root component %s is not a directory", prefix)
			}
		}
		if !present {
			continue
		}
		found := ""
		_, err := view.FilesSelected(root, func(f gitx.TreeFile) bool {
			if initializationIssuePath(f.Path, []string{root}) {
				found = f.Path
			}
			return false
		})
		if err != nil {
			return "", err
		}
		if found != "" {
			return found, nil
		}
	}
	return "", nil
}

func initializationFilesystemEvidence(worktree string, roots []string) (string, error) {
	for _, root := range roots {
		present := true
		parts := strings.Split(root, "/")
		for i := range parts {
			prefix := filepath.Join(worktree, filepath.FromSlash(strings.Join(parts[:i+1], "/")))
			info, err := os.Lstat(prefix)
			if os.IsNotExist(err) {
				present = false
				break
			}
			if err != nil {
				return "", err
			}
			if !info.IsDir() {
				return "", fmt.Errorf("issue root component %s is not a directory", prefix)
			}
		}
		if !present {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(worktree, filepath.FromSlash(root)))
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			p := path.Join(root, entry.Name())
			if initializationIssuePath(p, roots) {
				return p, nil
			}
		}
	}
	return "", nil
}
