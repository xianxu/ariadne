// Command weave compiles a repo's agentic context from its layer DAG: it walks
// the layers (construct/deps), composes each layer's intents into an ordered
// []Action (the pure planner), and applies them to the filesystem.
//
//	weave              compile the current working directory's repo
//	weave --dry-run    print the planned []Action; mutate nothing
//
// The pure core (intent/, layer/, plan/) never touches disk; weave's only IO is
// the walk (reading manifests/deps/prose) and plan.Apply (the mutations),
// behind weavefs.FS (ARCH-PURE). M3 adds the `skills`/`skill` subcommands and
// `depend-on`; this M2 binary is the compile path only.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func main() {
	if err := buildRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the cobra command. Extracted from main so the wiring is
// testable. The root command IS the compile action (no subcommand needed);
// --dry-run flips it to print-only.
func buildRoot() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:           "weave",
		Short:         "Compile a repo's agentic context from its layer DAG",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return run(weavefs.OSFS{}, root, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the planned actions; mutate nothing")
	return cmd
}

// run is the compile pipeline: walk → Plan → (Apply | print). Injecting fs +
// out keeps it testable against a t.TempDir-rooted OSFS and a buffer.
//
// root is canonicalized to its physical form up front (filepath.EvalSymlinks ≈
// pwd -P) so it lives in the SAME namespace as the layer Paths the walk
// canonicalizes — without this, on macOS (/tmp → /private/tmp) Apply would
// compute a relative symlink target between a logical dst-dir and a physical
// upstream src that resolves wrong when the OS follows the link (the exact bug
// setup.sh's pwd -P guards against, lines 39-45).
func run(fs weavefs.FS, root string, dryRun bool, out io.Writer) error {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	layers, err := walk.Walk(fs, root)
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	actions, err := plan.Plan(layers)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if dryRun {
		fmt.Fprint(out, formatActions(actions))
		return nil
	}
	if err := plan.Apply(fs, root, actions); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	fmt.Fprintf(out, "weave: applied %d action(s) to %s\n", len(actions), root)
	return nil
}

// formatActions renders a []Action as one line per action for --dry-run. Pure
// (string in/out), so it's unit-tested directly.
func formatActions(actions []plan.Action) string {
	if len(actions) == 0 {
		return "weave: no actions\n"
	}
	var b []byte
	for _, a := range actions {
		switch act := a.(type) {
		case plan.Symlink:
			b = append(b, fmt.Sprintf("symlink   %s -> %s\n", act.Dst, act.Src)...)
		case plan.WriteFile:
			b = append(b, fmt.Sprintf("writefile %s (%d bytes)\n", act.Path, len(act.Content))...)
		case plan.Mkdir:
			b = append(b, fmt.Sprintf("mkdir     %s\n", act.Path)...)
		case plan.MergeSettings:
			b = append(b, fmt.Sprintf("merge     %s -> %s\n", act.Source, act.Target)...)
		case plan.ToolDep:
			b = append(b, fmt.Sprintf("tool      %s (%s)\n", act.Path, act.Owner)...)
		default:
			b = append(b, fmt.Sprintf("unknown   %T\n", a)...)
		}
	}
	return string(b)
}
