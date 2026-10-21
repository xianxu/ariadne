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

func verifyMoveRegistration(regs map[string]*syncRegistration, root string, m sessionMoveFile, before *syncRegistration) error {
	r := regs[filepath.Join(root, filepath.FromSlash(m.New))]
	if r == nil {
		return fmt.Errorf("new registration missing: %s", m.New)
	}
	if before != nil && (r.ID != before.ID || r.Owner != before.Owner) {
		return fmt.Errorf("registration identity or fallback owner changed: %s", m.New)
	}
	if old := regs[filepath.Join(root, filepath.FromSlash(m.Old))]; old != nil {
		return fmt.Errorf("old registration remains: %s (ID %s)", m.Old, old.ID)
	}
	return nil
}

func runSessionMove(d deps, out io.Writer, s sessionMoveSnapshot, p sessionMovePlan, step func(string) error) error {
	after := func(e string) error {
		if step != nil {
			return step(e)
		}
		return nil
	}
	abs := func(rel string) string { return filepath.Join(s.Root, filepath.FromSlash(rel)) }
	check := func(m sessionMoveFile, live bool) error {
		paths := []string{m.Old}
		if !live {
			paths = append(paths, m.New)
		}
		for _, rel := range paths {
			if err := checkMoveLock(s.Root, rel); err != nil {
				return err
			}
		}
		v, err := moveRead(s.Root, m.Old)
		if err != nil {
			return err
		}
		if !bytes.Equal(v.Bytes, m.Before) {
			return fmt.Errorf("source changed: %s; preserved", m.Old)
		}
		return nil
	}
	checkSources := func() error {
		current := map[string]moveSource{}
		if err := readMoveTree(s.Root, sidedoc.SideDir(s.Old), current); err != nil {
			return err
		}
		for rel, now := range current {
			was, ok := s.Files[rel]
			if !ok || !bytes.Equal(was.Bytes, now.Bytes) {
				return fmt.Errorf("source tree changed: %s; old copies preserved", rel)
			}
		}
		return nil
	}

	prepare := func() error {
		if p.CleanupOnly {
			if err := d.ss.Rename(abs(s.Old), abs("session.md")); err != nil {
				return fmt.Errorf("cannot confirm prior root rename; old copies preserved: %w", err)
			}
			regs, err := d.ss.RegistrationSnapshot()
			if err != nil {
				return err
			}
			rootPair := sessionMoveFile{Old: s.Old, New: "session.md"}
			if err := verifyMoveRegistration(regs, s.Root, rootPair, s.Registrations["session.md"]); err != nil {
				return err
			}
			for _, m := range p.Files {
				if regs[abs(m.Old)] != nil {
					return fmt.Errorf("old registration conflict: %s (ID %s); preserve old copy", m.Old, regs[abs(m.Old)].ID)
				}
				if err := check(m, true); err != nil {
					return err
				}
			}
			return checkSources()
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
		if err := verifyMoveIgnores(s.Root, s, p); err != nil {
			return err
		}
		for _, dir := range p.Directories {
			dest, err := safeMovePath(s.Root, dir)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			if err := after("mkdir " + dir); err != nil {
				return err
			}
		}
		return nil
	}
	copyOne := func(m sessionMoveFile) error {
		if err := check(m, false); err != nil {
			return err
		}
		if err := copyMoveFile(s.Root, m); err != nil {
			return err
		}
		if err := after("copy " + m.New); err != nil {
			return err
		}
		return nil
	}
	beforeSwitch := func() error {
		if err := verifyMoveIgnores(s.Root, s, p); err != nil {
			return err
		}
		for _, m := range p.Files {
			if err := check(m, false); err != nil {
				return err
			}
			v, err := moveRead(s.Root, m.New)
			if err != nil {
				return err
			}
			if !bytes.Equal(v.Bytes, m.After) {
				return fmt.Errorf("new copy changed before switch: %s; old copies preserved", m.New)
			}
		}
		return checkSources()
	}
	switchOne := func(m sessionMoveFile) error {
		if err := check(m, false); err != nil {
			return err
		}
		if s.Registrations[m.Old] != nil || s.Registrations[m.New] != nil {
			if err := d.ss.Rename(abs(m.Old), abs(m.New)); err != nil {
				return fmt.Errorf("switch unconfirmed at %s: %w; keep copies and rerun", m.Old, err)
			}
			if err := after("rename " + m.New); err != nil {
				return err
			}
		}
		return nil
	}
	verifyAll := func() error {
		regs, err := d.ss.RegistrationSnapshot()
		if err != nil {
			return err
		}
		for _, m := range p.Files {
			if r := regs[abs(m.Old)]; r != nil {
				return fmt.Errorf("old registration conflict: %s ID %s; old copy preserved", m.Old, r.ID)
			}
			if s.Registrations[m.Old] == nil && s.Registrations[m.New] == nil && regs[abs(m.New)] != nil {
				return fmt.Errorf("unexpected destination registration: %s ID %s", m.New, regs[abs(m.New)].ID)
			}
			if s.Registrations[m.Old] != nil || s.Registrations[m.New] != nil {
				before := s.Registrations[m.Old]
				if before == nil {
					before = s.Registrations[m.New]
				}
				if err := verifyMoveRegistration(regs, s.Root, m, before); err != nil {
					return err
				}
			}
			if err := check(m, true); err != nil {
				return err
			}
		}
		return checkSources()
	}

	removeOne := func(m sessionMoveFile) error {
		if err := verifyMoveDestinationIgnored(s.Root, m.New); err != nil {
			return err
		}
		if err := check(m, true); err != nil {
			return err
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
		return nil
	}
	state := sessionMoveState{Total: len(p.Files), Effect: movePrepare}
	for state.Effect != moveDone {
		var err error
		switch state.Effect {
		case movePrepare:
			err = prepare()
		case moveCopy:
			err = copyOne(p.Files[state.Copies])
		case moveSwitch:
			if state.Bindings == 0 {
				err = beforeSwitch()
			}
			if err == nil {
				err = switchOne(p.Files[state.Bindings])
			}
		case moveVerify:
			err = verifyAll()
			if err == nil {
				fmt.Fprintln(out, "new session in use; cleaning up old copies")
			}
		case moveRemove:
			err = removeOne(p.Files[state.Removed])
		default:
			return fmt.Errorf("session move has no authorized effect")
		}
		outcome := moveSucceeded
		if p.CleanupOnly && state.Effect == movePrepare && err == nil {
			outcome = moveResumeConfirmed
		}
		if err != nil {
			outcome = moveFailed
			if state.Effect == moveSwitch {
				outcome = moveUnconfirmed
			}
		}
		next, modelErr := transitionMove(state, outcome)
		if modelErr != nil {
			return modelErr
		}
		state = next
		if err != nil {
			return err
		}
	}
	removeEmpty(abs(sidedoc.SideDir(s.Old)))
	fmt.Fprintln(out, "migrated: reopen session.md; decisions belong in tracked artifacts")
	return nil
}

func renderSessionMove(out io.Writer, s sessionMoveSnapshot, p sessionMovePlan, apply bool) {
	mode := "dry run"
	if apply {
		mode = "apply"
	}
	if p.CleanupOnly {
		fmt.Fprintln(out, "new session in use; apply confirms prior rename before source-only cleanup")
	}
	fmt.Fprintf(out, "status move: %s -> session.md (%s)\n", s.Old, mode)
	if !p.CleanupOnly {
		for _, dir := range p.Directories {
			fmt.Fprintf(out, "directory %s\n", dir)
		}
	}
	for _, m := range p.Files {
		if p.CleanupOnly {
			fmt.Fprintf(out, "remove unchanged %s (keep live %s)\n", m.Old, m.New)
			continue
		}
		fmt.Fprintf(out, "copy %s -> %s", m.Old, m.New)
		if !bytes.Equal(m.Before, m.After) {
			fmt.Fprint(out, " (rewrite links/index)")
		}
		if s.Registrations[m.Old] != nil || s.Registrations[m.New] != nil {
			fmt.Fprint(out, " (Couch rename)")
		}
		fmt.Fprintln(out)
	}
	if !p.CleanupOnly && !bytes.Equal(s.Ignore, p.Ignores) {
		fmt.Fprintln(out, "update .gitignore: session.md, session.side/ and local companions")
	}
	if !p.CleanupOnly {
		for _, n := range p.Notes {
			fmt.Fprintln(out, n)
		}
	}
	if !apply {
		quiet := "close all affected editors first"
		if p.CleanupOnly {
			quiet = "keep old leftovers quiet; the new tree may stay live"
		}
		fmt.Fprintf(out, "next (%s): tl-lead status move %s --apply\n", quiet, s.Old)
	}
}

func moveSession(d deps, out io.Writer, root, old string, apply bool, step func(string) error) error {
	d.ss = boundedMoveSync(d.ss)
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
