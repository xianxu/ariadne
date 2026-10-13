package main

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// assetMaterializationPaths lists metadata only: checkout may materialize many
// issue bundles, whose aggregate bytes need not fit a single snapshot batch.
func assetMaterializationPaths(root string, view *gitx.TrunkView, issuesDir string) ([]string, error) {
	if canonical, err := filepath.EvalSymlinks(root); err == nil {
		root = canonical
	}
	var prefixes []string
	for _, dir := range []string{issuesDir, vocab.ArchiveSubdir(historyDir(), vocab.ArchiveIssues)} {
		rel, err := gitx.InsideRoot(root, dir)
		if err != nil {
			return nil, fmt.Errorf("asset directory %s: %w; set the issue and history directories inside this repository", dir, err)
		}
		prefixes = append(prefixes, path.Join(rel, "assets")+"/")
	}
	var paths []string
	_, err := view.FilesSelected("", func(f gitx.TreeFile) bool {
		for _, prefix := range prefixes {
			if strings.HasPrefix(f.Path, prefix) {
				paths = append(paths, f.Path)
				break
			}
		}
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("list asset paths at %s: %w; fetch that commit before retrying", view.Ref(), err)
	}
	return paths, nil
}

// Check both current and incoming attributes before checkout or merge. The
// additional refs describe other snapshots whose attributes may participate.
func verifyAssetMaterialization(ctx context.Context, root string, view *gitx.TrunkView, issuesDir string, attributeRefs ...string) error {
	paths, err := assetMaterializationPaths(root, view, issuesDir)
	if err != nil {
		return err
	}
	refs := append([]string{"", view.Ref()}, attributeRefs...)
	for _, p := range paths {
		seen := map[string]bool{}
		for _, ref := range refs {
			if !seen[ref] {
				seen[ref] = true
				if err := verifyAssetAttributesAt(ctx, root, p, ref); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// A checkout hook or other concurrent writer can alter bytes despite safe
// attributes. Compare regular, bounded files to their pinned Git contents.
func verifyMaterializedAssets(root string, view *gitx.TrunkView, issuesDir string) error {
	paths, err := assetMaterializationPaths(root, view, issuesDir)
	if err != nil {
		return err
	}
	for _, p := range paths {
		if err := assetDirectories(root, path.Dir(p), false); err != nil {
			return fmt.Errorf("materialized asset %s: %w; restore its regular directory and file from %s before retrying", p, err, view.Ref())
		}
		got, err := readAssetRegular(root, p, issueassets.MaxFileBytes)
		if err != nil {
			return fmt.Errorf("materialized asset %s: %w; restore this file from %s before retrying", p, err, view.Ref())
		}
		want, err := view.Read(p)
		if err != nil {
			return fmt.Errorf("read pinned asset %s at %s: %w; fetch that commit before retrying", p, view.Ref(), err)
		}
		if !bytes.Equal(got, want) {
			return fmt.Errorf("materialized asset %s differs from %s; restore this file from that commit before retrying", p, view.Ref())
		}
	}
	return nil
}

// A fast-forward or three-way merge can use attributes from either side;
// validate the union of paths under both snapshots before local edits are set aside.
func verifyAssetTransition(env *trackerEnv, target, issuesDir string) error {
	head, err := env.git("rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("read checkout head before asset materialization: %w; restore a valid HEAD before retrying", err)
	}
	for _, ref := range []string{head, target} {
		if err := verifyAssetMaterialization(env.ctx, env.root, env.main.ViewOf(ref), issuesDir, head, target); err != nil {
			return err
		}
	}
	return nil
}
