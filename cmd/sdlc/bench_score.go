package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// Scoring keeps at most three command logs, each stopped at this explicit
// 8 MiB diagnostic budget (plus one bounded pipe-copy chunk in the managed
// owner's log). No hidden output is forwarded to the TL or worker.
const benchScoreLogBudget = 8 << 20
const benchScoreCommandTimeout = 15 * time.Minute
const benchScoreBudget = 45 * time.Minute

type benchScoreCommand struct {
	Role, Root, WorkRoot, Head, LogPath string
	Argv                                []string
}
type benchScoreRun struct {
	Exit            int
	TimedOut, Quiet bool
}
type benchScoreDeps struct {
	Execute func(context.Context, benchScoreCommand) (benchScoreRun, error)
}

// The caller owns Run admission/retirement and records the returned unavailable
// status even when err supplies additional private diagnostics. This function
// never changes Run/store state or rereads the descriptor's original patch.
func evaluateBenchAcceptance(ctx context.Context, task bench.Task, workRoot, candidate, observerDir string) (bench.Acceptance, error) {
	return evaluateBenchAcceptanceWith(ctx, task, workRoot, candidate, observerDir, benchScoreDeps{})
}

func evaluateBenchAcceptanceWith(ctx context.Context, task bench.Task, workRoot, candidate, observerDir string, d benchScoreDeps) (a bench.Acceptance, retErr error) {
	ctx, cancel := context.WithTimeout(ctx, benchScoreBudget)
	defer cancel()
	a = bench.Acceptance{Status: "unavailable", Reason: "acceptance evaluation did not complete"}
	fail := func(reason string, cause error) (bench.Acceptance, error) {
		a.Status = "unavailable"
		a.Reason = reason
		a.CommandPassed = nil
		return a, fmt.Errorf("%s: %w", reason, cause)
	}
	if err := bench.ValidateTask(task); err != nil {
		return fail("invalid acceptance task", err)
	}
	if !regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`).MatchString(candidate) {
		return fail("candidate must be a full local commit pin", errors.New("invalid candidate object ID"))
	}
	for _, p := range []string{workRoot, observerDir} {
		physical, err := filepath.EvalSymlinks(p)
		if err != nil || !filepath.IsAbs(p) || filepath.Clean(p) != p || physical != p {
			return fail("scoring roots must be canonical physical directories", errors.New("invalid scoring root"))
		}
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			return fail("scoring roots must be directories", errors.New("invalid scoring root"))
		}
	}
	source, err := filepath.EvalSymlinks(task.SourceRepo)
	if err != nil {
		return fail("source is unavailable", err)
	}
	for _, pin := range []struct{ root, oid string }{{source, task.ParentOID}, {source, task.LandingOID}, {workRoot, candidate}} {
		kind, err := runBenchPrepareCommand(ctx, pin.root, nil, "git", "--no-replace-objects", "-c", "protocol.allow=never", "-c", "protocol.file.allow=never", "cat-file", "-t", pin.oid)
		if err != nil {
			return fail("acceptance pins must name available local commits", err)
		}
		if strings.TrimSpace(string(kind)) != "commit" {
			return fail("acceptance pins must name local commits", errors.New("non-commit object pin"))
		}
	}
	for _, protected := range []string{source, workRoot} {
		if observerDir == protected || strings.HasPrefix(observerDir, protected+string(filepath.Separator)) {
			return fail("scoring scratch must be outside source and worker trees", errors.New("scoring root overlaps protected input"))
		}
	}
	patch, err := readBenchPatch(filepath.Join(observerDir, "acceptance.patch"))
	if err != nil {
		return fail("captured acceptance patch is unavailable", err)
	}
	digest := sha256.Sum256(patch)
	a.PatchSHA256 = hex.EncodeToString(digest[:])
	if a.PatchSHA256 != task.Acceptance.PatchSHA256 {
		return fail("captured acceptance patch digest differs", errors.New("immutable patch mismatch"))
	}
	if err := verifyBenchPatchPaths(ctx, observerDir, patch, task.Acceptance.Paths); err != nil {
		return fail("captured acceptance patch scope differs", err)
	}
	scoreDir, err := os.MkdirTemp(observerDir, "score-")
	if err != nil {
		return fail("cannot create owned scoring scratch", err)
	}
	if err := os.Mkdir(filepath.Join(scoreDir, "logs"), 0700); err != nil {
		return fail("cannot create scoring logs", err)
	}
	// Private bounded summary/logs survive disposable tree cleanup. It contains
	// no claim that an unavailable command was an assertion failure.
	defer func() {
		raw, err := json.Marshal(a)
		if err == nil {
			err = os.WriteFile(filepath.Join(scoreDir, "acceptance.json"), append(raw, '\n'), 0600)
		}
		if err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()
	if d.Execute == nil {
		d.Execute = runBenchScoreCommand
	}
	for _, phase := range []struct{ role, source, pin string }{{"parent", source, task.ParentOID}, {"landing", source, task.LandingOID}, {"candidate", workRoot, candidate}} {
		bundle := filepath.Join(scoreDir, phase.role)
		tree, err := prepareBenchScoreTree(ctx, bundle, phase.source, phase.pin)
		if err != nil {
			return fail(phase.role+" scoring copy is unavailable", err)
		}
		quiet := true
		cleanup := func() error {
			if !quiet {
				return nil
			}
			return cleanupBenchScoreTree(context.WithoutCancel(ctx), bundle, tree)
		}
		owner, err := readRunnerOwner(filepath.Join(tree, ".git"))
		if err != nil {
			return fail(phase.role+" scoring lease is unavailable", errors.Join(err, cleanup()))
		}
		gitdir, err := parseRunnerOwner(owner, tree, true)
		if err != nil {
			return fail(phase.role+" scoring lease is invalid", errors.Join(err, cleanup()))
		}
		leasePath := filepath.Join(gitdir, "locked")
		// Landing already owns exactly the hidden patch: reverse --check binds
		// those bytes without modifying its solution. Other copies apply once.
		args := []string{"apply", "--check"}
		if phase.role == "landing" {
			args = append(args, "--reverse")
		}
		if _, err := runBenchPrepareCommand(ctx, tree, patch, "git", args...); err != nil {
			return fail(phase.role+" hidden acceptance patch cannot be applied or verified", errors.Join(err, cleanup()))
		}
		if phase.role != "landing" {
			if _, err := runBenchPrepareCommand(ctx, tree, patch, "git", "apply"); err != nil {
				return fail(phase.role+" hidden acceptance patch failed", errors.Join(err, cleanup()))
			}
		}
		result, runErr := d.Execute(ctx, benchScoreCommand{Role: phase.role, Root: tree, WorkRoot: workRoot, Head: candidate, LogPath: filepath.Join(scoreDir, "logs", phase.role+".log"), Argv: append([]string(nil), task.Acceptance.Argv...)})
		quiet = result.Quiet
		if !quiet {
			runErr = errors.Join(&runnerCleanupUnconfirmed{Detail: fmt.Sprintf("hidden acceptance bundle %s retained; native Git worktree lease %s; reconcile the owned scorer before removing either", bundle, leasePath)}, runErr)
		}
		cleanupErr := cleanup()
		if runErr != nil || cleanupErr != nil || !quiet || result.TimedOut || result.Exit < 0 || result.Exit > 125 || ctx.Err() != nil {
			return fail(phase.role+" acceptance execution is unavailable", errors.Join(runErr, cleanupErr, ctx.Err(), errors.New("execution/startup/cleanup must be confirmed")))
		}
		switch phase.role {
		case "parent":
			if result.Exit == 0 {
				return fail("hidden acceptance is invalid: parent passed", errors.New("expected parent-red control"))
			}
			a.ParentRed = true
		case "landing":
			if result.Exit != 0 {
				return fail("hidden acceptance is invalid: landing failed", errors.New("expected landing-green control"))
			}
			a.LandingGreen = true
		case "candidate":
			passed := result.Exit == 0
			a.Status = "command"
			a.Reason = ""
			a.CommandPassed = &passed
		}
	}
	if err := bench.ValidateAcceptance(a); err != nil {
		return fail("acceptance result is inconsistent", err)
	}
	return a, nil
}

// Each bundle owns brand-new object storage. Its linked worktree gets a real
// Git lock recognized by the existing managed-run artifact lifetime owner.
// No original/source/work object directory, alternates or remote is reused.
func prepareBenchScoreTree(ctx context.Context, bundle, source, pin string) (string, error) {
	files, err := gitx.ReadLocalTree(ctx, source, pin, func(f gitx.TreeFile) bool { return !benchProcessArtifact(f.Path) && f.Path != benchPublicContextPath })
	if err != nil {
		return "", err
	}
	if err := validateBenchTree(files); err != nil {
		return "", err
	}
	if err := os.Mkdir(bundle, 0700); err != nil {
		return "", err
	}
	anchor, tree := filepath.Join(bundle, "anchor"), filepath.Join(bundle, "tree")
	if err := os.Mkdir(anchor, 0700); err != nil {
		return "", err
	}
	if err := materializeBenchTree(anchor, files); err != nil {
		return "", err
	}
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.name", "Model Bench Scorer"}, {"config", "user.email", "model-bench@localhost.invalid"}, {"config", "core.hooksPath", os.DevNull}, {"add", "-f", "-A"}, {"commit", "-qm", "bench: isolated acceptance copy"}} {
		if _, err := runBenchPrepareCommand(ctx, anchor, nil, "git", args...); err != nil {
			return "", err
		}
	}
	seed, err := runBenchPrepareCommand(ctx, anchor, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	committed, err := gitx.ReadLocalTree(ctx, anchor, strings.TrimSpace(string(seed)), nil)
	if err != nil {
		return "", err
	}
	byPath := map[string]gitx.TreeFile{}
	for _, f := range committed {
		byPath[f.Path] = f
	}
	for _, f := range files {
		actual, ok := byPath[f.Path]
		if !ok || actual.Mode != f.Mode || !bytes.Equal(actual.Content, f.Content) {
			return "", fmt.Errorf("Git changed scoring input %s", f.Path)
		}
	}
	if _, err := runBenchPrepareCommand(ctx, anchor, nil, "git", "worktree", "add", "--detach", "--no-checkout", tree, "HEAD"); err != nil {
		return "", err
	}
	if err := materializeBenchTree(tree, files); err != nil {
		return "", err
	}
	if _, err := runBenchPrepareCommand(ctx, anchor, nil, "git", "worktree", "lock", "--reason", "model-bench hidden acceptance", tree); err != nil {
		return "", err
	}
	return tree, nil
}

func cleanupBenchScoreTree(ctx context.Context, bundle, tree string) error {
	anchor := filepath.Join(bundle, "anchor")
	if _, err := runBenchPrepareCommand(ctx, anchor, nil, "git", "worktree", "unlock", tree); err != nil {
		return err
	}
	if _, err := runBenchPrepareCommand(ctx, anchor, nil, "git", "worktree", "remove", "--force", tree); err != nil {
		return err
	}
	return os.RemoveAll(bundle)
}

// A budget breach cancels the same managed owner that holds the queue/slot,
// rather than abandoning a still-writing process or silently truncating proof.
type benchScoreOutput struct {
	remaining int
	exceeded  bool
	cancel    context.CancelFunc
}

func (w *benchScoreOutput) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		w.exceeded = true
		w.cancel()
		return 0, errors.New("hidden acceptance output exceeds 8 MiB budget")
	}
	w.remaining -= len(p)
	return len(p), nil
}

func runBenchScoreCommand(ctx context.Context, c benchScoreCommand) (benchScoreRun, error) {
	helper, err := prepareSuiteHelper(ctx, c.WorkRoot, "")
	if err != nil {
		return benchScoreRun{Exit: -1, Quiet: true}, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	output := &benchScoreOutput{remaining: benchScoreLogBudget, cancel: cancel}
	lifetime := &runnerArtifactLifetime{retirable: true}
	d := localRunDeps{Root: c.WorkRoot, WorkflowAnchor: c.WorkRoot, Stderr: output, StrictGroupCleanup: true, runnerLifetime: lifetime}
	run, err := runUnderSlotBatch(ctx, d, c.Root, helper, filepath.Dir(c.LogPath), "bench-acceptance-"+c.Role, c.Head, 1, suitePrioritySubject{}, [][]string{c.Argv}, c.LogPath, benchScoreCommandTimeout)
	if output.exceeded {
		err = errors.Join(err, errors.New("hidden acceptance output exceeds 8 MiB budget"))
	}
	return benchScoreRun{Exit: run.Exit, TimedOut: run.TimedOut, Quiet: lifetime.mayRetire()}, err
}

var _ io.Writer = (*benchScoreOutput)(nil)
