package layergraph

import "path/filepath"

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

// discoverEdges follows construct/deps transitively from root, returning the
// edge map Resolve consumes: each discovered layer dir → the layer dirs it
// depends on (substrate targets that pass the _seen_or_add layer filter). A BFS
// mirroring discover_ancestors' substrate walk, but construct/deps-only.
func discoverEdges(fs FS, root string) (map[string][]string, error) {
	edges := map[string][]string{}
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
			return nil, err
		}
		for _, dep := range targets {
			// _seen_or_add filters: a dep counts as a layer only if it ships a
			// base.manifest, and the root is never its own ancestor.
			if dep == root {
				continue // target-self-exclusion
			}
			if !hasManifest(fs, dep) {
				continue // base.manifest-existence filter
			}
			edges[cur] = append(edges[cur], dep)
			if !visited[dep] {
				queue = append(queue, dep)
			}
		}
	}
	return edges, nil
}

// substrateTargets ports lib-deps.sh:deps_substrate_targets — parses
// repoRoot/construct/deps and resolves each `substrate` row to an absolute,
// physical-path peer dir. Relative targets resolve against repoRoot; absolute
// targets are taken verbatim; the result is canonicalized via the parent dir
// (pwd -P semantics) so an absent peer still resolves syntactically, while an
// unresolvable parent is dropped (present-skip). The ParseDeps grammar (which
// rows are substrate) is reused — only the resolution is added here.
func substrateTargets(fs FS, repoRoot string) ([]string, error) {
	depsPath := filepath.Join(repoRoot, "construct", "deps")
	content, err := fs.ReadFile(depsPath)
	if err != nil {
		return nil, nil // no construct/deps ⇒ no edges (lib-deps: [[ -f ]] || return 0)
	}
	rels, err := ParseDeps(string(content))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, target := range rels {
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
		out = append(out, filepath.Join(parent, filepath.Base(raw)))
	}
	return out, nil
}

// hasManifest reports whether dir ships construct/base.manifest (the
// _seen_or_add layer filter).
func hasManifest(fs FS, dir string) bool {
	_, err := fs.Stat(filepath.Join(dir, "construct", "base.manifest"))
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
