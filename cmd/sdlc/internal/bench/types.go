// Package bench is the multi-agent benchmark harness (#119): it freezes a live
// issue into an immutable task, runs different coding agents against it in
// isolated worktrees, and grades each via measured objective signals plus blind
// head-to-head judging.
//
// The package is a pure core (task/run-record/scorecard/rubric structs and the
// deterministic functions over them — unit-tested with no IO mocks) wrapped by a
// thin IO shell (Store, Worktreer, Runner, Measurer). Per ARCH-PURE, nothing in
// the pure files touches os/exec/time/git; all such access is injected.
package bench

// Mode is how an agent is driven during a benchmark run. Only autonomous is
// wired day-one (#119); interactive/live are accommodated by the Responder seam.
type Mode string

const (
	ModeAutonomous  Mode = "autonomous"
	ModeInteractive Mode = "interactive"
	ModeLive        Mode = "live"
)
