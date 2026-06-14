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
//     WriteFile whose live content equals weave's; a ToolDep whose substrate row
//     (derivative) / go.mod tool directive (owner) is already present. This is
//     the parity proof.
//   - EXPECTED — the pre-registered/deferred ledger: the verbs setup.sh ran that
//     weave does NOT yet lower (seed, merge→settings.json [M4]). Their live
//     output is present; weave defers; the ledger SHRINKS as each lands — `tool`
//     left it in M3 part 2 (it now lowers + classifies, MATCH/UNEXPECTED). Not a
//     failure.
//   - UNEXPECTED — anything else: a file-op verb weave mis-emits, a symlink
//     pointing somewhere different, a target setup.sh produced but weave's plan
//     diverges on. The harness FAILS on these.
package golden

import (
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/gomodx"
	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
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
// (the pre-registered ledger of verbs weave doesn't lower yet). Pure — reads
// only its Input.
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

	case plan.ToolDep:
		return classifyToolDep(root, act, obs)

	default:
		// MergeSettings is a typed PLACEHOLDER the planner does not emit today;
		// reaching here means a lowering started emitting one without the harness
		// learning to classify it. Flag loudly rather than silently pass.
		return Divergence{Unexpected, fmt.Sprintf("%T", a), "",
			"weave emitted an action the golden harness does not classify yet"}
	}
}

// classifyToolDep classifies a ToolDep against the live tree, mirroring
// applyToolDep's bimodal split (ARCH-DRY — same Owner==root test, same probe):
//
//   - Derivative (Owner != root): the probe is the live construct/deps. MATCH
//     when it carries the `substrate <rel-to-owner>` row weave would append
//     (detected via layer.ParseDeps, the same parser Apply uses for its
//     present-check). UNEXPECTED when the file is absent or the row is missing.
//   - Owner self-walk (Owner == root): the probe is the live go.mod. MATCH when
//     it already declares the `tool <module>/<Path>` directive weave would add.
//     UNEXPECTED when go.mod is absent or the directive is missing.
//
// The detail names the verb "tool" so the ledger reads consistently with the
// pre-M3 EXPECTED line it replaces.
func classifyToolDep(root string, act plan.ToolDep, obs map[string]Observed) Divergence {
	if act.Owner != root {
		// Derivative: check construct/deps for the substrate row.
		rel, err := filepath.Rel(root, act.Owner)
		if err != nil {
			return Divergence{Unexpected, "tool", "construct/deps",
				fmt.Sprintf("cannot compute relpath of %s from %s: %v", act.Owner, root, err)}
		}
		depsAbs := filepath.Join(root, "construct", "deps")
		o := obs[depsAbs]
		if !o.Exists {
			return Divergence{Unexpected, "tool", "construct/deps",
				fmt.Sprintf("weave would add `substrate %s`, but construct/deps absent in live", rel)}
		}
		if !depsHasSubstrate(o.Content, rel) {
			return Divergence{Unexpected, "tool", "construct/deps",
				fmt.Sprintf("weave would add `substrate %s`, but live construct/deps lacks it", rel)}
		}
		return Divergence{Match, "tool", "construct/deps",
			fmt.Sprintf("substrate %s present (derivative depends on tool owner)", rel)}
	}

	// Owner self-walk: check go.mod for the tool directive.
	gomodAbs := filepath.Join(root, "go.mod")
	o := obs[gomodAbs]
	module := gomodx.ModuleLine(o.Content)
	want := module + "/" + act.Path
	switch {
	case !o.Exists:
		return Divergence{Unexpected, "tool", "go.mod", "weave would add a tool directive, but go.mod absent in live"}
	case module == "":
		return Divergence{Unexpected, "tool", "go.mod", "weave would add a tool directive, but go.mod has no module line"}
	case !gomodx.HasTool(o.Content, want):
		return Divergence{Unexpected, "tool", "go.mod",
			fmt.Sprintf("weave would add `tool %s`, but live go.mod lacks it", want)}
	default:
		return Divergence{Match, "tool", "go.mod",
			fmt.Sprintf("tool %s present (owner self-walk)", want)}
	}
}

// depsHasSubstrate reports whether the construct/deps content declares a
// `substrate <rel>` row, reusing layer.ParseDeps so the harness reads the file
// the same way Apply does (ARCH-DRY).
func depsHasSubstrate(content, rel string) bool {
	rows, err := layer.ParseDeps(content)
	if err != nil {
		return false
	}
	for _, r := range rows {
		if r == rel {
			return true
		}
	}
	return false
}

// classifyDeferred ledgers one deferred-verb Intent (seed/merge) as an EXPECTED
// divergence: setup.sh produced its output, weave does not lower it yet
// (Seed=content copy [deferred], Merge=settings.json [M4]). The detail notes
// whether setup.sh's output is present in live, so the ledger reads as a
// checklist that shrinks as lowerings land. (Tool left this ledger in M3 part 2.)
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
// the deferral, for the ledger line. (tool is no longer here — it lowers to a
// ToolDep, classified by classifyAction.)
func deferredLabel(k intent.Kind) (verb, milestone string) {
	switch k {
	case intent.Seed:
		return "seed", "content-copy deferred"
	case intent.Merge:
		return "merge", "M4 settings backend"
	default:
		return "deferred", "unknown"
	}
}

// IsDeferred reports whether kind is a verb weave does not lower to a
// filesystem Action yet — the verbs the gatherer collects for the EXPECTED
// ledger. Skill is excluded: it feeds the M3 SkillIndex, not a file-op slot, so
// it has no live file-op target to ledger. Prose IS lowered (composed AGENTS.md)
// once a manifest declares it, so it is not deferred. Tool is NO LONGER deferred
// (M3 part 2): it lowers to a ToolDep, classified directly by classifyAction.
func IsDeferred(k intent.Kind) bool {
	switch k {
	case intent.Seed, intent.Merge:
		return true
	default:
		return false
	}
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
