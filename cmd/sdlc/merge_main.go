package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Legacy worktrees retain their manual integration workflow. Always fetch:
// before checks, and after their potentially long-running suite has finished.
func requireLegacyMainIntegrated(git func(...string) (string, error), head string) (string, error) {
	if out, err := git("fetch", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return "", fmt.Errorf("fetch main before landing: %w\n%s", err, out)
	}
	main, err := git("rev-parse", "--verify", "origin/main^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve fetched main: %w", err)
	}
	main = strings.TrimSpace(main)
	if !landingOIDValid(main) {
		return "", fmt.Errorf("resolve fetched main: invalid commit %q", main)
	}
	if out, err := git("merge-base", "--is-ancestor", main, head); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return "", fmt.Errorf("main moved: this legacy worktree requires manual integration; run `git merge origin/main`, resolve any conflicts, run `sdlc local-check run --name suite` after pushing with `sdlc pr`, then retry `sdlc merge`")
		}
		return "", fmt.Errorf("check main ancestry before landing: %w\n%s", err, out)
	}
	return main, nil
}
