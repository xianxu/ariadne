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
		Use: "dependencies", Short: "Restore base repositories and install their Brewfiles",
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
			return startup.Dependencies(weavefs.OSFS{}, restored.Layers, runner, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report known dependency operations without cloning or installing")
	return cmd
}
