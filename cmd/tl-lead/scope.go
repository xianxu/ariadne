package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// hubScope adapts existing on-disk ownership. Its zero value preserves the
// legacy machine view; an enabled scope never treats a tie as ownership.
type hubScope struct {
	self    string
	hubs    []hub
	enabled bool
}

// readHubScope resolves the checkout from the caller's vantage directory.
// An unreadable peer claim could outrank a local claim, so it refuses too.
func readHubScope(vantage string) (hubScope, error) {
	top, err := git(vantage, "rev-parse", "--show-toplevel")
	if err != nil {
		// Legacy callers may be outside Git. Do not mistake a present claim
		// in their ancestry for absence just because Git could not resolve it.
		dir, absErr := filepath.Abs(vantage)
		if absErr != nil {
			return hubScope{}, absErr
		}
		for {
			if _, claimErr := os.Lstat(filepath.Join(dir, ".hub")); !os.IsNotExist(claimErr) {
				return hubScope{}, fmt.Errorf("hub scope: cannot resolve checkout: %w", err)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return hubScope{}, nil
	}
	_, err = os.Lstat(filepath.Join(top, ".hub"))
	if os.IsNotExist(err) {
		return hubScope{}, nil
	}
	if err != nil {
		return hubScope{}, fmt.Errorf("hub scope: %w", err)
	}
	hubs, err := readHubs(filepath.Dir(top))
	if err != nil {
		return hubScope{}, fmt.Errorf("hub scope: %w", err)
	}
	// readHubs intentionally keeps legacy team listing semantics: a dangling
	// .hub symlink reads as ENOENT there. Scope cannot grant ownership when
	// an existing claim path was skipped for that reason (or changed mid-read).
	observed := map[string]bool{}
	for _, h := range hubs {
		observed[h.Name] = true
	}
	entries, err := os.ReadDir(filepath.Dir(top))
	if err != nil {
		return hubScope{}, fmt.Errorf("hub scope: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() || observed[e.Name()] {
			continue
		}
		path := filepath.Join(filepath.Dir(top), e.Name(), ".hub")
		if _, claimErr := os.Lstat(path); !os.IsNotExist(claimErr) {
			return hubScope{}, fmt.Errorf("hub scope: %s .hub could not be read", e.Name())
		}
	}
	self := filepath.Base(top)
	found := false
	for _, h := range hubs {
		if h.Err != nil {
			return hubScope{}, fmt.Errorf("hub scope: %s .hub unreadable: %w", h.Name, h.Err)
		}
		if h.Name == self {
			found = true
		}
	}
	if !found {
		return hubScope{}, fmt.Errorf("hub scope: %s .hub could not be read", self)
	}
	return hubScope{self: self, hubs: hubs, enabled: true}, nil
}

func (s hubScope) owner(slot FleetSlot) ownership { return ownerOf(slot.Repo, slot.Slot, s.hubs) }
func (s hubScope) owns(slot FleetSlot) bool {
	if !s.enabled {
		return true
	}
	owner, sole := s.owner(slot).sole()
	return sole && owner == s.self
}

// repoNames is the pure, deduplicated set of explicit repository claims.
// An exact slot claim includes that repository's landings too.
func (s hubScope) repoNames() []string {
	names := map[string]bool{}
	for _, h := range s.hubs {
		if h.Name == s.self {
			for _, c := range h.Claims {
				names[c.Repo] = true
			}
		}
	}
	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
