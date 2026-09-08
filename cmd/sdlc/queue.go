// queue.go — `sdlc queue`: the advisory work queue, stored on the trunk
// (ariadne#209).
//
// This file is the IO shell and nothing else. The line format, the document
// round-trip, and the merge semantics are pure and live in internal/queue; the
// compare-and-swap retry lives in gitx.TrunkFile. What is left here is argument
// parsing, wiring one to the other, and rendering — including the refusal, which
// is a handoff rather than a dead end (ARCH-PURE).
package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queue"
)

// SPINE GUARD (#176), decided rather than defaulted.
//
// Every lifecycle verb calls guardSpineRepo, which refuses in a brain repo and in
// a repo with no workshop/issues/. The write subcommands here ARE that class: they
// commit and push to the trunk, and a queue of issue refs is meaningless where
// there are no issues — while the brain charter excludes SDLC process artifacts
// outright.
//
// The bare LIST is deliberately NOT guarded, matching the charter's own carve-out
// that reads are unaffected. It is the same read/write asymmetry this feature
// already applies offline: reading a queue that happens to exist costs nothing and
// refusing it would only obstruct.
//
// queuePath is the one file per repo. Not configurable: a queue whose location
// varies is a queue two checkouts can disagree about, which is the whole thing
// this verb exists to prevent.
const queuePath = "workshop/queue.md"

// trunkStore is the seam the verb consumes. Declared HERE, in the consumer,
// rather than exported from gitx — Go's idiom, and it keeps the fake local to
// the tests that need it.
type trunkStore interface {
	ReadDegraded(path string) ([]byte, string, error)
	Update(path, msg string, transform func([]byte) ([]byte, error)) error
}

// NewQueueCmd returns the cobra command for `sdlc queue`.
func NewQueueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "queue",
		Short:         "Advisory work queue on the trunk — what's next, and why",
		Long:          "Placeholder — replaced by helptext.MustGet(\"queue\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error {
			s, err := openTrunkStore()
			if err != nil {
				return err
			}
			return runQueueList(c.OutOrStdout(), c.ErrOrStderr(), s)
		},
	}
	cmd.AddCommand(newQueueAddCmd(), newQueueRemoveCmd(), newQueueMoveCmd())
	return cmd
}

func openTrunkStore() (trunkStore, error) {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return nil, err
	}
	return gitx.NewTrunkFile(root, "origin", "main")
}

func newQueueAddCmd() *cobra.Command {
	var tag string
	var project bool
	c := &cobra.Command{
		Use:           "add <ref> <why-now>",
		Short:         "Append an entry to the queue",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			guardSpineRepo(c.ErrOrStderr()) // #176 — writes to the trunk
			s, err := openTrunkStore()
			if err != nil {
				return err
			}
			kind := queue.KindIssue
			if project {
				kind = queue.KindProject
			}
			return runQueueEdit(c.OutOrStdout(), c.ErrOrStderr(), s, queue.Intent{
				Op: queue.OpAdd, Ref: args[0], WhyNow: args[1], Tag: tag, Kind: kind,
			})
		},
	}
	c.Flags().StringVar(&tag, "tag", "", "project tag for grouping")
	c.Flags().BoolVar(&project, "project", false,
		"this is a project line (an area to work in), not a next action")
	return c
}

func newQueueRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "remove <ref>",
		Short:         "Drop an entry from the queue",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			guardSpineRepo(c.ErrOrStderr()) // #176 — writes to the trunk
			s, err := openTrunkStore()
			if err != nil {
				return err
			}
			return runQueueEdit(c.OutOrStdout(), c.ErrOrStderr(), s,
				queue.Intent{Op: queue.OpRemove, Ref: args[0]})
		},
	}
}

func newQueueMoveCmd() *cobra.Command {
	var before, after string
	c := &cobra.Command{
		Use:           "move <ref> --before <ref> | --after <ref>",
		Short:         "Reorder an entry",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			if (before == "") == (after == "") {
				return errors.New("move needs exactly one of --before or --after")
			}
			guardSpineRepo(c.ErrOrStderr()) // #176 — writes to the trunk
			s, err := openTrunkStore()
			if err != nil {
				return err
			}
			in := queue.Intent{Op: queue.OpMove, Ref: args[0], Anchor: before}
			if after != "" {
				in.Anchor, in.After = after, true
			}
			return runQueueEdit(c.OutOrStdout(), c.ErrOrStderr(), s, in)
		},
	}
	c.Flags().StringVar(&before, "before", "", "place it before this ref")
	c.Flags().StringVar(&after, "after", "", "place it after this ref")
	return c
}

// runQueueList renders the trunk's queue. Warnings go to stderr so the listing
// itself stays pipeable.
func runQueueList(stdout, stderr io.Writer, s trunkStore) error {
	b, warn, err := s.ReadDegraded(queuePath)
	if err != nil {
		return err
	}
	if warn != "" {
		cwarn(stderr, warn)
	}
	doc := queue.Parse(b)
	entries := doc.Entries()
	if len(entries) == 0 {
		cok(stderr, "The queue is empty.")
	}
	for _, l := range entries {
		fmt.Fprintln(stdout, l.String())
	}
	// A hand-edit that near-misses the format (a hyphen where the separator is an
	// em-dash, say) parses as prose and is preserved in the file — but it would
	// then be absent from this listing with no signal, so the operator sees a
	// queue that silently omits the line they just wrote. Preserving it is right;
	// hiding it is not.
	if n := doc.UnrecognizedItems(); n > 0 {
		cwarn(stderr, fmt.Sprintf(
			"%d line(s) look like entries but do not parse, so they are NOT listed above "+
				"(they are preserved in the file). Format: `- <ref> — <why-now> [tag]`, "+
				"with an em-dash separator.", n))
	}
	return nil
}

// runQueueEdit applies one intent to the trunk.
//
// Intent.Apply IS the transform handed to Update, which is what makes a
// concurrent edit survive: on a rejected push Update re-reads the trunk and
// re-runs this closure against the base the peer just created.
func runQueueEdit(stdout, stderr io.Writer, s trunkStore, in queue.Intent) error {
	// BEFORE any git call. TrunkFile.Update fetches and reads the trunk before it
	// invokes the transform, so validating only inside Intent.Apply meant a
	// malformed ref cost a network round trip — and offline, the fetch failure
	// masked the real cause with "origin unreachable". The claim is only true if
	// something outside the transform makes it true.
	if err := in.Validate(); err != nil {
		return queueRefusal(stderr, s, in, err)
	}

	var applied queue.Applied
	err := s.Update(queuePath, queueCommitMessage(in), func(old []byte) ([]byte, error) {
		next, a, err := in.Apply(queue.Parse(old))
		if err != nil {
			return nil, err
		}
		applied = a
		return next.Render(), nil
	})
	if err != nil {
		return queueRefusal(stderr, s, in, err)
	}
	if applied.Note != "" {
		cwarn(stderr, applied.Note)
	}
	cok(stderr, queueCommitMessage(in))
	return nil
}

// queueCommitMessage is both the commit subject and the success line, so what
// the operator reads is what landed on the trunk.
func queueCommitMessage(in queue.Intent) string {
	switch in.Op {
	case queue.OpAdd:
		return "queue: add " + in.Ref
	case queue.OpRemove:
		return "queue: remove " + in.Ref
	case queue.OpMove:
		rel := "before"
		if in.After {
			rel = "after"
		}
		return fmt.Sprintf("queue: move %s %s %s", in.Ref, rel, in.Anchor)
	default:
		return "queue: edit"
	}
}

// queueRefusal renders a failed edit as a HANDOFF.
//
// The design deliberately does not resolve the unresolvable cells in code — a
// move whose anchor a peer deleted has no defensible position to land on. What it
// owes instead is enough context for the operator or an agent to re-derive the
// edit: the trunk's current state, and the intent that could not be applied. That
// is what `sdlc --help` means by errors being next-action specs.
func queueRefusal(stderr io.Writer, s trunkStore, in queue.Intent, cause error) error {
	fmt.Fprintln(stderr)
	cwarn(stderr, "could not apply: "+queueCommitMessage(in))
	fmt.Fprintf(stderr, "  %v\n\n", cause)

	if b, warn, rerr := s.ReadDegraded(queuePath); rerr == nil {
		if warn != "" {
			fmt.Fprintf(stderr, "  (this snapshot is itself stale: %s)\n", warn)
		}
		entries := queue.Parse(b).Entries()
		fmt.Fprintf(stderr, "  the queue on the trunk right now (%d entries):\n", len(entries))
		for _, l := range entries {
			fmt.Fprintf(stderr, "    %s\n", l.String())
		}
		fmt.Fprintln(stderr)
	}
	switch {
	case errors.Is(cause, queue.ErrAnchorMissing):
		fmt.Fprintf(stderr, "  the anchor is gone — pick one from the list above and re-run.\n")
	case errors.Is(cause, queue.ErrSubjectMissing):
		fmt.Fprintf(stderr, "  %s is no longer queued; it may already have been done.\n", in.Ref)
	}
	// queueCommitMessage already carries the "queue: " prefix; adding another
	// produced "queue: queue: add X refused", which the seed run surfaced.
	return fmt.Errorf("%s refused", queueCommitMessage(in))
}
