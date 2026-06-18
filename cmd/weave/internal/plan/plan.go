package plan

import (
	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
)

// Plan lowers a foundation-first []layer.Layer into the ordered []Action that
// realizes the composed agentic context. Pure: it computes Actions from
// in-memory Layers and never touches disk (ARCH-PURE); a later IO seam (part 2)
// applies them. Layers arrive in resolved order (foundation first, the
// consuming repo last and self-included — see layergraph.Resolve / layer.Layer).
//
// Lowering is one switch over intent.Kind, ported from walk_manifest's dispatch
// (ARCH-DRY — construct/setup.sh:320):
//
//   - Prose composes ACROSS layers under the visibility algebra (#99): every
//     layer's EXPORT prose foundation-first, then the LEAF's INTERNAL prose LAST
//     (an ancestor's internal prose is excluded — that's the @AGENTS.local.md /
//     parley bug fix). The composed prose is written to EACH per-harness ENTRY
//     FILE (Option B, #107): the caller passes entryFiles (CLAUDE.md / AGENTS.md /
//     GEMINI.md for the Union; one for a lean --target). There is NO `## Skills`
//     menu — every harness discovers its skill DIR natively (skills are lowered
//     separately as <dir>/<name> symlinks). Empty prose ⇒ no entry-file Action.
//   - Symlink/Scaffold/Touch/Seed lower near-identity per intent (the dominant
//     file-op case): Symlink → Symlink{upstream/Source, Target}; Scaffold →
//     Mkdir{Target}; Touch → empty WriteFile{Target}; Seed →
//     Seed{upstream/Source, Target} (a content-tracking real-file copy whose
//     bytes the IO seam reads from the upstream source — see plan.applySeed).
//   - Merge lowers to a MergeSettings{Source, Target} — the settings cascade
//     (ported from setup.sh's `merge` case). The planner records the path facts;
//     Apply reads Source + the sibling settings.local.json off disk and runs
//     settingsx.Merge (the merge-settings.sh port) to write Target.
//   - Skill is DEFERRED (M3 skill serving): it emits no Action and must not
//     error — a manifest carrying it still compiles. Skill feeds the SkillIndex
//     (the menu), not the filesystem-op list.
//
// DEFERRED to the part-2 IO walk (NOT this pure unit), per the plan's M2
// carry-forward notes:
//   - The self-reference filter (walk_manifest:315 — skip an entry whose
//     upstream/source == target/target on a self-walk). It needs absolute
//     on-disk paths the pure planner doesn't resolve; Resolve emits root last
//     and self, so the IO walk knows which layer is the self-walk.
//   - The two _seen_or_add filters (base.manifest-existence, target-self-
//     exclusion) and substrate path resolution (repo-root-relative + absolute +
//     present-skip, ported from deps_substrate_targets). All IO concerns.
func Plan(layers []layer.Layer, entryFiles []string) ([]Action, error) {
	var actions []Action

	// The leaf Lₙ is the LAST layer (layergraph.Resolve emits root last + self-
	// included). 𝒜(R) selects every layer's EXPORTS plus the LEAF's INTERNALS
	// only — an ancestor's internal artifacts never reach R (the visibility axis,
	// workshop/targets/base-layer-mechanics.md). leafIdx anchors both the
	// prose composition and the per-intent export/leaf filter below.
	leafIdx := len(layers) - 1

	// Prose composes across all layers per the algebra:
	//   prose(R) = ⟦export-prose(L₀)⟧ ∥ … ∥ ⟦export-prose(Lₙ)⟧ ∥ ⟦internal-prose(Lₙ)⟧
	// i.e. every layer's EXPORT prose foundation-first, then the LEAF's INTERNAL
	// prose LAST. Ancestor internal prose is excluded; leaf internal is included
	// last. The one composition is fanned to each per-harness entry file (#107);
	// there is no skill menu (each harness discovers its skill dir natively).
	var fragments []string
	for _, l := range layers { // export prose, foundation-first (incl. the leaf's export)
		for _, f := range l.ProseFragments {
			if f.Visibility == intent.Export {
				fragments = append(fragments, f.Content)
			}
		}
	}
	if leafIdx >= 0 { // the leaf's internal prose LAST (excludes every ancestor's)
		for _, f := range layers[leafIdx].ProseFragments {
			if f.Visibility == intent.Internal {
				fragments = append(fragments, f.Content)
			}
		}
	}
	if body := composeProse(fragments); body != "" {
		for _, ef := range entryFiles { // one composition, fanned to each per-harness entry file
			actions = append(actions, WriteFile{Path: ef, Content: body})
		}
	}

	// File-op intents lower per intent, in layer (foundation-first) order, under
	// the SAME 𝒜(R) filter: an intent participates iff it is an EXPORT or it
	// belongs to the LEAF (so an ancestor's internal is excluded; the leaf's
	// internal is included). Today every non-prose intent is export, so this is
	// behavior-preserving — but the filter must be uniform across kinds (the
	// composition algebra is type-uniform; visibility picks the operands).
	for i, l := range layers {
		for _, in := range l.Intents {
			if !participates(in.Visibility, i, leafIdx) {
				continue
			}
			switch in.Kind {
			case intent.Symlink:
				// create_symlink "$upstream/$source" "$TARGET_DIR/$target"
				actions = append(actions, Symlink{Src: joinPath(l.Path, in.Source), Dst: in.Target})
			case intent.Scaffold:
				// create_scaffold "$TARGET_DIR/$target"
				actions = append(actions, Mkdir{Path: in.Target})
			case intent.Touch:
				// create-if-missing (setup.sh:347 `if [[ ! -f ]] then touch`).
				// Lowers to Touch (NOT WriteFile{content:""}) so Apply never
				// clobbers an existing, content-bearing file (e.g. the
				// accumulated workshop/lessons.md) — the divergence the
				// golden-diff harness surfaced.
				actions = append(actions, Touch{Path: in.Target})
			case intent.Seed:
				// create_seed "$upstream/$source" "$TARGET_DIR/$target": a
				// content-tracking real-file copy. The planner records only the
				// path FACTS (Src = absolute upstream path, Dst = target-relative)
				// — it does NOT read the upstream bytes (ARCH-PURE; the bytes live
				// on disk and are read by applySeed in the IO seam). This mirrors
				// Symlink's lowering (same joinPath(l.Path, in.Source) for the
				// absolute source); applySeed does the content-compare + write.
				actions = append(actions, Seed{Src: joinPath(l.Path, in.Source), Dst: in.Target})
			case intent.Prose:
				// Handled above (composes across layers); nothing per-intent.
			case intent.Merge:
				// The settings cascade (setup.sh's `merge` case): lower to a
				// MergeSettings{Source, Target}. Source is the layer's base
				// settings (settings.ariadne.json), Target the composed
				// settings.json. The planner records only the path facts (pure);
				// Apply reads Source + the sibling settings.local.json off disk,
				// runs settingsx.Merge (the merge-settings.sh port), writes Target.
				actions = append(actions, MergeSettings{Source: in.Source, Target: in.Target})
			case intent.Skill:
				// TODO(M3): feeds the SkillIndex (agent-agnostic skill serving),
				// not the filesystem-op list. No Action here.
			}
		}
	}

	return actions, nil
}

// participates reports whether an intent at layer index i (in foundation-first
// order, leafIdx = the leaf Lₙ) is in the selected multiset 𝒜(R). It delegates to
// intent.Selected (the single source of truth for the visibility-axis rule,
// ARCH-DRY) — the type picks the compose operator, visibility picks the operands
// (workshop/targets/base-layer-mechanics.md).
func participates(v intent.Visibility, i, leafIdx int) bool {
	return intent.Selected(v, i == leafIdx)
}

// joinPath joins an upstream layer path and a source relpath with a single
// separator. It is a pure string join — NOT filepath.Join — because the
// planner must stay IO-free (ARCH-PURE) and path cleaning/abs-resolution is the
// IO seam's job. setup.sh likewise just string-concatenates "$upstream/$source".
func joinPath(base, rel string) string {
	if base == "" {
		return rel
	}
	if rel == "" {
		return base
	}
	return base + "/" + rel
}
