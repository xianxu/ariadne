// project — the vocabulary of a project: its data shape, its lifecycle funnel,
// and the laws the funnel must satisfy. Single source of truth for the `project`
// noun (ariadne#180). sdlc reads the exported JSON; humans and the LLM read this
// file directly. The prose companion (construct/datatype/project.md) cites this
// file as schema authority — a drift test binds the two.
//
// The organizing insight (#180 Spec): the project lifecycle is the issue
// lifecycle one level up. A project is a structured, TIME-BOUND push for a major
// change, across repos — not merely a container of issues; it carries a deadline
// set at commit.
package project

import "list"

// ── categories: the single concrete source of status membership ──
// forming   = pre-baseline (no deadline yet)
// committed = baseline set (deadline + planned finish), not yet broken down
// executing = broken down, live portfolio (paused keeps its baseline)
// terminal  = closed
categories: {
	forming:   ["ideation", "defined"]
	committed: ["committed"]
	executing: ["executing", "paused"]
	terminal:  ["done", "dropped"]
}

#Forming:  or(categories.forming)
#Terminal: or(categories.terminal)
#Status:   or(list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]))

// ── when: one-line semantics per status (the documented-value source) ──
when: {
	ideation:  "idea captured; PRD not yet written (ideation lives in parley, linked via sources)"
	defined:   "PRD exists in the project file; not yet committed to a timeline"
	committed: "baseline set (deadline + planned finish + parallelism intent); not yet broken down"
	executing: "PRD broken down into issues across repos; work in flight"
	paused:    "execution suspended; committed baseline stays intact"
	done:      "done_when met; retro + fog-factor ledger row recorded"
	dropped:   "no longer worth pursuing"
}

// ── discovery: where instances of this noun live, PER REPO. The cue declares
// the per-repo home; cross-repo resolution owns the walk across peers (#171) —
// same division of labor as resolveRepoDir. Repo-relative. ──
discovery: {
	home: "workshop/projects" // repo-relative home for project instances
	glob: "*.md"
	// archive: done/dropped projects move under here (per-kind subdir derived in
	// Go by pkg/vocab.ArchiveSubdir — projects land in <archive>/projects, the
	// #181 layout; operator decision 2026-07-15: archive, don't stay in place).
	archive: "workshop/history"
}

// ── scaffold: the on-disk creation template `sdlc project new` writes. The
// fractal file: sections grow through the gated stages (PRD at define,
// Estimate at commit, Breakdown at breakdown, Log throughout). ──
#ScaffoldSection: {
	name:  string
	seed?: string
}
scaffold: sections: [...#ScaffoldSection] & [
	{name: "PRD"},
	{name: "Estimate"},
	{name: "Breakdown", seed: "- [ ]"},
	{name: "Log"},
]

// ── #Project: the data shape of a project record ──
#Project: {
	type:      "project"
	name:      string // slug; matches filename without .md
	goal:      string // one sentence: why this project exists
	done_when: string // the MVP boundary, falsifiable
	status:    #Status
	// The commit-time baseline (the time-bound attribute distinguishing project
	// from an issue container). Optional pre-commit; compiled-required after.
	// YAML date literals decode as strings (#124 lesson: accept what real
	// frontmatter parses to, don't self-vet only).
	deadline?:       string | null
	planned_finish?: string | null
	operator?:       string | null
	// issue refs ("repo#id"); the MVP commitment. explicitly_out is the
	// load-bearing half of the scoping conversation.
	mvp_scope?:      [...string] | null
	explicitly_out?: [...string] | null
	// compiled guard: a LIVE post-commit project (committed/executing/paused)
	// carries the baseline. `done` is exempt: a properly-run project reaches
	// done via executing (so it still has one), but a record archived from the
	// pre-baseline era honestly has none — requiring it would force fabricated
	// dates on legacy migrations (ariadne#171 M1). A dropped project may have
	// died pre-commit, so it was never required either.
	if status == "committed" || status == "executing" || status == "paused" {
		deadline!:       string
		planned_finish!: string
	}
	// OPEN (#124 precedent): allow organically-growing frontmatter (created/
	// updated/sources/…) so instance conformance doesn't false-positive on a
	// valid-but-unmodeled field.
	...
}

// ── lifecycle: the transition table (the verbs). Guards are NAMED here; their
// implementations live in sdlc's guard registry (internal/project/guards.go),
// which refuses transitions naming a guard it doesn't implement. ──
#Transition: {
	from:   #Status
	to:     #Status
	event:  string
	guards: [...string]
}

lifecycle: [...#Transition] & [
	// the funnel
	{from: "ideation", to: "defined", event: "define", guards: ["prd-present"]},
	{from: "defined", to: "committed", event: "commit", guards: ["phase-a-estimate", "baseline-set", "reality-check"]},
	{from: "committed", to: "executing", event: "breakdown", guards: ["issues-cover-prd"]},
	// close is a dedicated verb (`sdlc project close`) owning retro + ledger +
	// archive; set-status refuses →done and points at it. Deliberately NO
	// paused→done edge: a paused project resumes before it closes.
	{from: "executing", to: "done", event: "close", guards: ["retro-recorded", "fog-factor-recorded"]},
	// pause/resume (baseline survives)
	{from: "executing", to: "paused", event: "pause"},
	{from: "paused", to: "executing", event: "resume"},
	// drop at any pre-terminal stage; once executing, a retro is owed
	{from: "ideation", to: "dropped", event: "drop"},
	{from: "defined", to: "dropped", event: "drop"},
	{from: "committed", to: "dropped", event: "drop"},
	{from: "executing", to: "dropped", event: "drop", guards: ["retro-recorded"]},
	{from: "paused", to: "dropped", event: "drop", guards: ["retro-recorded"]},
]

// ── laws: named assertions the graph shape doesn't already guarantee.
// Each evaluates to a concrete value when satisfied, or ⊥ (a vet failure) when not. ──
_froms: [for t in lifecycle {t.from}]
_tos: [for t in lifecycle {t.to}]

laws: {
	// every status carries a non-empty `when`
	"documented-value": {
		for s in list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]) {
			(s): when[s] & !=""
		}
	}
	// every non-initial status is reachable (ideation is the entry point)
	"reachable": {
		for s in list.Concat([["defined"], categories.committed, categories.executing, categories.terminal]) {
			(s): list.Contains(_tos, s) & true
		}
	}
	// every non-terminal status is escapable
	"escapable": {
		for s in list.Concat([categories.forming, categories.committed, categories.executing]) {
			(s): list.Contains(_froms, s) & true
		}
	}
}
