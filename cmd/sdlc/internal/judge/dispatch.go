package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// AgentCLI names a coding-agent CLI. The default is "claude"; the
// shell script supports "codex" and "gemini" via $AGENT_CMD. We mirror
// that surface so `make check-*` shims (env-driven) and `sdlc judge`
// (flag-driven) target the same agents.
type AgentCLI string

const (
	AgentClaude AgentCLI = "claude"
	AgentCodex  AgentCLI = "codex"
	AgentGemini AgentCLI = "gemini"
)

// DispatchOptions configures one invocation.
type DispatchOptions struct {
	Agent        AgentCLI
	Prompt       string
	AllowedTools string // for claude; ignored by codex/gemini
	IsSandbox    bool   // if true, codex/gemini get auto-approve flags
	Stdout       io.Writer
	Stderr       io.Writer
	// Timeout is the caller's limit for the whole dispatch, retry included
	// (#300: sized by ReviewTimeout); zero is the 30-minute default.
	// WF_REVIEW_TIMEOUT, when set, overrides it.
	Timeout time.Duration
}

// ownerBinDir is the directory of the running sdlc binary — i.e. the owner
// `bin/` (e.g. .../ariadne/bin), resolved from os.Executable(). The single
// source for "where do sibling tools (sdlc, weave, …) live", consumed by both
// Run (to build the subprocess PATH) and Dispatch (to diagnose launch failures).
// Works unchanged from a downstream repo: the binary is .../ariadne/bin/sdlc
// regardless of cwd (#138).
func ownerBinDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// binAugmentedEnv returns env with binDir prepended to its PATH entry (or a
// synthesized PATH= entry when none exists), so a spawned agent can resolve
// `sdlc` and its sibling owner-bin tools even when the spawning shell's startup
// files never put that dir on PATH (#138). No-op when binDir is empty/".". Pure.
func binAugmentedEnv(binDir string, env []string) []string {
	if binDir == "" || binDir == "." {
		return env
	}
	out := make([]string, 0, len(env)+1)
	found := false
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, "PATH="); ok {
			found = true
			out = append(out, "PATH="+binDir+string(os.PathListSeparator)+v)
		} else {
			out = append(out, e)
		}
	}
	if !found {
		out = append(out, "PATH="+binDir)
	}
	return out
}

// ProcessOutput preserves the two process channels across the replaceable Run
// seam. Stdout is the agent's semantic response; Stderr is harness diagnostics
// and progress that Dispatch may route to the operator's diagnostic sink.
type ProcessOutput struct {
	Stdout []byte
	Stderr []byte
}

// Run is the package-level subprocess shim. Tests replace it with a
// fake to assert the right command line / capture without spawning a
// real agent process. Production execs the binary — with the owner bin/
// prepended to PATH so the agent can resolve `sdlc` (#138).
//
// onStart (nil ok) is invoked once with the child PID immediately after a
// successful launch — before the (potentially minutes-long) Wait — so a caller
// can report liveness while the agent runs (#140). We hand-roll Start→Wait
// instead of CombinedOutput to get that hook and preserve stdout/stderr as
// distinct semantic and diagnostic channels (#201).
var Run = runProcess

// Five seconds is the maximum graceful shutdown/pipe-drain interval. Tests use
// a shorter interval while exercising the same actual subprocess boundary.
var reviewShutdownGrace = 5 * time.Second

func runProcess(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	configureReviewProcess(cmd)
	cmd.Cancel = func() error { return terminateReviewProcess(cmd, false) }
	cmd.WaitDelay = reviewShutdownGrace
	if dir, err := ownerBinDir(); err == nil {
		cmd.Env = binAugmentedEnv(dir, os.Environ())
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		return ProcessOutput{}, err
	}
	if onStart != nil {
		onStart(cmd.Process.Pid)
	}

	// WaitDelay bounds both graceful cancellation and inherited pipe draining.
	// Its escalation kills only the direct child; finish by killing remaining
	// members of our own process group before returning. Wait reaps the direct
	// child and joins the output-copy goroutines, so no runner work outlives us.
	err := cmd.Wait()
	_ = terminateReviewProcess(cmd, true)
	return ProcessOutput{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

// BuildArgs returns the argv (binary name + flags + final prompt) for
// invoking the chosen agent. Exposed for tests + --dry-run callers
// that want to print the would-be command line.
func BuildArgs(opts DispatchOptions) (name string, args []string, err error) {
	switch opts.Agent {
	case AgentClaude, "":
		args = []string{
			"-p",
			// #300: the whole run as events, so a late message can't erase
			// the verdict and a failed run is told apart from its prose.
			"--output-format", "stream-json", "--verbose",
			"--allowedTools", opts.AllowedTools,
			"--permission-mode", "bypassPermissions",
			opts.Prompt,
		}
		return "claude", args, nil

	case AgentCodex:
		args = []string{"exec"}
		if opts.IsSandbox {
			args = append(args, "--full-auto")
		}
		args = append(args, opts.Prompt)
		return "codex", args, nil

	case AgentGemini:
		args = []string{}
		if opts.IsSandbox {
			args = append(args, "--yolo")
		}
		args = append(args, "-p", opts.Prompt)
		return "gemini", args, nil

	default:
		return "", nil, fmt.Errorf("unknown agent: %q (supported: claude, codex, gemini)", opts.Agent)
	}
}

// Dispatch invokes the agent CLI with the given prompt and returns semantic
// stdout only (#300: on an error from a retry, the text still carries the
// first attempt, so a caller that shows output on error loses no evidence). Captured process stderr is forwarded to opts.Stderr when one is
// configured; it never enters verdict parsing or durable review artifacts.
// Outcome classification is the caller's responsibility via Classify().
//
// Exit-code policy (review I3):
//
//   - subprocess fails to launch (binary missing, permission denied,
//     ctx cancelled, unknown agent name) → return error.
//   - subprocess runs and exits non-zero (any output) → swallow the
//     exit error, return the output for Classify(). Matches shell's
//     `|| true` and lets agents emit "found X violations, exit 1"
//     without us treating it as a binary-launch failure.
//
// In particular: empty-output + non-zero exit is *not* a launch failure
// — Classify() will mark it as Failure based on the empty-output rule.
// This keeps the binary/agent failure modes cleanly separated.
//
// A run whose text carries no verdict (#271: a reviewer ending on "I'll wait
// for the background run") is dispatched once more, inside the same deadline;
// if that one has none either, both runs' text is returned, labelled, for the
// caller's fail-safe and the sidecar (#300).
func Dispatch(ctx context.Context, opts DispatchOptions) (output string, err error) {
	timeout, err := EffectiveTimeout(os.Getenv("WF_REVIEW_TIMEOUT"), opts.Timeout)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	first, err := dispatchOnce(ctx, opts)
	if err != nil || HasVerdict(first) {
		return first, err
	}
	if opts.Stderr != nil {
		fmt.Fprintln(opts.Stderr, "    … the reviewer ended without a verdict; dispatching it once more")
	}
	again := opts
	again.Prompt = opts.Prompt + retryNotice
	second, err := dispatchOnce(ctx, again)
	if err != nil {
		// Once a retry has started, nothing may drop attempt 1: the returned
		// text carries it with the error. Running out of time is the fail-safe
		// (no verdict, so the caller halts and the sidecar keeps the text);
		// any other error is returned beside that text for the caller to show.
		kept := "## Attempt 1 (ended without a verdict)\n\n" + first + "\n\n## Attempt 2 (retry; " + err.Error() + ")\n"
		if errors.Is(err, context.DeadlineExceeded) {
			return kept, nil
		}
		return kept, err
	}
	if HasVerdict(second) {
		return second, nil
	}
	return "## Attempt 1 (ended without a verdict)\n\n" + first + "\n\n## Attempt 2 (retry; also without a verdict)\n\n" + second, nil
}

// retryNotice is appended to a retried review's prompt (#300).
const retryNotice = "\n\nNOTE: a previous attempt of this review ended without a verdict. You run " +
	"non-interactively and receive no notifications: run every command in the foreground, " +
	"and end your final message with the verdict and findings blocks."

// dispatchOnce runs the agent once and returns its review text.
func dispatchOnce(ctx context.Context, opts DispatchOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("review interrupted: %w", err)
	}
	name, args, err := BuildArgs(opts)
	if err != nil {
		return "", err
	}

	// pid is written once by Run's onStart (at child launch) and read by the
	// heartbeat loop; atomic because the two live on different goroutines.
	var pid atomic.Int64
	onStart := func(p int) { pid.Store(int64(p)) }

	// No progress sink → run synchronously, exactly as before. This is the fast
	// path (unit tests, quick dispatches) and stays free of goroutines/tickers.
	if opts.Stderr == nil {
		out, runErr := Run(ctx, onStart, name, args...)
		return classifyRunResult(ctx, out, runErr, name, opts.Agent, nil)
	}

	// Progress path (#140): the agent can run for minutes. Run it on a background
	// goroutine and, until it returns, emit a heartbeat to opts.Stderr every
	// heartbeatInterval showing elapsed + agent + child PID — the automated form
	// of the operator's manual `ps` inspection. The captured output and the
	// exit-code policy are identical to the fast path, so verdict parsing and
	// classification downstream are untouched.
	type runResult struct {
		out    ProcessOutput
		runErr error
	}
	done := make(chan runResult, 1)
	go func() {
		out, runErr := Run(ctx, onStart, name, args...)
		done <- runResult{out, runErr}
	}()

	start := time.Now()
	ticks, stop := newHeartbeatTicker(heartbeatInterval)
	defer stop()
	for {
		select {
		case r := <-done:
			return classifyRunResult(ctx, r.out, r.runErr, name, opts.Agent, opts.Stderr)
		case <-ticks:
			fmt.Fprintln(opts.Stderr, heartbeatLine(sinceStart(start), string(opts.Agent), int(pid.Load())))
		}
	}
}

// classifyRunResult is the single process→semantic transition shared by the
// synchronous and heartbeat paths (ARCH-DRY). It forwards diagnostic stderr,
// returns semantic stdout, and applies Dispatch's existing exit-code policy: a
// non-zero exit is swallowed so Classify can interpret the response, while a
// real launch failure returns a diagnosable error naming owner bin/ + PATH.
func classifyRunResult(ctx context.Context, out ProcessOutput, runErr error, name string, agent AgentCLI, diagnostics io.Writer) (string, error) {
	if diagnostics != nil && len(out.Stderr) > 0 {
		_, _ = diagnostics.Write(out.Stderr)
	}
	// Cancellation invalidates even a complete-looking response or a clean exit.
	// Check before the intentional nonzero-exit compatibility rule below.
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("review interrupted: %w", err)
	}
	if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
		return "", fmt.Errorf("review interrupted: %w", runErr)
	}
	// Only claude answers as a stream; codex and gemini prose is never parsed
	// as events, so a quoted event line in their review can't fail it.
	run := AgentRun{Tail: string(out.Stdout)}
	if agent == AgentClaude || agent == "" {
		run = ReadStream(out.Stdout)
	}
	_, exited := runErr.(*exec.ExitError)
	if err := RunFailure(agent, run, out.Stderr, exited); err != nil {
		return "", err
	}
	if exited {
		return run.Text(), nil
	}
	if runErr != nil {
		dir, derr := ownerBinDir()
		if derr != nil || dir == "" {
			dir = "?"
		}
		return run.Text(), fmt.Errorf("dispatch %s (owner bin %q prepended to PATH=%s): %w", name, dir, os.Getenv("PATH"), runErr)
	}
	return run.Text(), nil
}

// FormatCommandLine returns a shell-safe rendering of the would-be
// command, suitable for printing under --dry-run. It does NOT actually
// exec anything.
func FormatCommandLine(opts DispatchOptions) (string, error) {
	name, args, err := BuildArgs(opts)
	if err != nil {
		return "", err
	}
	parts := []string{name}
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " "), nil
}

// shellQuote wraps strings containing whitespace or shell metacharacters
// in single quotes (with internal single quotes escaped). Used only for
// --dry-run display; production Dispatch passes args through exec
// directly so quoting isn't an exec-safety concern.
func shellQuote(s string) string {
	if !strings.ContainsAny(s, " \t\n'\"$`\\|&;<>(){}*?[]#~=") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ReviewTimeout sizes a review's limit by the lines it reads (#300, D3 of the
// 2026-10-09 evidence: reviews of large windows routinely outran a flat 30
// minutes): 30 minutes up to 500 added lines, 15 more per further 1,000
// (rounded up), at most the 2-hour ceiling WF_REVIEW_TIMEOUT also has.
func ReviewTimeout(addedLines int) time.Duration {
	d := 30 * time.Minute
	if extra := addedLines - 500; extra > 0 {
		d += time.Duration((extra+999)/1000) * 15 * time.Minute
	}
	return min(d, 2*time.Hour)
}

// EffectiveTimeout is the dispatch's limit, the one statement of its
// precedence: WF_REVIEW_TIMEOUT when set, else the caller's sized limit, else
// the 30-minute default. Callers that report the limit read it from here.
func EffectiveTimeout(env string, sized time.Duration) (time.Duration, error) {
	if env == "" && sized > 0 {
		return sized, nil
	}
	return reviewTimeout(env)
}

// reviewTimeout keeps the operating bound independent of command callers.
// A parent context with an earlier deadline remains authoritative.
func reviewTimeout(value string) (time.Duration, error) {
	if value == "" {
		return 30 * time.Minute, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil || d < time.Second || d > 2*time.Hour {
		return 0, fmt.Errorf("WF_REVIEW_TIMEOUT must be a Go duration from 1s through 2h (unset, the limit scales with the review window)")
	}
	return d, nil
}
