// localcheck_run.go — the IO shell that runs one declared local check on an
// exact head and records the outcome (#345). The rules live in localchecks.go.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

// localRunDeps is what a local run touches: GitHub (to record), git in the
// repository root (worktree and log location), and where to report.
type localRunDeps struct {
	GH     ghCaller
	Repo   string // owner/name
	Root   string // the checkout's top level; Makefile.workflow resolves from here
	Git    func(args ...string) (string, error)
	PR     int // 0: no PR known, so a failure posts no comment
	Stderr io.Writer
	// SkipRecord is for a diagnostic baseline on main. It keeps the log but
	// never publishes a commit status or PR comment for that baseline.
	SkipRecord bool
	// minute is the unit of timeout_minutes; tests shrink it.
	minute time.Duration
}

const (
	localCheckLogKeep   = 20 // newest per-head log directories kept
	localCheckStaleAge  = 24 * time.Hour
	localCheckTmpPrefix = "sdlc-local-"
	localCheckTailLines = 200
)

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
	rec.Host, _ = os.Hostname()
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
	sweepStaleLocalWorktrees(d)

	wt, cleanup, err := detachedWorktree(d.Git, sha)
	if err != nil {
		return rec, fmt.Errorf("local check %s: %w", c.Name, err)
	}
	defer cleanup()

	fmt.Fprintf(d.Stderr, "==> local check %s on %s: %s\n    in %s\n    log %s\n", c.Name, abbrevSHA(sha), strings.Join(c.Command, " "), wt, rec.LogPath)
	heldBy, release, err := acquireSuiteSlot(ctx, d, wt, c.Weight, helper)
	if err != nil {
		return rec, fmt.Errorf("local check %s: %w", c.Name, err)
	}
	defer release()

	if err := os.MkdirAll(filepath.Dir(rec.LogPath), 0o755); err != nil {
		return rec, err
	}
	logf, err := os.Create(rec.LogPath)
	if err != nil {
		return rec, err
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(c.TimeoutMinutes)*d.minute)
	defer cancel()
	cmd := exec.CommandContext(runCtx, c.Command[0], c.Command[1:]...)
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
	// The slot covers the suite's run only (#383): free it before the log,
	// the GitHub writes and whatever the caller (merge) does next.
	release()
	rec.Duration = time.Since(start)
	_ = logf.Close()
	var exitErr *exec.ExitError
	switch {
	case runCtx.Err() == context.DeadlineExceeded:
		rec.TimedOut = true
	case ctx.Err() != nil:
		return rec, fmt.Errorf("local check %s interrupted: %w", c.Name, ctx.Err())
	case runErr == nil:
		rec.Exit = 0
	case errors.As(runErr, &exitErr):
		rec.Exit = exitErr.ExitCode()
	default:
		fmt.Fprintf(d.Stderr, "local check %s did not start: %v\n", c.Name, runErr)
	}
	if rec.LogSHA256, rec.Tail, err = hashAndTail(rec.LogPath, localCheckTailLines); err != nil {
		return rec, err
	}
	pruneLocalCheckLogs(logDir, localCheckLogKeep)
	if d.SkipRecord {
		fmt.Fprintf(d.Stderr, "==> local check %s: %s on %s (diagnostic baseline, unrecorded)\n", c.Name, rec.statusDescription(), abbrevSHA(sha))
		return rec, nil
	}

	if err := d.GH.PostStatus(ctx, d.Repo, sha, localStatusContext(c.Name), rec.state(), rec.statusDescription()); err != nil {
		return rec, fmt.Errorf("local check %s %s but was NOT recorded on %s: %w", c.Name, rec.state(), abbrevSHA(sha), err)
	}
	if !rec.passed() && d.PR > 0 {
		if err := d.GH.PRComment(ctx, d.Repo, d.PR, rec.failureComment()); err != nil {
			fmt.Fprintf(d.Stderr, "[!] local check %s: failure recorded, but the PR comment failed: %v\n", c.Name, err)
		}
	}
	fmt.Fprintf(d.Stderr, "==> local check %s: %s, recorded as %s on %s\n", c.Name, rec.statusDescription(), localStatusContext(c.Name), abbrevSHA(sha))
	return rec, nil
}

// detachedWorktree checks sha out in a fresh temporary worktree and returns
// its path and a cleanup that removes it (a run's, or a probe's, view of an
// exact commit whatever the slot has checked out).
func detachedWorktree(git func(args ...string) (string, error), sha string) (string, func(), error) {
	wt, err := os.MkdirTemp("", localCheckTmpPrefix+"*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_, _ = git("worktree", "remove", "--force", wt)
		_ = os.RemoveAll(wt)
		_, _ = git("worktree", "prune")
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
	cmd.Dir = dir
	cmd.Stderr = d.Stderr
	out, err := cmd.Output()
	state := strings.TrimSpace(string(out))
	if err != nil {
		release()
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
	sc := bufio.NewScanner(io.TeeReader(f, h))
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		tail = append(tail, sc.Text())
		if len(tail) > n {
			tail = tail[1:]
		}
	}
	if err := sc.Err(); err != nil {
		// An over-long line stops the scan; hash the rest so the digest covers the file.
		if _, cerr := io.Copy(h, f); cerr != nil {
			return "", nil, cerr
		}
	}
	return hex.EncodeToString(h.Sum(nil)), tail, nil
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
