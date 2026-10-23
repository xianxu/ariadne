// project — the vocabulary of a project: its data shape, its lifecycle funnel,
// and the laws the funnel must satisfy. Single source of truth for the `project`
// noun (ariadne#180). sdlc reads the exported JSON; humans and the LLM read this
// file directly. The prose companion (construct/datatype/project.md) cites this
// file as schema authority — a drift test binds the two.
//
// The organizing insight (#329, superseding #180's shipping plan): a project is
// a TL-driven coordination plane for cross-issue work — a Charter, Trains with
// cross-train edges, append-only Decisions — not a time-bound push. No deadline:
// the forecast is derived from the trains and the calibration ledger, never
// typed, and a date (when a team needs one) never drives status.
package project

import "list"

// ── categories: the single concrete source of status membership ──
// forming   = charter written, no train dispatched yet
// executing = trains running (executing), or every MVP issue landed and the
//             done_when audit is the remaining work (closing)
// terminal  = closed (audit passed, archived) or dropped
categories: {
	forming:   ["defined"]
	executing: ["executing", "closing"]
	terminal:  ["closed", "dropped"]
}

#Forming:  or(categories.forming)
#Terminal: or(categories.terminal)
#Status:   or(list.Concat([categories.forming, categories.executing, categories.terminal]))

// ── when: one-line semantics per status (the documented-value source) ──
when: {
	defined:   "Charter written (problem, cross-issue requirements, done_when, scope); no train dispatched yet"
	executing: "the first train is dispatched; the TL grooms Trains and appends Decisions"
	closing:   "every mvp_scope issue has landed; the done_when audit is the remaining work"
	closed:    "done_when audit passed; retro recorded and the file archived"
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

// ── scaffold: the on-disk creation template `sdlc project new` writes.
// Charter on top (rarely changes), Trains (the TL's plan, rewritten at each
// grooming), Breakdown (the checkbox list), Log, and Decisions last (append-only,
// cross-issue only). ──
#ScaffoldSection: {
	name:      string
	seed?:     string
	optional?: bool // written by `project new`; `project validate` does not require it
}
scaffold: sections: [...#ScaffoldSection] & [
	{name: "Charter"},
	{name: "Trains", seed: "| Train | Cars, in order | Slot |\n|---|---|---|"},
	{name: "Breakdown", seed: "- [ ]"},
	{name: "Log"},
	{name: "Decisions"},
]

// ── #Project: the data shape of a project record ──
#Project: {
	type:      "project"
	name:      string // slug; matches filename without .md
	goal:      string // one sentence: why this project exists
	done_when: string // the MVP boundary, falsifiable
	status:    #Status
	// No deadline (#329): a forecast is derived, never typed, so the shipping
	// plan's baseline fields are banned in a live record. Archived records are
	// frozen history and never validated against this schema.
	// The constraint is a sentinel string, not _|_, so cue's diagnostic names
	// the field and carries the remedy.
	deadline?:       "retired by ariadne#329: the forecast is derived (sdlc project forecast)"
	planned_finish?: "retired by ariadne#329: the forecast is derived (sdlc project forecast)"
	operator?:       string | null
	// the TL who drives this project (a slot handle, or a list of them).
	// A TL-owned project is written only by its TL: close and milestone-close
	// skip it (ariadne#326).
	tl?: string | [...string] | null
	// whether `sdlc merge` ticks this project's rows when a referenced issue
	// lands (ariadne#373). Absent: yes without tl:, and off for a TL-owned
	// project until the operator flips the default.
	landing_tick?: bool | null
	// issue refs ("repo#id"); the MVP commitment. explicitly_out is the
	// load-bearing half of the scoping conversation.
	mvp_scope?:      [...string] | null
	explicitly_out?: [...string] | null
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
	{from: "defined", to: "executing", event: "execute", guards: ["trains-present"]},
	{from: "executing", to: "closing", event: "wrap", guards: ["mvp-landed"]},
	// new MVP work filed while closing sends the project back to executing
	{from: "closing", to: "executing", event: "reopen"},
	// close is a dedicated verb (`sdlc project close`) owning retro + archive;
	// set-status refuses →closed and points at it.
	{from: "closing", to: "closed", event: "close", guards: ["retro-recorded"]},
	// drop at any pre-terminal stage; once executing, a retro is owed
	{from: "defined", to: "dropped", event: "drop"},
	{from: "executing", to: "dropped", event: "drop", guards: ["retro-recorded"]},
	{from: "closing", to: "dropped", event: "drop", guards: ["retro-recorded"]},
]

// ── laws: named assertions the graph shape doesn't already guarantee.
// Each evaluates to a concrete value when satisfied, or ⊥ (a vet failure) when not. ──
_froms: [for t in lifecycle {t.from}]
_tos: [for t in lifecycle {t.to}]

laws: {
	// every status carries a non-empty `when`
	"documented-value": {
		for s in list.Concat([categories.forming, categories.executing, categories.terminal]) {
			(s): when[s] & !=""
		}
	}
	// every non-initial status is reachable (defined is the entry point)
	"reachable": {
		for s in list.Concat([categories.executing, categories.terminal]) {
			(s): list.Contains(_tos, s) & true
		}
	}
	// every non-terminal status is escapable
	"escapable": {
		for s in list.Concat([categories.forming, categories.executing]) {
			(s): list.Contains(_froms, s) & true
		}
	}
}
