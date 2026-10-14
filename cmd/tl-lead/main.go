// Command tl-lead prints the procedure for the TL running a project from a Couch hub.
//
// It is a skill binary in the shape of `sdlc --help` and `couch --skill`: the
// prose ships inside the binary (help.md, embedded), and the xx-tl-lead skill
// is a pointer stub that tells the agent to run `tl-lead --help`. The prose
// cannot drift from the tooling it describes because both release together.
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

// buildRoot assembles the single-command tree; with no arguments it prints the
// procedure, the same as --help.
func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "tl-lead",
		Short:         "The procedure for the TL running a project from a Couch hub",
		Long:          rootHelp,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	return root
}
