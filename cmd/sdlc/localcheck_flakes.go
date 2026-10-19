// localcheck_flakes.go — the known-flake step of a recorded suite (#394). A
// failed suite attempt is classified against the registry of open flake
// issues; when only registered tests failed, those tests alone are rerun once
// on the same head, as a linked attempt of their own, and the outcome is
// classified again. The rules are internal/flakes; this is their IO shell.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// knownFlakeRerunTimeout bounds the targeted rerun: a few named tests, never
// the suite's budget. Tests shrink it.
var knownFlakeRerunTimeout = 15 * time.Minute

// knownFlakeRerunArgv builds the rerun's command; tests script it.
var knownFlakeRerunArgv = flakes.RerunCommand

// suiteRerunCheck is the check name a known-flake rerun is saved under (#394
// M2 review, BR-19): never "suite", so no reader that asks for suite runs (the
// saved-pass reuse, test-stats, the weekly snapshot) can take a rerun of a few
// tests for one. Its parent attempt (rerun_of) carries the gate's outcome.
const suiteRerunCheck = "suite-rerun"

// suiteGateRun is where the step runs: the suite's worktree and stores.
type suiteGateRun struct {
	check                localCheck
	sha, wt, helper      string
	subject              suitePrioritySubject // the suite run's queue priority (#452): its rerun shares it
	logDir, storeRoot    string
	failedID             string
	failedFacts          SuiteLogFacts
	repo, slot           string
	agent, model, effort *string
}

// knownFlakeStep returns the gate's final outcome for a failed suite attempt.
// Every effect it runs (the rerun) is saved as an attempt linked to the
// failed one. An error is a rerun that was interrupted or could not be run
// and recorded: the caller then records the failed attempt as unjudged and
// posts nothing, so nothing turns green.
func knownFlakeStep(ctx context.Context, d localRunDeps, g suiteGateRun) (flakes.Outcome, error) {
	in := flakes.Input{
		Attempt: flakes.Result{Failures: g.failedFacts.Failures, Unattributed: g.failedFacts.Unattributed, Subtests: g.failedFacts.Subtests},
		Policy:  g.check.knownFlakePolicy(),
	}
	if in.Policy != flakes.PolicyOff {
		if d.FlakeRegistry == nil {
			in.RegistryErr = errors.New("no known-flake registry is available to this run")
		} else {
			in.Registry, in.RegistryErr = d.FlakeRegistry(ctx)
		}
	}
	o := flakes.Classify(in)
	fmt.Fprintf(d.Stderr, "==> local check %s: %s\n", g.check.Name, o.Reason)
	if o.Kind != flakes.NeedRerun {
		return o, nil
	}
	// Whole top-level tests (#466): a registered subtest reruns its test.
	argv, err := knownFlakeRerunArgv(flakes.RerunTests(o.Known))
	if err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun: %w", err)
	}
	w, err := beginSuiteAttempt(g.storeRoot)
	if err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun: no attempt record: %w", err)
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(filepath.Dir(w.LogPath))
		}
	}()
	fmt.Fprintf(d.Stderr, "==> local check %s: rerunning %s on %s (known flakes %s)\n", g.check.Name, strings.Join(argv, " "), abbrevSHA(g.sha), strings.Join(o.Issues(), ", "))
	// Weight 1: a few named tests, not the sharded suite. Its own lease and
	// slot, released when it exits (#383), so cancel reaches it too.
	// Bounded by the check's own timeout as well (its declared envelope).
	timeout := min(knownFlakeRerunTimeout, time.Duration(g.check.TimeoutMinutes)*d.minute)
	run, err := runUnderSlot(ctx, d, g.wt, g.helper, g.logDir, g.check.Name, g.sha, 1, g.subject, argv, w.LogPath, timeout)
	if err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun: %w", err)
	}
	logSHA, _, err := hashAndTail(w.LogPath, localCheckTailLines)
	if err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun log: %w", err)
	}
	rerunLog, err := os.ReadFile(w.LogPath)
	if err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun log: %w", err)
	}
	facts := parseSuiteLogFacts(string(rerunLog))
	// The rerun proves only what its log shows passing (#394 M2 review): a
	// timeout, or a named test with no PASS line, is inconclusive.
	if run.TimedOut {
		facts.Unattributed = append(facts.Unattributed, "the rerun timed out")
	}
	if run.Exit == 0 && !run.TimedOut {
		for _, t := range flakes.RerunMissing(string(rerunLog), flakes.Tests(o.Known)) {
			facts.Unattributed = append(facts.Unattributed, t+" did not pass in the rerun log")
		}
	}
	in.Rerun = &flakes.Result{Passed: run.Exit == 0 && !run.TimedOut, Failures: facts.Failures, Unattributed: facts.Unattributed, Subtests: facts.Subtests}
	o = flakes.Classify(in)
	queueSeconds, wallSeconds := run.QueueWait.Seconds(), run.Duration.Seconds()
	attempt := SuiteAttempt{Version: suiteAttemptVersion, ID: w.ID, Repo: g.repo, SHA: g.sha, Check: suiteRerunCheck, Slot: g.slot,
		Agent: g.agent, Model: g.model, Effort: g.effort,
		StartedAt: run.Start, FinishedAt: run.Finished, QueuedAt: ptrTime(run.QueueStart), AcquiredAt: ptrTime(run.Acquired),
		QueueSeconds: &queueSeconds, WallSeconds: &wallSeconds, Exit: run.Exit, TimedOut: run.TimedOut, LogSHA256: logSHA,
		Failures: facts.Failures, Units: facts.Units, Outcome: string(o.Kind), KnownFlakes: o.Known, RerunOf: g.failedID}
	if _, err := publishSuiteAttempt(g.storeRoot, w, attempt); err != nil {
		return flakes.Outcome{}, fmt.Errorf("known-flake rerun: its attempt was not saved: %w", err)
	}
	published = true
	fmt.Fprintf(d.Stderr, "==> local check %s: %s (rerun attempt %s)\n", g.check.Name, o.Reason, w.ID)
	return o, nil
}

// gateFlakeRegistry is the registry a suite run in root reads (#394 D4): the
// repository's open issues' flake_tests, from a fresh tracker read.
func gateFlakeRegistry(root string) func(ctx context.Context) (flakes.Registry, error) {
	return func(ctx context.Context) (flakes.Registry, error) {
		r, _, err := loadFlakeRegistry(ctx, filepath.Join(root, envOr("WF_ISSUES_DIR", "workshop/issues")))
		return r, err
	}
}

// revalidateKnownFlake rereads the registry just before a known-flake landing
// is posted (#394 M2 review, BR-12): the rerun took minutes, and its issue may
// have closed. The policy is the run's own declaration, read from main when
// the run began; republish rereads both (republishRevalidates).
func revalidateKnownFlake(ctx context.Context, d localRunDeps, c localCheck, o flakes.Outcome) flakes.Outcome {
	if o.Kind != flakes.KnownFlakePassed && o.Kind != flakes.KnownFlakeAccept {
		return o
	}
	var reg flakes.Registry
	regErr := errors.New("no known-flake registry is available to this run")
	if d.FlakeRegistry != nil {
		reg, regErr = d.FlakeRegistry(ctx)
	}
	next := flakes.Revalidate(o, reg, regErr, c.knownFlakePolicy())
	if next.Kind != o.Kind {
		fmt.Fprintf(d.Stderr, "==> local check %s: not landing: %s\n", c.Name, next.Reason)
	}
	return next
}
