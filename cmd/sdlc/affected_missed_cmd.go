package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

func newAffectedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "affected",
		Short: "Curate the affected suite's declared ownership (#616)",
		Long: `The affected suite (#473) selects packages from the import graph plus the
ownership a repository declares in .sdlc/affected.json (owners, sidecars,
traceability files). A declaration the graph cannot infer and nobody wrote
shows up as a missed mapping: a failure the merge's full suite (or a landing
train's split) caught that the branch's affected selection would not have run.

  sdlc affected missed list [--json]
  sdlc affected missed dismiss --id ID --reason TEXT

The curation loop: a missed mapping names a package and the changed paths that
broke it. Declare the ownership on main (an owners rule, a sidecar's reads, or
a traceability entry) and the next list drops the entry, because main's
mapping now selects that package for those paths.`,
		SilenceUsage: true,
	}
	missed := &cobra.Command{
		Use:   "missed",
		Short: "List or dismiss recorded missed mappings",
		Long: `Missed mappings (#616) are recorded in one ledger per clone, in the common git
directory (sdlc/affected-missed.json), by two producers: a merge whose
integrated suite fails where main's baseline passes, and a landing train's
split attribution. Only a narrowed, group-free selection is evidence; a full
selection, a red main, a bare test name and a registered flake record nothing.
Entries are deduped by package and changed-path set, capped at 200 (the oldest
go first), and keep any members a later sdlc adds.`,
		SilenceUsage: true,
	}
	var asJSON bool
	list := &cobra.Command{
		Use:   "list [--json]",
		Short: "Show open missed mappings, dropping those main's mapping now covers",
		Long: `List the open missed mappings. First, each entry is planned again against
origin/main's affected.json and declarations over main's package graph: an
entry whose paths now select its package is covered, and list removes it. If
the graph or the declarations cannot be read, list shows every entry and says
why nothing was pruned.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := missedRepoRoot()
			if err != nil {
				return err
			}
			return runMissedList(cmd.Context(), cmd.OutOrStdout(), root, asJSON, mainMissedCoverage(root))
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "print the open entries as JSON")
	var id, reason string
	dismiss := &cobra.Command{
		Use:   "dismiss --id ID --reason TEXT",
		Short: "Remove one missed mapping, with the reason",
		Long: `Remove one entry from the missed-mapping ledger: the failure was not a
missing declaration (a flake, an environment problem), or the declaration is
not worth making.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := missedRepoRoot()
			if err != nil {
				return err
			}
			return runMissedDismiss(cmd.Context(), cmd.OutOrStdout(), root, id, reason)
		},
	}
	dismiss.Flags().StringVar(&id, "id", "", "the entry's id, as list prints it")
	dismiss.Flags().StringVar(&reason, "reason", "", "why it is not a missing declaration")
	missed.AddCommand(list, dismiss)
	cmd.AddCommand(missed)
	return cmd
}

func missedRepoRoot() (string, error) {
	out, err := execGitRunner{}.GitInDir(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("affected missed: not in a git checkout: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// missedCoverage builds, once per list, the predicate "main's mapping now selects this entry's package for its
// paths"; an error means nothing can be pruned.
type missedCoverage func(ctx context.Context) (func(affected.MissedEntry) bool, func(), error)

// mainMissedCoverage plans entries against origin/main: its config and declarations over its own graph, loaded once.
func mainMissedCoverage(root string) missedCoverage {
	return func(ctx context.Context) (func(affected.MissedEntry) bool, func(), error) {
		r := boundedLandingGit(ctx, execGitRunner{})
		git := func(args ...string) (string, error) { return landingGit(r, root, args...) }
		out, err := git("rev-parse", "--verify", "origin/main^{commit}")
		if err != nil {
			return nil, nil, fmt.Errorf("pin origin/main: %w", err)
		}
		main := strings.TrimSpace(out)
		cleanupGit, retire := retiringCleanupGit(ctx, root)
		at, cleanup, err := affectedSelectionAt(ctx, root, git, cleanupGit, main, main, nil)
		done := func() {
			if cleanup != nil {
				cleanup()
			}
			retire()
		}
		if err != nil {
			done()
			return nil, nil, err
		}
		return func(e affected.MissedEntry) bool {
			sel := affected.Plan(e.Paths, at.graph, at.cfg, at.mapping, affectedIsDoc)
			return !sel.Full && slices.Contains(sel.GoPkgs, e.Package)
		}, done, nil
	}
}

func runMissedList(ctx context.Context, w io.Writer, root string, asJSON bool, coverage missedCoverage) error {
	store, err := missedStoreFor(root)
	if err != nil {
		return err
	}
	var pruned []affected.MissedEntry
	var open affected.MissedLedger
	covered, done, covErr := coverage(ctx)
	if covErr == nil {
		defer done()
	}
	err = store.update(ctx, time.Now(), func(l *affected.MissedLedger) error {
		if covErr == nil {
			pruned = l.Prune(covered)
		}
		open = *l
		if len(pruned) == 0 {
			return errNoLedgerChange
		}
		return nil
	})
	if err != nil && !errors.Is(err, errNoLedgerChange) {
		return err
	}
	if asJSON {
		entries := open.Entries
		if entries == nil {
			entries = []affected.MissedEntry{}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}
	if covErr != nil {
		fmt.Fprintf(w, "nothing pruned: main's mapping is unavailable (%v)\n", covErr)
	}
	for _, e := range pruned {
		fmt.Fprintf(w, "covered now, removed: %s %s (%s)\n", e.ID, e.Package, e.Test)
	}
	if len(open.Entries) == 0 {
		fmt.Fprintln(w, "no open missed mappings")
		return nil
	}
	for _, e := range open.Entries {
		fmt.Fprintf(w, "%s  %s  %s\n", e.ID, e.Package, e.Test)
		paths := strings.Join(e.Paths, ", ")
		if more := e.PathsTotal - len(e.Paths); more > 0 {
			paths += fmt.Sprintf(" (+%d more)", more)
		}
		fmt.Fprintf(w, "    paths: %s\n", paths)
		fmt.Fprintf(w, "    %d hit(s), last %s by %s on %s (main %s)\n", e.Hits, e.LastSeen.Format(time.RFC3339), e.Source, abbrevSHA(e.Head), abbrevSHA(e.Main))
	}
	fmt.Fprintln(w, "→ declare the ownership on main (an owners rule, a sidecar's reads, or a traceability entry), or `sdlc affected missed dismiss --id ID --reason TEXT`")
	return nil
}

// errNoLedgerChange aborts an update that changed nothing, so a read-only list never rewrites the file.
var errNoLedgerChange = errors.New("no ledger change")

func runMissedDismiss(ctx context.Context, w io.Writer, root, id, reason string) error {
	if id == "" || strings.TrimSpace(reason) == "" {
		return errors.New("affected missed dismiss: --id and --reason are required")
	}
	store, err := missedStoreFor(root)
	if err != nil {
		return err
	}
	var gone affected.MissedEntry
	if err := store.update(ctx, time.Now(), func(l *affected.MissedLedger) error {
		gone, err = l.Dismiss(id)
		return err
	}); err != nil {
		return fmt.Errorf("affected missed dismiss: %w", err)
	}
	fmt.Fprintf(w, "dismissed %s (%s, %s): %s\n", gone.ID, gone.Package, gone.Test, reason)
	return nil
}
