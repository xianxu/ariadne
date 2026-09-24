package acquire

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// GitRunner is the acquisition process boundary. Completed Git failures retain
// ExitCode so a missing value can be distinguished from failed IO. Cancellation
// and resource limits invalidate that predicate result and do not expose it.
type GitRunner interface {
	Run(context.Context, string, ...string) (string, error)
	RunOwned(context.Context, string, string, ...string) (string, error)
}

// Client holds the Git boundary for one acquisition operation. Its zero value
// uses real Git; injected clients share the same restore/probe/publication code.
type Client struct {
	Git    GitRunner
	Policy *Policy
	// MaxLayers includes the host; zero leaves traversal unlimited.
	MaxLayers int
	// MaxDeclarationBytes also requires ordinary declaration files when positive.
	MaxDeclarationBytes int64
}

// ExecGit opts into exact stdout and bounded execution for refresh. Zero options
// retain acquisition's combined, whitespace-trimmed output behavior.
type ExecGit struct {
	ExtraFiles []*os.File
	Raw        bool
	Timeout    time.Duration
	// MaxOutputBytes bounds stdout and stderr together; zero leaves output unlimited.
	MaxOutputBytes int
}

func (g ExecGit) Run(ctx context.Context, dir string, args ...string) (string, error) {
	if g.Raw || g.Timeout > 0 || g.MaxOutputBytes > 0 {
		return g.runBounded(ctx, dir, "", args...)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.ExtraFiles = g.ExtraFiles
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v in %s: %w: %s", args, dir, err, strings.TrimSpace(string(b)))
	}
	return strings.TrimSpace(string(b)), nil
}
func (g ExecGit) RunOwned(ctx context.Context, dir, stage string, args ...string) (string, error) {
	if stage == "" {
		return "", fmt.Errorf("producer stage is required")
	}
	if g.Raw || g.Timeout > 0 || g.MaxOutputBytes > 0 {
		return g.runBounded(ctx, dir, stage, args...)
	}
	var output bytes.Buffer
	runner := weavefs.ExecRunner{Context: ctx, Stdout: &output, Stderr: &output, ExtraFiles: g.ExtraFiles}
	err := runner.RunOwned(dir, append([]string{"git"}, args...), stage)
	if err != nil {
		return "", fmt.Errorf("git %v in %s: %w: %s", args, dir, err, strings.TrimSpace(output.String()))
	}
	return strings.TrimSpace(output.String()), nil
}
func (c Client) gitOwned(ctx context.Context, dir, stage string, args ...string) (string, error) {
	runner := c.Git
	if runner == nil {
		runner = ExecGit{}
	}
	out, err := runner.RunOwned(ctx, dir, stage, args...)
	return strings.TrimSuffix(out, "\n"), err
}

func (c Client) git(ctx context.Context, dir string, args ...string) (string, error) {
	runner := c.Git
	if runner == nil {
		runner = ExecGit{}
	}
	out, err := runner.Run(ctx, dir, args...)
	return strings.TrimSuffix(out, "\n"), err
}

// Origin inspects an existing repository. An empty result means Git successfully
// read its local configuration and confirmed that remote.origin.url is absent.
// Plain non-repository directories are the caller's separate local-only case.
func Origin(ctx context.Context, dir string) (string, error) { return (Client{}).Origin(ctx, dir) }
func (c Client) Origin(ctx context.Context, dir string) (string, error) {
	if _, err := c.git(ctx, dir, "config", "--local", "--list"); err != nil {
		return "", fmt.Errorf("inspect repository config in %s: %w", dir, err)
	}
	origin, err := c.git(ctx, dir, "config", "--local", "--get", "remote.origin.url")
	if err != nil {
		var code interface{ ExitCode() int }
		if errors.As(err, &code) && code.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("inspect origin in %s: %w", dir, err)
	}
	return origin, nil
}

// capturedGitOutput bounds memory even when a child ignores a closed output pipe.
// Cancellation is shared by both streams and terminates the runner's process group.
type gitOutputBudget struct {
	mu          sync.Mutex
	limit, used int
	overflow    bool
	cancel      context.CancelFunc
}
type capturedGitOutput struct {
	data   bytes.Buffer
	budget *gitOutputBudget
}

func (b *capturedGitOutput) Write(p []byte) (int, error) {
	budget := b.budget
	budget.mu.Lock()
	defer budget.mu.Unlock()
	n := len(p)
	if budget.limit > 0 && len(p) > budget.limit-budget.used {
		p = p[:budget.limit-budget.used]
		budget.overflow = true
		budget.cancel()
	}
	budget.used += len(p)
	b.data.Write(p)
	return n, nil
}

// rawGitError retains process identity for errors.As without rendering argv or
// stderr, either of which can contain credentials supplied by Git or a remote.
type rawGitError struct {
	operation, dir, reason string
	cause                  error
}

func (e *rawGitError) Error() string {
	return fmt.Sprintf("git %s in %s: %s", e.operation, e.dir, e.reason)
}
func (e *rawGitError) Unwrap() error { return e.cause }

func (g ExecGit) runBounded(ctx context.Context, dir, stage string, args ...string) (string, error) {
	if g.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	budget := &gitOutputBudget{limit: g.MaxOutputBytes, cancel: cancel}
	stdout := &capturedGitOutput{budget: budget}
	stderr := &capturedGitOutput{budget: budget}
	if !g.Raw {
		stderr = stdout
	}
	runner := weavefs.ExecRunner{Context: runCtx, Stdout: stdout, Stderr: stderr, ExtraFiles: g.ExtraFiles}
	argv := append([]string{"git"}, args...)
	var err error
	if stage == "" {
		err = runner.Run(dir, argv)
	} else {
		err = runner.RunOwned(dir, argv, stage)
	}
	reason := "process failed"
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			reason = exit.Error()
		}
	}
	// Resource failures invalidate the process result. Do not retain ExitCode in
	// their error chain: predicate callers interpret exit 1 as a valid negative
	// answer, which must never hide truncated output or interrupted execution.
	if ctx.Err() != nil {
		err = ctx.Err()
		reason = err.Error()
	}
	if budget.overflow {
		limitErr := fmt.Errorf("output limit %d bytes exceeded", g.MaxOutputBytes)
		err = errors.Join(ctx.Err(), limitErr)
		reason = limitErr.Error()
	}
	if err != nil {
		if g.Raw {
			operation := "operation"
			// Only the command name is diagnostic; never print user-controlled arguments.
			if len(args) > 0 {
				switch args[0] {
				case "config", "rev-parse", "status", "diff", "fetch", "merge", "rebase", "merge-base", "show", "ls-tree", "symbolic-ref", "for-each-ref", "clone", "cat-file", "check-ref-format", "rev-list":
					operation = args[0]
				}
			}
			return "", &rawGitError{operation: operation, dir: dir, reason: reason, cause: err}
		}
		return "", fmt.Errorf("git %v in %s: %w: %s", args, dir, err, strings.TrimSpace(stdout.data.String()))
	}
	if g.Raw {
		return stdout.data.String(), nil
	}
	return strings.TrimSpace(stdout.data.String()), nil
}
