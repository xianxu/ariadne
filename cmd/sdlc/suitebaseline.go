package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// diagnoseSuiteFailure runs the exact same required check on the main commit
// only after the integrated head failed. The diagnostic run is deliberately
// unrecorded: main's published status must describe a landing, not this probe.
func diagnoseSuiteFailure(ctx context.Context, gateErr error, main string, deps localRunDeps) error {
	var failed *localCheckFailure
	if !errors.As(gateErr, &failed) || failed.Check.Name != suiteCheckName {
		return gateErr
	}
	if failed.Bootstrap {
		return fmt.Errorf("%w\n  suite comparison unverified: bootstrap main has no test target; no main baseline was run", gateErr)
	}
	// The baseline run prunes old local-check logs. Capture the failed head's
	// log first so a busy clone cannot evict the evidence under comparison.
	mergedLog, err := os.ReadFile(failed.Run.LogPath)
	if err != nil {
		return fmt.Errorf("%w\n  merged suite log unreadable: %v", gateErr, err)
	}
	deps.SkipRecord = true
	deps.PR = 0
	baseline, err := runLocalCheck(ctx, deps, failed.Check, main)
	if err != nil {
		return fmt.Errorf("%w\n  baseline on main could not run: %w", gateErr, err)
	}
	baselineLog, err := os.ReadFile(baseline.LogPath)
	if err != nil {
		return fmt.Errorf("%w\n  main baseline log unreadable: %v", gateErr, err)
	}
	merged := failingTests(string(mergedLog))
	base := failingTests(string(baselineLog))
	var labels []string
	if !baseline.passed() && len(base) == 0 {
		labels = append(labels, "main is red (unparsed); introduced failures cannot be determined")
	} else if len(merged) == 0 {
		labels = append(labels, "merged suite failed (unparsed); introduced failures cannot be determined")
	} else {
		introduced, preexisting := classifyFailures(merged, base)
		if len(introduced) > 0 {
			labels = append(labels, "introduced: "+strings.Join(introduced, ", "))
		}
		if len(preexisting) > 0 {
			label := "pre-existing: " + strings.Join(preexisting, ", ")
			// #394: a failure main shares is a persistent main failure, never
			// flake-eligible without same-head fail→pass evidence; the gate's
			// classifier decides it, from the unrecorded baseline.
			o := flakes.Classify(flakes.Input{Attempt: flakes.Result{Failures: merged},
				MainBaseline: &flakes.Result{Passed: baseline.passed(), Failures: base}})
			if o.Kind == flakes.PersistentMain {
				label += " (a persistent main failure, never a known flake)"
			}
			labels = append(labels, label)
		}
	}
	if baseline.passed() {
		labels = append(labels, "main baseline passed")
	} else {
		labels = append(labels, "main baseline failed; log "+baseline.LogPath)
	}
	return fmt.Errorf("%w\n  suite comparison: %s", gateErr, strings.Join(labels, "; "))
}
