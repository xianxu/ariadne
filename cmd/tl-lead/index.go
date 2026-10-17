package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// syncIndex reconciles the session file's discussions: index (#424,
// collaboration-modes.md "The index") through one whole-file write: it drops
// each entry whose link the operator deleted and appends each of add. The
// frontmatter is outside every section, so the write is couch's write --all,
// named by the hash of the read it was computed from, and retried like
// writeSection. An entry whose file is missing is named, never dropped: the
// operator removes it.
func syncIndex(d deps, out io.Writer, file string, add []string) error {
	for range writeTries {
		text, hash, err := d.ss.ReadAll(file)
		if err != nil {
			return refusal(file, "discussions: index", err)
		}
		next, r, err := sidedoc.ReconcileIndex(text, add)
		if err != nil {
			return fmt.Errorf("refused: %s's discussions: index: %v", file, err)
		}
		if r.Changed() {
			err = writeBody(next, func(path string) error { return d.ss.WriteAll(file, path, hash) })
			if errors.Is(err, ErrStale) {
				continue
			}
			if err != nil {
				return refusal(file, "discussions: index", err)
			}
		}
		reportIndex(out, file, r)
		return nil
	}
	return exitError{6, fmt.Errorf("refused: %s is still stale after %d reads; someone is editing it", file, writeTries)}
}

func reportIndex(out io.Writer, file string, r sidedoc.Report) {
	if len(r.Dropped) > 0 {
		fmt.Fprintf(out, "index: dropped %s (its link is gone)\n", strings.Join(r.Dropped, ", "))
	}
	if len(r.Added) > 0 {
		fmt.Fprintf(out, "index: added %s\n", strings.Join(r.Added, ", "))
	}
	for _, e := range r.Entries {
		if _, err := os.Stat(filepath.Join(filepath.Dir(file), filepath.FromSlash(e))); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(out, "index: %s names a missing file; remove the entry by hand\n", e)
		}
	}
}

func newIndexCmd(d deps) *cobra.Command {
	var add []string
	c := &cobra.Command{
		Use:   "index FILE [--add DOC]...",
		Short: "Reconcile the discussions: index: drop entries whose link is gone; --add a side doc the TL just linked",
		Long: `The session file's discussions: frontmatter lists the side docs waiting for the
operator's review. The rule is one-way: deleting a link drops its entry on the
next reconcile, and removing an entry means reviewed (the link stays, and no
reconcile adds it back). instructions link adds its answer doc itself, and
submit reconciles; use --add for a discussion side doc linked in another
section (Needs you, Threads), with the path as the link writes it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncIndex(d, cmd.OutOrStdout(), args[0], add)
		},
	}
	c.Flags().StringArrayVar(&add, "add", nil, "a side doc to list, relative to FILE (repeatable)")
	return c
}
