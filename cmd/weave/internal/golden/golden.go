// Package golden is weave's golden-diff harness core: a PURE classifier that
// proves weave's intended file-op output matches what construct/setup.sh already
// produced on the live repos. The live repos' current on-disk state IS setup.sh's
// output (they were set up by it), so the harness compares weave's INTENDED
// actions (a dry-run Plan, never applied) against the live filesystem and
// classifies every divergence (ARCH-PURE — the classifier takes a SNAPSHOT of
// observed state, never touches disk; the IO gatherer in the weave CLI fills the
// snapshot).
//
// Three classes, per the M2 "Done when" + the plan's ## Revisions
// pre-registered divergences:
//
//   - MATCH — weave's action already realized in live: a Symlink whose live link
//     points exactly where weave would link; a Mkdir whose dir exists; a
//     WriteFile whose live content equals weave's; a MergeSettings whose
//     recomputed merge SEMANTICALLY equals the live settings.json. This is the
//     parity proof.
//   - EXPECTED — the pre-registered/deferred ledger: the verbs setup.sh ran that
//     weave does NOT yet lower. The ledger SHRANK as each landed — `merge` in M4
//     and `seed` in M5 — so as of M5 it is EMPTY (every setup.sh file-op verb now
//     lowers + classifies MATCH/UNEXPECTED). The `tool` verb is RETIRED, not
//     deferred (#95 M5: ownership is location-based, weave never edits go.mod).
//     The class is retained for any future deferred verb. Not a failure.
//   - UNEXPECTED — anything else: a file-op verb weave mis-emits, a symlink
//     pointing somewhere different, a target setup.sh produced but weave's plan
//     diverges on. The harness FAILS on these.
package golden

import (
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/settingsx"
)

// Class is a divergence classification.
type Class int

const (
	// Match — weave's intended op is already realized in the live tree.
	Match Class = iota
	// Expected — a pre-registered/deferred divergence (setup.sh did it; weave
	// defers). Ledgered, not a failure.
	Expected
	// Unexpected — a real divergence; the harness fails on any of these.
	Unexpected
)

func (c Class) String() string {
	switch c {
	case Match:
		return "MATCH"
	case Expected:
		return "EXPECTED"
	case Unexpected:
		return "UNEXPECTED"
	default:
		return "?"
	}
}

// Observed is a snapshot of the live filesystem at one target path. The IO
// gatherer fills it (Lstat/Readlink/ReadFile); the classifier reads it. Keeping
// observation in a plain struct is what makes the classifier pure (ARCH-PURE):
// classification is a pure function of (planned action, observed state).
type Observed struct {
	Exists     bool
	IsSymlink  bool
	IsDir      bool
	LinkTarget string // result of Readlink, when IsSymlink
	Content    string // file bytes, when a WriteFile target is compared
}

// Input is everything the classifier needs for one repo: weave's planned
// Actions, the deferred-verb Intents weave does NOT lower yet (so they can be
// ledgered EXPECTED rather than silently dropped), and a snapshot of the live
// FS keyed by ABSOLUTE target path. RepoRoot is the live repo's absolute root,
// against which each repo-relative action/intent Target is resolved to look up
// Observed.
type Input struct {
	RepoRoot string
	Actions  []plan.Action
	Deferred []intent.Intent
	Observed map[string]Observed
}

// Divergence is one classified probe: a verb acting on a repo-relative Path,
// its Class, and a human-readable Detail explaining the classification (the
// ledger line).
type Divergence struct {
	Class  Class
	Verb   string // symlink | mkdir | writefile | seed | merge | tool
	Path   string // repo-relative target
	Detail string
}

// Classify is the pure core: it turns an Input into the per-repo divergence
// ledger. Each weave Action becomes one Divergence (MATCH or UNEXPECTED against
// the observed live state); each deferred Intent becomes one EXPECTED Divergence
// (the pre-registered ledger of verbs weave doesn't lower yet — now just seed).
// Pure — reads only its Input.
func Classify(in Input) []Divergence {
	var divs []Divergence

	for _, a := range in.Actions {
		divs = append(divs, classifyAction(in.RepoRoot, a, in.Observed))
	}
	for _, intn := range in.Deferred {
		divs = append(divs, classifyDeferred(in.RepoRoot, intn, in.Observed))
	}
	return divs
}

// classifyAction classifies one weave Action against the observed state at its
// target. MATCH when the live tree already realizes the action; UNEXPECTED
// otherwise (a real, harness-failing divergence).
func classifyAction(root string, a plan.Action, obs map[string]Observed) Divergence {
	switch act := a.(type) {
	case plan.Symlink:
		abs := filepath.Join(root, act.Dst)
		o := obs[abs]
		// weave's Apply computes a RELATIVE link target: rel(dir(dst), src).
		// Live MATCH requires an existing symlink pointing at that exact rel.
		want, err := filepath.Rel(filepath.Dir(abs), act.Src)
		if err != nil {
			return Divergence{Unexpected, "symlink", act.Dst,
				fmt.Sprintf("cannot compute relpath of %s from %s: %v", act.Src, filepath.Dir(abs), err)}
		}
		switch {
		case !o.Exists:
			return Divergence{Unexpected, "symlink", act.Dst,
				fmt.Sprintf("weave would link -> %s, but nothing present in live", want)}
		case !o.IsSymlink:
			return Divergence{Unexpected, "symlink", act.Dst,
				"weave would symlink, but a non-symlink (regular file/dir) occupies the slot in live"}
		case o.LinkTarget != want:
			return Divergence{Unexpected, "symlink", act.Dst,
				fmt.Sprintf("live link -> %s, weave would link -> %s", o.LinkTarget, want)}
		default:
			return Divergence{Match, "symlink", act.Dst,
				fmt.Sprintf("link -> %s", want)}
		}

	case plan.Mkdir:
		abs := filepath.Join(root, act.Path)
		o := obs[abs]
		switch {
		case !o.Exists:
			return Divergence{Unexpected, "mkdir", act.Path, "weave would mkdir, but dir absent in live"}
		case !o.IsDir:
			return Divergence{Unexpected, "mkdir", act.Path, "weave would mkdir, but a non-dir occupies the slot in live"}
		default:
			return Divergence{Match, "mkdir", act.Path, "dir exists"}
		}

	case plan.Touch:
		abs := filepath.Join(root, act.Path)
		o := obs[abs]
		// Touch is create-if-missing (setup.sh:347): mere EXISTENCE satisfies it,
		// regardless of content (a Touch target like workshop/lessons.md
		// accumulates real content over time). Only an ABSENT target diverges.
		if !o.Exists {
			return Divergence{Unexpected, "touch", act.Path, "weave would create-if-missing, but file absent in live"}
		}
		return Divergence{Match, "touch", act.Path, "file exists (touch is create-if-missing)"}

	case plan.Seed:
		// Seed is a content-TRACKING copy (create_seed): MATCH iff the live
		// target exists AND its content equals the upstream SOURCE's content
		// (what applySeed would write). The probe is TWO files: the target
		// (Dst, root-relative) and the source (Src, an ABSOLUTE upstream path,
		// keyed by its own abs path in Observed — Gather reads both).
		//   - Absent source → mirrors applySeed's non-fatal skip: weave would do
		//     nothing, so a present-or-absent target is not a divergence; MATCH
		//     with a note. (We can't fault a target weave wouldn't touch.)
		//   - Source present, target absent → UNEXPECTED (weave would seed it,
		//     but setup.sh's output isn't there — a real gap).
		//   - Source present, content drift → UNEXPECTED (weave would refresh to
		//     the source bytes; live differs).
		//   - Source present, content equal → MATCH.
		dstAbs := filepath.Join(root, act.Dst)
		dstO := obs[dstAbs]
		srcO := obs[act.Src]
		if !srcO.Exists {
			return Divergence{Match, "seed", act.Dst,
				"upstream source absent — weave would skip (non-fatal), nothing to diverge"}
		}
		switch {
		case !dstO.Exists:
			return Divergence{Unexpected, "seed", act.Dst,
				"weave would seed (copy upstream content), but the target is absent in live"}
		case dstO.Content != srcO.Content:
			return Divergence{Unexpected, "seed", act.Dst,
				fmt.Sprintf("content drift (live %d bytes, upstream source %d bytes)", len(dstO.Content), len(srcO.Content))}
		default:
			return Divergence{Match, "seed", act.Dst, "content matches upstream source"}
		}

	case plan.WriteFile:
		abs := filepath.Join(root, act.Path)
		o := obs[abs]
		switch {
		case !o.Exists:
			return Divergence{Unexpected, "writefile", act.Path, "weave would write, but file absent in live"}
		case o.Content != act.Content:
			return Divergence{Unexpected, "writefile", act.Path,
				fmt.Sprintf("content drift (live %d bytes, weave %d bytes)", len(o.Content), len(act.Content))}
		default:
			return Divergence{Match, "writefile", act.Path, "content matches"}
		}

	case plan.MergeSettings:
		return classifyMergeSettings(root, act, obs)

	default:
		// Reaching here means a lowering started emitting an Action the harness
		// does not classify yet. Flag loudly rather than silently pass.
		return Divergence{Unexpected, fmt.Sprintf("%T", a), "",
			"weave emitted an action the golden harness does not classify yet"}
	}
}

// classifyMergeSettings classifies a MergeSettings against the live tree. The
// probe is THREE observed files: the base (act.Source), the optional sibling
// local (<dir(Target)>/settings.local.json), and the live target (act.Target —
// which IS merge-settings.sh's output). The classifier RECOMPUTES weave's merge
// from the observed base + local (settingsx.Merge — the same pure port Apply
// uses, ARCH-DRY) and SEMANTICALLY compares it to the live target:
//
//   - MATCH iff the live settings.json parses + deep-equals weave's merge output.
//     The compare is on PARSED JSON, NOT bytes — merge-settings.sh (jq/python)
//     key ordering need not match weave's, and a semantically-equal file is not a
//     divergence.
//   - UNEXPECTED when the base is absent (a setup/port error), the live target is
//     absent, weave's merge errors, or the two are not semantically equal.
//
// The local file's presence is read from Observed: an absent/empty local takes
// settingsx.Merge's local-absent path (base with meta keys stripped).
func classifyMergeSettings(root string, act plan.MergeSettings, obs map[string]Observed) Divergence {
	baseAbs := filepath.Join(root, act.Source)
	targetAbs := filepath.Join(root, act.Target)
	localAbs := filepath.Join(filepath.Dir(targetAbs), "settings.local.json")

	baseO := obs[baseAbs]
	if !baseO.Exists {
		return Divergence{Unexpected, "merge", act.Target,
			fmt.Sprintf("weave would merge %s, but base %s absent in live", act.Target, act.Source)}
	}
	targetO := obs[targetAbs]
	if !targetO.Exists {
		return Divergence{Unexpected, "merge", act.Target,
			"weave would write the merged settings, but the target is absent in live"}
	}

	var local []byte
	if localO := obs[localAbs]; localO.Exists {
		local = []byte(localO.Content)
	}
	merged, err := settingsx.Merge([]byte(baseO.Content), local)
	if err != nil {
		return Divergence{Unexpected, "merge", act.Target,
			fmt.Sprintf("weave's merge failed: %v", err)}
	}

	eq, err := settingsx.SemanticEqual([]byte(targetO.Content), merged)
	if err != nil {
		return Divergence{Unexpected, "merge", act.Target,
			fmt.Sprintf("cannot compare live target to weave's merge (parse error): %v", err)}
	}
	if !eq {
		return Divergence{Unexpected, "merge", act.Target,
			"live settings.json is NOT semantically equal to weave's merge output (a port gap)"}
	}
	return Divergence{Match, "merge", act.Target,
		"merged settings.json semantically equals weave's merge output"}
}

// classifyDeferred ledgers one deferred-verb Intent as an EXPECTED divergence:
// setup.sh produced its output, weave does not lower it yet. The detail notes
// whether setup.sh's output is present in live, so the ledger reads as a
// checklist that shrinks as lowerings land. As of M5 the ledger is EMPTY —
// every setup.sh verb now lowers + classifies (merge→MergeSettings in M4,
// seed→Seed in M5; the `tool` verb is RETIRED, not deferred). The function is
// retained for the classifier's generic shape (any future deferred verb
// re-enters here via IsDeferred); deferredLabel falls through to a generic label.
func classifyDeferred(root string, in intent.Intent, obs map[string]Observed) Divergence {
	verb, milestone := deferredLabel(in.Kind)
	abs := filepath.Join(root, in.Target)
	present := obs[abs].Exists
	detail := fmt.Sprintf("weave defers %s (%s)", verb, milestone)
	if present {
		detail += "; setup.sh output present in live"
	} else {
		detail += "; not present in live (nothing for weave to omit)"
	}
	return Divergence{Expected, verb, in.Target, detail}
}

// deferredLabel returns the verb name + the milestone/status that will retire
// the deferral, for the ledger line. As of M5 NO setup.sh verb is deferred —
// merge lowers to a MergeSettings, seed to a Seed, all classified by
// classifyAction (and the `tool` verb is retired, not deferred). It is left as a
// generic fallthrough so a future deferred verb can re-register here.
func deferredLabel(k intent.Kind) (verb, milestone string) {
	return "deferred", "unknown"
}

// IsDeferred reports whether kind is a verb weave does not lower to a
// filesystem Action yet — the verbs the gatherer collects for the EXPECTED
// ledger. As of M5 NOTHING is deferred: every setup.sh file-op verb lowers to
// an Action classified by classifyAction (symlink/scaffold/touch since M2,
// merge→MergeSettings in M4, seed→Seed in M5; the `tool` verb is retired). Skill
// is excluded (it feeds the SkillIndex, not a file-op slot) and Prose IS lowered
// (the composed AGENTS.md) — neither is "deferred". The function is retained so a
// future deferred verb can re-enter the EXPECTED ledger by returning true here.
func IsDeferred(k intent.Kind) bool {
	return false
}

// HasUnexpected reports whether any divergence is UNEXPECTED — the harness's
// fail signal (the gatherer/subcommand exits non-zero on true).
func HasUnexpected(divs []Divergence) bool {
	for _, d := range divs {
		if d.Class == Unexpected {
			return true
		}
	}
	return false
}
