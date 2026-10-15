package main

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// Only Git's completed negative query means absent; failed probes preserve metadata.
func optionalBranchConfig(git gitFn, args ...string) (string, error) {
	out, err := git(args...)
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return "", nil
	}
	return out, err
}

// cleanupBranchConfig is ancillary to retirement of a landed branch. Existing
// config is its retry marker: the done-card merge/reconcile sweep finds it again.
// No retry receipt or sandbox permission is created. Live refs/checkouts win.
func cleanupBranchConfig(git gitFn, branch string, stderr io.Writer) {
	if err := removeObsoleteBranchConfig(git, branch); err != nil {
		location, locErr := git("rev-parse", "--git-path", "config")
		if locErr != nil {
			location = "shared Git config (path probe failed: " + locErr.Error() + ")"
		}
		cwarn(stderr, fmt.Sprintf("branch %s metadata cleanup deferred in %s: %v; the next writable sdlc merge or sdlc issue recovery reconcile retries automatically", branch, location, err))
	}
}

func removeObsoleteBranchConfig(git gitFn, branch string) error {
	config, err := optionalBranchConfig(git, "config", "--local", "--get-regexp", `^branch\.`+regexp.QuoteMeta(branch)+`\.`)
	if err != nil || config == "" {
		return err
	}
	tip, err := optionalBranchConfig(git, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil || tip != "" {
		return err
	}
	out, err := git("worktree", "list", "--porcelain", "-z")
	if err != nil {
		return err
	}
	trees, err := workspace.ParseWorktrees([]byte(out))
	if err != nil {
		return err
	}
	for _, tree := range trees {
		if tree.Branch == branch {
			return nil
		}
	}
	_, err = git("config", "--local", "--remove-section", "branch."+branch)
	return err
}
