package main

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// verifyIndexedAssetBundles pins the staged tree before a local merge commit.
// The body indexes and content names must describe its bytes, and materialized
// files must still equal those bytes. This validates preservation, not admission.
func verifyIndexedAssetBundles(env *trackerEnv, issuesDir string) error {
	tree, err := env.git("write-tree")
	if err != nil {
		return fmt.Errorf("pin merged asset index: %w; resolve the index before retrying", err)
	}
	view := env.main.ViewOf(tree)
	for _, dir := range []string{issuesDir, vocab.ArchiveSubdir(historyDir(), vocab.ArchiveIssues)} {
		home, err := gitx.InsideRoot(env.root, dir)
		if err != nil {
			return err
		}
		owners := map[string]bool{}
		docs := map[string][]gitx.TreeFile{}
		assets := path.Join(home, "assets")
		var invalidRoot bool
		files, err := view.FilesSelected(home, func(f gitx.TreeFile) bool {
			if f.Path == assets {
				invalidRoot = true
			}
			if strings.HasPrefix(f.Path, assets+"/") {
				id, _, _ := strings.Cut(strings.TrimPrefix(f.Path, assets+"/"), "/")
				owners[id] = true
			}
			_, _, ok := issue.ParseFilename(path.Base(f.Path))
			return path.Dir(f.Path) == home && ok
		})
		if err != nil {
			return fmt.Errorf("read merged issue index: %w; restore readable issue details before retrying", err)
		}
		if invalidRoot {
			return fmt.Errorf("asset directory %s is a file or symlink; restore a regular directory before committing", assets)
		}
		for _, f := range files {
			id, _, _ := issue.ParseFilename(path.Base(f.Path))
			docs[id] = append(docs[id], f)
			refs, err := issueassets.ParseRefs(id, issueAssetBody(f.Content))
			if err != nil {
				return fmt.Errorf("%s: %w", f.Path, err)
			}
			if len(refs) > 0 {
				owners[id] = true
			}
		}
		ids := make([]string, 0, len(owners))
		for id := range owners {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			if len(docs[id]) > 1 {
				return fmt.Errorf("multiple details index assets for issue %s; reconcile duplicate files before committing", id)
			}
			detail := path.Join(home, id+"-missing.md")
			var body []byte
			if len(docs[id]) == 1 {
				detail, body = docs[id][0].Path, docs[id][0].Content
			}
			bundle, err := assetFilesFromView(view, detail, id, body)
			if err != nil {
				return err
			}
			var paths []string
			for p := range assetBundleWrite(bundle, detail) {
				if err := verifyAssetAttributes(env.ctx, env.root, p); err != nil {
					return err
				}
				if err := verifyAssetAttributesAt(env.ctx, env.root, p, tree); err != nil {
					return err
				}
				paths = append(paths, p)
			}
			sort.Strings(paths)
			if err := verifyMaterializedAssetPaths(env.root, view, paths); err != nil {
				return err
			}
		}
	}
	return nil
}
