// Command sdlc is the unified SDLC checkpoint binary for ariadne.
//
// Subcommands are checkpoint guards — they defend known commit moments
// (close, merge, push, milestone-close) against drift. Subcommands are
// added incrementally when the same drift recurs at a stage; the binary
// does not model the SDLC as a state machine.
//
// Help disclosure is progressive:
//
//	sdlc --help              top-level skill narrative + verb list
//	sdlc <verb> --help       per-checkpoint contract + flags + examples
//	sdlc --index             emits SKILL.md content (regenerator)
//
// Design rationale: workshop/issues/000031-sdlc-checkpoint-binary.md +
// docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
)

func main() {
	var indexFlag bool

	root := &cobra.Command{
		Use:           "sdlc",
		Short:         "SDLC checkpoint binary — guards known commit moments against drift",
		Long:          helptext.MustGet("root"),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Flags().BoolVar(&indexFlag, "index", false,
		"emit SKILL.md content to stdout (regenerates construct/local/sdlc/SKILL.md)")

	root.RunE = func(cmd *cobra.Command, args []string) error {
		if indexFlag {
			_, err := fmt.Fprint(cmd.OutOrStdout(), helptext.MustGet("index"))
			return err
		}
		return cmd.Help()
	}

	closeCmd := NewCloseCmd()
	closeCmd.Long = helptext.MustGet("close")
	root.AddCommand(closeCmd)

	stateCmd := NewStateCmd()
	stateCmd.Long = helptext.MustGet("state")
	root.AddCommand(stateCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
