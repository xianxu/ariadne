// Command sdlc is the unified SDLC checkpoint binary for ariadne.
//
// Subcommands are checkpoint guards — they defend known commit moments
// (close, merge, push, milestone-close) against drift. Subcommands are
// added incrementally when the same drift recurs at a stage; the binary
// does not model the SDLC as a state machine.
//
// Help disclosure is progressive — `sdlc --help` is the single workflow
// contract (the xx-sdlc skill is a static pointer to it):
//
//	sdlc --help              top-level workflow contract + verb list
//	sdlc <verb> --help       per-checkpoint contract + flags + examples
//
// Design rationale: workshop/issues/000031-sdlc-checkpoint-binary.md +
// docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md.
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// renderLong loads a command's embedded help text and substitutes the model-derived
// placeholders from the issue vocabulary at runtime (#125): the lifecycle FACTS
// derive live from pkg/vocab (← construct/vocabulary/issue.cue) so the help prose
// can't drift. The {{ARCH_STAR}} idiom (cmd/sdlc/internal/judge/review.go). No-op
// when the text carries no placeholder — so EVERY command-Long load can route through
// it uniformly (none can forget the substitution). The editorial framing around each
// placeholder stays hand-written in the helptext `.md`.
func renderLong(name string) string {
	m := vocab.Issue()
	return strings.NewReplacer(
		"{{LIFECYCLE}}", m.RenderLifecycleHelp(),
		"{{STATUS_NAMES}}", m.StatusNames(" | "),
		"{{STATUS_GLOSS}}", m.StatusGloss(),
	).Replace(helptext.MustGet(name))
}

func main() {
	if err := buildRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the full cobra command tree. Extracted from main so
// the tree itself is testable (e.g. the flat→group alias wiring for
// set-status / fetch — #56 M2).
//
// The verb list in `sdlc --help` is cobra's auto-generated "Available
// Commands" — NOT a hand-maintained block in root.md, which drifts (#56).
// Sorting is disabled so that list renders in workflow order (the order the
// AddCommand calls below run); hidden commands are auto-omitted, so the
// deprecated aliases never appear. The single source is the registry here.
func buildRoot() *cobra.Command {
	disableCobraCommandSorting()

	root := &cobra.Command{
		Use:           "sdlc",
		Short:         "SDLC checkpoint binary — guards known commit moments against drift",
		Long:          renderLong("root"),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.HiddenDefaultCmd = true // keep `completion` out of the verb list

	root.RunE = func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	}

	// add registers a command, wiring its Long help from the embedded
	// helptext and overriding Short with the crisp workflow-facing
	// one-liner shown in `sdlc --help`.
	add := func(c *cobra.Command, longKey, short string) {
		if longKey != "" {
			c.Long = renderLong(longKey)
		}
		c.Short = short
		root.AddCommand(c)
	}

	// Workflow order (claim → ship), which is the order the verb list renders.
	add(NewClaimCmd(), "claim", "Start work: flip an open issue to working + broadcast the claim")
	add(NewStartPlanCmd(), "start-plan", "Enter planning: deliver the architecture principles to design against (#75)")
	add(NewChangeCodeCmd(), "change-code", "Enter implementation after the structural + plan-quality gates")
	add(NewIssueCmd(), "issue", "Create + manage issues (new / set-status / list / show)")
	add(NewActualCmd(), "actual", "Compute an issue's focused dev-hours via active-time-v3 (#68)")
	add(NewActiveTimeCmd(), "active-time", "Per-issue active-time attribution table (the v3 engine, standalone)")
	add(NewCloseCmd(), "close", "Close an issue or milestone (ACTUAL + VERIFIED + atlas/project sweep)")
	add(NewMilestoneCloseCmd(), "milestone-close", "Close one milestone + auto-dispatch its review")
	add(NewPRCmd(), "pr", "Open a pull request from a feature branch")
	add(NewMergeCmd(), "merge", "Merge the PR, archive done issues, clean up")
	add(NewPushCmd(), "push", "Ship from main (clean tree + pre-merge judges + archive)")
	add(NewStateCmd(), "state", "Inspect workflow state (branch, working issues, drift)")
	add(NewResolveCmd(), "resolve", "Resolve a symbolic artifact ref (ariadne#11, #15 M4) to its current path(s) — read-only")
	add(NewOpenCmd(), "open", "Resolve a ref and open the primary artifact in $EDITOR")
	add(NewJudgeCmd(), "judge", "Run an LLM-judge check against the diff (fresh-context)")
	add(NewArchPrinciplesCmd(), "arch-principles", "Print the ARCH-* architecture principles (single source; pull for non-gate work)")
	add(NewEstimateSourceCmd(), "estimate-source", "Name the shared estimate method + the repo-local calibration source (pull)")
	add(NewProcessManualCmd(), "process-manual", "Unroll every injection source into a linked process manual (#153)")

	// Hidden: deprecated aliases + the start stub. Order is irrelevant —
	// they're omitted from the verb list.
	fetchCmd := NewFetchCmd()
	fetchCmd.Long = renderLong("fetch")
	fetchCmd.Hidden = true
	fetchCmd.Deprecated = "use `sdlc issue new --from-github N`" // #56 M2
	root.AddCommand(fetchCmd)

	flatSetStatus := NewSetStatusCmd()
	flatSetStatus.Long = renderLong("set-status")
	flatSetStatus.Hidden = true
	flatSetStatus.Deprecated = "use `sdlc issue set-status`" // #56 M2
	root.AddCommand(flatSetStatus)

	root.AddCommand(NewStartCmd()) // hidden migration stub (#39)
	root.AddCommand(newPropagateBaseCmd())

	wrapRepoLockCommands(root)

	return root
}

var cobraSortingOnce sync.Once

func disableCobraCommandSorting() {
	cobraSortingOnce.Do(func() {
		cobra.EnableCommandSorting = false
	})
}
