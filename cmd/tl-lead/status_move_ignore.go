package main

import (
	"fmt"
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
	for _, rel := range sortedKeys(paths) {
		if _, err := git(root, "check-ignore", "--quiet", "--", rel); err != nil {
			return fmt.Errorf("cannot confirm destination %s is ignored; fix Git ignore rules before retrying: %w", rel, err)
		}
	}
	return nil
}
