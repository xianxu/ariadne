package acquire

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GitRunner is the acquisition process boundary. Its errors retain an ExitCode
// method when Git ran, so a missing value can be distinguished from failed IO.
type GitRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

// Client holds the Git boundary for one acquisition operation. Its zero value
// uses real Git; injected clients share the same restore/probe/publication code.
type Client struct{ Git GitRunner }
type ExecGit struct{}

func (ExecGit) Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v in %s: %w: %s", args, dir, err, strings.TrimSpace(string(b)))
	}
	return strings.TrimSpace(string(b)), nil
}
func (c Client) git(ctx context.Context, dir string, args ...string) (string, error) {
	runner := c.Git
	if runner == nil {
		runner = ExecGit{}
	}
	return runner.Run(ctx, dir, args...)
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
