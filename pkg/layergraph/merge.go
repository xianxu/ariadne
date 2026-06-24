package layergraph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MergeByName resolves files matching *<ext> across an ordered list of dirs
// (given foundation-first, with the leaf's project-local dir LAST), keyed by
// basename without <ext>, where a later dir's same-named file SHADOWS an
// earlier one (local/leaf wins). Returns name → absolute winning path.
//
// This is the single "merge *.X across the layer graph, keyed by filename,
// leaf-wins" primitive (ARCH-DRY): cmd/datatype consumes it for *.md prototypes
// and cmd/vocabulary for *.cue models, so two DAG-aware tools can never diverge
// on shadow semantics. Callers layer their own per-file read on top (frontmatter
// parse for datatype, `cue export` for vocabulary).
//
// A *missing* dir is silently skipped (a layer need not contribute); any OTHER
// ReadDir error propagates (silently dropping a layer's whole set would corrupt
// the merged view). Subdirectories and non-matching files are ignored — so test
// fixtures under a `testdata/` subdir are never picked up. Pure over the
// injected FS (ARCH-PURE): the only IO is ReadDir per dir.
func MergeByName(fs FS, dirs []string, ext string) (map[string]string, error) {
	byName := map[string]string{}
	for _, dir := range dirs {
		ents, err := fs.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // no such dir ⇒ nothing here (a layer need not contribute)
			}
			return nil, fmt.Errorf("read dir %s: %w", dir, err)
		}
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			fn := e.Name()
			if !strings.HasSuffix(fn, ext) {
				continue
			}
			byName[strings.TrimSuffix(fn, ext)] = filepath.Join(dir, fn)
		}
	}
	return byName, nil
}
