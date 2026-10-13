package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func projectWorkspaceGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func projectWorkspaceFixture(t *testing.T) (string, string, string) {
	t.Helper()
	fleet, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	primary := filepath.Join(fleet, "ariadne")
	if err := os.MkdirAll(primary, 0755); err != nil {
		t.Fatal(err)
	}
	projectWorkspaceGit(t, primary, "init", "-b", "main")
	projectWorkspaceGit(t, primary, "-c", "core.hooksPath=/dev/null", "commit", "--allow-empty", "-m", "init")
	slot := filepath.Join(fleet, "worktree", "ariadne-slot1")
	projectWorkspaceGit(t, primary, "worktree", "add", "-b", "main-slot1", slot)
	return fleet, primary, slot
}

func TestProjectWorkspaceFindUsesCurrentCheckout(t *testing.T) {
	fleet, primary, slot := projectWorkspaceFixture(t)
	writeFleetProject(t, fleet, "ariadne", "primary-only", "executing", "ariadne#18", "100h")
	sibling := filepath.Join(fleet, "ariadne-sibling")
	projectWorkspaceGit(t, filepath.Join(fleet, "ariadne"), "worktree", "add", "-b", "ordinary-sibling", sibling)
	writeFleetProject(t, fleet, "ariadne-sibling", "duplicate", "executing", "ariadne#18", "40h")
	current := writeFleetProject(t, filepath.Dir(slot), filepath.Base(slot), "current-only", "executing", "ariadne#18", "10h")
	peer := writeFleetProject(t, fleet, "metis", "peer", "executing", "ariadne#18", "20h")
	for _, ref := range []string{"#18", "ariadne#18"} {
		matches, _, err := discoverProjectsForRef(ref, slot)
		if err != nil {
			t.Fatalf("%s: %v", ref, err)
		}
		if len(matches) != 2 {
			t.Fatalf("want current + peer, got %+v", matches)
		}
		found := map[string]bool{}
		for _, m := range matches {
			found[m.Path] = true
			if m.RepoDir == primary {
				t.Fatal("primary offered as peer write")
			}
			if m.Path == current && m.Repo != "ariadne" {
				t.Fatalf("slot leaked as repo: %+v", m)
			}
		}
		if !found[current] || !found[peer] {
			t.Fatalf("wrong matches: %+v", matches)
		}
	}
}

func TestProjectWorkspaceFindOrdinarySibling(t *testing.T) {
	fleet, primary, _ := projectWorkspaceFixture(t)
	sibling := filepath.Join(fleet, "ariadne-sibling")
	projectWorkspaceGit(t, primary, "worktree", "add", "-b", "ordinary-sibling", sibling)
	writeFleetProject(t, fleet, "ariadne", "stale", "executing", "ariadne#18", "40h")
	current := writeFleetProject(t, fleet, "ariadne-sibling", "current", "executing", "ariadne#18", "40h")
	matches, _, err := discoverProjectsForRef("ariadne#18", sibling)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Path != current || matches[0].Repo != "ariadne" || matches[0].RepoDir != sibling {
		t.Fatalf("want current ordinary checkout once: %+v", matches)
	}
}

// The forecast resolves cars from the checkout holding the project: a bare #N
// takes that checkout's canonical repo name (not the slot directory's), and
// every card lookup runs from that checkout's root.
func TestProjectWorkspaceForecastResolvesFromCurrentCheckout(t *testing.T) {
	_, _, slot := projectWorkspaceFixture(t)
	subject := writeTrainProject(t, slot, "subject", "| Train | Cars, in order | Slot |\n|---|---|---|\n| A | #18 → #19 | :1 |")
	writeForecastLedger(t)
	var roots []string
	stubForecastLookup(t, map[string]issueMeta{"ariadne#18": {Status: "done", EstimateHours: 4}, "ariadne#19": {Status: "open", EstimateHours: 2}}, &roots)
	d, err := readProject(subject)
	if err != nil {
		t.Fatal(err)
	}
	f, _, err := forecastForProject(context.Background(), d, subject, "")
	if err != nil {
		t.Fatal(err)
	}
	if f.Remaining != 2 || len(f.Notes) != 0 {
		t.Fatalf("want only ariadne#19's 2h at ×1 (no notes), got %+v", f)
	}
	for _, r := range roots {
		if r != slot {
			t.Fatalf("lookup ran from %q, want the slot %q", r, slot)
		}
	}
}

func TestProjectWorkspaceDiscoverySurfacesBrokenPeerGit(t *testing.T) {
	fleet, _, slot := projectWorkspaceFixture(t)
	writeFleetProject(t, filepath.Dir(slot), filepath.Base(slot), "current", "executing", "ariadne#18", "40h")
	peer := filepath.Join(fleet, "broken-peer")
	if err := os.MkdirAll(peer, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, ".git"), []byte("gitdir: missing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _, err := discoverProjectsForRef("#18", slot)
	if err == nil {
		t.Fatal("malformed peer Git evidence must not silently produce partial discovery")
	}
}
