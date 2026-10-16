// Command tl-lead prints the procedure for the TL running a project from a Couch hub.
//
// It is a skill binary in the shape of `sdlc --help` and `couch --skill`: the
// prose ships inside the binary (help.md, embedded), and the xx-tl-lead skill
// is a pointer stub that tells the agent to run `tl-lead --help`. The prose
// cannot drift from the tooling it describes because both release together.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/tlrules"
)

//go:embed help.md
var rootHelp string

// deps are the seams to the world: couch's session-sync verbs and sdlc's JSON
// contracts. Tests pass stateful fakes.
type deps struct {
	ss   SessionSync
	sdlc SDLC
}

func main() {
	d := deps{ss: couchSessionSync{bin: "couch"}, sdlc: execSDLC{bin: "sdlc"}}
	if err := buildRoot(d).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		var ee exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		os.Exit(1)
	}
}

// buildRoot assembles the command tree; with no arguments it prints the
// procedure, the same as --help.
func buildRoot(d deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "tl-lead",
		Short:         "The procedure for the TL running a project from a Couch hub",
		Long:          tlrules.ComposeHelp(rootHelp),
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.AddCommand(newStatusCmd(d), newTrainsCmd(d))
	return root
}
