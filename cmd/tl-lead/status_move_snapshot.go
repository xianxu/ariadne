package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

func sessionMovePaths(root, old string) (string, string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", err
	}
	top, err := git(root, "rev-parse", "--show-toplevel")
	if err != nil || top != root {
		return "", "", fmt.Errorf("run from a hub checkout root")
	}
	if _, err := os.Stat(filepath.Join(root, ".hub")); err != nil {
		return "", "", fmt.Errorf("%s has no .hub", root)
	}
	if old == "" {
		matches, err := filepath.Glob(filepath.Join(root, "teams", "*", "tl-status-*.md"))
		if err != nil {
			return "", "", err
		}
		switch len(matches) {
		case 0:
			old = "session.md"
		case 1:
			old = matches[0]
		default:
			return "", "", fmt.Errorf("ambiguous legacy sessions: %s; pass OLD explicitly", strings.Join(matches, ", "))
		}
	}
	if filepath.IsAbs(old) {
		old, err = filepath.Rel(root, old)
		if err != nil {
			return "", "", err
		}
	}
	old = filepath.ToSlash(old)
	if _, err := safeMovePath(root, old); err != nil {
		return "", "", err
	}
	if !strings.HasSuffix(old, ".md") {
		return "", "", fmt.Errorf("OLD must be a Markdown session file")
	}
	return root, old, nil
}

// Check every existing path component, including ancestors of a missing leaf.
func safeMovePath(root, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) || filepath.ToSlash(filepath.Clean(rel)) != rel || rel == ".." || strings.HasPrefix(rel, "../") || strings.ContainsAny(rel, "\x00\\") {
		return "", fmt.Errorf("path %q must stay inside the hub checkout", rel)
	}
	p := root
	for _, part := range strings.Split(rel, "/") {
		p = filepath.Join(p, part)
		fi, err := os.Lstat(p)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink refused: %s", p)
		}
	}
	return p, nil
}

func moveRead(root, rel string) (moveSource, error) {
	p, err := safeMovePath(root, rel)
	if err != nil {
		return moveSource{}, err
	}
	fi, err := os.Lstat(p)
	if err != nil {
		return moveSource{}, err
	}
	if !fi.Mode().IsRegular() {
		return moveSource{}, fmt.Errorf("not a regular file: %s", p)
	}
	b, err := os.ReadFile(p)
	return moveSource{b, fi.Mode().Perm(), fi.ModTime()}, err
}

func readMoveTree(root, dir string, into map[string]moveSource) error {
	p, err := safeMovePath(root, dir)
	if err != nil {
		return err
	}
	return filepath.WalkDir(p, func(p string, e fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) && p == filepath.Join(root, filepath.FromSlash(dir)) {
			return nil
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if e.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink refused: %s", rel)
		}
		if e.IsDir() {
			return nil
		}
		if strings.HasSuffix(rel, ".lock") {
			return exitError{3, fmt.Errorf("operator lock: %s; wait for submit and close the editor", rel)}
		}
		if strings.HasSuffix(rel, ".clobbered") {
			return fmt.Errorf("resolve clobbered evidence: %s", rel)
		}
		// Couch owns imported legacy metadata. Leave its companions alone.
		base := filepath.Base(rel)
		if strings.HasPrefix(base, ".") && (strings.HasSuffix(base, ".shadow") || strings.HasSuffix(base, ".sync") || strings.HasSuffix(base, ".flock") || strings.HasSuffix(base, ".round")) {
			return nil
		}
		v, err := moveRead(root, rel)
		if err != nil {
			return err
		}
		into[rel] = v
		return nil
	})
}

func readSessionMove(d deps, root, old string) (sessionMoveSnapshot, error) {
	s := sessionMoveSnapshot{Root: root, Old: old, Files: map[string]moveSource{}, Dest: map[string]moveSource{}, Existing: map[string]bool{}, Registrations: map[string]*syncRegistration{}}
	abs := func(p string) string { return filepath.Join(root, filepath.FromSlash(p)) }

	src, err := moveRead(root, old)
	if errors.Is(err, os.ErrNotExist) && old != "session.md" {
		src, err = moveRead(root, "session.md")
	} else if err == nil {
		s.Files[old] = src
	}
	if err != nil {
		return s, err
	}
	fields, err := moveSessionFields(string(src.Bytes))
	if err != nil {
		return s, fmt.Errorf("%s: %w", old, err)
	}
	if old != "session.md" {
		if err := readMoveTree(root, sidedoc.SideDir(old), s.Files); err != nil {
			return s, err
		}
		if v, err := moveRead(root, filepath.ToSlash(seenPath(old))); err == nil {
			s.Files[filepath.ToSlash(seenPath(old))] = v
		} else if !errors.Is(err, os.ErrNotExist) {
			return s, err
		}
	}
	for _, rel := range []string{"session.md", ".session.md.ss-seen"} {
		if v, err := moveRead(root, rel); err == nil {
			s.Dest[rel] = v
		} else if !errors.Is(err, os.ErrNotExist) {
			return s, err
		}
	}
	if err := readMoveTree(root, "session.side", s.Dest); err != nil {
		return s, err
	}
	if v, err := moveRead(root, ".gitignore"); err == nil {
		s.Ignore = v.Bytes
		s.IgnoreExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return s, err
	}
	paths := map[string]bool{old: true, "session.md": true}
	for rel := range s.Files {
		if strings.HasSuffix(rel, ".md") {
			paths[rel] = true
			to, _ := moveTarget(old, rel)
			paths[to] = true
		}
	}
	for rel := range s.Dest {
		if strings.HasSuffix(rel, ".md") {
			paths[rel] = true
		}
	}
	for _, rel := range sortedKeys(paths) {
		r, err := d.ss.LookupRegistration(abs(rel))
		if err != nil {
			return s, err
		}
		s.Registrations[rel] = r
		if err := checkMoveLock(d, root, rel, r != nil); err != nil {
			return s, err
		}
	}
	oldReg, newReg := s.Registrations[old], s.Registrations["session.md"]
	if oldReg == nil && newReg == nil {
		owner := strings.Trim(fields["couch-owner"], `"`)
		if owner == "" {
			owner = strings.Trim(fields["owner"], `"`)
		}
		if owner == "" {
			owner = "SLOT"
		}
		return s, fmt.Errorf("session is unregistered; run couch --session-sync register %s --owner %s, then rerun", abs(old), owner)
	}
	if oldReg != nil && newReg != nil && old != "session.md" {
		return s, fmt.Errorf("destination session registration is occupied")
	}
	for rel, v := range s.Files {
		to, _ := moveTarget(old, rel)
		if a, b := s.Registrations[rel], s.Registrations[to]; a != nil && b != nil && rel != to {
			return s, fmt.Errorf("destination registration occupied: %s", to)
		}
		if strings.HasSuffix(rel, ".md") {
			if f, ok := sidedoc.Fields(string(v.Bytes)); ok && f["type"] == "session-sync" && s.Registrations[rel] == nil && s.Registrations[to] == nil {
				return s, fmt.Errorf("nested session unregistered; run couch --session-sync register %s --owner SLOT, then rerun", abs(rel))
			}
		}
	}
	// Discover referenced targets without scanning unrelated repository files.
	for rel, v := range s.Files {
		s.Existing[rel] = true
		if !strings.HasSuffix(rel, ".md") {
			continue
		}
		sidedoc.RebaseLinks(string(v.Bytes), rel, rel, func(p string) (string, bool) { return moveTarget(old, p) }, func(p string) bool {
			if at, err := safeMovePath(root, p); err == nil {
				if _, err := os.Lstat(at); err == nil {
					s.Existing[p] = true
				}
			}
			return false
		})
	}
	// No tracked destination may be silently turned into local session state.
	tracked, err := git(root, "ls-files", "--", "session.md", "session.side", ".session.md.ss-seen")
	if err != nil {
		return s, err
	}
	if tracked != "" {
		return s, fmt.Errorf("destination is tracked: %s", tracked)
	}
	return s, nil
}

func checkMoveLock(d deps, root, rel string, registered bool) error {
	for _, suffix := range []string{".lock", ".clobbered"} {
		p, err := safeMovePath(root, rel+suffix)
		if err != nil {
			return err
		}
		if _, err := os.Lstat(p); err == nil {
			if suffix == ".lock" {
				return exitError{3, fmt.Errorf("operator lock on %s; wait for submit and close the editor", rel)}
			}
			return fmt.Errorf("resolve clobbered evidence: %s", rel)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if registered {
		st, err := d.ss.Status(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		if st.Holder != "" {
			return exitError{3, fmt.Errorf("operator holds %s; wait for submit and close the editor", rel)}
		}
	}
	return nil
}
