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
// collaboration-modes.md "The index"): it drops each entry whose link the
// operator deleted and appends each of add. It writes only the frontmatter,
// through couch's frontmatter scope (#426, pair#643), named by the hash of
// the frontmatter read it was computed from, and retried like writeSection.
// Both mark only the frontmatter seen, so the operator's unread section edits
// stay in the TL's diff; the body's links are read with Peek, which moves no
// baseline. A body edit after that Peek can only delay a drop to the next
// reconcile: entries are never added from links. An entry whose file is
// missing is named, never dropped: the operator removes it.
func syncIndex(d deps, out io.Writer, file string, add []string) error {
	if text, err := d.ss.Peek(file); err == nil {
		if _, r, err := sidedoc.ReconcileIndex(text, add); err == nil && !r.Changed() {
			reportIndex(out, file, r)
			return nil
		}
	}
	for range writeTries {
		block, hash, err := d.ss.ReadFrontmatter(file)
		if err != nil {
			return refusal(file, "discussions: index", err)
		}
		text, err := d.ss.Peek(file)
		if err != nil {
			return err
		}
		_, body, ok := sidedoc.FrontBlock(text)
		if !ok {
			return fmt.Errorf("refused: %s has no closed frontmatter block", file)
		}
		next, r, err := sidedoc.ReconcileIndex(block+body, add)
		if err != nil {
			return fmt.Errorf("refused: %s's discussions: index: %v", file, err)
		}
		if r.Changed() {
			front, _, _ := sidedoc.FrontBlock(next)
			err = writeBody(front, func(path string) error { return d.ss.WriteFrontmatter(file, path, hash) })
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
section (Needs you, Threads), with the path as the link writes it.

The write changes only the frontmatter (couch's read/write --frontmatter,
ariadne#426): the operator's unread section edits stay in your diff.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncIndex(d, cmd.OutOrStdout(), args[0], add)
		},
	}
	c.Flags().StringArrayVar(&add, "add", nil, "a side doc to list, relative to FILE (repeatable)")
	return c
}
