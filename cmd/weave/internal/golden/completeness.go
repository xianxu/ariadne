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
//   - prose   → composed into SOME per-harness entry file (a plan.WriteFile for
//     CLAUDE.md/AGENTS.md/GEMINI.md — idx.entryFile).
//   - skill   → the per-harness skill-DIR symlinks (Option B, #107): ≥1
//     plan.Symlink under .claude/skills (claude) and/or .agents/skills
//     (codex/gemini). No menu backend. A `skill <dir>` row expands to N name-keyed
//     links; every target emits links into its own skill dir, so a covered intent
//     yields skillLinks ≥ 1 regardless of which target ran — see coverIntent's
//     intent.Skill case. This is what lets BOTH the Union and a lean `--target`
//     report zero under-production.
func CheckCompleteness(layers []layer.Layer, actions []plan.Action) []Uncovered {
	idx := indexActions(actions)

	// The leaf Lₙ is the last layer (layergraph.Resolve emits root last + self). Only
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
			if u, missing := coverIntent(in, idx); missing {
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
	skillLinks  int             // count of plan.Symlink under a per-harness skill dir
	entryFile   bool            // a plan.WriteFile for SOME per-harness entry file exists
}

func indexActions(actions []plan.Action) actionIndex {
	idx := actionIndex{
		symlinkDsts: map[string]bool{},
		seedDsts:    map[string]bool{},
		mkdirPaths:  map[string]bool{},
		touchPaths:  map[string]bool{},
		mergeTgts:   map[string]bool{},
	}
	// The per-harness entry files (Option B): prose is covered if it lands in ANY of
	// them (CLAUDE.md / AGENTS.md / GEMINI.md). Reuse the face registry as the single
	// source of truth (ARCH-DRY).
	entryFiles := map[string]bool{}
	for _, ef := range plan.TargetAll.EntryFiles() {
		entryFiles[ef] = true
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
		case plan.WriteFile:
			if entryFiles[act.Path] {
				idx.entryFile = true
			}
		}
	}
	return idx
}

// coverIntent reports whether weave's plan covers one manifest Intent, returning
// the Uncovered gap when it does not.
func coverIntent(in intent.Intent, idx actionIndex) (Uncovered, bool) {
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
	case intent.Prose:
		if !idx.entryFile {
			return mk("no per-harness entry-file WriteFile composed (the prose fragment would be dropped)")
		}
	case intent.Skill:
		// Option B (#107): a `skill <dir>` row is covered by the per-harness skill DIR
		// symlinks — .claude/skills (claude) and/or .agents/skills (codex/gemini). No
		// menu backend. Every target emits ≥1 skill symlink (into its own dir), so a
		// covered intent yields skillLinks ≥ 1 regardless of which target ran.
		if idx.skillLinks == 0 {
			return mk("no per-harness skill-dir symlinks produced this intent")
		}
	}
	return Uncovered{}, false
}

// underSkills reports whether a repo-relative path is under ANY per-harness skill
// dir (.claude/skills or .agents/skills), so the index can count those symlinks for
// skill coverage. Reuses the face registry (ARCH-DRY).
func underSkills(rel string) bool {
	s := filepath.ToSlash(rel)
	for _, dir := range plan.TargetAll.SkillDirs() {
		if strings.HasPrefix(s, dir+"/") {
			return true
		}
	}
	return false
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
