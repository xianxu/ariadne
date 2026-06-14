package plan

import (
	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
)

// Plan lowers a foundation-first []layer.Layer into the ordered []Action that
// realizes the composed agentic context. Pure: it computes Actions from
// in-memory Layers and never touches disk (ARCH-PURE); a later IO seam (part 2)
// applies them. Layers arrive in resolved order (foundation first, the
// consuming repo last and self-included — see layer.Resolve / layer.Layer).
//
// Lowering is one switch over intent.Kind, ported from walk_manifest's dispatch
// (ARCH-DRY — construct/setup.sh:320):
//
//   - Prose composes ACROSS layers: every layer's ProseFragments are gathered
//     foundation-first and emitted as ONE WriteFile{AGENTS.md, composeProse}.
//     This is the @AGENTS.local.md fix. No fragments anywhere ⇒ no AGENTS.md
//     Action (we don't write an empty file).
//   - Symlink/Scaffold/Touch lower near-identity per intent (the dominant
//     file-op case): Symlink → Symlink{upstream/Source, Target}; Scaffold →
//     Mkdir{Target}; Touch → empty WriteFile{Target}.
//   - Merge/Tool/Skill are DEFERRED (Merge=M4 settings cascade, Tool=the exec
//     seam, Skill=M3 skill serving). They emit no Action here and must not
//     error — a manifest carrying them still compiles. Their landing types are
//     MergeSettings / ToolDep (action.go); Skill has no Action yet (it feeds the
//     M3 SkillIndex, not the filesystem-op list).
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
func Plan(layers []layer.Layer) ([]Action, error) {
	var actions []Action

	// Prose composes across all layers, foundation-first → one AGENTS.md.
	var fragments []string
	for _, l := range layers {
		fragments = append(fragments, l.ProseFragments...)
	}
	if body := composeProse(fragments); body != "" {
		actions = append(actions, WriteFile{Path: "AGENTS.md", Content: body})
	}

	// File-op intents lower per intent, in layer (foundation-first) order.
	for _, l := range layers {
		for _, in := range l.Intents {
			switch in.Kind {
			case intent.Symlink:
				// create_symlink "$upstream/$source" "$TARGET_DIR/$target"
				actions = append(actions, Symlink{Src: joinPath(l.Path, in.Source), Dst: in.Target})
			case intent.Scaffold:
				// create_scaffold "$TARGET_DIR/$target"
				actions = append(actions, Mkdir{Path: in.Target})
			case intent.Touch:
				// touch "$TARGET_DIR/$source" (target defaults to source)
				actions = append(actions, WriteFile{Path: in.Target, Content: ""})
			case intent.Seed:
				// TODO(part-2): create_seed is a content-tracking real-file copy
				// — it reads the upstream file's bytes (IO). Lowers to a
				// WriteFile{Target, <upstream content>} once the IO seam reads
				// the source; deferred here (the pure planner has no content).
			case intent.Prose:
				// Handled above (composes across layers); nothing per-intent.
			case intent.Merge:
				// TODO(M4): the settings cascade — lower to MergeSettings via
				// the settings backend (ports merge-settings.sh). Not this unit.
			case intent.Tool:
				// TODO(later): lower to ToolDep via the exec seam (ports
				// ensure_go_tool_dependency). Not this unit.
			case intent.Skill:
				// TODO(M3): feeds the SkillIndex (agent-agnostic skill serving),
				// not the filesystem-op list. No Action here.
			}
		}
	}

	return actions, nil
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
