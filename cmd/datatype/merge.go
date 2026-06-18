package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
// layergraph.Walk returns), so iterating them in order means a leafward layer's
// write OVERWRITES an ancestor's same-filename key (local/leaf shadows shared).
// The leaf's project-local datatype/ dir (leafLocal) is overlaid LAST, so it
// wins over both an ancestor's and the leaf's own construct/datatype/ entry.
// Result is sorted by name.
//
// Pure over the injected FS (ARCH-PURE): the only IO is ReadDir per dir +
// ReadFile per prototype, all through fs. A missing dir is tolerated (a layer
// need not define any prototypes) — it contributes nothing, never an error.
// Reads ancestors' SOURCE prototype dirs directly (not their materialized
// copies), so a compile is independent of whether ancestors were compiled first
// (plan-review I3 — re-weave order is immaterial to correctness).
func mergeTypes(fs layergraph.FS, roots []string, leafLocal string) ([]TypeProto, error) {
	byName := map[string]TypeProto{}

	// Foundation-first: leafward writes overwrite the filename key.
	for _, root := range roots {
		dir := filepath.Join(root, "construct", "datatype")
		if err := overlayDir(fs, dir, byName); err != nil {
			return nil, err
		}
	}
	// The leaf's project-local datatype/ overlay wins over everything above.
	if err := overlayDir(fs, leafLocal, byName); err != nil {
		return nil, err
	}

	protos := make([]TypeProto, 0, len(byName))
	for _, p := range byName {
		protos = append(protos, p)
	}
	sort.Slice(protos, func(i, j int) bool { return protos[i].Name < protos[j].Name })
	return protos, nil
}

// overlayDir reads dir's *.md prototypes and writes each into byName keyed by
// filename-without-`.md`, overwriting any existing entry (the same-filename
// shadow). A *missing* dir is silently skipped (a layer need not define
// prototypes); any OTHER ReadDir error (permissions, ENOTDIR — dir exists as a
// file) is propagated, since silently dropping a layer's whole prototype set
// would corrupt both the eager description and the apply-time list/show
// (M2-review Important). Subdirectories and non-`.md` files are ignored.
func overlayDir(fs layergraph.FS, dir string, byName map[string]TypeProto) error {
	ents, err := fs.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no such dir ⇒ nothing here (a layer need not have prototypes)
		}
		return fmt.Errorf("read datatype dir %s: %w", dir, err)
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		fn := e.Name()
		if !strings.HasSuffix(fn, ".md") {
			continue
		}
		path := filepath.Join(dir, fn)
		content, err := fs.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read prototype %s: %w", path, err)
		}
		name := strings.TrimSuffix(fn, ".md")
		byName[name] = TypeProto{
			Name:        name,
			Description: frontmatter.Description(string(content)),
			BodyPath:    path,
		}
	}
	return nil
}
