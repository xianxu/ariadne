// suiteincomplete.go — a suite the Go test alarm cut short is incomplete, not
// failed (#630): it is recorded with the host's load, retried once on the
// packages that did not finish or had a named failure, and never sent to a
// main-baseline diagnosis or a train split.
package main

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// incompleteRun is what the alarm left of a run: where it fired, what was
// running, and the packages the one retry runs.
type incompleteRun struct {
	Alarmed  []string // "<process> in <package>"
	Running  []string
	Packages []string // sorted: alarmed packages and every package with a named failure
}

var (
	alarmExplainedReason = regexp.MustCompile(`^(packages|shard[0-9]+) exited without a failing test$`)
	packageNoTestReason  = regexp.MustCompile(`^package (\S+) failed without a failing test$`)
)

// incompleteEligible is which runs the rule may judge: an untyped test run
// that exited nonzero on its own (it is called before any outcome is set, so
// that is also "did not pass"). A typed runner's run (a declared contract) is
// v4 evidence, and a v5 attempt never carries a contract, so it keeps today's
// classification; so does a run sdlc's own deadline stopped.
func incompleteEligible(check, contract string, rec localRunRecord) bool {
	return isTestCheck(check) && contract == "" && rec.Exit > 0 && !rec.TimedOut
}

// classifyIncomplete applies #630's rule to a run's log facts: at least one
// process hit the alarm, every process that failed did, and every failure the
// log cannot name is one the alarm explains (tests that never ran, an alarmed
// process or package that failed without a test). A build or setup failure, a
// typed runner's evidence, or a parse warning keeps today's classification.
func classifyIncomplete(f SuiteLogFacts) (incompleteRun, bool) {
	var run incompleteRun
	if len(f.Warnings) != 0 || f.Typed.Seen {
		return run, false
	}
	alarmedPkg, alarmedLabel := map[string]bool{}, map[string]bool{}
	for _, p := range f.Processes {
		if p.Exit != 0 && len(p.Alarmed) == 0 {
			return run, false // a process failed on its own
		}
		for _, pkg := range p.Alarmed {
			alarmedPkg[pkg], alarmedLabel[p.Label] = true, true
			run.Alarmed = append(run.Alarmed, p.Label+" in "+pkg)
		}
		run.Running = append(run.Running, p.Running...)
	}
	if len(run.Alarmed) == 0 {
		return run, false
	}
	for _, why := range f.Unattributed {
		switch m, n := alarmExplainedReason.FindStringSubmatch(why), packageNoTestReason.FindStringSubmatch(why); {
		case why == "listed tests never ran" || why == "a test timed out":
		case m != nil && alarmedLabel[m[1]]:
		case n != nil && alarmedPkg[n[1]]:
		default:
			return run, false
		}
	}
	pkgs := map[string]bool{}
	for pkg := range alarmedPkg {
		pkgs[pkg] = true
	}
	for _, name := range append(slices.Clone(f.Failures), f.Subtests...) {
		if pkg, _, ok := flakes.SplitTest(name); ok { // a typed name only comes from typed evidence, refused above
			pkgs[pkg] = true
		}
	}
	for pkg := range pkgs {
		run.Packages = append(run.Packages, pkg)
	}
	slices.Sort(run.Packages)
	return run, true
}

// String is the one-line report: where the alarm fired and what the retry runs.
func (r incompleteRun) String() string {
	running := ""
	if len(r.Running) != 0 {
		running = " (running " + strings.Join(r.Running, ", ") + ")"
	}
	return "the Go test alarm stopped " + strings.Join(r.Alarmed, ", ") + running + "; unfinished or failed: " + strings.Join(r.Packages, " ")
}

// goIdentities reports whether every name is a Go test: a v5 attempt carries
// only what a v2 one does, never a typed runner's identity.
func goIdentities(names []string) bool {
	for _, name := range names {
		if id, err := flakes.ParseIdentity(name); err == nil && id.Kind() != flakes.KindGo {
			return false
		}
	}
	return true
}

// validLoadAverage is a recorded load: three finite, non-negative averages.
func validLoadAverage(load []float64) bool {
	if len(load) != 3 {
		return false
	}
	for _, v := range load {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return false
		}
	}
	return true
}

// suiteIncompleteError is a run the alarm cut short and its one retry did not
// complete either (#630). Execution evidence, never a test result: the merge
// gate runs no main baseline for it and the train neither splits nor blames.
type suiteIncompleteError struct {
	Check, Head, LogPath string
	Run                  incompleteRun
	Load                 []float64
	Baseline             bool // main's own diagnostic baseline was the incomplete run
	Retried              bool // this is the one retry, cut short again: there is no other
}

func (e *suiteIncompleteError) Error() string {
	load := "host load unavailable"
	if len(e.Load) == 3 {
		load = fmt.Sprintf("host load %.1f %.1f %.1f", e.Load[0], e.Load[1], e.Load[2])
	}
	consequence := "not a test result, so no main baseline was run; rerun " + localCheckRunCommand(e.Check, e.Head) + " when the host is less loaded"
	if e.Baseline {
		consequence = "main's baseline is unknown, so no failure is attributed to main or to the change"
	}
	again := ""
	if e.Retried {
		again = " again on its one retry"
	}
	return fmt.Sprintf("local check %s on %s incomplete (timeout)%s: %s; %s; %s; log %s", e.Check, abbrevSHA(e.Head), again, e.Run, load, consequence, e.LogPath)
}

// hostLoad is the 1, 5 and 15 minute load for a report, or nil when it cannot
// be read: the run's outcome never depends on it.
func (d localRunDeps) hostLoad() []float64 {
	read := d.loadAverage
	if read == nil {
		read = hostLoadAverage
	}
	load, err := read()
	if err != nil {
		return nil
	}
	return load[:]
}

// incompleteRetryOutcome judges an incomplete run's one retry (#630): it lands
// only when it exited cleanly with nothing failed and nothing unnamed; a retry
// the alarm cut short again is incomplete again; anything else is a failure
// of the retried packages, named when the log names it.
func incompleteRetryOutcome(retry incompleteRun, exit int, timedOut bool, child SuiteLogFacts) flakes.Outcome {
	pkgs := strings.Join(retry.Packages, " ")
	switch {
	case retryPassed(exit, timedOut, child):
		return flakes.Outcome{Kind: flakes.IncompleteRetried, Reason: "the retry of " + pkgs + " passed"}
	case exit > 0 && !timedOut:
		if again, ok := classifyIncomplete(child); ok {
			return flakes.Outcome{Kind: flakes.IncompleteTimeout, Reason: "the retry was incomplete too: " + again.String()}
		}
	}
	if len(child.Failures) == 0 || len(child.Unattributed) != 0 {
		return flakes.Outcome{Kind: flakes.Unattributable, Unknown: child.Failures, Reason: "the retry of " + pkgs + " failed without naming every failure"}
	}
	return flakes.Outcome{Kind: flakes.Unknown, Unknown: child.Failures, Reason: "the retry of " + pkgs + " failed: " + strings.Join(child.Failures, ", ")}
}

// retryPassed is the one definition of a clean retry: exit 0 within its time,
// and its log names no failure, failed nothing, and left nothing unnamed. A
// "--- FAIL pkg Test" line names a failure without marking the log failed.
func retryPassed(exit int, timedOut bool, child SuiteLogFacts) bool {
	return exit == 0 && !timedOut && !child.Failed && len(child.Failures) == 0 && len(child.Subtests) == 0 && len(child.Unattributed) == 0
}

// incompleteRetryCommands is the retry's one command (#630): the affected
// command when main declares one, else the declared suite command, under the
// affected env contract naming the packages to run. A suite command that
// ignores the list reruns whole: slower, never wrong.
func incompleteRetryCommands(suite, affectedCommand []string, pkgs []string) [][]string {
	command := suite
	if len(affectedCommand) != 0 {
		command = affectedCommand
	}
	return [][]string{affectedEnvCommand(command, pkgs, nil)}
}

// affectedCommandOnMain is the affected command origin/main declares, or nil
// when it declares none or cannot be read: the retry then runs the suite's
// own command under the same package list (#630).
func (d localRunDeps) affectedCommandOnMain() []string {
	main, err := d.Git("rev-parse", "--verify", "origin/main^{commit}")
	if err != nil {
		return nil
	}
	cfg, ok, err := readAffectedConfig(d.Git, strings.TrimSpace(main))
	if err != nil || !ok {
		return nil
	}
	return cfg.Command
}
