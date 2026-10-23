package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

//go:embed bench.md
var benchHelp string

var benchRunID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var benchCandidate = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func newBenchCmd(d deps) *cobra.Command {
	command := &cobra.Command{Use: "bench", Short: "Record public bench TL decisions and finish proposals", Long: benchHelp, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	var decisionRun, kind, reason string
	decision := &cobra.Command{Use: "decision", Short: "Record one bench TL intervention", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if !benchRunID.MatchString(decisionRun) {
			return fmt.Errorf("--run-id must name the exact bench run")
		}
		if kind != "approve" && kind != "replan" && kind != "ruling" && kind != "stop" {
			return fmt.Errorf("--kind must be approve, replan, ruling or stop")
		}
		if strings.TrimSpace(reason) == "" || len(reason) > 4096 || !utf8.ValidString(reason) || strings.ContainsRune(reason, 0) {
			return fmt.Errorf("--reason must contain 1..4096 bytes of UTF8 explanation")
		}
		return forwardBenchCommand(cmd, d, []string{"bench", "decision", "--run-id", decisionRun, "--kind", kind, "--reason", reason})
	}}
	decision.Flags().StringVar(&decisionRun, "run-id", "", "exact run ID from the provisioned public context (required)")
	decision.Flags().StringVar(&kind, "kind", "", "approve, replan, ruling or stop (required)")
	decision.Flags().StringVar(&reason, "reason", "", "reason for this intervention (required)")
	var finishRun, candidate string
	finish := &cobra.Command{Use: "finish", Short: "Propose a frozen candidate for would-land verification", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if !benchRunID.MatchString(finishRun) {
			return fmt.Errorf("--run-id must name the exact bench run")
		}
		if !benchCandidate.MatchString(candidate) {
			return fmt.Errorf("--candidate must be a full lowercase Git object ID")
		}
		return forwardBenchCommand(cmd, d, []string{"bench", "finish", "--run-id", finishRun, "--candidate", candidate})
	}}
	finish.Flags().StringVar(&finishRun, "run-id", "", "exact run ID from the provisioned public context (required)")
	finish.Flags().StringVar(&candidate, "candidate", "", "full candidate commit object ID (required)")
	command.AddCommand(decision, finish)
	return command
}

func forwardBenchCommand(cmd *cobra.Command, d deps, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), sdlcCallTimeout)
	defer cancel()
	run := d.benchCommand
	if run == nil {
		run = func(ctx context.Context, args []string) ([]byte, error) { return runBenchCommand(ctx, "sdlc", args) }
	}
	out, err := run(ctx, args)
	if len(out) > docRepositoryOutputLimit {
		return errors.Join(err, fmt.Errorf("sdlc bench output exceeds limit"))
	}
	if len(out) > 0 {
		if _, writeErr := cmd.OutOrStdout().Write(out); writeErr != nil {
			return errors.Join(err, writeErr)
		}
	}
	return err
}

// runBenchCommand reuses the managed, context-bound SDLC command owner. Unlike
// execSDLC.run, it has bounded stdout and diagnostics and retains exit identity.
func runBenchCommand(ctx context.Context, bin string, args []string) ([]byte, error) {
	out, err := docRepositoryOutput(ctx, "", bin, args...)
	var child *exec.ExitError
	if errors.As(err, &child) {
		code := child.ExitCode()
		if code < 1 {
			code = 1
		}
		return out, exitError{code: code, err: err}
	}
	return out, err
}
