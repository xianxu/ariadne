package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

const benchHelp = `Bench stewardship adds these overrides to the normal TL rules.
Read tl-lead --help for the complete procedure; its design, scope and reporting
rules remain in force.

Use the fixed TL, worker and reviewer models provisioned for this run. No model
fallback or agent replacement: ask the observer to stop if the binding changes.
The bench TL never codes and never lands. It approves plans, re-plans, rules or
stops under the normal stewardship rules. Record those decisions with the exact
categories approve, replan, ruling and stop.

Run these commands from the provisioned bench hub root with its exact --run-id.
decision records an intervention; finish submits a candidate proposal. SDLC owns
the public context, actor/generation checks and real suite/close-review evidence.
A proposal does not grant a pass or authorize a merge. The observer decides when
the frozen candidate would land and owns scoring, measurement and the ledger.

M1 uses trusted agents with prepared-repository isolation. It does not prevent
same-user host reads or GitHub access. Keep hidden acceptance and observer records
outside the TL conversation. Unexpected operator input stops this autonomous run.
`

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
