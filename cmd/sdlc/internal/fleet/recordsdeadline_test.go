package fleet

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// trackedRepo makes <fleet>/<name>: a migrated repository with a bare origin
// that holds an issue tracker (the manifest plus cards), the cutover marker on
// main, and the tracker fetched. ARCH-MOCK: real Git end to end.
func trackedRepo(t *testing.T, fleetRoot, name string, cards ...map[string][]byte) string {
	t.Helper()
	root := filepath.Join(fleetRoot, name)
	origin := filepath.Join(t.TempDir(), name+".git")
	testfix.Git(t, "", "init", "-q", "-b", "main", root)
	testfix.Git(t, "", "init", "-q", "--bare", "-b", "main", origin)
	configure(t, root)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	// The tracker branch, built in a scratch repository and pushed.
	scratch := filepath.Join(t.TempDir(), "tracker")
	testfix.Git(t, "", "init", "-q", "-b", "issue-tracker", scratch)
	configure(t, scratch)
	files := map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes()}
	for _, set := range cards {
		for p, b := range set {
			files[p] = b
		}
	}
	for p, b := range files {
		full := filepath.Join(scratch, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	testfix.Git(t, scratch, "add", "-A")
	testfix.Git(t, scratch, "commit", "-qm", "tracker")
	testfix.Git(t, scratch, "push", "-q", origin, "HEAD:refs/heads/issue-tracker")
	tip := strings.TrimSpace(testfix.Capture(t, scratch, "rev-parse", "HEAD"))
	testfix.Git(t, root, "fetch", "-q", "origin", "+refs/heads/issue-tracker:refs/remotes/origin/issue-tracker")
	marker := filepath.Join(root, filepath.FromSlash(tracker.CutoverMarkerPath))
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, tracker.CutoverMarkerBytes(tip), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", "--", tracker.CutoverMarkerPath)
	testfix.Git(t, root, "commit", "-qm", "migrate: cutover marker")
	testfix.Git(t, root, "push", "-q", "-u", "origin", "main")
	return root
}

func configure(t *testing.T, dir string) {
	t.Helper()
	for _, kv := range [][2]string{{"user.name", "t"}, {"user.email", "t@t"}, {"commit.gpgsign", "false"}, {"core.hooksPath", os.DevNull}} {
		testfix.Git(t, dir, "config", kv[0], kv[1])
	}
}

// hangRemote points root's origin at an ssh transport that never answers.
func hangRemote(t *testing.T, root string) {
	t.Helper()
	hang := filepath.Join(t.TempDir(), "hang.sh")
	if err := os.WriteFile(hang, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_SSH_COMMAND", hang) // the environment wins over core.sshCommand
	testfix.Git(t, root, "remote", "set-url", "origin", "ssh://tracker.invalid/repo.git")
}

// freshRecords gives the test its own records cache (the package cache lives
// for the process).
func freshRecords(t *testing.T) {
	t.Helper()
	prev := repoRecords
	repoRecords = newRecordsCache(loadRepoRecords).get
	t.Cleanup(func() { repoRecords = prev })
}

func shortDeadline(t *testing.T, d time.Duration) {
	t.Helper()
	prev := recordsReadDeadline
	recordsReadDeadline = d
	t.Cleanup(func() { recordsReadDeadline = prev })
}

// #290: a tracked repository whose remote never answers degrades to unknown
// with the timeout as its reason, within the read deadline, while another
// repository in the same inventory reads present — one dead remote never
// stalls or degrades the others.
func TestHangingRemoteDegradesOnlyItsRepository(t *testing.T) {
	fleetRoot := t.TempDir()
	healthy := trackedRepo(t, fleetRoot, "healthy")
	stuck := trackedRepo(t, fleetRoot, "stuck")
	hangRemote(t, stuck)
	freshRecords(t)
	shortDeadline(t, 500*time.Millisecond)

	start := time.Now()
	inv, err := CollectInventory(context.Background(), fleetRoot, InventoryOptions{Git: execGitReader{}, Machine: func() (MachineIdentity, error) {
		return MachineIdentity{Fingerprint: meFP, Name: "here"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("inventory took %v with a %v read deadline", elapsed, recordsReadDeadline)
	}
	states := map[string]TreeRow{}
	for _, r := range inv.Rows {
		states[filepath.Base(r.TreePath)] = r
	}
	if h := states[filepath.Base(healthy)]; h.ClaimsState != ClaimsPresent {
		t.Fatalf("healthy repository: %s %q", h.ClaimsState, h.ClaimsError)
	}
	if s := states[filepath.Base(stuck)]; s.ClaimsState != ClaimsUnknown || !strings.Contains(s.ClaimsError, "deadline") {
		t.Fatalf("hanging repository: %s %q, want unknown naming the deadline", s.ClaimsState, s.ClaimsError)
	}
}

// #290: the concurrent warm-up changes nothing observable: a warmed inventory
// equals a sequential one (lookups supplied, so no warm-up), each from its
// own cache.
func TestWarmedInventoryEqualsSequential(t *testing.T) {
	fleetRoot := t.TempDir()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		trackedRepo(t, fleetRoot, name)
	}
	machine := func() (MachineIdentity, error) { return MachineIdentity{Fingerprint: meFP, Name: "here"}, nil }
	collect := func(sequential bool) string {
		freshRecords(t)
		opts := InventoryOptions{Git: execGitReader{}, Machine: machine}
		if sequential {
			opts.LookupIssues = func(root, id string) ([]IssueRecord, error) { return LookupRepoIssues(context.Background(), root, id) }
			opts.LookupClaims = func(root string) RepoClaims { return LookupRepoClaims(context.Background(), root) }
		}
		inv, err := CollectInventory(context.Background(), fleetRoot, opts)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(inv)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	warmed, sequential := collect(false), collect(true)
	if warmed != sequential {
		t.Fatalf("warmed inventory differs from sequential:\n%s\n%s", warmed, sequential)
	}
	if !strings.Contains(warmed, `"claims_state":"present"`) {
		t.Fatalf("fixture did not exercise tracker reads: %s", warmed)
	}
}
