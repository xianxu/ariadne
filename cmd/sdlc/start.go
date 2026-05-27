// start.go — `sdlc start` migration stub.
//
// As of #39, `sdlc start` no longer exists as a working verb. The
// pre-#39 behavior (claim + worktree creation in one shot) was split
// into two separately-timed verbs because in practice the planning
// step takes hours-to-days between claim and the moment you actually
// want to start coding. Composing them was the wrong default.
//
// This stub stays so invocations of the old verb get a clear
// migration message instead of cobra's generic "unknown command".
// Remove after one cycle once external scripts / muscle memory have
// migrated.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewStartCmd returns a cobra command that fails fast with a
// migration message. Hidden from `--help` listings so docs don't
// suggest it.
func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "start",
		Short:  "REMOVED — use `sdlc claim` (early) + `sdlc change-code` (later); see #39",
		Hidden: true,
		Args:   cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(
				"`sdlc start` was removed (#39). Use:\n" +
					"  sdlc claim --issue N        # claim the issue (commit + push)\n" +
					"  sdlc change-code --issue N  # later, after the plan is written\n" +
					"The two transitions happen at different times; composing them\n" +
					"is the wrong default for solo + planning-heavy work.")
		},
	}
}
