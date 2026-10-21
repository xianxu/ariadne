package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// A copied ignore file can change the policy after the pre-copy Git check.
// This one-time migration accepts no side-tree policy files, including copies
// left by an interrupted attempt; the root ignore policy remains authoritative.
func rejectMoveIgnoreFiles(s sessionMoveSnapshot) error {
	for _, rel := range sortedKeys(s.Files) {
		dest, ok := moveTarget(s.Old, rel)
		if ok && strings.HasPrefix(dest, "session.side/") && path.Base(dest) == ".gitignore" {
			return fmt.Errorf("ignore policy file refused: %s; remove it from the side tree before retrying", rel)
		}
	}
	for _, rel := range sortedKeys(s.Dest) {
		if strings.HasPrefix(rel, "session.side/") && path.Base(rel) == ".gitignore" {
			return fmt.Errorf("ignore policy file refused: %s; remove it from the destination side tree before retrying", rel)
		}
	}
	return nil
}

// Check effective Git policy after appending any missing rules: later
// negations can override a rule that is already present in .gitignore.
func verifyMoveIgnores(root string, s sessionMoveSnapshot, p sessionMovePlan) error {
	// Refresh destination observations so late files cannot change policy or
	// escape the effective-ignore check between copying and cleanup.
	current := map[string]moveSource{}
	if err := readMoveTree(root, "session.side", current); err != nil {
		return err
	}
	if err := rejectMoveIgnoreFiles(sessionMoveSnapshot{Dest: current}); err != nil {
		return err
	}
	paths := map[string]bool{}
	for _, dir := range p.Directories {
		paths[dir+"/"] = true
	}
	for rel := range current {
		paths[rel] = true
	}
	for _, m := range p.Files {
		paths[m.New] = true
	}
	// On retry some sources may already have been removed. Their destination
	// copies still need protection before any further copy or cleanup.
	for rel := range s.Dest {
		paths[rel] = true
	}
	return verifyMoveIgnoredPaths(root, sortedKeys(paths))
}

// Cleanup checks the retiring source's destination without rescanning the live
// destination tree, where editor locks and new content are now expected.
func verifyMoveDestinationIgnored(root, rel string) error {
	return verifyMoveIgnoredPaths(root, []string{rel})
}

func verifyMoveIgnoredPaths(root string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	input := strings.NewReader(strings.Join(paths, "\x00") + "\x00")
	out, err := gitWithInput(root, input, "check-ignore", "-z", "--stdin")
	if err != nil {
		// Exit 1 means no path matched. It is an ordinary policy refusal,
		// while other failures must retain their Git diagnostic.
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return fmt.Errorf("cannot confirm destination ignore policy; fix Git ignore rules before retrying: %w", err)
		}
	}
	ignored := map[string]bool{}
	for _, rel := range strings.Split(out, "\x00") {
		ignored[rel] = true
	}
	for _, rel := range paths {
		if !ignored[rel] {
			return fmt.Errorf("cannot confirm destination %s is ignored; fix Git ignore rules before retrying", rel)
		}
	}
	return nil
}
