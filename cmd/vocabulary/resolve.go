package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// resolveVocab DAG-merges this repo's construct/vocabulary/*.cue across the
// layer graph (foundation-first, leaf-wins by filename), then overlays the
// leaf's project-local vocabulary/ dir LAST. Returns noun name → winning .cue
// absolute path. Mirrors cmd/datatype's resolveTypes; the shadow policy lives in
// the shared layergraph.MergeByName (ARCH-DRY). Test fixtures under
// construct/vocabulary/testdata/ are not picked up (MergeByName skips subdirs).
func resolveVocab() (map[string]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}
	root, err := findRepoRoot(cwd)
	if err != nil {
		return nil, err
	}
	fs := layergraph.OSFS{}
	roots, err := layergraph.Walk(fs, root)
	if err != nil {
		return nil, fmt.Errorf("walk layer graph from %s: %w", root, err)
	}
	dirs := make([]string, 0, len(roots)+1)
	for _, r := range roots {
		dirs = append(dirs, filepath.Join(r, "construct", "vocabulary"))
	}
	dirs = append(dirs, filepath.Join(root, "vocabulary")) // leaf project-local
	return layergraph.MergeByName(fs, dirs, ".cue")
}

// findRepoRoot returns the nearest ancestor of start with a construct/ dir (the
// layer-root signal), else the nearest with .git. A compact twin of
// cmd/datatype's findRepoRoot — candidates for a shared pkg helper later (noted,
// not extracted: two ~20-line path walks, below the DRY-extraction threshold).
func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs %s: %w", start, err)
	}
	for d := dir; ; {
		if isDir(filepath.Join(d, "construct")) {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			break
		}
		d = p
	}
	for d := dir; ; {
		if isDir(filepath.Join(d, ".git")) {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			break
		}
		d = p
	}
	return "", fmt.Errorf("no repo root above %s (no ancestor has construct/ or .git)", start)
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
