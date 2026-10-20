package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// statusArgs preserves the explicit FILE form and prepends the supplied
// default only when the command's other required arguments are present.
func statusArgs(args []string, explicitArity int, defaultFile string) ([]string, error) {
	switch len(args) {
	case explicitArity:
		return args, nil
	case explicitArity - 1:
		return append([]string{defaultFile}, args...), nil
	default:
		return nil, fmt.Errorf("accepts %d or %d arg(s), received %d", explicitArity-1, explicitArity, len(args))
	}
}

// defaultStatusPath uses the current checkout (including a worktree), never
// the enclosing workspace or the current subdirectory. Resolution reads no
// session content and never creates a session file.
func defaultStatusPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := git(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("refused: omitted FILE requires a git checkout with .hub at its root; pass FILE explicitly: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".hub")); err != nil {
		return "", fmt.Errorf("refused: omitted FILE requires .hub at checkout root %s; pass FILE explicitly: %w", root, err)
	}
	return filepath.Join(root, "session.md"), nil
}

// withDefaultStatusFile normalizes before the existing argument validator,
// and passes those same arguments to the existing behavior. Explicit FILE
// calls never consult git or .hub. Cobra validates Args before calling RunE.
func withDefaultStatusFile(c *cobra.Command, explicitArity int) *cobra.Command {
	validate, run := c.Args, c.RunE
	var normalized []string
	c.Args = func(cmd *cobra.Command, args []string) error {
		var err error
		normalized, err = statusArgs(args, explicitArity, "")
		if err != nil {
			return err
		}
		if len(args) == explicitArity-1 {
			file, err := defaultStatusPath()
			if err != nil {
				return err
			}
			normalized, _ = statusArgs(args, explicitArity, file)
		}
		if validate != nil {
			return validate(cmd, normalized)
		}
		return nil
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return run(cmd, normalized)
	}
	return c
}
