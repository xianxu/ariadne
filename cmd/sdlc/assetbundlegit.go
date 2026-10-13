package main

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

// assetFilesFromView inventories only this issue's directory. The body remains
// the index: BuildBundle rejects unknown entries rather than publishing them.
func assetFilesFromView(view *gitx.TrunkView, detailRel, id string, body []byte) (issueassets.Bundle, error) {
	home := path.Dir(detailRel)
	assets := path.Join(home, "assets")
	dir := path.Join(assets, id)
	// Include ancestor entries: a symlink replacing assets/ or its ID folder
	// is a blob at that exact path and is not listed by a trailing-slash query.
	files, err := view.FilesSelected(home, func(f gitx.TreeFile) bool {
		return f.Path == assets || f.Path == dir || strings.HasPrefix(f.Path, dir+"/")
	})
	if err != nil {
		return issueassets.Bundle{}, fmt.Errorf("cannot read asset bundle %s: %w; restore this bundle from its last valid commit before retrying", dir, err)
	}
	inventory := map[string][]byte{}
	for _, f := range files {
		if f.Path == assets || f.Path == dir {
			return issueassets.Bundle{}, fmt.Errorf("asset directory %s is a file or symlink: replace it with a regular directory", f.Path)
		}
		if f.Mode != "100644" {
			return issueassets.Bundle{}, fmt.Errorf("asset %s is not an ordinary file: restore a regular asset file", f.Path)
		}
		inventory[strings.TrimPrefix(f.Path, path.Dir(detailRel)+"/")] = f.Content
	}
	return issueassets.BuildBundle(id, issueAssetBody(body), inventory)
}
func assetBundleWrite(bundle issueassets.Bundle, detailRel string) map[string][]byte {
	out := map[string][]byte{}
	for p, b := range bundle.Files {
		out[path.Join(path.Dir(detailRel), p)] = b
	}
	return out
}
func assetBundlesEqual(a, b issueassets.Bundle) bool {
	if len(a.Files) != len(b.Files) {
		return false
	}
	for p, v := range a.Files {
		w, ok := b.Files[p]
		if !ok || !bytes.Equal(v, w) {
			return false
		}
	}
	return true
}
func assetBundlePaths(detailRel string, bundles ...issueassets.Bundle) []string {
	seen := map[string]bool{detailRel: true}
	for _, b := range bundles {
		for p := range assetBundleWrite(b, detailRel) {
			seen[p] = true
		}
	}
	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// validateAssetPublication inspects only affected issue bundles in the actual
// merged candidate. Existing evidence can move under a tightened policy; only
// newly admitted bytes require permission.
func validateAssetPublication(ctx context.Context, root string, base, next *gitx.TrunkView, changes []string, issuesDir, historyDir string) error {
	type owner struct{ home, id string }
	affected := map[owner]bool{}
	homes := []string{issuesDir, path.Join(historyDir, "issues")}
	for _, p := range changes {
		for _, home := range homes {
			if path.Dir(p) == home {
				if id := issueIDPrefix(path.Base(p)); id != "" {
					affected[owner{home, id}] = true
				}
			}
			assets := path.Join(home, "assets")
			if p == assets {
				entries, err := next.FilesSelected(home, func(f gitx.TreeFile) bool { return f.Path == assets })
				if err != nil {
					return err
				}
				if len(entries) > 0 {
					return fmt.Errorf("asset directory %s is a file or symlink: replace it with a regular directory", p)
				}
			}
			prefix := assets + "/"
			if strings.HasPrefix(p, prefix) {
				id, _, _ := strings.Cut(strings.TrimPrefix(p, prefix), "/")
				affected[owner{home, id}] = true
			}
		}
	}
	type bundlePair struct {
		o             owner
		before, after issueassets.Bundle
	}
	var pairs []bundlePair
	known := map[string][]byte{}
	for o := range affected {
		read := func(v *gitx.TrunkView) (issueassets.Bundle, error) {
			docs, err := v.FilesSelected(o.home, func(f gitx.TreeFile) bool {
				return path.Dir(f.Path) == o.home && strings.HasSuffix(f.Path, ".md") && issueIDPrefix(path.Base(f.Path)) == o.id
			})
			if err != nil {
				return issueassets.Bundle{}, err
			}
			if len(docs) > 1 {
				return issueassets.Bundle{}, fmt.Errorf("multiple details for issue %s: reconcile the duplicate files before publication", o.id)
			}
			detail := path.Join(o.home, o.id+"-missing.md")
			var body []byte
			if len(docs) == 1 {
				detail = docs[0].Path
				body = docs[0].Content
			}
			return assetFilesFromView(v, detail, o.id, body)
		}
		before, err := read(base)
		if err != nil {
			return err
		}
		after, err := read(next)
		if err != nil {
			return err
		}
		pairs = append(pairs, bundlePair{o, before, after})
		for p, b := range before.Files {
			known[p] = b
		}
	}
	needPolicy := false
	for _, pair := range pairs {
		for p, b := range issueassets.BundleDelta(pair.before, pair.after).Writes {
			if old, ok := known[p]; !ok || !bytes.Equal(old, b) {
				needPolicy = true
			}
			if err := verifyAssetAttributesAt(ctx, root, path.Join(pair.o.home, p), next.Ref()); err != nil {
				return err
			}
		}
	}

	if needPolicy {
		policy, err := resolveAssetPolicy(ctx, root)
		if err != nil {
			return err
		}
		return requireAssetPolicy(policy)
	}
	return nil
}

// issueAssetBody keeps tracker metadata outside the Markdown evidence index.
// Raw body fragments are also accepted for pure adapter fixtures.
func issueAssetBody(raw []byte) []byte {
	if _, body, err := issue.Parse(string(raw)); err == nil {
		return []byte(body)
	}
	return raw
}
