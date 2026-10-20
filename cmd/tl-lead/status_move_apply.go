package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/sidedoc"
)

func newStatusMoveCmd(d deps) *cobra.Command {
	var apply bool
	c := &cobra.Command{Use: "move [OLD] [--apply]", Short: "Copy the old hub session to repo-root session.md, switch Couch registration, then clean up; dry run by default", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root, err := git(cwd, "rev-parse", "--show-toplevel")
		if err != nil {
			return err
		}
		old := ""
		if len(args) > 0 {
			old = args[0]
		}
		return moveSession(d, cmd.OutOrStdout(), root, old, apply, nil)
	}}
	c.Flags().BoolVar(&apply, "apply", false, "copy, switch and clean up while all affected editors are closed")
	return c
}

func copyMoveFile(root string, m sessionMoveFile) error {
	p, err := safeMovePath(root, m.New)
	if err != nil {
		return err
	}
	if old, err := os.ReadFile(p); err == nil {
		if !bytes.Equal(old, m.After) {
			return fmt.Errorf("destination changed: %s", m.New)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := createExclusive(p, m.After); err != nil {
			return err
		}
	} else {
		return err
	}
	if err := os.Chmod(p, m.Mode); err != nil {
		return err
	}
	return os.Chtimes(p, m.Mtime, m.Mtime)
}

func removeMoveSource(root string, m sessionMoveFile) error {
	v, err := moveRead(root, m.Old)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(v.Bytes, m.Before) {
		return fmt.Errorf("source changed: %s; preserved", m.Old)
	}
	return os.Remove(filepath.Join(root, filepath.FromSlash(m.Old)))
}

func verifyMoveRegistration(d deps, root string, m sessionMoveFile, before *syncRegistration) error {
	r, err := d.ss.LookupRegistration(filepath.Join(root, filepath.FromSlash(m.New)))
	if err != nil {
		return err
	}
	if r == nil {
		return fmt.Errorf("new registration missing: %s", m.New)
	}
	if before != nil && (r.ID != before.ID || r.Owner != before.Owner) {
		return fmt.Errorf("registration identity or owner changed: %s", m.New)
	}
	old, err := d.ss.LookupRegistration(filepath.Join(root, filepath.FromSlash(m.Old)))
	if err != nil {
		return err
	}
	if old != nil {
		return fmt.Errorf("old registration remains: %s", m.Old)
	}
	return nil
}

func runSessionMove(d deps, out io.Writer, s sessionMoveSnapshot, p sessionMovePlan, step func(string) error) error {
	phase := movePreparing
	advance := func(e string) error {
		next, err := transitionMove(phase, e)
		if err == nil {
			phase = next
		}
		return err
	}
	after := func(e string) error {
		if step != nil {
			return step(e)
		}
		return nil
	}
	abs := func(rel string) string { return filepath.Join(s.Root, filepath.FromSlash(rel)) }
	check := func(m sessionMoveFile) error {
		for _, rel := range []string{m.Old, m.New} {
			r, err := d.ss.LookupRegistration(abs(rel))
			if err != nil {
				return err
			}
			if err := checkMoveLock(d, s.Root, rel, r != nil); err != nil {
				return err
			}
		}
		v, err := moveRead(s.Root, m.Old)
		if err != nil {
			return err
		}
		if !bytes.Equal(v.Bytes, m.Before) {
			return fmt.Errorf("source changed: %s; rerun dry run", m.Old)
		}
		return nil
	}
	if !bytes.Equal(s.Ignore, p.Ignores) {
		if _, err := safeMovePath(s.Root, ".gitignore"); err != nil {
			return err
		}
		var err error
		if s.IgnoreExists {
			err = replaceIfUnchanged(abs(".gitignore"), s.Ignore, p.Ignores)
		} else {
			err = createExclusive(abs(".gitignore"), p.Ignores)
		}
		if err != nil {
			return err
		}
		if err := after("ignore"); err != nil {
			return err
		}
	}
	for _, m := range p.Files {
		if err := check(m); err != nil {
			return err
		}
		if err := copyMoveFile(s.Root, m); err != nil {
			return err
		}
		if err := after("copy " + m.New); err != nil {
			return err
		}
	}
	if err := advance("copied"); err != nil {
		return err
	}
	for _, m := range p.Files {
		if err := check(m); err != nil {
			return err
		}
		if s.Registrations[m.Old] != nil || s.Registrations[m.New] != nil {
			if err := d.ss.Rename(abs(m.Old), abs(m.New)); err != nil {
				return fmt.Errorf("switch unconfirmed at %s: %w; keep both copies, inspect Couch lookup, then rerun the identical apply", m.Old, err)
			}
			if err := after("rename " + m.New); err != nil {
				return err
			}
			if err := verifyMoveRegistration(d, s.Root, m, s.Registrations[m.Old]); err != nil {
				return err
			}
		}
		v, err := moveRead(s.Root, m.New)
		if err != nil {
			return err
		}
		if !bytes.Equal(v.Bytes, m.After) {
			return fmt.Errorf("new copy changed: %s; preserve both copies", m.New)
		}
	}
	// Verify every copy again before authorizing any deletion. Later edits are
	// still subject to the documented closed-editor maintenance requirement.
	for _, m := range p.Files {
		if err := check(m); err != nil {
			return err
		}
		v, err := moveRead(s.Root, m.New)
		if err != nil {
			return err
		}
		if !bytes.Equal(v.Bytes, m.After) {
			return fmt.Errorf("destination changed: %s", m.New)
		}
	}
	current := map[string]moveSource{}
	if err := readMoveTree(s.Root, sidedoc.SideDir(s.Old), current); err != nil {
		return err
	}
	for rel, now := range current {
		was, ok := s.Files[rel]
		if !ok || !bytes.Equal(was.Bytes, now.Bytes) {
			return fmt.Errorf("source tree changed: %s; preserve both copies and rerun", rel)
		}
	}
	if err := advance("cleanup"); err != nil {
		return err
	}
	fmt.Fprintln(out, "new session in use; cleaning up old copies")
	for _, m := range p.Files {
		if err := check(m); err != nil {
			return err
		}
		v, err := moveRead(s.Root, m.New)
		if err != nil {
			return err
		}
		if !bytes.Equal(v.Bytes, m.After) {
			return fmt.Errorf("new session in use; destination changed during cleanup: %s; old copy preserved", m.New)
		}
		if m.Old == s.Old {
			remaining := map[string]moveSource{}
			if err := readMoveTree(s.Root, sidedoc.SideDir(s.Old), remaining); err != nil {
				return err
			}
			if len(remaining) > 0 {
				return fmt.Errorf("new session in use; source tree changed during cleanup; old session preserved, inspect %s", sidedoc.SideDir(s.Old))
			}
		}
		if err := removeMoveSource(s.Root, m); err != nil {
			return fmt.Errorf("new session in use; cleanup incomplete: %w; rerun", err)
		}
		if err := after("remove " + m.Old); err != nil {
			return err
		}
	}
	removeEmpty(abs(sidedoc.SideDir(s.Old)))
	if err := advance("complete"); err != nil {
		return err
	}
	fmt.Fprintln(out, "migrated: reopen session.md; decisions belong in tracked artifacts")
	return nil
}

func renderSessionMove(out io.Writer, s sessionMoveSnapshot, p sessionMovePlan, apply bool) {
	mode := "dry run"
	if apply {
		mode = "apply"
	}
	fmt.Fprintf(out, "status move: %s -> session.md (%s)\n", s.Old, mode)
	for _, m := range p.Files {
		fmt.Fprintf(out, "copy %s -> %s", m.Old, m.New)
		if !bytes.Equal(m.Before, m.After) {
			fmt.Fprint(out, " (rewrite links/index)")
		}
		if s.Registrations[m.Old] != nil || s.Registrations[m.New] != nil {
			fmt.Fprint(out, " (Couch rename)")
		}
		fmt.Fprintln(out)
	}
	if !bytes.Equal(s.Ignore, p.Ignores) {
		fmt.Fprintln(out, "update .gitignore: session.md, session.side/ and local companions")
	}
	for _, n := range p.Notes {
		fmt.Fprintln(out, n)
	}
	if !apply {
		fmt.Fprintf(out, "next (close all affected editors first): tl-lead status move %s --apply\n", s.Old)
	}
}

func moveSession(d deps, out io.Writer, root, old string, apply bool, step func(string) error) error {
	root, old, err := sessionMovePaths(root, old)
	if err != nil {
		return err
	}
	s, err := readSessionMove(d, root, old)
	if err != nil {
		return err
	}
	if old == "session.md" || len(s.Files) == 0 {
		if s.Registrations["session.md"] == nil {
			return fmt.Errorf("session.md has no registration")
		}
		if old != "session.md" && apply {
			if err := d.ss.Rename(filepath.Join(root, old), filepath.Join(root, "session.md")); err != nil {
				return err
			}
		}
		fmt.Fprintln(out, "already migrated: session.md; reopen the editor at this path")
		return nil
	}
	p, err := planSessionMove(s)
	if err != nil {
		return err
	}
	renderSessionMove(out, s, p, apply)
	if !apply {
		return nil
	}
	if err := runSessionMove(d, out, s, p, step); err != nil {
		return err
	}
	return nil
}
