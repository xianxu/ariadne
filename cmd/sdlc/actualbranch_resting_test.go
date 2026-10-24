package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func restingProofFixture() gatestate.MeasurementInputs {
	start, split, end := "2026-10-23T21:35:40-07:00", "2026-10-24T04:36:48Z", "2026-10-24T04:46:48Z"
	return gatestate.MeasurementInputs{Version: 3, Method: gatestate.MeasurementMethodBranchV1, Repository: "github.com/xianxu/pair", CodeHead: strings.Repeat("a", 40), TrackerHead: strings.Repeat("b", 40), Since: start, Until: end, GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true,
		Engagements: []gatestate.MeasurementEngagement{{Start: start, End: end, Began: start, Roots: []string{"/pair"}}},
		Sources:     []gatestate.MeasurementSource{{Harness: "codex", ID: "main", CWD: "/pair", Digest: strings.Repeat("c", 64), Started: start, Bytes: 123, Windows: []gatestate.MeasurementInterval{{Start: split, End: end}}, DispatchWindows: []gatestate.MeasurementInterval{{Start: start, End: end}}}},
		Branches:    []gatestate.MeasurementBranchEvidence{{Root: "/pair", Unavailable: "fixture", Intervals: []gatestate.MeasurementBranchInterval{{Source: "codex/main", Start: start, End: split, Branch: "main-slot4", Kind: "resting-excluded"}, {Source: "codex/main", Start: split, End: end, Branch: "000856-cast", Kind: "branch"}}}}}
}

func TestRestingExcludedSavedProofAndVisibleMinutes(t *testing.T) {
	m := restingProofFixture()
	if err := validateOwnedBranchProof(m, "856"); err != nil {
		t.Fatalf("valid resting exclusion refused: %v", err)
	}
	notes := measurementBranchWarnings(m)
	var out bytes.Buffer
	printActual(&out, actualResult{Status: actualMeasured, Issue: "856", Hours: 1.0 / 6, Warnings: notes}, "")
	for _, part := range []string{"excluded and unassigned", "1.13 branch-window minutes", "main-slot4", "codex/main"} {
		if !strings.Contains(out.String(), part) {
			t.Fatalf("missing visible exclusion %q: %s", part, out.String())
		}
	}
	for _, mutation := range []string{"selected resting", "own branch", "other branch", "unknown branch"} {
		t.Run(mutation, func(t *testing.T) {
			bad := restingProofFixture()
			switch mutation {
			case "selected resting":
				bad.Sources[0].Windows[0].Start = bad.Since
			case "own branch":
				bad.Branches[0].Intervals[0].Branch = "000856-cast"
			case "other branch":
				bad.Branches[0].Intervals[0].Branch = "000850-cast"
			case "unknown branch":
				bad.Branches[0].Intervals[0].Branch = "feature"
			}
			if err := validateOwnedBranchProof(bad, "856"); err == nil {
				t.Fatal("tampered resting proof accepted")
			}
		})
	}
}

func TestOwnedActualRestingExclusionReplayAndManualPartition(t *testing.T) {
	ownedFixtureAt(t, 0)
	t.Setenv("HOME", t.TempDir())
	pathA, cardA, detailA, bodyA := seededIssue(t, "000336", "active-a")
	pathB, cardB, detailB, bodyB := seededIssue(t, "000335", "reserved-b")
	r := newTrackerRepo(t, map[string]string{pathA: cardA, pathB: cardB}, map[string]string{detailA: bodyA, detailB: bodyB})
	worker := filepath.Join(t.TempDir(), "worker")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000336-active-a", worker)
	owner := hostClaimant(t, worker)
	owner.Workspace = "ariadne:1"
	for i, id := range []string{"000336", "000335"} {
		minute := i + 1
		ownedFixtureAt(t, minute)
		repo, err := tracker.NewRepository(t.Context(), r.root, "origin")
		if err != nil {
			t.Fatal(err)
		}
		snap, err := repo.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		card, ok := snap.Card(id)
		if !ok {
			t.Fatal(id)
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
	ownedFixtureAt(t, 10)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#336: work")
	testfix.Git(t, r.root, "merge", "--ff-only", "000336-active-a")
	ownedFixtureAt(t, 12)
	testfix.Git(t, worker, "switch", "-q", "-c", "main-slot1", "main")
	ownedFixtureAt(t, 20)
	testfix.Git(t, worker, "switch", "-q", "-c", "000335-reserved-b", "main")
	ownedFixtureAt(t, 31)
	testfix.Git(t, worker, "commit", "--allow-empty", "-qm", "#335: work")
	source := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", "owner.jsonl")
	if err := os.MkdirAll(filepath.Dir(source), 0700); err != nil {
		t.Fatal(err)
	}
	data := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"owner\",\"cwd\":%q}}\n", worker)
	for _, minute := range []int{2, 5, 10, 12, 15, 19, 20, 25, 30, 31} {
		data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"work\"}}\n", minute)
	}
	if err := os.WriteFile(source, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	inputs, err := resolveOwnedActualInputs(context.Background(), worker, "", "335", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(inputs.Warnings, " "), "excluded and unassigned 8.00 branch-window minutes") {
		t.Fatalf("fresh warning missing: %v", inputs.Warnings)
	}
	got := computeOwnedActualFor(t.Context(), worker, "335", "", inputs)
	if got.Status != actualMeasured || got.Inputs == nil {
		t.Fatalf("measurement failed: %+v", got)
	}
	var out, diagnostics bytes.Buffer
	opts := activetime.Options{Context: t.Context(), Dirs: []string{filepath.Dir(source)}, GitRepo: worker, Issues: []string{"336", "335"}, SinceISO: "2026-10-14T00:00:00Z", UntilISO: "2026-10-14T00:40:00Z", ThresholdMin: 15, CommitWeight: 1}
	if code := runActiveTime(opts, &out, &diagnostics); code != 0 {
		t.Fatalf("manual partition: %d %s", code, diagnostics.String())
	}
	for _, want := range []string{"#335: 0.18 hr  (11.0 min)", "#336: 0.13 hr  (8.0 min)"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("wrong manual credit: %s", out.String())
		}
	}
	if !strings.Contains(out.String(), "8.00 branch-window minutes") {
		t.Fatalf("manual exclusion invisible: %s", out.String())
	}
	if binary := os.Getenv("WF_SDLC_591_BINARY"); binary != "" {
		cmd := exec.CommandContext(t.Context(), binary, "active-time", "--git-repo", worker, "--dir", filepath.Dir(source), "--issue", "336", "--issue", "335", "--since", opts.SinceISO, "--until", opts.UntilISO)
		cmd.Dir = worker
		raw, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(raw), "8.00 branch-window minutes") || !strings.Contains(string(raw), "#335: 0.18 hr  (11.0 min)") {
			t.Fatalf("executable proof: %v %s", err, raw)
		}
	}
	logPath := strings.TrimSpace(testfix.Capture(t, worker, "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD"))
	if err := os.Remove(logPath); err != nil {
		t.Fatal(err)
	}
	replay, err := resolveOwnedReplay(t.Context(), worker, "335", *got.Inputs, *inputs.Ledger)
	if err != nil {
		t.Fatal(err)
	}
	again := computeOwnedActualFor(t.Context(), worker, "335", "", replay)
	if again.Status != actualMeasured || again.Hours != got.Hours || !strings.Contains(strings.Join(replay.Warnings, " "), "8.00 branch-window minutes") {
		t.Fatalf("replay changed result or hid exclusion: %+v %v", again, replay.Warnings)
	}
}
