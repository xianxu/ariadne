package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/startup"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func buildDependencies() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use: "dependencies", Short: "Restore base repositories and install their Brewfiles",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			restored, err := acquire.Restore(cmd.Context(), root, dryRun)
			for _, missing := range restored.Missing {
				fmt.Fprintf(cmd.OutOrStdout(), "weave: missing source %s (preview incomplete)\n", missing)
			}
			if err != nil {
				return err
			}
			runner := weavefs.ExecRunner{Context: cmd.Context(), Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr()}
			return startup.Dependencies(weavefs.OSFS{}, restored.Layers, runner, runtime.GOOS, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report known dependency operations without cloning or installing")
	return cmd
}
