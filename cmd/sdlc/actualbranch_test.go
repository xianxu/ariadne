package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// #580 reconstructs the documented parley #418 -> #322 ordering. The earlier
// close is deliberately not an adopted credit: it must not be the only thing
// preventing the later reservation from absorbing the first issue's tail.
func TestOwnedActualConcurrentClaimDoesNotCreditForeignBranchTail(t *testing.T) {
	ownedFixtureAt(t, 0)
	t.Setenv("HOME", t.TempDir())
	pathA, cardA, detailA, bodyA := seededIssue(t, "000336", "active-a")
	pathB, cardB, detailB, bodyB := seededIssue(t, "000335", "reserved-b")
	r := newTrackerRepo(t, map[string]string{pathA: cardA, pathB: cardB}, map[string]string{detailA: bodyA, detailB: bodyB})
	worker := filepath.Join(t.TempDir(), "worker")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000336-active-a", worker)
	owner := hostClaimant(t, worker)
	owner.Workspace = "ariadne:1"
	claim := func(id string, minute int) {
		t.Helper()
		ownedFixtureAt(t, minute)
		repo, err := tracker.NewRepository(t.Context(), r.root, "origin")
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := repo.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		card, ok := snapshot.Card(id)
		if !ok {
			t.Fatal("missing fixture card", id)
		}
		raw, err := issue.SetCardClaimant(card.Raw, owner)
		if err != nil {
			t.Fatal(err)
		}
		raw, err = issue.SetCardField(raw, "started", fmt.Sprintf("2026-10-14T00:%02d:00Z", minute))
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.UpdateCard(card, raw, "claim-"+id, func(string, string) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	claim("000336", 1)
	claim("000335", 2)
	for _, step := range []struct {
		minute int
		msg    string
	}{{10, "#336: code"}, {15, "#336: close"}, {19, "#336: verification and landing tail"}} {
		ownedFixtureAt(t, step.minute)
		testfix.Git(t, worker, "commit", "--allow-empty", "-qm", step.msg)
	}
	ownedFixtureAt(t, 20)
	testfix.Git(t, worker, "switch", "-q", "-c", "000335-reserved-b", "main")
	ownedFixtureAt(t, 31)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: work")
	source := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", "owner.jsonl")
	if err := os.MkdirAll(filepath.Dir(source), 0700); err != nil {
		t.Fatal(err)
	}
	data := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"owner\",\"cwd\":%q}}\n", worker)
	for _, minute := range []int{2, 5, 10, 15, 19, 20, 25, 30, 31} {
		data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"work\"}}\n", minute)
	}
	if err := os.WriteFile(source, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	inputs, err := resolveOwnedActualInputs(context.Background(), worker, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	boundary := time.Date(2026, 10, 14, 0, 20, 0, 0, time.UTC)
	if len(inputs.Windows[source]) == 0 {
		t.Fatal("B lost its witnessed activity")
	}
	for _, window := range inputs.Windows[source] {
		if window.Start.Before(boundary) {
			t.Fatalf("B credited A's unadopted close/test/landing tail: %s -> %s; B starts at %s", window.Start, window.End, boundary)
		}
	}
}
