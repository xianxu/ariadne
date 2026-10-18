package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/pkg/workspace"
)

func TestResolveResourceRootsUsesGlobalPairStore(t *testing.T) {
	fleet := t.TempDir()
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("PAIR_DATA_DIR", filepath.Join(dataHome, "pair", "repos", "one-repo"))
	t.Setenv("COUCH_STORE_DIR", filepath.Join(dataHome, "pair", "couch"))
	slot := filepath.Join(fleet, "worktree", "ariadne-slot1", ".couch")
	team := filepath.Join(fleet, "ops", "teams", "devinfra")
	for _, dir := range []string{slot, team} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	roots, warnings := resolveResourceRoots(context.Background(), workspace.Identity{FleetRoot: fleet}, nil, weeklySnapshotFlags{})
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	has := func(token, path string) bool {
		for _, root := range roots {
			if root.Token == token && root.Path == path {
				return true
			}
		}
		return false
	}
	if !has("pair-data", filepath.Join(dataHome, "pair")) || has("pair-data", os.Getenv("PAIR_DATA_DIR")) || !has("slot-store", slot) || !has("ops-worktree", team) {
		t.Fatalf("wrong root binding: %+v", roots)
	}
}
