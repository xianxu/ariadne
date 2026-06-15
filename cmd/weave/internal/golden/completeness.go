package golden

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
)

// completeness.go is the Phase-1b completeness check (M2-review I-1): an
// INDEPENDENT proof that weave's plan covers EVERY path setup.sh would produce,
// catching UNDER-production — a manifest entry setup.sh honors but weave's
// lowering silently DROPS. The golden classifier (golden.go) cannot see this:
// it classifies only the paths weave PLANS, so a verb whose lowering is a no-op
// (e.g. the pre-M5 `seed` TODO that emitted nothing) leaves no Action to
// classify and slips through as an invisible gap. This check closes that hole.
//
// Design (the simpler, sufficient option (b) from the M5 spec — enumerate the
// MANIFEST'S declared managed set, not a re-run of setup.sh):
//
//   - setup.sh's full managed set for a repo IS its walked manifest: every
//     base.manifest entry, foundation-first, that survives the self-reference
//     filter. The walk already produces exactly that as each layer's typed
//     Intents (walk.loadLayer drops the self-referential file-shape entries
//     before they land on the Layer — the SAME filter setup.sh's walk_manifest
//     applies), so the layers' Intents ARE the authoritative "what setup.sh
//     would produce" enumeration. No setup.sh re-run, no temp checkout.
//   - The check is then: does weave's PLAN (the []plan.Action) cover each Intent?
//     Crucially it enumerates from Intents (the manifest) and matches against
//     Actions (weave's output) — DIFFERENT sources — so a verb whose lowering
//     drops the entry shows up as an uncovered Intent. (Were it to re-derive the
//     expected set from the same Plan, it would be circular and catch nothing.)
//
// Pure (ARCH-PURE): a function of (layers, actions). The CLI's IO seam walks +
// plans, then hands both here. ZERO uncovered intents on the ariadne self-walk
// is the M5 "no under-production" proof.

// Uncovered is one manifest Intent that weave's plan does NOT cover — an
// under-production gap. Verb/Source/Target name the offending manifest row;
// Reason explains why no Action covers it.
type Uncovered struct {
	Verb   string
	Source string
	Target string
	Reason string
}

// CheckCompleteness enumerates every managed path the walked layers' manifests
// declare (the setup.sh-equivalent managed set) and returns the Intents weave's
// plan does NOT cover. Empty result ⇒ weave covers everything setup.sh would
// produce (no under-production). Pure: it reads only (layers, actions).
//
// Coverage per verb (how a manifest row is "covered" by an Action):
//   - symlink → a plan.Symlink with the SAME Dst (target).
//   - seed    → a plan.Seed with the same Dst.
//   - scaffold→ a plan.Mkdir with the same Path.
//   - touch   → a plan.Touch with the same Path.
//   - merge   → a plan.MergeSettings with the same Target.
//   - tool    → a plan.ToolDep (the go.mod / construct/deps edit; not a
//     path-keyed slot, so coverage is "≥1 ToolDep planned for this layer's
//     owner" — see toolCovered).
//   - prose   → composed into the one AGENTS.md (a plan.WriteFile{Path:"AGENTS.md"}).
//   - skill   → EITHER skill backend (mutually exclusive per --target): the
//     .claude/skills/<name> symlink backend (≥1 plan.Symlink under
//     .claude/skills/, the claude target) OR the AGENTS.md `## Skills` menu (the
//     codex/agy target). A `skill <dir>` row expands to N name-keyed links, so
//     coverage is "the symlink backend emitted links" OR "AGENTS.md carries the
//     menu section" — see coverIntent's intent.Skill case. Counting the menu
//     (not just any AGENTS.md write) is what lets `--target codex` — which emits
//     no symlinks — still report zero under-production.
func CheckCompleteness(layers []layer.Layer, actions []plan.Action) []Uncovered {
	idx := indexActions(actions)

	// The leaf Lₙ is the last layer (layer.Resolve emits root last + self). Only
	// intents in 𝒜(R) — every layer's exports + the leaf's internals — are PLANNED,
	// so only those must be covered. An ANCESTOR's internal is deliberately
	// excluded by the planner (not dropped), so flagging it as under-produced
	// would be a false positive. Filter to participating intents up front, using
	// the SAME selection rule the planner uses (intent.Selected, ARCH-DRY).
	leafIdx := len(layers) - 1

	var out []Uncovered
	seen := map[string]bool{} // dedup by verb+target (a verb repeated across layers)
	for i, l := range layers {
		for _, in := range l.Intents {
			if !intent.Selected(in.Visibility, i == leafIdx) {
				continue // ancestor's internal — excluded from 𝒜(R), not under-produced
			}
			verb := verbName(in.Kind)
			key := verb + "\x00" + in.Target
			if seen[key] {
				continue
			}
			seen[key] = true
			if u, missing := coverIntent(l, in, idx); missing {
				out = append(out, u)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Verb != out[j].Verb {
			return out[i].Verb < out[j].Verb
		}
		return out[i].Target < out[j].Target
	})
	return out
}

// actionIndex is the precomputed coverage sets over weave's planned Actions,
// keyed for O(1) lookup by the cover-checks.
type actionIndex struct {
	symlinkDsts map[string]bool // every plan.Symlink.Dst
	seedDsts    map[string]bool // every plan.Seed.Dst
	mkdirPaths  map[string]bool // every plan.Mkdir.Path
	touchPaths  map[string]bool // every plan.Touch.Path
	mergeTgts   map[string]bool // every plan.MergeSettings.Target
	toolOwners  map[string]bool // every plan.ToolDep.Owner
	skillLinks  int             // count of plan.Symlink under .claude/skills/
	agentsMD    bool            // a plan.WriteFile for AGENTS.md exists
	skillMenu   bool            // that AGENTS.md carries a `## Skills` menu section
}

func indexActions(actions []plan.Action) actionIndex {
	idx := actionIndex{
		symlinkDsts: map[string]bool{},
		seedDsts:    map[string]bool{},
		mkdirPaths:  map[string]bool{},
		touchPaths:  map[string]bool{},
		mergeTgts:   map[string]bool{},
		toolOwners:  map[string]bool{},
	}
	for _, a := range actions {
		switch act := a.(type) {
		case plan.Symlink:
			idx.symlinkDsts[act.Dst] = true
			if underSkills(act.Dst) {
				idx.skillLinks++
			}
		case plan.Seed:
			idx.seedDsts[act.Dst] = true
		case plan.Mkdir:
			idx.mkdirPaths[act.Path] = true
		case plan.Touch:
			idx.touchPaths[act.Path] = true
		case plan.MergeSettings:
			idx.mergeTgts[act.Target] = true
		case plan.ToolDep:
			idx.toolOwners[act.Owner] = true
		case plan.WriteFile:
			if act.Path == "AGENTS.md" {
				idx.agentsMD = true
				if strings.Contains(act.Content, "## Skills") {
					idx.skillMenu = true
				}
			}
		}
	}
	return idx
}

// coverIntent reports whether weave's plan covers one manifest Intent, returning
// the Uncovered gap when it does not. l is the declaring layer (for tool owner /
// skill backend reasoning).
func coverIntent(l layer.Layer, in intent.Intent, idx actionIndex) (Uncovered, bool) {
	mk := func(reason string) (Uncovered, bool) {
		return Uncovered{Verb: verbName(in.Kind), Source: in.Source, Target: in.Target, Reason: reason}, true
	}
	switch in.Kind {
	case intent.Symlink:
		if !idx.symlinkDsts[in.Target] {
			return mk("no plan.Symlink targets this path")
		}
	case intent.Seed:
		if !idx.seedDsts[in.Target] {
			return mk("no plan.Seed targets this path (lowering dropped the entry?)")
		}
	case intent.Scaffold:
		if !idx.mkdirPaths[in.Target] {
			return mk("no plan.Mkdir creates this dir")
		}
	case intent.Touch:
		if !idx.touchPaths[in.Target] {
			return mk("no plan.Touch creates this file")
		}
	case intent.Merge:
		if !idx.mergeTgts[in.Target] {
			return mk("no plan.MergeSettings writes this target")
		}
	case intent.Tool:
		if !idx.toolOwners[l.Path] {
			return mk("no plan.ToolDep for this layer's owner (the go.mod/construct/deps edit)")
		}
	case intent.Prose:
		if !idx.agentsMD {
			return mk("no AGENTS.md WriteFile composed (the prose fragment would be dropped)")
		}
	case intent.Skill:
		// The two skill backends are MUTUALLY EXCLUSIVE per --target, and a `skill
		// <dir>` row is covered by EITHER: the .claude/skills/<name> symlinks (the
		// claude target emits ≥1) OR the AGENTS.md `## Skills` menu section (the
		// codex/agy target). Counting the menu specifically — not just any AGENTS.md
		// write — is what lets `--target codex` report zero under-production despite
		// emitting no symlinks.
		if idx.skillLinks == 0 && !idx.skillMenu {
			return mk("neither .claude/skills symlinks nor an AGENTS.md `## Skills` menu (no skill backend produced this intent)")
		}
	}
	return Uncovered{}, false
}

// underSkills reports whether a repo-relative path is under .claude/skills/ (a
// skill-backend symlink), so the index can count those for skill coverage.
func underSkills(rel string) bool {
	return strings.HasPrefix(filepath.ToSlash(rel), ".claude/skills/")
}

// verbName maps a Kind to its manifest verb word, for the report.
func verbName(k intent.Kind) string {
	switch k {
	case intent.Symlink:
		return "symlink"
	case intent.Seed:
		return "seed"
	case intent.Scaffold:
		return "scaffold"
	case intent.Touch:
		return "touch"
	case intent.Merge:
		return "merge"
	case intent.Tool:
		return "tool"
	case intent.Prose:
		return "prose"
	case intent.Skill:
		return "skill"
	default:
		return "?"
	}
}

// RenderCompleteness formats the completeness result for one repo: a header, the
// uncovered gaps (each a failing line), and a one-line verdict. Pure (string
// in/out). ZERO uncovered ⇒ a clean "0 under-production" verdict.
func RenderCompleteness(repo string, uncovered []Uncovered) string {
	var b strings.Builder
	fmt.Fprintf(&b, "== completeness: %s ==\n", repo)
	for _, u := range uncovered {
		fmt.Fprintf(&b, "  UNDER-PRODUCED %-8s %s — %s\n", u.Verb, u.Target, u.Reason)
	}
	fmt.Fprintf(&b, "verdict: %d setup.sh-produced path(s) NOT planned by weave\n", len(uncovered))
	return b.String()
}
