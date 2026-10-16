// Command tl-worker prints the rules for a worker slot in a TL-run Couch fleet.
//
// It is a skill binary in the shape of `sdlc --help` and `couch --skill`: the
// prose ships inside the binary (help.md, embedded), and the xx-tl-worker skill
// is a pointer stub that tells the agent to run `tl-worker --help`. The prose
// cannot drift from the tooling it describes because both release together.
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/tlrules"
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
// rules, the same as --help.
func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "tl-worker",
		Short:         "The rules for a worker slot in a TL-run Couch fleet",
		Long:          tlrules.ComposeHelp(rootHelp),
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
