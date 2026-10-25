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
		if id, err := flakes.ParseIdentity(name); err == nil && id.Kind() != flakes.KindGo {
			return run, false // a typed runner's failure: not this rule's evidence
		}
		if pkg, _, ok := flakes.SplitTest(name); ok {
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
	return "the Go test alarm stopped " + strings.Join(r.Alarmed, ", ") + "; retry runs " + strings.Join(r.Packages, " ")
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
	Retried              bool
}

func (e *suiteIncompleteError) Error() string {
	load := "host load unavailable"
	if len(e.Load) == 3 {
		load = fmt.Sprintf("host load %.1f %.1f %.1f", e.Load[0], e.Load[1], e.Load[2])
	}
	again := ""
	if e.Retried {
		again = " again on its retry"
	}
	return fmt.Sprintf("local check %s on %s incomplete (timeout)%s: %s; %s; not a test result, so no main baseline was run; rerun %s when the host is less loaded; log %s",
		e.Check, abbrevSHA(e.Head), again, e.Run, load, localCheckRunCommand(e.Check, e.Head), e.LogPath)
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
