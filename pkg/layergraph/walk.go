package layergraph

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"path/filepath"
)

// Walk discovers the transitive construct/deps layer graph for the repo at root
// (which must be absolute) and returns its layer roots foundation-first, leaf
// last, deduped, and physical-path canonicalized (absolute). It is the SINGLE
// source of truth for "what is repo R's layer graph" (ARCH-DRY): cmd/weave
// loads each layer's rich manifest/prose on top of this topology, and any other
// DAG-aware subsystem (cmd/datatype) consumes the same ordered roots, so they
// never diverge on topology.
//
// It ports two shell behaviors verbatim:
//   - deps_substrate_targets (lib-deps.sh): per `substrate` row, repo-root-
//     relative OR absolute-path resolution, then physical-path canonicalization
//     (pwd -P) with present-skip — an unresolvable parent is dropped silently.
//   - discover_ancestors' two _seen_or_add filters (setup.sh): a candidate
//     counts as a layer ONLY if it ships construct/base.manifest, and the
//     target repo is never its own ancestor. construct/deps is the sole edge
//     source (no go.mod / go list).
func Walk(fs FS, root string) ([]string, error) {
	root = physical(root) // canonicalize so self-comparisons match (pwd -P)

	edges, err := discoverEdges(fs, root)
	if err != nil {
		return nil, err
	}
	return Resolve(root, edges)
}

// DeclaredSubstrate is one substrate a layer declares (#294), whether or not it
// is checked out.
type DeclaredSubstrate struct {
	Path    string // physical, absolute (substrateTargets' resolution)
	Owner   string // the layer root whose construct/deps declares it
	Source  string // the row's source column, "" when absent
	Present bool   // exists on disk
}

// DeclaredSubstrates walks construct/deps from root over present layers
// (transitively, as Walk does) and returns every declared substrate, present
// or absent, deduplicated by Path, in discovery order. It is Walk's own
// traversal (declaredGraph), so the two cannot diverge: Walk's layers minus
// the root are exactly its present entries. Errors are Walk's: a malformed
// row, or a present substrate without construct/base.manifest.
func DeclaredSubstrates(fs FS, root string) ([]DeclaredSubstrate, error) {
	_, declared, err := declaredGraph(fs, physical(root))
	return declared, err
}

// discoverEdges is declaredGraph's edge map, which Resolve consumes.
func discoverEdges(fs FS, root string) (map[string][]string, error) {
	edges, _, err := declaredGraph(fs, root)
	return edges, err
}

// declaredGraph follows construct/deps transitively from root — a BFS
// mirroring discover_ancestors' substrate walk, construct/deps-only — and
// returns both views of it: the edge map (each discovered layer dir → the
// layer dirs it depends on, the substrate targets passing the _seen_or_add
// layer filter) and every declared substrate in discovery order.
func declaredGraph(fs FS, root string) (map[string][]string, []DeclaredSubstrate, error) {
	edges := map[string][]string{}
	var declared []DeclaredSubstrate
	seen := map[string]bool{}
	visited := map[string]bool{}
	queue := []string{root}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true

		targets, err := substrateTargets(fs, cur)
		if err != nil {
			return nil, nil, err
		}
		for _, target := range targets {
			dep := target.Path
			// _seen_or_add filters: a dep counts as a layer only if it ships a
			// base.manifest, and the root is never its own ancestor.
			if dep == root {
				continue // target-self-exclusion
			}
			present := pathExists(fs, dep)
			if !seen[dep] {
				seen[dep] = true
				declared = append(declared, DeclaredSubstrate{Path: dep, Owner: cur, Source: target.Source, Present: present})
			}
			if !hasManifest(fs, dep) {
				// A declared `substrate` row is a layer edge (ParseDeps yields only
				// substrate rows). If its target is PRESENT on disk but ships no
				// construct/base.manifest, it is a broken layer edge — the old
				// silent skip dropped the ENTIRE transitive chain below it with no
				// signal (#155: a fresh-bootstrapped derivative under-compiled to a
				// 1-action no-op). Fail loud and actionable. An ABSENT target keeps
				// the silent present-skip (a peer simply not checked out).
				if present {
					return nil, nil, fmt.Errorf(
						"substrate %s (declared in %s) is present but not a compilable layer: "+
							"missing %s — seed its base.manifest (`weave link` does this) or fix the construct/deps path",
						dep, filepath.Join(cur, "construct", "deps"),
						filepath.Join(dep, "construct", "base.manifest"))
				}
				continue // absent peer — present-skip (not checked out)
			}
			edges[cur] = append(edges[cur], dep)
			if !visited[dep] {
				queue = append(queue, dep)
			}
		}
	}
	return edges, declared, nil
}

// substrateRow is one resolved `substrate` row.
type substrateRow struct{ Path, Source string }

// substrateTargets ports lib-deps.sh:deps_substrate_targets — parses
// repoRoot/construct/deps and resolves each `substrate` row to an absolute,
// physical-path peer dir, keeping its source column. Relative targets resolve
// against repoRoot; absolute targets are taken verbatim; the result is
// canonicalized via the parent dir (pwd -P semantics) so an absent peer still
// resolves syntactically, while an unresolvable parent is dropped
// (present-skip). The ParseRows grammar is reused — only the resolution is
// added here.
func substrateTargets(fs FS, repoRoot string) ([]substrateRow, error) {
	content, found, err := readDeclaration(fs, filepath.Join(repoRoot, "construct", "deps"))
	if err != nil || !found {
		return nil, err // no construct/deps ⇒ no edges (lib-deps: [[ -f ]] || return 0)
	}
	rows, err := ParseRows(string(content))
	if err != nil {
		return nil, err
	}
	var out []substrateRow
	for _, row := range rows {
		if row.Kind != "substrate" {
			continue
		}
		target := row.Path
		var raw string
		if filepath.IsAbs(target) {
			raw = target // raw="$target"
		} else {
			raw = filepath.Join(repoRoot, target) // raw="$repo_root/$target"
		}
		// parent="$(cd "$(dirname "$raw")" && pwd -P)"; present-skip if empty.
		parent := physicalOrEmpty(filepath.Dir(raw))
		if parent == "" {
			continue // unresolvable parent — skipped silently (present-peers)
		}
		out = append(out, substrateRow{Path: filepath.Join(parent, filepath.Base(raw)), Source: row.Source})
	}
	return out, nil
}

// readDeclaration reads a construct/deps document, bounded (#294): through the
// FS's safe reader when it has one (OSFS: ReadDeclaration's ordinary-file,
// no-follow, no-FIFO rules), else ReadFile with the byte limit. found is false
// when there is none.
func readDeclaration(fs FS, path string) ([]byte, bool, error) {
	if r, ok := fs.(DeclarationReader); ok {
		b, err := r.ReadDeclaration(path)
		if errors.Is(err, iofs.ErrNotExist) {
			return nil, false, nil
		}
		return b, err == nil, err
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		return nil, false, nil // as before: an unreadable declaration is none
	}
	if int64(len(b)) > DeclarationLimit {
		return nil, false, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, DeclarationLimit)
	}
	return b, true, nil
}

// hasManifest reports whether dir ships construct/base.manifest (the
// _seen_or_add layer filter).
func hasManifest(fs FS, dir string) bool {
	_, err := fs.Stat(filepath.Join(dir, "construct", "base.manifest"))
	return err == nil
}

// pathExists reports whether path is present on disk (dir or file). It
// distinguishes a substrate target that is checked out but not a compilable
// layer (present, no base.manifest → a #155 loud error) from one that simply
// isn't present (absent → a silent present-skip).
func pathExists(fs FS, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}

// physical canonicalizes path to its physical form (EvalSymlinks ≈ pwd -P),
// falling back to the input when it can't be resolved (e.g. doesn't exist yet).
// This is a real-disk concern, intentionally NOT routed through FS — it
// preserves the macOS /tmp→/private/tmp semantics the relative-symlink targets
// depend on (the bug setup.sh's pwd -P guards against).
func physical(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

// physicalOrEmpty is physical that returns "" when the path can't be resolved —
// the present-skip signal for an absent parent dir (cd … && pwd -P || true).
func physicalOrEmpty(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return ""
}
