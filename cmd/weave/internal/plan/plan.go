package plan

import (
	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
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
//     foundation-first and emitted (with the skill menu, below) as ONE
//     WriteFile{AGENTS.md, composeAgentsBody}. This is the @AGENTS.local.md fix.
//   - The skill menu (the agent-agnostic floor's always-on discovery face) is
//     appended as a `## Skills` section to that same AGENTS.md body. The menu
//     is computed by the IO seam (walk.GatherSkills → skill.Build) and passed
//     in; an empty menu adds nothing. Neither prose NOR a menu ⇒ no AGENTS.md
//     Action (we don't write an empty file).
//   - Symlink/Scaffold/Touch lower near-identity per intent (the dominant
//     file-op case): Symlink → Symlink{upstream/Source, Target}; Scaffold →
//     Mkdir{Target}; Touch → empty WriteFile{Target}.
//   - Tool lowers to a ToolDep{Owner, Path} (Owner = the declaring layer's
//     Path, Path = the tool path within it). The planner records the facts;
//     Apply decides derivative (append `substrate` to construct/deps) vs owner
//     (`go mod edit -tool`) by comparing Owner to the repo root. Ported from
//     ensure_go_tool_dependency.
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
func Plan(layers []layer.Layer, menu []skill.MenuItem) ([]Action, error) {
	var actions []Action

	// Prose composes across all layers, foundation-first; the skill menu (the
	// agent-agnostic floor's always-on face) appends below it — one AGENTS.md.
	var fragments []string
	for _, l := range layers {
		fragments = append(fragments, l.ProseFragments...)
	}
	if body := composeAgentsBody(fragments, menu); body != "" {
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
				// create-if-missing (setup.sh:347 `if [[ ! -f ]] then touch`).
				// Lowers to Touch (NOT WriteFile{content:""}) so Apply never
				// clobbers an existing, content-bearing file (e.g. the
				// accumulated workshop/lessons.md) — the divergence the
				// golden-diff harness surfaced.
				actions = append(actions, Touch{Path: in.Target})
			case intent.Seed:
				// TODO(part-2): create_seed is a content-tracking real-file copy
				// — it reads the upstream file's bytes (IO). Lowers to a
				// WriteFile{Target, <upstream content>} once the IO seam reads
				// the source; deferred here (the pure planner has no content).
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
			case intent.Tool:
				// Lower to ToolDep, recording the FACTS the IO seam needs:
				// Owner is this layer's absolute path (setup.sh's `upstream`),
				// Path is the tool's path within that owner module (`source`).
				// The planner stays pure — it does NOT decide derivative-vs-owner
				// (that compares Owner to the compiling repo root, which lives in
				// Apply) and does NOT compute the substrate relpath / read go.mod
				// (both IO). One ToolDep per `tool` intent, in layer order.
				actions = append(actions, ToolDep{Owner: l.Path, Path: in.Source})
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
