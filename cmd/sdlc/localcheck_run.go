// localcheck_run.go — the IO shell that runs one declared local check on an
// exact head and records the outcome (#345). The rules live in localchecks.go.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// localRunDeps is what a local run touches: GitHub (to record), git in the
// repository root (worktree and log location), and where to report.
type localRunDeps struct {
	GH         ghCaller
	Repo       string // owner/name
	Root       string // the checkout's top level; Makefile.workflow resolves from here
	Git        func(args ...string) (string, error)
	CleanupGit func(args ...string) (string, error) // independent bounded recovery IO
	PR         int                                  // 0: no PR known, so a failure posts no comment
	Stderr     io.Writer
	// SkipRecord is for a diagnostic baseline on main. It keeps the log but
	// never publishes a commit status or PR comment for that baseline.
	SkipRecord bool
	// FlakeRegistry reads the known-flake registry (#394) when a suite fails:
	// the open issues' flake_tests, from a fresh tracker read. nil means no
	// registry is available, which blocks like an unknown failure.
	FlakeRegistry func(ctx context.Context) (flakes.Registry, error)
	// minute is the unit of timeout_minutes; tests shrink it.
	minute time.Duration
}

const (
	localCheckLogKeep   = 20 // newest per-head log directories kept
	localCheckStaleAge  = 24 * time.Hour
	localCheckTmpPrefix = "sdlc-local-"
	localCheckTailLines = 200
)

var localCheckCleanupTimeout = landingRecoveryTimeout // tests shrink the existing recovery ceiling

// errUnboundedSuite: the machine-wide suite lock could not be taken, so the
// check does not run (#319/#341 bound heavy suites machine-wide).
var errUnboundedSuite = errors.New("refusing to run an unbounded suite")

// runLocalCheck runs c on sha in a fresh detached worktree, under one of the
// machine-wide suite slots, keeps the log, and normally posts the outcome as
// the localStatusContext(c.Name) status on sha. SkipRecord makes a diagnostic
// run on main local-only. A failed check is a record, not an error; the error
// is for a run that could not happen or, when required, could not be recorded.
func runLocalCheck(ctx context.Context, d localRunDeps, c localCheck, sha string) (localRunRecord, error) {
	rec := localRunRecord{Name: c.Name, SHA: sha, Exit: -1}
	rec.Host = hostName()
	if d.minute == 0 {
		d.minute = time.Minute
	}
	helper, err := prepareSuiteHelper(ctx, d.Root)
	if err != nil {
		return rec, fmt.Errorf("local check %s: %w", c.Name, err)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return rec, err
	}
	rec.LogPath = filepath.Join(logDir, sha, c.Name+".log")
	storeRoot := suiteStoreRoot(logDir)
	sweepStaleLocalWorktrees(d)

	cleanupGit := d.CleanupGit
	if cleanupGit == nil {
		cleanupGit = d.Git
	}
	wt, cleanup, err := detachedWorktreeWithCleanup(d.Git, cleanupGit, sha)
	if err != nil {
		return rec, fmt.Errorf("local check %s: %w", c.Name, err)
	}
	defer cleanup()

	writer, err := beginSuiteAttempt(storeRoot)
	publishing := false
	if err != nil {
		fmt.Fprintf(d.Stderr, "[!] local check %s: attempt log unavailable; using legacy log path: %v\n", c.Name, err)
	} else {
		defer func() {
			if !publishing {
				_ = os.RemoveAll(filepath.Dir(writer.LogPath))
			}
		}()
		rec.LogPath = writer.LogPath
	}
	fmt.Fprintf(d.Stderr, "==> local check %s on %s: %s\n    in %s\n    log %s\n", c.Name, abbrevSHA(sha), strings.Join(c.Command, " "), wt, rec.LogPath)
	run, err := runUnderSlot(ctx, d, wt, helper, logDir, c.Name, sha, c.Weight, c.Command, rec.LogPath, time.Duration(c.TimeoutMinutes)*d.minute)
	if err != nil {
		return rec, err
	}
	rec.Exit, rec.TimedOut, rec.Duration = run.Exit, run.TimedOut, run.Duration
	start, finished, queueStart, queueAcquired, queueWait := run.Start, run.Finished, run.QueueStart, run.Acquired, run.QueueWait
	if rec.LogSHA256, rec.Tail, err = hashAndTail(rec.LogPath, localCheckTailLines); err != nil {
		return rec, err
	}
	queueSeconds, wallSeconds := queueWait.Seconds(), rec.Duration.Seconds()
	saved := false // a saved attempt is what `local-check republish` can post later
	var gateErr error
	if writer != nil {
		attempt := SuiteAttempt{Version: suiteAttemptVersion, ID: writer.ID, Repo: d.Repo, SHA: sha, Check: c.Name,
			StartedAt: start.UTC(), FinishedAt: finished, MainBaseline: d.SkipRecord,
			QueuedAt: ptrTime(queueStart.UTC()), AcquiredAt: ptrTime(queueAcquired.UTC()),
			QueueSeconds: &queueSeconds, WallSeconds: &wallSeconds,
			Exit: rec.Exit, TimedOut: rec.TimedOut, LogSHA256: rec.LogSHA256}
		if identity, idErr := workspace.Resolve(execGitRunner{}, d.Root, ""); idErr == nil && identity.Address != nil {
			attempt.Slot = *identity.Address
		}
		if trailer, trailerErr := d.Git("show", "-s", "--format=%(trailers:key=Agent,valueonly)", sha); trailerErr == nil {
			attempt.Agent, attempt.Model, attempt.Effort = suiteAgentFields(trailer)
		}
		// A log the parser cannot read names no test, so a failure in it is
		// unattributable: never a known flake (#394).
		facts := SuiteLogFacts{Unattributed: []string{"the suite log could not be parsed"}}
		if info, statErr := os.Stat(writer.LogPath); statErr == nil && info.Size() <= 64<<20 {
			if log, readErr := os.ReadFile(writer.LogPath); readErr == nil {
				facts = parseSuiteLogFacts(string(log))
				attempt.Failures, attempt.Units = facts.Failures, facts.Units
				for _, warning := range facts.Warnings {
					fmt.Fprintf(d.Stderr, "[!] local check %s: log parser: %s\n", c.Name, warning)
				}
			}
		} else {
			fmt.Fprintf(d.Stderr, "[!] local check %s: log too large for timing/failure parser; raw log retained\n", c.Name)
		}
		var stepErr error
		switch {
		case rec.passed():
			attempt.Outcome = string(flakes.Passed)
		case c.Name == suiteCheckName && !d.SkipRecord:
			// #394: classify; when only registered flakes failed, rerun them.
			var o flakes.Outcome
			o, stepErr = knownFlakeStep(ctx, d, suiteGateRun{check: c, sha: sha, wt: wt, helper: helper, logDir: logDir,
				storeRoot: storeRoot, failedID: writer.ID, failedFacts: facts, repo: d.Repo, slot: attempt.Slot,
				agent: attempt.Agent, model: attempt.Model, effort: attempt.Effort})
			if stepErr == nil {
				attempt.Outcome, attempt.KnownFlakes = string(o.Kind), o.Known
				rec.Outcome, rec.KnownFlakeIssues = o.Kind, o.Issues()
			}
		}
		gateErr = stepErr
		publishing = true // retain failure residue for dead-owner cleanup if save fails
		if _, saveErr := publishSuiteAttempt(storeRoot, writer, attempt); saveErr != nil {
			fmt.Fprintf(d.Stderr, "[!] local check %s: attempt result was not fully saved: %v\n", c.Name, saveErr)
		} else {
			saved = true
			if attempt.Exit != 0 || attempt.TimedOut { // where publishSuiteAttempt kept it
				rec.LogPath = filepath.Join(storeRoot, "failed", writer.ID+".log")
			} else {
				rec.LogPath = filepath.Join(logDir, sha, c.Name+".log")
			}
			fmt.Fprintf(d.Stderr, "    saved log %s\n", rec.LogPath)
			if pruneErr := pruneSuiteAttempts(storeRoot, time.Now()); pruneErr != nil {
				fmt.Fprintf(d.Stderr, "[!] local check %s: attempt pruning: %v\n", c.Name, pruneErr)
			}
		}
	}
	pruneLocalCheckLogs(logDir, localCheckLogKeep)
	if gateErr != nil {
		// An interrupted or unrecordable known-flake rerun records nothing new:
		// the failed attempt is kept (unjudged) and no status is posted.
		return rec, fmt.Errorf("local check %s failed on %s and its known-flake rerun did not complete: %w; log %s", c.Name, abbrevSHA(sha), gateErr, rec.LogPath)
	}
	if d.SkipRecord {
		fmt.Fprintf(d.Stderr, "==> local check %s: %s on %s (diagnostic baseline, unrecorded)\n", c.Name, rec.statusDescription(), abbrevSHA(sha))
		return rec, nil
	}

	if err := postLocalCheckRecord(ctx, d, rec); err != nil {
		if saved {
			return rec, fmt.Errorf("local check %s %s but was NOT recorded on %s: %w; the result is saved: publish it with %q (no rerun needed)", c.Name, rec.state(), abbrevSHA(sha), err, republishCommand(c.Name, sha))
		}
		return rec, fmt.Errorf("local check %s %s but was NOT recorded on %s: %w", c.Name, rec.state(), abbrevSHA(sha), err)
	}
	fmt.Fprintf(d.Stderr, "==> local check %s: %s, recorded as %s on %s\n", c.Name, rec.statusDescription(), localStatusContext(c.Name), abbrevSHA(sha))
	return rec, nil
}

// suiteSlotRun is the outcome of one command run under a machine-wide suite slot.
type suiteSlotRun struct {
	Exit                                  int // -1: timed out or did not start
	TimedOut                              bool
	Duration, QueueWait                   time.Duration
	QueueStart, Acquired, Start, Finished time.Time
}

// runUnderSlot runs argv in wt under one of the machine-wide suite slots
// (weight of them), its output to logPath and d.Stderr, within timeout. The
// run lease (#400), owned signals and the slot last exactly the command and
// are released the moment it exits (#383), so a suite and a known-flake rerun
// (#394) are both cancellable and neither holds a slot while it records. An
// interrupt is an error and records nothing.
func runUnderSlot(ctx context.Context, d localRunDeps, wt, helper, logDir, check, sha string, weight int, argv []string, logPath string, timeout time.Duration) (suiteSlotRun, error) {
	run := suiteSlotRun{Exit: -1}
	queueStart := time.Now()
	// Own signals only while our queue/suite resources exist. The surrounding
	// Git/network preparation and bookkeeping retain default termination.
	ownedCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	// The lease (#400) spans exactly the stretch where SIGTERM is ours, so
	// `local-check cancel` always lands on the managed interruption path.
	pidStarted, _ := processStartTime(os.Getpid())
	lease := runLease{Root: canonicalRoot(d.Root), Check: check, SHA: sha, PID: os.Getpid(), PIDStarted: pidStarted, Started: time.Now().UTC()}
	runsDir := localCheckRunsDir(logDir)
	sweepStaleRunLeases(runsDir)
	releaseLease, err := holdRunLease(runsDir, lease)
	if err != nil {
		releaseLease = func() {}
		fmt.Fprintf(d.Stderr, "[!] local check %s: no run lease, so `sdlc local-check cancel` cannot find this run: %v\n", check, err)
	}
	defer releaseLease()
	heldBy, release, err := acquireSuiteSlot(ownedCtx, d, wt, weight, helper)
	if err != nil {
		return run, fmt.Errorf("local check %s: %w", check, err)
	}
	queueAcquired := time.Now()
	// Derive the recorded duration from the same wall-clock endpoints that are
	// serialized. time.Sub on live values can use their monotonic readings,
	// which need not equal the persisted timestamp difference.
	run.QueueStart, run.Acquired = queueStart.UTC(), queueAcquired.UTC()
	run.QueueWait = run.Acquired.Sub(run.QueueStart)
	defer release()

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return run, err
	}
	logf, err := os.Create(logPath)
	if err != nil {
		return run, err
	}
	runCtx, cancel := context.WithTimeout(ownedCtx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.Dir = wt
	cmd.WaitDelay = 5 * time.Second
	cmd.Env = os.Environ()
	if heldBy != "" {
		cmd.Env = append(cmd.Env, "WF_TEST_LOCK_HELD_BY="+heldBy)
	}
	out := io.MultiWriter(logf, d.Stderr)
	cmd.Stdout, cmd.Stderr = out, out
	start := time.Now()
	runErr := cmd.Run()
	// WaitDelay bounds inherited-pipe draining after the parent exits;
	// stop remaining children before closing their log or removing their tree.
	_ = processgroup.Terminate(cmd, true)
	// The slot covers the run only (#383): start its release before the log,
	// the GitHub writes and whatever the caller (merge) does next. The watcher
	// frees it on the holder's exit (at once with kqueue, within its 0.5 s
	// poll elsewhere).
	release()
	// Signal ownership ends here, so the lease does too: a cancel waiting on it
	// returns once the check has stopped and its slot is freed.
	releaseLease()
	interrupted := ownedCtx.Err() // stopSignals itself cancels ownedCtx.
	stopSignals()
	run.Start, run.Duration = start.UTC(), time.Since(start)
	run.Finished = time.Now().UTC()
	_ = logf.Close()
	var exitErr *exec.ExitError
	switch {
	case runCtx.Err() == context.DeadlineExceeded:
		run.TimedOut = true
	case interrupted != nil:
		return run, fmt.Errorf("local check %s interrupted: %w", check, interrupted)
	case runErr == nil:
		run.Exit = 0
	case errors.As(runErr, &exitErr):
		run.Exit = exitErr.ExitCode()
	default:
		fmt.Fprintf(d.Stderr, "local check %s did not start: %v\n", check, runErr)
	}
	return run, nil
}

func suiteAgentFields(raw string) (*string, *string, *string) {
	parts := strings.Split(strings.TrimSpace(raw), "/")
	if len(parts) == 1 && parts[0] != "" && !strings.ContainsAny(parts[0], " \t\r\n") {
		return &parts[0], nil, nil
	}
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" || strings.ContainsAny(strings.Join(parts, ""), " \t\r\n") {
		return nil, nil, nil
	}
	return &parts[0], &parts[1], &parts[2]
}

// localCheckCleanupGit preserves the existing Git seam with an independent
// recovery budget: cancellation must not skip or indefinitely stall retirement.
func localCheckCleanupGit(ctx context.Context, r gitRunner, root string) func(...string) (string, error) {
	r = newBoundedLandingGit(context.WithoutCancel(commandContext(ctx)), r, localCheckCleanupTimeout)
	return func(args ...string) (string, error) {
		out, err := r.GitInDir(root, args...)
		return string(out), err
	}
}

// detachedWorktree checks sha out in a fresh temporary worktree and returns
// its path and a cleanup that removes it (a run's, or a probe's, view of an
// exact commit whatever the slot has checked out).
func detachedWorktree(git func(args ...string) (string, error), sha string) (string, func(), error) {
	return detachedWorktreeWithCleanup(git, git, sha)
}

func detachedWorktreeWithCleanup(git, cleanupGit func(args ...string) (string, error), sha string) (string, func(), error) {
	wt, err := os.MkdirTemp("", localCheckTmpPrefix+"*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_, _ = cleanupGit("worktree", "remove", "--force", wt)
		_ = os.RemoveAll(wt)
		_, _ = cleanupGit("worktree", "prune")
	}
	if out, err := git("worktree", "add", "--detach", wt, sha); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("worktree at %s: %w\n%s", sha, err, out)
	}
	return wt, cleanup, nil
}

// localCheckLogDir is <git common dir>/sdlc/local-checks: never committed, shared
// by the slot's worktrees, gone with the clone.
func localCheckLogDir(d localRunDeps) (string, error) {
	out, err := d.Git("rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("local check: git common dir: %w", err)
	}
	dir := strings.TrimSpace(out)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(d.Root, dir)
	}
	return filepath.Join(dir, "sdlc", "local-checks"), nil
}

// acquireSuiteSlot takes one of the machine-wide full-suite slots through the
// prepared helper Makefile.workflow uses, so a local check and `make test`
// share one cap. Capability preflight has already run before any worktree
// mutation.
//
// The slot lasts one run (#383): the helper watches a per-run holder child (a
// `cat` on a pipe from this process), not this process, so release (close the
// pipe, reap the child) frees it, and so does this process dying by any route
// (the kernel closes the pipe). A later run in the same process starts a new
// holder and queues for a slot like anyone else. It returns the pid to export
// as WF_TEST_LOCK_HELD_BY (the holder's, which the slot record names) and the
// release, which is idempotent.
//
// weight is the slots the run takes at once (#346): passed as --slots, so a
// make test inside it, which then sees WF_TEST_LOCK_HELD_BY and goes reentrant,
// still runs at its declared weight (#332).
func acquireSuiteSlot(ctx context.Context, d localRunDeps, dir string, weight int, helper string) (string, func(), error) {
	holder := exec.Command("cat")
	processgroup.Configure(holder)
	pipe, err := holder.StdinPipe()
	if err != nil {
		return "", nil, fmt.Errorf("%w: suite slot holder: %v", errUnboundedSuite, err)
	}
	if err := holder.Start(); err != nil {
		return "", nil, fmt.Errorf("%w: suite slot holder: %v", errUnboundedSuite, err)
	}
	var once sync.Once
	release := func() {
		once.Do(func() {
			_ = pipe.Close()
			_ = holder.Wait()
		})
	}
	pid := fmt.Sprint(holder.Process.Pid)
	cmd := exec.CommandContext(ctx, "python3", helper, "acquire", "--watch", pid, "--slots", strconv.Itoa(max(weight, 1)))
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	cmd.Dir = dir
	// The run's worktree is a temp directory; name the owning checkout in the
	// slot record so the queue's held-by lines identify the slot (#400).
	cmd.Env = append(os.Environ(), "WF_TEST_LOCK_OWNER="+d.Root)
	cmd.Stderr = d.Stderr
	out, err := cmd.Output()
	state := strings.TrimSpace(string(out))
	if err != nil {
		release()
		if ctx.Err() != nil {
			return "", nil, fmt.Errorf("%w: suite lock interrupted: %w", errUnboundedSuite, ctx.Err())
		}
		return "", nil, fmt.Errorf("%w: suite lock helper: %v", errUnboundedSuite, err)
	}
	switch state {
	case "acquired":
		return pid, release, nil
	case "reentrant":
		// Nested under a make that holds the slot: no slot of ours to hold.
		release()
		return os.Getenv("WF_TEST_LOCK_HELD_BY"), func() {}, nil
	case "skipped":
		release()
		return "", func() {}, nil
	default:
		release()
		return "", nil, fmt.Errorf("%w: suite lock answered %q (see the [test-lock] lines above)", errUnboundedSuite, state)
	}
}

// hashAndTail reads the log once: its sha256 and its last n lines.
func hashAndTail(path string, n int) (string, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	h := sha256.New()
	var tail []string
	r := bufio.NewReaderSize(io.TeeReader(f, h), 64<<10)
	line := make([]byte, 0, 64<<10)
	add := func() {
		if n > 0 {
			tail = append(tail, string(bytes.TrimSuffix(line, []byte{'\r'})))
			if len(tail) > n {
				tail = tail[1:]
			}
		}
		line = line[:0]
	}
	for {
		chunk, readErr := r.ReadSlice('\n')
		if len(line) < 1<<20 {
			room := (1 << 20) - len(line)
			part := bytes.TrimSuffix(chunk, []byte{'\n'})
			if len(part) > room {
				part = part[:room]
			}
			line = append(line, part...)
		}
		switch readErr {
		case nil:
			add()
		case bufio.ErrBufferFull:
			continue
		case io.EOF:
			if len(line) > 0 {
				add()
			}
			return hex.EncodeToString(h.Sum(nil)), tail, nil
		default:
			return "", nil, readErr
		}
	}
}

// pruneLocalCheckLogs keeps the newest keep per-head directories under dir.
func pruneLocalCheckLogs(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type aged struct {
		name string
		mod  time.Time
	}
	var dirs []aged
	for _, e := range entries {
		if info, err := e.Info(); err == nil && e.IsDir() {
			dirs = append(dirs, aged{e.Name(), info.ModTime()})
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].mod.After(dirs[j].mod) })
	for i := keep; i < len(dirs); i++ {
		_ = os.RemoveAll(filepath.Join(dir, dirs[i].name))
	}
}

// sweepStaleLocalWorktrees collects worktrees a killed run left behind: temp
// directories past the longest possible run, then git's record of them.
func sweepStaleLocalWorktrees(d localRunDeps) {
	matches, _ := filepath.Glob(filepath.Join(os.TempDir(), localCheckTmpPrefix+"*"))
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && time.Since(info.ModTime()) > localCheckStaleAge {
			_ = os.RemoveAll(m)
		}
	}
	_, _ = d.Git("worktree", "prune")
}
