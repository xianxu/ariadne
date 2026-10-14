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

func prepareCIGateFactsContext(ctx context.Context, git func(...string) (string, error), base, head string) (*ciGateFacts, error) {
	hasWF, err := headHasWorkflows(git, head)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w", err)
	}
	decl, err := readLocalChecks(git, base)
	if err != nil {
		return nil, fmt.Errorf("CI gate: %w\n  → fix the declaration on main (it is under the local-gates guard)", err)
	}
	decl, err = withDefaultSuiteContext(ctx, git, base, decl)
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

// Each polling read gets a fresh deadline from the remaining budget. A suite
// does not inherit that deadline: its complete runtime is excluded afterward.
type ciReadBudget struct {
	started   time.Time
	suiteTime time.Duration
	expired   bool
}

func (b *ciReadBudget) remaining() time.Duration {
	return ciWaitLimit - (ciWaitNow().Sub(b.started) - b.suiteTime)
}

func readWithinCIBudget[T any](ctx context.Context, b *ciReadBudget, read func(context.Context) (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if b.expired || b.remaining() <= 0 {
		b.expired = true
		return zero, context.DeadlineExceeded
	}
	ioCtx, cancel := context.WithTimeout(ctx, b.remaining())
	defer cancel()
	value, err := read(ioCtx)
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	if ioCtx.Err() != nil || b.remaining() <= 0 {
		b.expired = true
		return zero, context.DeadlineExceeded
	}
	return value, err
}

type ciBudgetGH struct {
	ghCaller
	budget *ciReadBudget
}

func (g ciBudgetGH) PRHeadOID(ctx context.Context, repo string, number int) (string, error) {
	return readWithinCIBudget(ctx, g.budget, func(ctx context.Context) (string, error) { return g.ghCaller.PRHeadOID(ctx, repo, number) })
}
func (g ciBudgetGH) PRChecks(ctx context.Context, repo string, number int) ([]prCheck, error) {
	return readWithinCIBudget(ctx, g.budget, func(ctx context.Context) ([]prCheck, error) { return g.ghCaller.PRChecks(ctx, repo, number) })
}
func (g ciBudgetGH) JobRun(ctx context.Context, repo string, id int64) (jobRun, error) {
	return readWithinCIBudget(ctx, g.budget, func(ctx context.Context) (jobRun, error) { return g.ghCaller.JobRun(ctx, repo, id) })
}
func (g ciBudgetGH) CommitStatuses(ctx context.Context, repo, sha string) (map[string]string, error) {
	return readWithinCIBudget(ctx, g.budget, func(ctx context.Context) (map[string]string, error) { return g.ghCaller.CommitStatuses(ctx, repo, sha) })
}

func waitCIGate(ctx context.Context, gh ghCaller, repo string, number int, base, head string, git func(args ...string) (string, error), run localCheckRunner, noCI bool, stderr io.Writer, gitForContext ...func(context.Context) func(...string) (string, error)) error {
	if noCI {
		return runCIGate(ctx, gh, repo, number, base, head, git, run, true, stderr)
	}
	budget := &ciReadBudget{started: ciWaitNow()}
	var lastPending error
	timeout := func() error {
		return fmt.Errorf("CI wait exceeded %v: %w", ciWaitLimit, errors.Join(context.DeadlineExceeded, lastPending))
	}
	budgetGit := func(args ...string) (string, error) {
		// Detached probe cleanup must remain possible after the read budget or
		// caller expires. Only these exact cleanup commands get fresh authority.
		cleanup := len(args) == 2 && args[0] == "worktree" && args[1] == "prune" || len(args) == 4 && args[0] == "worktree" && args[1] == "remove" && args[2] == "--force"
		if cleanup {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), landingRecoveryTimeout)
			defer cancel()
			read := git
			if len(gitForContext) > 0 {
				read = gitForContext[0](cleanupCtx)
			}
			return read(args...)
		}
		return readWithinCIBudget(ctx, budget, func(ioCtx context.Context) (string, error) {
			read := git
			if len(gitForContext) > 0 {
				read = gitForContext[0](ioCtx)
			}
			return read(args...)
		})
	}
	// The default-suite facts include a real Makefile probe, not only Git
	// reads. Keep the whole preparation inside the remaining read deadline.
	facts, err := readWithinCIBudget(ctx, budget, func(ioCtx context.Context) (*ciGateFacts, error) {
		return prepareCIGateFactsContext(ioCtx, budgetGit, base, head)
	})
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if budget.expired || budget.remaining() <= 0 {
		return timeout()
	}
	if err != nil {
		return err
	}
	var timedRun localCheckRunner
	if run != nil {
		timedRun = func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			if err := ctx.Err(); err != nil {
				return localRunRecord{}, err
			}
			if budget.expired || budget.remaining() <= 0 {
				budget.expired = true
				return localRunRecord{}, context.DeadlineExceeded
			}
			start := ciWaitNow()
			rec, err := run(ctx, c, sha)
			budget.suiteTime += ciWaitNow().Sub(start)
			return rec, err
		}
	}
	boundedGH := ciBudgetGH{ghCaller: gh, budget: budget}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if budget.expired || budget.remaining() <= 0 {
			return timeout()
		}
		err := runPreparedCIGate(ctx, boundedGH, repo, number, base, head, budgetGit, timedRun, false, stderr, facts)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if budget.expired || budget.remaining() <= 0 {
			return timeout()
		}
		var verdict *ciGateError
		if !errors.As(err, &verdict) || (verdict.Reason != ciPending && verdict.Reason != ciNotRegistered) {
			return err
		}
		lastPending = err
		elapsed := ciWaitLimit - budget.remaining()
		fmt.Fprintf(stderr, "waiting for CI on PR #%d (%v of %v): %s\n", number, elapsed.Round(time.Second), ciWaitLimit, err)
		if err := ciWaitPause(ctx, min(ciWaitInterval, budget.remaining())); err != nil {
			return err
		}
	}
}
