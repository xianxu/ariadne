package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

//go:embed helptext/bench.md
var benchCLIHelp string

func NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "bench", Short: "Metis step backends and bound TL evidence", Long: benchCLIHelp, Args: cobra.NoArgs}
	step := &cobra.Command{Use: "step", Short: "Execute one Metis-owned step", Args: cobra.NoArgs}
	for _, name := range []string{"prepare", "session", "score"} {
		name := name
		step.AddCommand(&cobra.Command{Use: name, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			return runBenchStep(cmd.Context(), name, cmd.OutOrStdout())
		}})
	}
	cmd.AddCommand(step)
	addBenchPublicCommands(cmd)
	return cmd
}

func addBenchPublicCommands(parent *cobra.Command) {
	var runID, kind, reason, candidate string
	decision := &cobra.Command{Use: "decision", Short: "Record a decision from the exact bound TL conversation", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		id, err := recordBenchProposal(cmd.Context(), "decision", runID, kind, reason, "", benchCouchDeps{})
		if err == nil {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "bench decision retained:", id)
		}
		return err
	}}
	decision.Flags().StringVar(&runID, "run-id", "", "exact public run identity")
	decision.Flags().StringVar(&kind, "kind", "", "approve, replan, ruling or stop")
	decision.Flags().StringVar(&reason, "reason", "", "decision rationale")
	var finishID string
	finish := &cobra.Command{Use: "finish", Short: "Propose reviewed code for independent would-land verification", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		id, err := recordBenchProposal(cmd.Context(), "finish", finishID, "", "", candidate, benchCouchDeps{})
		if err == nil {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "bench candidate proposed:", id)
		}
		return err
	}}
	finish.Flags().StringVar(&finishID, "run-id", "", "exact public run identity")
	finish.Flags().StringVar(&candidate, "candidate", "", "full reviewed code commit; metadata head is not the candidate")
	parent.AddCommand(decision, finish)
}

func requireBenchObserver(ctx context.Context) error {
	root, err := gitx.RepoTopLevelContext(ctx)
	if err != nil {
		return nil
	}
	c, err := loadBenchContext(ctx, root)
	if err != nil {
		return err
	}
	if c != nil {
		return errors.New("bench observer effects must run outside the prepared agent checkouts")
	}
	return nil
}

func readBenchInput(path string, limit int) ([]byte, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	physical, err := filepath.EvalSymlinks(path)
	if err != nil || physical != path {
		return nil, errors.New("bench input must be an existing physical path without symlinks")
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("bench input must be a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > limit {
		return nil, errors.New("bench input exceeds its bound")
	}
	return raw, nil
}
