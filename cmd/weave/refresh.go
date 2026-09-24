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
	var rebase bool
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

Run on a workspace's resting branch when ready to adopt newer sources. Numbered
slots refresh their private substrates; a primary refresh includes its declared
shared peers. No other slot is refreshed. Data mount repositories are not Git
refresh subjects. Changed dependency declarations refuse before branch updates;
reconcile those declarations and prepare any missing checkouts separately.

Refresh does not reserve an issue or change machine-wide tool installation policy.
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
			return refresh.Run(cmd.Context(), root, client, rebase, cmd.OutOrStdout(), func() error {
				compileClient := client
				compileClient.Git = acquire.ExecGit{ExtraFiles: runner.ExtraFiles}
				return compilePrepared(cmd.Context(), weavefs.OSFS{}, root, plan.TargetAll, false, cmd.OutOrStdout(), compileClient, runner)
			})
		},
	}
	cmd.Flags().BoolVar(&rebase, "rebase", false, "allow rebasing local commits onto captured origin/main instead of fast-forward only")
	return cmd
}
