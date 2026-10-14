package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// ciGateError preserves the decision class through the user-facing refusal.
// Pollers must not infer retryability from the wording of an error.
type ciGateError struct {
	Reason  ciReason
	Message string
}

func (e *ciGateError) Error() string { return e.Message }

// Facts describe immutable Git objects. A wait probes them once, not on every
// GitHub poll (the default suite probe materializes a detached worktree).
type ciGateFacts struct {
	decl         localCheckDecl
	hasWorkflows bool
	required     []string
}

func prepareCIGateFacts(git func(...string) (string, error), base, head string) (*ciGateFacts, error) {
	hasWF, err := headHasWorkflows(git, head)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w", err)
	}
	decl, err := readLocalChecks(git, base)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w\n  → fix the declaration on main (it is under the local-gates guard)", err)
	}
	decl, err = withDefaultSuite(git, base, decl)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w", err)
	}
	required, err := requiredChecks(git, base)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w", err)
	}
	return &ciGateFacts{decl: decl, hasWorkflows: hasWF, required: required}, nil
}

const ciWaitLimit = 20 * time.Minute
const ciWaitInterval = 30 * time.Second

var ciWaitNow = time.Now
var ciWaitPause = func(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func waitCIGate(ctx context.Context, gh ghCaller, repo string, number int, base, head string, git func(args ...string) (string, error), run localCheckRunner, noCI bool, stderr io.Writer) error {
	if noCI {
		return runCIGate(ctx, gh, repo, number, base, head, git, run, true, stderr)
	}
	facts, err := prepareCIGateFacts(git, base, head)
	if err != nil {
		return err
	}
	started := ciWaitNow()
	var suiteTime time.Duration
	var timedRun localCheckRunner
	if run != nil {
		timedRun = func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			start := ciWaitNow()
			rec, err := run(ctx, c, sha)
			suiteTime += ciWaitNow().Sub(start)
			return rec, err
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := runPreparedCIGate(ctx, gh, repo, number, base, head, git, timedRun, false, stderr, facts)
		var verdict *ciGateError
		if !errors.As(err, &verdict) || (verdict.Reason != ciPending && verdict.Reason != ciNotRegistered) {
			return err
		}
		elapsed := ciWaitNow().Sub(started) - suiteTime
		if elapsed >= ciWaitLimit {
			return fmt.Errorf("CI wait exceeded %v: %w", ciWaitLimit, err)
		}
		fmt.Fprintf(stderr, "waiting for CI on PR #%d (%v of %v): %s\n", number, elapsed.Round(time.Second), ciWaitLimit, err)
		if err := ciWaitPause(ctx, min(ciWaitInterval, ciWaitLimit-elapsed)); err != nil {
			return err
		}
	}
}
