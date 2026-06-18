// Package walk is weave's IO seam in front of the pure compiler core: given a
// repo root, it discovers the layer DAG topology via the shared pkg/layergraph
// walk (the SINGLE source of truth for "what is repo R's layer graph" — ARCH-DRY;
// weave and any other DAG-aware subsystem consume the same ordered roots), then
// loads each layer's base.manifest into typed Intents and its `prose` files into
// ProseFragments. Everything mutating/reading lives behind weavefs.FS (ARCH-PURE
// — intent/, layer/, and plan/ stay pure; the walk + plan.Apply are the only IO).
//
// The transitive construct/deps walk itself — deps_substrate_targets'
// repo-root-relative + absolute + present-skip resolution and discover_ancestors'
// two _seen_or_add filters (a candidate is a layer ONLY if it ships
// construct/base.manifest; the root is never its own ancestor) — now lives in
// pkg/layergraph (moved here in #115 M1). This file keeps only the rich
// per-layer load weave needs on top of that topology.
package walk

import (
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

// Walk builds the foundation-first []layer.Layer for the repo at root (which
// must be absolute). It delegates the DAG topology to layergraph.Walk (which
// canonicalizes root + the layer roots to their physical form, emits the root
// last + self-included, and dedups diamonds), then loads each layer's manifest
// and prose. Returns the layers in application order.
//
// weavefs.FS satisfies layergraph.FS structurally (a superset interface), so it
// is passed through directly — no adapter needed.
func Walk(fs weavefs.FS, root string) ([]layer.Layer, error) {
	order, err := layergraph.Walk(fs, root)
	if err != nil {
		return nil, err
	}

	// layergraph.Walk canonicalized root; loadLayer's self-reference filter
	// compares a layer's Source against root/Target, so use the same physical
	// root the walk resolved its layer paths against (order[len-1] is root,
	// canonicalized + emitted last).
	canonRoot := root
	if len(order) > 0 {
		canonRoot = order[len(order)-1]
	}

	layers := make([]layer.Layer, 0, len(order))
	for _, dir := range order {
		l, err := loadLayer(fs, canonRoot, dir)
		if err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}
	return layers, nil
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
		// guaranteed it via the base.manifest filter for ancestors; root may
		// lack one).
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
		//   - merge — not file-shape (rename), per setup.sh.
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
// read-only intents (prose/skill) and the rename intent (merge) are NOT
// file-shape and bypass the filter — see loadLayer.
func isFileShape(k intent.Kind) bool {
	switch k {
	case intent.Symlink, intent.Seed, intent.Scaffold, intent.Touch:
		return true
	default:
		return false
	}
}
