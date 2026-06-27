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

// actual_hours can be explicitly marked not applicable when a close has no
// measured time to feed velocity calibration. Keep the accepted spelling closed.
#ActualNotApplicable: "N/A"

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
	// id: a 6-digit zero-padded number. UNQUOTED in real frontmatter (`id: 000124`),
	// which cue's YAML loader reads as an (octal) int — so accept int|string rather
	// than reject every real file (#124: the #122 schema was only ever self-vetted).
	id:     int | string
	status: #Status
	// estimate/actual: present-but-empty (`estimate_hours:`) parses as YAML null, so
	// allow null alongside a positive number (#124 — real files leave these blank).
	// Closed issues may use the exact N/A sentinel when measured actuals do not
	// apply; arbitrary strings still fail instance validation (#135).
	estimate_hours?: (number & >0) | null
	actual_hours?:   (number & >0) | #ActualNotApplicable | null
	// compiled guard: a done issue must carry measured actuals (a positive number)
	// or the explicit not-applicable sentinel, not null/absent.
	if status == "done" {
		actual_hours!: (number & >0) | #ActualNotApplicable
	}
	// OPEN (#124): allow organically-growing frontmatter (target/references/
	// related/created/… and future fields) so instance-conformance vetting at the
	// fail-closed merge gate doesn't false-positive on a valid-but-unmodeled field.
	// A bad `status` value (the #Status enum) and a typo'd *required* field
	// (statuss: → status absent) are still caught; only optional-field typos slip,
	// which is the right trade for a gate (don't train --no-validate).
	...
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

	// #122 M4: triage an unstarted issue without ever working it
	{from: "open", to: "wontfix", event: "abandon"}, // reject at triage
	{from: "open", to: "punt", event: "defer"},      // defer at triage
	// #122 M4: reopen from any terminal (resume a deferred / reconsider a rejected)
	{from: "punt", to: "working", event: "reopen"},
	{from: "wontfix", to: "working", event: "reopen"},
	// #122 M4: abandon/defer while blocked (don't force an unblock-first detour)
	{from: "blocked", to: "wontfix", event: "abandon"},
	{from: "blocked", to: "punt", event: "defer"},
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
