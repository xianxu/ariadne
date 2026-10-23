package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type benchCommandRouteKey struct{}

// The production execution path is unchanged when no replay context exists.
// Audit begins before Cobra argument validation and ends on every returned
// error, including refusals before a reviewer starts.
func executeWithBenchBoundary(ctx context.Context, root *cobra.Command, args []string) error {
	repo, err := gitx.RepoTopLevelContext(ctx)
	if err != nil {
		return root.ExecuteContext(ctx)
	}
	c, err := loadBenchContext(ctx, repo)
	if err != nil {
		return err
	}
	if command, _, findErr := root.Find(args); findErr == nil {
		path := strings.Fields(command.CommandPath())
		if len(path) > 0 {
			path = path[1:]
		}
		ctx = context.WithValue(ctx, benchCommandRouteKey{}, path)
	}
	return withBenchCommand(ctx, c, repo, args, func() error {
		// Find resolves aliases and flags before checking the actual command.
		if command, _, findErr := root.Find(args); findErr == nil {
			path := strings.Fields(command.CommandPath())
			if len(path) > 0 {
				path = path[1:]
			}
			if err := guardBenchCommand(c, path); err != nil {
				return err
			}
		}
		return root.ExecuteContext(ctx)
	})
}
