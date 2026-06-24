package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/xianxu/ariadne/pkg/frontmatter"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

// TypeProto is one resolved datatype prototype after the DAG merge: the most
// leafward layer that defines a given filename owns the entry. Name is the
// filename without `.md` (the authoritative type name — NOT the `type:`
// frontmatter, which product.md sets to "type"). Description is the prototype's
// frontmatter `description:` (the matching surface). BodyPath is the absolute
// path of the winning prototype file, so `datatype show <name>` can print the
// resolved body.
type TypeProto struct {
	Name        string
	Description string
	BodyPath    string
}

// mergeTypes computes the DAG-merged datatype set for a repo: the union over
//
//	{each root's construct/datatype/*.md} ∪ {leafLocal/*.md}
//
// keyed by FILENAME minus `.md` (the naming trap — product.md → "product",
// though its `type:` field is "type"). roots MUST be foundation-first (the order
// layergraph.Walk returns), with leafLocal overlaid LAST, so a leafward layer's
// same-filename file SHADOWS an ancestor's (local/leaf wins). Result is sorted
// by name.
//
// The leaf-wins-by-filename shadow policy lives in layergraph.MergeByName
// (ARCH-DRY — the single merge primitive, shared with cmd/vocabulary so two
// DAG-aware tools never diverge). mergeTypes layers the datatype-specific read
// on top: each winning prototype's frontmatter `description:`. Pure over the
// injected FS (ARCH-PURE): the only IO is ReadDir per dir (in MergeByName) +
// ReadFile per winning prototype. Reads ancestors' SOURCE prototype dirs
// directly (not materialized copies), so re-weave order is immaterial.
func mergeTypes(fs layergraph.FS, roots []string, leafLocal string) ([]TypeProto, error) {
	dirs := make([]string, 0, len(roots)+1)
	for _, root := range roots {
		dirs = append(dirs, filepath.Join(root, "construct", "datatype"))
	}
	dirs = append(dirs, leafLocal)

	paths, err := layergraph.MergeByName(fs, dirs, ".md")
	if err != nil {
		return nil, err
	}

	protos := make([]TypeProto, 0, len(paths))
	for name, path := range paths {
		content, err := fs.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read prototype %s: %w", path, err)
		}
		protos = append(protos, TypeProto{
			Name:        name,
			Description: frontmatter.Description(string(content)),
			BodyPath:    path,
		})
	}
	sort.Slice(protos, func(i, j int) bool { return protos[i].Name < protos[j].Name })
	return protos, nil
}
