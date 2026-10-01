// relocation.go — the local record that `sdlc move` relocated an issue's work
// (#277). It is the positive evidence relocation requires: the owner's own move
// command wrote it, naming the source and destination worktrees.
//
// Lifecycle (ARCH-FUNERAL): one file per issue in the repository's git common
// dir (shared by its linked worktrees, never committed or published). sdlc move
// writes it before switching; the move's re-stamp removes it on success or when
// relocation does not apply; a failed re-stamp keeps it for the repair, which
// (`sdlc claim` at the destination) removes it. A later move of the same issue
// overwrites it. Bound: one small file per moved issue.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func relocationPath(root, id string) (string, error) {
	out, err := execGitRunner{}.GitInDir(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("locate %s's git directory: %v", root, err)
	}
	return filepath.Join(strings.TrimSpace(string(out)), "sdlc", "relocations", id+".json"), nil
}

func writeRelocation(root, id string, r issue.Relocation) error {
	p, err := relocationPath(root, id)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(p, raw, 0o644)
}

// readRelocation is nil when no move is recorded; an unreadable or malformed
// record is an error, never "no move".
func readRelocation(root, id string) (*issue.Relocation, error) {
	p, err := relocationPath(root, id)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var r issue.Relocation
	if err := json.Unmarshal(raw, &r); err != nil || r.From == "" || r.To == "" {
		return nil, fmt.Errorf("malformed relocation record %s", p)
	}
	return &r, nil
}

func removeRelocation(root, id string) {
	if p, err := relocationPath(root, id); err == nil {
		_ = os.Remove(p)
	}
}
