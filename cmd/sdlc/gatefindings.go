// gatefindings.go — the ONE routing line every fixer-facing findings refusal emits
// (ariadne#203).
//
// The `family:` machinery (#194) was built with three consumers — the reviewer
// (told to slug the RULE, not the symptom), the ledger (counts repeats), and the
// operator (the convergence line). The agent that actually FIXES the findings was
// not one of them. What it read at the moment it started fixing was "address the
// findings above and re-run", which reads as "address each of them" — the per-site
// patching the family counter exists to detect. parley.nvim#202 spent four boundary
// rounds against a cap of three that way, with `invariant-without-regression-guard`
// and `stale-restatement-of-moved-source` each surviving three separate rounds.
//
// This is a file rather than eight strings for the reason gatepersist.go is a file:
// eight hand-maintained copies of a rule about not hand-patching would refute
// themselves, and that tail diverged five times before extraction.
//
// It ROUTES rather than restates. The discipline itself has one statement, in
// ARCH-PURPOSE (cmd/sdlc/internal/judge/architecture.md) — the #128 pattern, where
// the constitution stopped restating ARCH-* definitions and began routing to
// `sdlc arch-principles`. ArchitectureBlock warns that "a marker alone would be a
// dangling pointer in a fresh-context subagent"; that constraint is why the JUDGES
// get the registry inlined, and why it does not bind HERE — these lines are read by
// the main thread, which already received the block from `sdlc start-plan` and can
// pull it any time. Restating the procedure here would make the drift guard vacuous.
package main

// fixTheClassLine is the routing line. Pure.
func fixTheClassLine() string {
	return "each finding names one instance — fix the CLASS it belongs to, not only that site: " +
		"ARCH-PURPOSE (`sdlc arch-principles`)"
}
