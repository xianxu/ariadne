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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
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

// DispatchEvidence holds nonsemantic, bounded attempt diagnostics for the caller.
type DispatchEvidence struct {
	Preparation      []reviewprepare.Result
	Protocol         []reviewprepare.ProtocolResult
	ReviewerRuns     []reviewusage.Record
	UsageDiagnostics []string
}

// DispatchOptions configures one invocation.
type DispatchOptions struct {
	// Usage accounting is independent of semantic gate admission. Requested
	// persistence must succeed before launch; later failures retain partial
	// receipts and are surfaced in Evidence without advancing a gate round.
	UsageIntent   *reviewusage.Intent
	UsageRecorder func(context.Context, reviewusage.Record) error
	// Preparation executes only in Checkout, after confinement and before each review.
	Preparation     *reviewprepare.Config
	PreparationHead string
	Evidence        *DispatchEvidence
	// ValidateResponse is pure admission against a caller-pinned ledger. Invalid
	// output consumes the same two-attempt budget as a missing verdict.
	ValidateResponse func(string) error

	Agent        AgentCLI
	Prompt       string
	AllowedTools string // for claude; ignored by codex/gemini
	Stdout       io.Writer
	Stderr       io.Writer
	// Timeout is the caller's limit for the whole dispatch, retry included
	// (#300: sized by ReviewTimeout); zero is the 30-minute default.
	// WF_REVIEW_TIMEOUT, when set, overrides it.
	Timeout time.Duration
	// Wrap, when set, transforms the agent argv last (#204: run it inside a
	// disposable checkout, under an OS write fence when one applies). It is
	// applied in BuildArgs, so the dispatch and the --dry-run line agree.
	Wrap func(name string, args []string) (string, []string)
	// Containment is what bounds the agent's writes; agentArgs maps it to each
	// CLI's own permission mode (#310).
	Containment Containment
	// BeforeRetry, when set, runs before the no-verdict retry's attempt 2 (#204:
	// it resets the review checkout, so attempt 2 never reads attempt 1's
	// reverted or half-mutated tree). An error ends the dispatch with attempt 1's
	// text kept.
	BeforeRetry func() error
	// Checkout is the review checkout the agent runs in (#204), empty when it
	// runs in the operator's tree. Dispatch confines the agent around it, or
	// around the working directory's repository (#339).
	Checkout string
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
	var stdout, stderr bytes.Buffer
	err := runCommand(ctx, onStart, "", &stdout, &stderr, name, args...)
	return ProcessOutput{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

// runCommand owns environment, cancellation and descendant cleanup for both
// reviewer and preparation processes. Callers choose capture and exit policy.
func runCommand(ctx context.Context, onStart func(int), dir string, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	configureReviewProcess(cmd)
	cmd.Cancel = func() error { return terminateReviewProcess(cmd, false) }
	cmd.WaitDelay = reviewShutdownGrace
	if bin, err := ownerBinDir(); err == nil {
		cmd.Env = binAugmentedEnv(bin, os.Environ())
	}
	if c, ok := ctx.Value(confinementKey{}).(*reviewbox.Confinement); ok {
		if cmd.Env == nil {
			cmd.Env = os.Environ()
		}
		cmd.Env = binAugmentedEnv(c.BinDir(), overlayEnv(cmd.Env, c.Env()))
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if onStart != nil {
		onStart(cmd.Process.Pid)
	}
	err := cmd.Wait()
	_ = terminateReviewProcess(cmd, true)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// BuildArgs returns the argv (binary name + flags + final prompt) for
// invoking the chosen agent. Exposed for tests + --dry-run callers
// that want to print the would-be command line.
func BuildArgs(opts DispatchOptions) (name string, args []string, err error) {
	name, args, err = agentArgs(opts)
	if err != nil || opts.Wrap == nil {
		return name, args, err
	}
	name, args = opts.Wrap(name, args)
	return name, args, nil
}

// Containment is what bounds a judge's writes (#310). Every judge is a
// reviewer, so no level grants an auto-approve mode (bypassPermissions,
// --full-auto, --yolo). Each maps to the CLI's own permission mode, scoped by
// what surrounds the agent. claude's Bash stays unrestricted at every level:
// only the fence stops its shell writes.
type Containment int

const (
	// ContainNone runs in the operator's tree: codex is read-only, claude's
	// Edit/Write are refused.
	ContainNone Containment = iota
	// ContainCheckout runs in a disposable checkout with no OS fence (#204):
	// the agent's sandbox, where it has one, may write only its working
	// directory, the checkout.
	ContainCheckout
	// ContainFenced runs under an OS write fence that already denies the
	// source tree, so the agent's own sandbox is off: a Seatbelt sandbox
	// cannot nest inside another (#204).
	ContainFenced
)

// fencedClaudeSettings turns off Claude Code's Bash sandbox under an outer fence.
const fencedClaudeSettings = `{"sandbox":{"enabled":false}}`

func agentArgs(opts DispatchOptions) (name string, args []string, err error) {
	switch opts.Agent {
	case AgentClaude, "":
		args = []string{
			"-p",
			// #300: the whole run as events, so a late message can't erase
			// the verdict and a failed run is told apart from its prose.
			"--output-format", "stream-json", "--verbose",
			"--allowedTools", opts.AllowedTools,
			// #310: only the allowed tools run; Edit/Write are refused.
			"--permission-mode", "dontAsk",
		}
		if opts.Containment == ContainFenced {
			args = append(args, "--settings", fencedClaudeSettings)
		}
		return "claude", append(args, opts.Prompt), nil

	case AgentCodex:
		sandbox := "read-only"
		switch opts.Containment {
		case ContainCheckout:
			sandbox = "workspace-write"
		case ContainFenced:
			sandbox = "danger-full-access"
		}
		return "codex", []string{"exec", "--sandbox", sandbox, opts.Prompt}, nil

	case AgentGemini:
		// Headless gemini refuses every tool that needs approval, and a
		// reviewer never gets --yolo, so it can read but not write.
		args = []string{}
		if opts.Containment == ContainFenced {
			args = append(args, "--sandbox=false")
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
//
// Every agent runs confined (#339): its TMPDIR is a directory Dispatch made
// and proved writable, git can't find or create a repository outside its
// directories, and a change outside them afterward is an error, with the text.
// A failure to confine it means the review did not run.
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
	conf, err := confine(opts)
	if err != nil {
		return "", fmt.Errorf("reviewer confinement (#339), the review did not run: %w", err)
	}
	defer conf.Close()
	ctx = context.WithValue(ctx, confinementKey{}, conf)
	output, err = dispatchAttempts(ctx, opts, conf)
	if cerr := conf.Check(); cerr != nil {
		return output, errors.Join(err, cerr)
	}
	return output, err
}

// confine prepares the agent's confinement (#339): around its review checkout
// when it has one, else around the working directory's repository.
func confine(opts DispatchOptions) (*reviewbox.Confinement, error) {
	if opts.Checkout != "" {
		return reviewbox.Confine(opts.Checkout, true)
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return reviewbox.Confine(wd, false)
}

// confinementKey carries the dispatch's confinement to Run, so the agent's
// environment is confined without changing Run's argv or signature.
type confinementKey struct{}

// overlayEnv returns env with each KEY=value of over replacing env's KEY. Pure.
func overlayEnv(env, over []string) []string {
	keys := map[string]bool{}
	for _, kv := range over {
		k, _, _ := strings.Cut(kv, "=")
		keys[k] = true
	}
	out := make([]string, 0, len(env)+len(over))
	for _, kv := range env {
		if k, _, _ := strings.Cut(kv, "="); !keys[k] {
			out = append(out, kv)
		}
	}
	return append(out, over...)
}

// dispatchAttempts is the dispatch proper: one attempt, and the no-verdict
// retry (#271, #300).
func dispatchAttempts(ctx context.Context, opts DispatchOptions, conf *reviewbox.Confinement) (string, error) {
	var first string
	var invalid error
	for attempt := 1; attempt <= 2; attempt++ {
		if attempt == 2 {
			if opts.Stderr != nil {
				fmt.Fprintln(opts.Stderr, "    … reviewer response invalid; dispatching it once more:", invalid)
			}
			if opts.BeforeRetry != nil {
				if err := opts.BeforeRetry(); err != nil {
					return retryFailure(first, err, opts.ValidateResponse != nil)
				}
			}
			if err := conf.Prepare(); err != nil {
				return retryFailure(first, err, opts.ValidateResponse != nil)
			}
		}
		if err := runPreparation(ctx, opts, attempt); err != nil {
			return "", err
		}
		current := opts
		if opts.UsageIntent != nil {
			intent := *opts.UsageIntent
			intent.Attempt = attempt
			current.UsageIntent = &intent
		}
		if attempt == 2 {
			current.Prompt += retryNotice + "\nResponse validation: " + invalid.Error()
		}
		output, err := dispatchOnce(ctx, current)
		if err != nil {
			if attempt == 1 || opts.ValidateResponse != nil {
				return output, err
			}
			kept, _ := retryFailure(first, err, false)
			if errors.Is(err, context.DeadlineExceeded) {
				return kept, nil
			}
			return kept, err
		}
		invalid = nil
		if !HasVerdict(output) {
			invalid = errors.New("missing verdict")
		}
		if invalid == nil && opts.ValidateResponse != nil {
			invalid = opts.ValidateResponse(output)
		}
		if invalid == nil {
			return output, nil
		}
		if opts.Evidence != nil {
			opts.Evidence.Protocol = append(opts.Evidence.Protocol, reviewprepare.ProtocolResult{Attempt: attempt, Reason: invalid.Error(), Output: truncateDiagnostic(output)})
		}
		if attempt == 1 {
			first = output
			continue
		}
		if opts.ValidateResponse != nil {
			return "", fmt.Errorf("review response protocol failed after two attempts: %w", invalid)
		}
		return "## Attempt 1 (ended without a verdict)\n\n" + first + "\n\n## Attempt 2 (retry; also without a verdict)\n\n" + output, nil
	}
	panic("unreachable")
}

func retryFailure(first string, err error, validated bool) (string, error) {
	if validated {
		return "", err
	}
	return "## Attempt 1 (ended without a verdict)\n\n" + first + "\n\n## Attempt 2 (not started: " + err.Error() + ")\n", err
}

func truncateDiagnostic(s string) string {
	if len(s) > reviewprepare.OutputLimit {
		return s[:reviewprepare.OutputLimit] + "\n[truncated]"
	}
	return s
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

	var accounting reviewusage.Record
	accounted := opts.UsageIntent != nil || opts.UsageRecorder != nil
	accountingIndex := -1
	diagnose := func(stage string, err error) {
		diagnostic := stage + ": " + strings.Join(strings.Fields(err.Error()), " ")
		if len(diagnostic) > 256 {
			diagnostic = diagnostic[:256]
		}
		if opts.Evidence != nil {
			opts.Evidence.UsageDiagnostics = append(opts.Evidence.UsageDiagnostics, diagnostic)
		}
		if opts.Stderr != nil {
			fmt.Fprintln(opts.Stderr, "    … reviewer usage coverage incomplete:", diagnostic)
		}
	}
	copyAccounting := func() {
		if opts.Evidence == nil {
			return
		}
		if accountingIndex < 0 {
			accountingIndex = len(opts.Evidence.ReviewerRuns)
			opts.Evidence.ReviewerRuns = append(opts.Evidence.ReviewerRuns, accounting)
		} else {
			opts.Evidence.ReviewerRuns[accountingIndex] = accounting
		}
	}
	if accounted {
		if opts.UsageIntent == nil || opts.UsageRecorder == nil {
			return "", errors.New("review usage requires intent and recorder together")
		}
		intent := *opts.UsageIntent
		intent.ID, err = reviewusage.NewID()
		if err != nil {
			return "", err
		}
		intent.Agent = string(opts.Agent)
		if intent.Agent == "" {
			intent.Agent = "claude"
		}
		if intent.Attempt == 0 {
			intent.Attempt = 1
		}
		intent.RequestedAt = time.Now().UTC().Format(time.RFC3339Nano)
		intent.Root = opts.Checkout
		if intent.Root == "" {
			intent.Root, err = os.Getwd()
			if err != nil {
				return "", err
			}
		}
		intent.Root, err = filepath.Abs(intent.Root)
		if err != nil {
			return "", err
		}
		// CLI metadata records the physical cwd. Retain the absolute path
		// when it cannot be resolved; resolution grants no new permissions.
		if physical, resolveErr := filepath.EvalSymlinks(intent.Root); resolveErr == nil {
			intent.Root = physical
		}
		accounting, err = reviewusage.NewRequested(intent)
		if err != nil {
			return "", fmt.Errorf("review usage requested intent: %w", err)
		}
		copyAccounting()
		if err := opts.UsageRecorder(ctx, accounting); err != nil {
			diagnose("requested persistence", err)
			return "", fmt.Errorf("review usage requested persistence failed; reviewer was not launched: %w", err)
		}
	}

	// pid is written once by Run's onStart (at child launch) and read by the
	// heartbeat loop; atomic because the two live on different goroutines.
	var pid atomic.Int64
	onStart := func(p int) {
		pid.Store(int64(p))
		if !accounted {
			return
		}
		next, err := reviewusage.Step(accounting, reviewusage.Event{Kind: reviewusage.Start, Intent: accounting.Intent, PID: p, At: time.Now().UTC().Format(time.RFC3339Nano)})
		if err != nil {
			diagnose("start confirmation", err)
			return
		}
		accounting = next
		copyAccounting()
		if err := opts.UsageRecorder(ctx, accounting); err != nil {
			diagnose("started persistence", err)
		}
	}
	// One returned-process path serves synchronous, heartbeat, failed and
	// cancelled runs. Only this accounting write gets a detached 2s budget.
	returned := func(out ProcessOutput, runErr error) (string, error) {
		if accounted {
			outcome := "returned"
			switch {
			case accounting.Phase == reviewusage.Requested:
				if runErr != nil {
					outcome = "launch-failed"
				}
			case errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(runErr, context.DeadlineExceeded):
				outcome = "deadline"
			case errors.Is(ctx.Err(), context.Canceled) || errors.Is(runErr, context.Canceled):
				outcome = "canceled"
			case runErr != nil:
				outcome = "exit-error"
			}
			id, identityErr := reviewusage.CaptureIdentity(string(opts.Agent), out.Stdout, out.Stderr)
			diagnostic := ""
			if identityErr != nil {
				diagnostic = identityErr.Error()
			}
			if outcome == "launch-failed" {
				id = ""
			}
			if outcome == "returned" && (opts.Agent == AgentClaude || opts.Agent == "") {
				run := ReadStream(out.Stdout)
				if run.Result != nil && run.Result.IsError {
					outcome = "agent-error"
				}
			}
			next, err := reviewusage.Step(accounting, reviewusage.Event{Kind: reviewusage.Finish, Intent: accounting.Intent, PID: accounting.PID, StartedAt: accounting.StartedAt, At: time.Now().UTC().Format(time.RFC3339Nano), SessionID: id, Outcome: outcome, Diagnostic: diagnostic})
			if err != nil {
				diagnose("finish confirmation", err)
			} else {
				accounting = next
				copyAccounting()
				cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
				if err := opts.UsageRecorder(cleanup, accounting); err != nil {
					diagnose("finished persistence", err)
				}
				cancel()
			}
		}
		return classifyRunResult(ctx, out, runErr, name, opts.Agent, opts.Stderr)
	}

	// No progress sink → run synchronously, exactly as before. This is the fast
	// path (unit tests, quick dispatches) and stays free of goroutines/tickers.
	if opts.Stderr == nil {
		out, runErr := Run(ctx, onStart, name, args...)
		return returned(out, runErr)
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
			return returned(r.out, r.runErr)
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
