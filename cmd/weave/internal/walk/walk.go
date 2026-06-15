// Package walk is weave's IO seam in front of the pure compiler core: given a
// repo root, it discovers the layer DAG by following each layer's
// construct/deps, resolves it foundation-first (layer.Resolve), then loads each
// layer's base.manifest into typed Intents and its `prose` files into
// ProseFragments. Everything mutating/reading lives behind weavefs.FS (ARCH-PURE
// — intent/, layer/, and plan/ stay pure; the walk + plan.Apply are the only IO).
//
// It ports two shell behaviors verbatim (ARCH-DRY; the part-3 golden-diff checks
// parity):
//   - deps_substrate_targets (lib-deps.sh): per `substrate` row, repo-root-
//     relative OR absolute-path resolution, then physical-path canonicalization
//     (pwd -P) with present-skip — an unresolvable parent is dropped silently.
//   - discover_ancestors' two _seen_or_add filters (setup.sh): a candidate
//     counts as a layer ONLY if it ships construct/base.manifest, and the target
//     repo is never its own ancestor. weave does NOT consult go.mod / go list
//     for edges (M1 step 9 — construct/deps is the sole edge source).
package walk

import (
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Walk builds the foundation-first []layer.Layer for the repo at root (which
// must be absolute). It discovers the edge set via construct/deps, resolves the
// DAG (root emitted last + self-included), then loads each layer's manifest and
// prose. Returns the layers in application order.
func Walk(fs weavefs.FS, root string) ([]layer.Layer, error) {
	root = physical(root) // canonicalize so self-comparisons match (pwd -P)

	edges, err := discoverEdges(fs, root)
	if err != nil {
		return nil, err
	}

	order, err := layer.Resolve(root, edges)
	if err != nil {
		return nil, err
	}

	layers := make([]layer.Layer, 0, len(order))
	for _, dir := range order {
		l, err := loadLayer(fs, root, dir)
		if err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}
	return layers, nil
}

// discoverEdges follows construct/deps transitively from root, returning the
// edge map layer.Resolve consumes: each discovered layer dir → the layer dirs
// it depends on (substrate targets that pass the _seen_or_add layer filter).
// A BFS mirroring discover_ancestors' substrate walk, but construct/deps-only.
func discoverEdges(fs weavefs.FS, root string) (map[string][]string, error) {
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
func substrateTargets(fs weavefs.FS, repoRoot string) ([]string, error) {
	depsPath := filepath.Join(repoRoot, "construct", "deps")
	content, err := fs.ReadFile(depsPath)
	if err != nil {
		return nil, nil // no construct/deps ⇒ no edges (lib-deps: [[ -f ]] || return 0)
	}
	rels, err := layer.ParseDeps(string(content))
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

// loadLayer reads dir's base.manifest into Intents (dropping self-reference
// entries) and each `prose` intent's file into ProseFragments. root is the
// consuming repo root, against which a Target resolves (a Source resolves
// against dir) for the self-reference comparison.
func loadLayer(fs weavefs.FS, root, dir string) (layer.Layer, error) {
	l := layer.Layer{Name: filepath.Base(dir), Path: dir}

	manifest, err := fs.ReadFile(filepath.Join(dir, "construct", "base.manifest"))
	if err != nil {
		// No manifest: a layer with no intents (the Resolve order already
		// guaranteed it via hasManifest for ancestors; root may lack one).
		return l, nil
	}
	intents, err := intent.ParseManifest(string(manifest))
	if err != nil {
		return l, err
	}

	for _, in := range intents {
		// Self-reference filter (walk_manifest:315): skip a FILE-SHAPE entry
		// whose upstream source == target target (would symlink/copy a file
		// onto its own canonical location, destroying it). It applies only to
		// the destructive file-ops; the read-only / rename intents bypass it:
		//   - merge/tool — not file-shape (rename / go.mod edit), per setup.sh.
		//   - prose/skill — weave's read-only semantic intents: the Source file
		//     is only READ (composed into AGENTS.md / indexed), never written
		//     to its own slot, so a self-walk MUST keep them — that is exactly
		//     how a repo contributes its OWN prose (the @AGENTS.local.md fix).
		if isFileShape(in.Kind) &&
			filepath.Join(dir, in.Source) == filepath.Join(root, in.Target) {
			continue
		}
		l.Intents = append(l.Intents, in)
		if in.Kind == intent.Prose {
			frag, err := fs.ReadFile(filepath.Join(dir, in.Source))
			if err != nil {
				// A declared-but-absent prose fragment is skipped (the layer
				// declared prose it doesn't ship); don't abort the compile.
				continue
			}
			// Tag the fragment with the intent's visibility (export|internal),
			// so the Planner can select 𝒜(R) — every layer's export prose plus
			// the leaf's internal prose only (the visibility axis, #99).
			l.ProseFragments = append(l.ProseFragments,
				layer.ProseFragment{Visibility: in.Visibility, Content: string(frag)})
		}
	}
	return l, nil
}

// isFileShape reports whether kind is a destructive file-shape op the
// self-reference filter guards (symlink/seed/scaffold/touch). The semantic
// read-only intents (prose/skill) and the rename/edit intents (merge/tool) are
// NOT file-shape and bypass the filter — see loadLayer.
func isFileShape(k intent.Kind) bool {
	switch k {
	case intent.Symlink, intent.Seed, intent.Scaffold, intent.Touch:
		return true
	default:
		return false
	}
}

// hasManifest reports whether dir ships construct/base.manifest (the
// _seen_or_add layer filter).
func hasManifest(fs weavefs.FS, dir string) bool {
	_, err := fs.Stat(filepath.Join(dir, "construct", "base.manifest"))
	return err == nil
}

// physical canonicalizes path to its physical form (EvalSymlinks ≈ pwd -P),
// falling back to the input when it can't be resolved (e.g. doesn't exist yet).
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
