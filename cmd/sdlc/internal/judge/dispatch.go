package judge

import (
	"bytes"
	"context"
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
var Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir, err := ownerBinDir(); err == nil {
		cmd.Env = binAugmentedEnv(dir, os.Environ())
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return ProcessOutput{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
	}
	if onStart != nil {
		onStart(cmd.Process.Pid)
	}
	err := cmd.Wait()
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
// stdout only. Captured process stderr is forwarded to opts.Stderr when one is
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
func Dispatch(ctx context.Context, opts DispatchOptions) (output string, err error) {
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
		return classifyRunResult(out, runErr, name, nil)
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
			return classifyRunResult(r.out, r.runErr, name, opts.Stderr)
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
func classifyRunResult(out ProcessOutput, runErr error, name string, diagnostics io.Writer) (string, error) {
	if diagnostics != nil && len(out.Stderr) > 0 {
		_, _ = diagnostics.Write(out.Stderr)
	}
	if _, ok := runErr.(*exec.ExitError); ok {
		return string(out.Stdout), nil
	}
	if runErr != nil {
		dir, derr := ownerBinDir()
		if derr != nil || dir == "" {
			dir = "?"
		}
		return string(out.Stdout), fmt.Errorf("dispatch %s (owner bin %q prepended to PATH=%s): %w", name, dir, os.Getenv("PATH"), runErr)
	}
	return string(out.Stdout), nil
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
