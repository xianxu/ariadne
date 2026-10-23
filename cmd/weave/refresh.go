package main

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/refresh"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func buildRefresh() *cobra.Command {
	var rebase, substrate, deferSettings, planning bool
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Explicitly refresh this repository and its substrate dependencies",
		Long: `Fetch origin/main for this repository and its declared substrate dependencies.
All checkouts must be clean, with no active Git operation. Before any branch
advances, verify that all current branches can fast-forward to captured targets.
Then update those same branches and run compile, even if already current.

--rebase permits replaying local commits onto the captured targets. Conflicts
stop for explicit resolution. Updates already completed remain after any error;
resolve the reported problem and run refresh again. No automatic stash or reset.

--substrate refreshes only the declared substrates and recompiles. It works on
a clean issue branch, preserving this repository's branch, HEAD, index and
origin/main ref. It cannot be combined with --rebase. Generated composition may
change during compile. With no substrates it only recompiles the current source;
merge current main separately to update helpers owned by that source branch.

Run the default mode on a workspace's resting branch to adopt newer sources. Numbered
slots refresh their private substrates; a primary refresh includes its declared
shared peers. No other slot is refreshed. Data mount repositories are not Git
refresh subjects. Changed dependency declarations refuse before branch updates;
reconcile those declarations and prepare any missing checkouts separately.

--planning is the SDLC planning-gate mode: defer every Seed and SeedOnce output
and protected Claude settings before publication. Stale seeded paths are named;
run weave compile on a dedicated consumer issue branch to regenerate them and
publish the diff as a separate change. Local-gates files require the operator
to apply local-gates-verified; agents never apply that label. Other generated
composition still runs; this is not a blanket tracked-file deferral.

--defer-settings retains the settings-only mode. Run weave refresh --substrate
from your shell to publish protected settings while preserving an issue branch.

Refresh does not reserve an issue. Package policy follows weave dependencies:
numbered slots, Couch contexts and SDLC calls only check packages and print the
operator's install command when missing. No brew call auto-updates metadata.
Ordinary weave compile preserves existing Git revisions; refresh is explicit.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) (retErr error) {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			client, runner, closeSetup, err := prepareSetup(cmd.Context(), root, false, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer func() { retErr = errors.Join(retErr, closeSetup()) }()
			// Refresh uses raw framed Git observations; compilation restores the
			// ordinary acquisition adapter while retaining the same setup lease.
			client.Git = acquire.ExecGit{ExtraFiles: runner.ExtraFiles, Raw: true, Timeout: 2 * time.Minute, MaxOutputBytes: 4 << 20}
			client.MaxLayers = 128
			client.MaxDeclarationBytes = 1 << 20
			return refresh.RunWithOptions(cmd.Context(), root, client, refresh.Options{Rebase: rebase, SubstrateOnly: substrate}, cmd.OutOrStdout(), func() error {
				compileClient := client
				compileClient.Git = acquire.ExecGit{ExtraFiles: runner.ExtraFiles}
				policy := publishAllOutputs
				if planning {
					policy = deferPlanningOutputs
				} else if deferSettings {
					policy = deferProtectedSettings
				}
				return compilePreparedWithPolicy(cmd.Context(), weavefs.OSFS{}, root, plan.TargetAll, false, cmd.OutOrStdout(), compileClient, runner, policy)
			})
		},
	}
	cmd.Flags().BoolVar(&rebase, "rebase", false, "allow rebasing local commits onto captured origin/main instead of fast-forward only")
	cmd.Flags().BoolVar(&substrate, "substrate", false, "refresh only declared substrates and compile, preserving this repository's branch and HEAD")
	cmd.Flags().BoolVar(&deferSettings, "defer-settings", false, "check Claude settings without writing them; operator refresh publishes them")
	cmd.Flags().BoolVar(&planning, "planning", false, "planning gate: defer seeded entrypoints and protected settings; regenerate explicitly on a dedicated consumer issue branch")
	cmd.MarkFlagsMutuallyExclusive("planning", "defer-settings")
	cmd.MarkFlagsMutuallyExclusive("substrate", "rebase")
	return cmd
}
