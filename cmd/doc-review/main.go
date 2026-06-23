// Command doc-review runs a fresh-context, read-only fact + reference review
// of a Markdown document using a SECOND agent from a DIFFERENT vendor (codex by
// default), then writes the reviewer's findings to a sidecar report.
//
// It is the fact-check path embedded in the `xx-fix` skill. Per AGENTS.md §3,
// the co-authoring agent carries confirmation bias, so a fresh-eyes review must
// be a *separate* agent with no conversation history. This binary dispatches one
// (read-only — it cannot edit the doc), captures its report, and hands the main
// agent a triage instruction. The main agent still owns the document; the report
// is advisory.
//
// Help is the operational contract referenced by xx-fix:
//
//	doc-review --help        what it does, usage, agents, output, the triage step
//	doc-review <agent> <file>   run the review (agent optional; defaults to codex)
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//go:embed help.md
var rootHelp string

func main() {
	if err := buildRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the single-command cobra tree. Extracted from main so the
// arg wiring is testable.
func buildRoot() *cobra.Command {
	f := reviewFlags{}

	root := &cobra.Command{
		Use:           "doc-review [agent] <file.md>",
		Short:         "Fresh-context, read-only fact + reference review of a Markdown doc by a second-vendor agent",
		Long:          rootHelp,
		Args:          cobra.RangeArgs(1, 2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, file, err := parseArgs(args)
			if err != nil {
				return err
			}
			f.Agent = agent
			f.File = file
			return runReview(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.Flags().BoolVar(&f.DryRun, "dry-run", false, "print the would-be agent command line + output path; do not run the reviewer")
	root.Flags().StringVar(&f.Out, "out", "", "override the report path (default: <file>-<agent>-check.md)")

	return root
}

// parseArgs maps positionals to (agent, file):
//
//	1 arg  → file; agent defaults to codex
//	2 args → agent, file (agent validated against the known set)
//
// A document literally named like an agent ("codex.md") is unaffected: with one
// positional we always treat it as the file.
func parseArgs(args []string) (agent AgentCLI, file string, err error) {
	switch len(args) {
	case 1:
		return DefaultAgent, args[0], nil
	case 2:
		a := AgentCLI(args[0])
		if !a.known() {
			return "", "", fmt.Errorf("unknown agent %q (supported: codex, gemini, claude)", args[0])
		}
		return a, args[1], nil
	default:
		return "", "", fmt.Errorf("expected <file> or <agent> <file>, got %d args", len(args))
	}
}
