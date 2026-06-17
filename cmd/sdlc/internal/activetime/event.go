// Package activetime is the native Go port of active-time-v3.py — the
// segment-anchored per-issue dev-hour attribution behind `sdlc actual` (#68,
// #110). It collapses what was a python3 subprocess + stdout-regex + script-
// resolution into an in-process engine: a pure core (gap-truncated active
// minutes, segment construction, the commit-weight/mention split) behind a thin
// IO seam (transcript .jsonl event loading + a git-log window loader).
//
// Semantics are preserved 1:1 with the Python so measured actuals don't shift;
// the only intentional divergences are cosmetic (the unattributed bucket renders
// as "unattributed", not the Python's "##unattributed").
package activetime

import "time"

// Event is one transcript line we attribute: a human user turn, or — when
// IncludeAssistant is set — an assistant text turn. Mentions holds the count of
// tracked-issue #N refs found in this single event's text.
//
// Asymmetry worth remembering (mirrors active-time-v3.py): a user event is
// dropped when its text is empty/whitespace, but an assistant event is ALWAYS
// emitted (even with no text and no mentions) — its timestamp is part of the
// active-time stream. See walkSessionEvents.
type Event struct {
	Time     time.Time
	Mentions map[string]int
}
