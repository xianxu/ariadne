package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
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
	for _, harness := range []string{"claude", "codex", "qoder"} {
		t.Run("lineage "+harness, func(t *testing.T) {
			repo, err := tracker.NewRepository(t.Context(), worker, "origin")
			if err != nil {
				t.Fatal(err)
			}
			inventory, err := repo.SnapshotAt(inputs.Snapshot.TrackerHead)
			if err != nil {
				t.Fatal(err)
			}
			sessions := []transcripts.SessionEvidence{{Harness: harness, ID: "root", File: "root", CWD: worker}, {Harness: harness, ID: "child", File: "child", ParentID: "root", Spawned: true, CWD: t.TempDir(), Started: time.Date(2026, 10, 14, 0, 3, 0, 0, time.UTC)}}
			eligible, _, err := transcripts.EligibleSources(sessions, []transcripts.SourceWindow{{Start: boundary.Add(-18 * time.Minute), End: boundary.Add(11 * time.Minute), Began: boundary.Add(-18 * time.Minute), Roots: []string{worker}}})
			if err != nil {
				t.Fatal(err)
			}
			selected, captures, _, _, err := selectOwnedBranchWindows(t.Context(), repo, inventory, "335", sessions, eligible, boundary.Add(-18*time.Minute), boundary.Add(11*time.Minute))
			if err != nil || len(captures) != 1 || captures[0].Root != worker {
				t.Fatalf("child cwd replaced its admitted root: %+v %v", captures, err)
			}
			for _, file := range []string{"root", "child"} {
				if len(selected[file]) != 1 || !selected[file][0].Start.Equal(boundary) || !selected[file][0].End.Equal(boundary.Add(11*time.Minute)) {
					t.Fatalf("%s branch partition for %s: %+v", harness, file, selected[file])
				}
			}
		})
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
	if binary := os.Getenv("WF_SDLC_580_BINARY"); binary != "" {
		command := exec.CommandContext(t.Context(), binary, "active-time", "--git-repo", worker, "--dir", filepath.Dir(source), "--issue", "336", "--issue", "335", "--since", opts.SinceISO, "--until", opts.UntilISO)
		command.Dir = worker
		got, err := command.CombinedOutput()
		if err != nil || !strings.Contains(string(got), "#335: 0.18 hr  (11.0 min)") || !strings.Contains(string(got), "#336: 0.28 hr  (17.0 min)") {
			t.Fatalf("private executable conformance: %v\n%s", err, got)
		}
	}
	measured := computeOwnedActualFor(t.Context(), worker, "335", "", inputs)
	if measured.Status != actualMeasured || measured.Inputs == nil || measured.Inputs.Version != 4 {
		t.Fatalf("branch actual adoption candidate lost proof: %+v", measured)
	}
	for _, mutation := range []string{"foreign witness", "widened selected window", "foreign ownership-only", "own ownership-only"} {
		t.Run(mutation, func(t *testing.T) {
			bad := *inputs
			raw, err := json.Marshal(inputs.Snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &bad.Snapshot); err != nil {
				t.Fatal(err)
			}
			if mutation == "foreign witness" {
				for i := range bad.Snapshot.Branches[0].Intervals {
					w := &bad.Snapshot.Branches[0].Intervals[i]
					if w.Kind == "branch" {
						w.Branch = "000336-active-a"
					}
				}
			} else if strings.HasSuffix(mutation, "ownership-only") {
				for i := range bad.Snapshot.Branches[0].Intervals {
					w := &bad.Snapshot.Branches[0].Intervals[i]
					if mutation == "foreign ownership-only" && w.Kind == "foreign-branch" || mutation == "own ownership-only" && w.Kind == "branch" {
						w.Kind = "ownership-only"
					}
				}
				if mutation == "foreign ownership-only" {
					bad.Snapshot.Sources[0].Windows[0].Start = bad.Snapshot.Branches[0].Intervals[0].Start
				}
			} else {
				bad.Snapshot.Sources[0].Windows[0].Start = "2026-10-14T00:19:00Z"
			}
			if replay, err := resolveOwnedReplay(t.Context(), worker, "335", bad.Snapshot, *inputs.Ledger); err == nil {
				got := computeOwnedActualFor(t.Context(), worker, "335", "", replay)
				t.Errorf("tampered saved branch proof accepted: status %v, %.1f minutes", got.Status, got.Hours*60)
			}
			got := computeOwnedActualFor(t.Context(), worker, "335", "", &bad)
			if got.Status != actualError || !strings.Contains(got.Detail, "branch proof") {
				t.Fatalf("tampered branch proof accepted: %+v", got)
			}
		})
	}
	// Expiring the selected worktree's history must not fall back to the
	// common checkout's log or silently choose B among the two reservations.
	logPath := strings.TrimSpace(testfix.Capture(t, worker, "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD"))
	rawLog, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(logPath); err != nil {
		t.Fatal(err)
	}
	// #603: the unknown branch is recorded as unattributed, never counted.
	if expired, err := resolveOwnedActualInputs(t.Context(), worker, "", "335", ""); err != nil || len(expired.Snapshot.Sources) != 0 || len(expired.Unattributed) == 0 || expired.Unattributed[0].Kind != gatestate.UnattributedUnknownBranch || !strings.Contains(expired.Unattributed[0].Reason, "concurrent claims") {
		t.Fatalf("expired history among concurrent claims: %+v %v", expired, err)
	}
	// A retained adoption replays the saved classifier and dispatch inputs,
	// even when the live history has expired and new source bytes have arrived.
	f, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("{\"timestamp\":\"2026-10-14T00:35:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"later\"}}\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	replayed, err := resolveOwnedReplay(t.Context(), worker, "335", *measured.Inputs, *inputs.Ledger)
	if err != nil {
		t.Fatal(err)
	}
	again := computeOwnedActualFor(t.Context(), worker, "335", "", replayed)
	if again.Status != actualMeasured || again.Hours != measured.Hours {
		t.Fatalf("expired live history changed adopted branch actual: before%+v after%+v", measured, again)
	}
	out.Reset()
	diagnostics.Reset()
	// #603: the unknown branch is not counted, so nothing is measured here.
	if code := runActiveTime(opts, &out, &diagnostics); code == 0 || strings.Contains(out.String(), "#335:") {
		t.Fatalf("explicit raw options bypassed missing tracked branch proof: code%d %s %s", code, out.String(), diagnostics.String())
	}
	// An expired prefix outside an explicit narrow window is irrelevant, but
	// asking for the broad window still cannot manufacture its missing branch.
	rows := strings.Split(strings.TrimSuffix(string(rawLog), "\n"), "\n")
	if err := os.WriteFile(logPath, []byte(strings.Join(rows[len(rows)-2:], "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	opts.Issues = []string{"335"}
	opts.SinceISO = "2026-10-14T00:25:00Z"
	out.Reset()
	diagnostics.Reset()
	if code := runActiveTime(opts, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "#335: 0.10 hr  (6.0 min)") {
		t.Fatalf("explicit narrowing did not remove irrelevant expired prefix: %d %s %s", code, out.String(), diagnostics.String())
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
	alias := filepath.Join(t.TempDir(), "worker-alias")
	if err := os.Symlink(worker, alias); err != nil {
		t.Fatal(err)
	}
	aliased := captureOwnedBranchEvidence(t.Context(), alias, since, until)
	if aliased.Unavailable != "" || aliased.Path != capture.Path || aliased.Digest != capture.Digest {
		t.Fatalf("alias selected a different worktree log: %+v", aliased)
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
	for _, mode := range []string{"append", "replace", "rewrite"} {
		t.Run("capture race "+mode, func(t *testing.T) {
			if err := os.WriteFile(capture.Path, before, 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			git := func(args ...string) (string, error) {
				calls++
				if calls == 4 {
					switch mode {
					case "append":
						if err := os.WriteFile(capture.Path, append(append([]byte(nil), before...), []byte("changed\n")...), 0600); err != nil {
							t.Fatal(err)
						}
					case "replace":
						replacement := filepath.Join(t.TempDir(), "HEAD")
						if err := os.WriteFile(replacement, before, 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.Rename(replacement, capture.Path); err != nil {
							t.Fatal(err)
						}
					case "rewrite":
						if err := os.Chtimes(capture.Path, time.Unix(1, 0), time.Unix(1, 0)); err != nil {
							t.Fatal(err)
						}
					}
				}
				return ownedActualGit(t.Context(), worker, append([]string{"--no-optional-locks"}, args...)...)
			}
			got := captureOwnedBranchEvidenceWithGit(t.Context(), worker, since, until, git)
			if got.Unavailable == "" || len(got.Intervals) > 0 {
				t.Fatalf("capture race accepted: %+v", got)
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
	if err := os.Rename(worker, filepath.Join(t.TempDir(), "removed-owner")); err != nil {
		t.Fatal(err)
	}
	if got := captureOwnedBranchEvidence(t.Context(), worker, since, until); got.Unavailable == "" || len(got.Intervals) > 0 {
		t.Fatalf("missing owner fell back to a primary log: %+v", got)
	}
}
