package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// The seen-set records the owed event IDs the TL has looked at (listed by
// `status submit` or `status owed`), so `--fold-owed` folds only those: a submit
// that lands mid-turn stays owed until a later listing shows it. It lives next
// to the status file as `.<name>.ss-seen` (the name ss.py used, so a cutover
// keeps it) and is bounded by the owed set: listing replaces it with the IDs
// owed now, and a fold removes what it folded.

func seenPath(file string) string {
	dir, base := filepath.Split(file)
	return filepath.Join(dir, "."+base+".ss-seen")
}

// loadSeen reads the seen-set. A missing file is an empty set; an unreadable
// or malformed one is an error, never silently empty (it gates folding).
func loadSeen(file string) ([]string, error) {
	raw, err := os.ReadFile(seenPath(file))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil, fmt.Errorf("%s is malformed (%v); delete it and list owed events again", seenPath(file), err)
	}
	return ids, nil
}

func saveSeen(file string, ids []string) error {
	ids = slices.Clone(ids)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	if ids == nil {
		ids = []string{}
	}
	raw, _ := json.Marshal(ids)
	return os.WriteFile(seenPath(file), raw, 0o644)
}

// foldable is the IDs both owed now and already seen, in owed order.
func foldable(owed, seen []string) []string {
	var out []string
	for _, id := range owed {
		if slices.Contains(seen, id) {
			out = append(out, id)
		}
	}
	return out
}

// without returns ids minus drop.
func without(ids, drop []string) []string {
	var out []string
	for _, id := range ids {
		if !slices.Contains(drop, id) {
			out = append(out, id)
		}
	}
	return out
}
