// issue — the vocabulary of a workshop issue: its data shape, its lifecycle, and
// the laws that lifecycle must satisfy. Single source of truth for the `issue`
// noun (ariadne#122). sdlc reads the exported JSON; humans and the LLM read this
// file directly. The `issue-lifecycle` target carries the *why* behind the shape.
package issue

import "list"

// ── categories: the single concrete source of status membership ──
// Stated once, here. #Status / #Active / #Terminal are DERIVED from these via
// or(), so there is nothing to keep in sync. Only `categories` (concrete data)
// reaches the exported JSON — CUE definitions (#) do not export.
categories: {
	open:     ["open"]                    // created, not yet started
	active:   ["working", "blocked"]      // started, not yet closed
	terminal: ["done", "wontfix", "punt"] // closed
}

#Active:   or(categories.active)
#Terminal: or(categories.terminal)
#Status:   or(list.Concat([categories.open, categories.active, categories.terminal]))

// ── when: one-line semantics per status (the documented-value source) ──
when: {
	open:    "created, not yet started"
	working: "actively in progress"
	blocked: "waiting on another tracked issue"
	done:    "complete and closed"
	wontfix: "rejected; will not be done"
	punt:    "deferred"
}

// ── #Issue: the data shape of an issue record ──
#Issue: {
	id:              string // zero-padded 6-digit, matches the filename
	status:          #Status
	estimate_hours?: number & >0
	actual_hours?:   number & >0
	// compiled guard: a done issue must carry measured actuals
	if status == "done" {
		actual_hours!: number & >0
	}
}

// ── lifecycle: the transition table (the verbs). Guards are NAMED here; their
// effectful implementations live in sdlc (the close gates). ──
#Transition: {
	from:   #Status
	to:     #Status
	event:  string
	guards: [...string]
}

lifecycle: [...#Transition] & [
	{from: "open", to: "working", event: "claim"},      // start work
	{from: "working", to: "blocked", event: "block"},   // hit a dependency
	{from: "blocked", to: "working", event: "unblock"}, // dependency cleared
	{from: "working", to: "done", event: "close", guards: ["actual-measured", "verified", "atlas-updated"]},
	{from: "blocked", to: "done", event: "close", guards: ["actual-measured", "verified", "atlas-updated"]},
	{from: "working", to: "wontfix", event: "abandon"}, // rejected mid-flight
	{from: "working", to: "punt", event: "defer"},      // deferred mid-flight
	{from: "done", to: "working", event: "reopen"},     // re-open a closed issue
]

// ── laws: named assertions the graph shape doesn't already guarantee.
// Each evaluates to a concrete value when satisfied, or ⊥ (a vet failure) when not. ──
_froms: [for t in lifecycle {t.from}]
_tos: [for t in lifecycle {t.to}]

laws: {
	// every status carries a non-empty `when`
	"documented-value": {
		for s in list.Concat([categories.open, categories.active, categories.terminal]) {
			(s): when[s] & !=""
		}
	}
	// every non-open status is reachable (appears as some transition's `to`)
	"reachable": {
		for s in list.Concat([categories.active, categories.terminal]) {
			(s): list.Contains(_tos, s) & true
		}
	}
	// every non-terminal status is escapable (appears as some transition's `from`)
	"escapable": {
		for s in list.Concat([categories.open, categories.active]) {
			(s): list.Contains(_froms, s) & true
		}
	}
}
