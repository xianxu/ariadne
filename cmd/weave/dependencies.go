package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/weave/internal/startup"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func buildDependencies() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use: "dependencies", Short: "Restore base repositories and prepare their Brewfiles",
		Long: `Restore transitive sources and prepare layer Brewfiles, foundation-first.
Numbered slots, Couch contexts and SDLC calls only check packages, never install.
Missing packages are listed with one install command for the operator's shell.
Hand runs outside these contexts retain bundle install --no-upgrade behavior.
Every brew call disables metadata auto-update with HOMEBREW_NO_AUTO_UPDATE=1.
Compile and refresh use this same policy. --dry-run reports without running brew.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) (retErr error) {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			client, runner, closeSetup, err := prepareSetup(cmd.Context(), root, dryRun, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer func() { retErr = errors.Join(retErr, closeSetup()) }()
			restored, err := restoreSetup(cmd.Context(), weavefs.OSFS{}, client, root, dryRun, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return startup.Dependencies(weavefs.OSFS{}, restored.Layers, runner, dependencyChecksOnly(client), dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report known dependency operations without cloning or installing")
	return cmd
}
