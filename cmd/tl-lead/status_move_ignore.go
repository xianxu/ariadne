package main

import "fmt"

// Check effective Git policy after appending any missing rules: later
// negations can override a rule that is already present in .gitignore.
func verifyMoveIgnores(root string, s sessionMoveSnapshot, p sessionMovePlan) error {
	paths := map[string]bool{}
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
