package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
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
	testfix.Git(t, r.root, "merge", "--ff-only", "000336-active-a")
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
	if len(inputs.Snapshot.Branches) == 0 || len(inputs.Snapshot.Sources[0].DispatchWindows) == 0 {
		t.Fatal("new measurement lost branch or original dispatch provenance")
	}
	if err := inputs.Snapshot.Validate(); err != nil {
		t.Fatal(err)
	}
	// The manual command must use the same selector, including when explicit
	// raw directories and wide timestamps would have bypassed ownership before.
	var out, diagnostics bytes.Buffer
	opts := activetime.Options{Context: t.Context(), Dirs: []string{filepath.Dir(source)}, GitRepo: worker, Issues: []string{"336", "335"}, SinceISO: "2026-10-14T00:00:00Z", UntilISO: "2026-10-14T00:40:00Z", ThresholdMin: 15, CommitWeight: 1}
	if code := runActiveTime(opts, &out, &diagnostics); code != 0 {
		t.Fatalf("tracked manual measurement failed: %d %s", code, diagnostics.String())
	}
	if !strings.Contains(out.String(), "#335: 0.18 hr  (11.0 min)") || !strings.Contains(out.String(), "#336: 0.28 hr  (17.0 min)") {
		t.Fatalf("manual attribution bypassed branch partition: %s", out.String())
	}
	// Expiring the selected worktree's history must not fall back to the
	// common checkout's log or silently choose B among the two reservations.
	logPath := strings.TrimSpace(testfix.Capture(t, worker, "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD"))
	if err := os.Remove(logPath); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveOwnedActualInputs(t.Context(), worker, "", "335", ""); err == nil || !strings.Contains(err.Error(), "--no-actual") || !strings.Contains(err.Error(), "concurrent claims") {
		t.Fatalf("expired history among concurrent claims: %v", err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := runActiveTime(opts, &out, &diagnostics); code == 0 || !strings.Contains(diagnostics.String(), "concurrent claims") {
		t.Fatalf("explicit raw options bypassed missing tracked branch proof: code%d %s", code, diagnostics.String())
	}
}

func TestOwnedBranchCaptureUsesExactWorktreeAndRejectsUnavailableHistory(t *testing.T) {
	ownedFixtureAt(t, 0)
	_, card, detail, body := seededIssue(t, "000335", "b")
	r := newTrackerRepo(t, map[string]string{"workshop/issue-cards/000335-b.md": card}, map[string]string{detail: body})
	worker := filepath.Join(t.TempDir(), "worker")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000335-b", worker)
	ownedFixtureAt(t, 10)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: work")
	since := time.Date(2026, 10, 14, 0, 1, 0, 0, time.UTC)
	until := since.Add(10 * time.Minute)
	capture := captureOwnedBranchEvidence(t.Context(), worker, since, until)
	if capture.Unavailable != "" || capture.Digest == "" || capture.Bytes == 0 || len(capture.Intervals) == 0 || capture.Intervals[0].Branch != "000335-b" {
		t.Fatalf("native capture = %+v", capture)
	}
	primaryLog := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD"))
	if capture.Path == primaryLog {
		t.Fatal("captured the primary log")
	}
	before, err := os.ReadFile(capture.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"truncated", before[:len(before)-1]},
		{"malformed", []byte("not a log\n")},
		{"limit", make([]byte, (4<<20)+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(capture.Path, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			got := captureOwnedBranchEvidence(t.Context(), worker, since, until)
			if got.Unavailable == "" || len(got.Intervals) != 0 {
				t.Fatalf("invalid capture accepted: %+v", got)
			}
		})
	}
	if err := os.WriteFile(capture.Path, before, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if got := captureOwnedBranchEvidence(ctx, worker, since, until); got.Unavailable == "" {
		t.Fatal("canceled capture succeeded")
	}
}
