package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func legacySpanningFixture(t *testing.T, modes ...string) (*trackerRepo, string, string, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	r := &trackerRepo{t: t, root: root, origin: origin}
	detail := "workshop/issues/000335-spanning.md"
	raw := "---\nid: 000335\nstatus: open\ncreated: 2026-10-14\n---\n\n# Spanning migration\n\n## Problem\n\nCount pre-cutover development.\n"
	writeRepoFile(t, root, detail, raw)
	r.git("add", detail)
	r.git("commit", "-qm", "#335: file")
	ownedFixtureAt(t, 1)
	raw = strings.Replace(raw, "status: open", "status: working\nstarted: 2026-10-14T00:01:00Z", 1)
	if len(modes) > 0 {
		raw = strings.Replace(raw, "started: 2026-10-14T00:01:00Z\n", "", 1)
		if modes[0] == "first" {
			raw = strings.Replace(raw, "status: working", "status: open", 1)
		}
	}
	writeRepoFile(t, root, detail, raw)
	r.git("commit", "--allow-empty", "-qam", "#335: claim legacy work")
	ownedFixtureAt(t, 5)
	r.git("commit", "--allow-empty", "-qm", "#335: legacy implementation")
	r.git("push", "-q", "-u", "origin", "main")
	ownedFixtureAt(t, 10)
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("dry: %v %s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v %s", err, stderr)
	}
	r.git("pull", "-q")
	ownerRoot := filepath.Join(t.TempDir(), "owner")
	testfix.Git(t, root, "worktree", "add", "-q", "-b", "000335-spanning", ownerRoot)
	owner := hostClaimant(t, ownerRoot)
	owner.Workspace = "ariadne:1"
	ownedFixtureAt(t, 11)
	cardPath := tracker.CardPath("000335", "spanning")
	ownedFixtureCard(t, r, cardPath, "post-cutover-claim", func(b []byte) ([]byte, error) { return issue.SetCardClaimant(b, owner) })
	writeSession := func(id, cwd string, minutes []int) string {
		p := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", id+".jsonl")
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		data := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", id, cwd)
		for _, m := range minutes {
			data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335 development\"}}\n", m)
		}
		if err := os.WriteFile(p, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	pre := writeSession("legacy-prefix", root, []int{2, 4, 6})
	writeSession("owned-suffix", ownerRoot, []int{12, 15, 20})
	ownedFixtureAt(t, 20)
	testfix.Git(t, ownerRoot, "commit", "--allow-empty", "-qm", "#335: owned implementation")
	return r, ownerRoot, cardPath, pre
}

func TestActualAcrossLegacyCutoverRetainsBothPartsAndReplays(t *testing.T) {
	r, ownerRoot, cardPath, pre := legacySpanningFixture(t)
	inputs, err := resolveOwnedActualInputsFor(t.Context(), ownerRoot, "", "335", "workshop/plans", actualCloseCandidate)
	if err != nil {
		t.Fatal(err)
	}
	if inputs.LegacyPrefix == nil || inputs.Snapshot.LegacyPrefix == nil || inputs.Snapshot.Since != "2026-10-14T00:01:00Z" {
		t.Fatalf("pre-cutover work lost: %+v", inputs)
	}
	if inputs.Snapshot.Version != gatestate.MeasurementCompactBranchInputsVersion || inputs.Snapshot.Method != gatestate.MeasurementMethodBranchHybridV1 {
		t.Fatalf("hybrid lacks explicit schema/interpretation fence: %+v", inputs.Snapshot)
	}
	if inputs.LegacyPrefix.Snapshot.Until != "2026-10-14T00:10:00Z" {
		t.Fatalf("wrong cutover bound: %+v", inputs.LegacyPrefix.Snapshot)
	}
	result := computeActualForPurpose(t.Context(), ownerRoot, "", "335", "", "workshop/plans", actualCloseCandidate)
	if result.Status != actualMeasured || result.Inputs == nil || result.Inputs.LegacyPrefix == nil || !strings.Contains(strings.Join(result.Warnings, " "), "legacy") {
		t.Fatalf("mixed measurement missing: %+v", result)
	}
	suffix := *inputs
	suffix.LegacyPrefix = nil
	suffix.Snapshot.LegacyPrefix = nil
	suffix.Snapshot.LegacyCutover = ""
	suffix.Snapshot.Version = gatestate.MeasurementBranchInputsVersion
	suffix.Snapshot.Method = gatestate.MeasurementMethodBranchV1
	suffix.Snapshot.Since = "2026-10-14T00:11:00Z"
	suffixResult := computeOwnedActualFor(t.Context(), ownerRoot, "335", "", &suffix)
	if suffixResult.Status != actualMeasured || result.Hours <= suffixResult.Hours {
		t.Fatalf("prefix did not contribute: combined=%+v suffix=%+v", result, suffixResult)
	}
	for _, w := range inputs.Snapshot.Engagements {
		at, _ := time.Parse(time.RFC3339Nano, w.Start)
		cut, _ := time.Parse(time.RFC3339Nano, inputs.LegacyPrefix.Snapshot.Until)
		if at.Before(cut) {
			t.Fatalf("overlapping ownership interval %+v", w)
		}
	}
	// A real post-cutover close retains the combined generation.
	saved := *result.Inputs
	gate := gatestate.Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: "2026-10-14T00:21:00Z", Reviewed: saved.CodeHead, ActualInputs: &saved}}}
	writeRepoFile(t, ownerRoot, "workshop/plans/000335-spanning-close-gate.md", gatestate.Render(gate, "fixture"))
	testfix.Git(t, ownerRoot, "add", "workshop/plans/000335-spanning-close-gate.md")
	ownedFixtureAt(t, 21)
	testfix.Git(t, ownerRoot, "commit", "-qm", "#335: close mixed measurement")
	evidence := strings.TrimSpace(testfix.Git(t, ownerRoot, "rev-parse", "HEAD"))
	ownedFixtureCard(t, r, cardPath, "close-mixed", func(b []byte) ([]byte, error) {
		return issue.SetCardCompletion(b, issue.Completion{Token: "close-mixed", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: evidence})
	})
	f, err := os.OpenFile(pre, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("{\"timestamp\":\"2026-10-14T00:07:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#335 appended later\"}}\n")
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	replay := computeActualForPurpose(context.Background(), ownerRoot, "", "335", "", "workshop/plans", actualHistorical)
	if replay.Status != actualMeasured || replay.Hours != result.Hours || replay.Inputs == nil || replay.Inputs.LegacyPrefix == nil {
		t.Fatalf("combined close replay changed: original=%+v replay=%+v", result, replay)
	}
}

func TestLegacyPrefixAuthorityAndReplayStayStrict(t *testing.T) {
	r, ownerRoot, cardPath, _ := legacySpanningFixture(t)
	repo, err := tracker.NewRepository(t.Context(), ownerRoot, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(testfix.Git(t, ownerRoot, "rev-parse", "HEAD"))
	prefix, cutover, err := legacyActualPrefixAuthority(t.Context(), ownerRoot, head, snapshot.Ref(), cardPath, hostClaimant(t, ownerRoot).Repository, "")
	if err != nil || prefix == nil {
		t.Fatalf("positive authority: %+v %v", prefix, err)
	}
	if absent, _, err := legacyActualPrefixAuthority(t.Context(), ownerRoot, head, snapshot.Ref(), tracker.CardPath("000999", "modern"), prefix.Repository, ""); err != nil || absent != nil {
		t.Fatalf("modern issue inherited legacy interval: %+v %v", absent, err)
	}
	if _, _, err := legacyActualPrefixAuthority(t.Context(), ownerRoot, head, strings.Repeat("a", 40), cardPath, prefix.Repository, ""); err == nil {
		t.Fatal("unavailable tracker pin admitted")
	}
	inputs, err := resolveOwnedActualInputsFor(t.Context(), ownerRoot, "", "335", "workshop/plans", actualCloseCandidate)
	if err != nil {
		t.Fatal(err)
	}
	saved := inputs.Snapshot
	if saved.LegacyCutover != cutover {
		t.Fatal("cutover commit not retained")
	}
	for _, kind := range []string{"cutover", "start", "code", "gap"} {
		t.Run(kind, func(t *testing.T) {
			bad := saved
			child := *saved.LegacyPrefix
			bad.LegacyPrefix = &child
			switch kind {
			case "cutover":
				bad.LegacyCutover = head
			case "start":
				bad.LegacyPrefix.Since = "2026-10-14T00:00:00Z"
			case "code":
				bad.LegacyPrefix.CodeHead = head
			case "gap":
				bad.LegacyPrefix.GapMinutes++
			}
			if err := validateRetainedLegacyPrefix(t.Context(), r.root, "335", bad); err == nil {
				t.Fatal("damaged retained prefix gained authority")
			}
		})
	}
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := legacyActualPrefixAuthority(canceled, ownerRoot, head, snapshot.Ref(), cardPath, prefix.Repository, ""); err == nil {
		t.Fatal("cancelled prefix authority accepted")
	}
}

func TestClipOwnedAfterCutoverRetainsDispatchOrigin(t *testing.T) {
	cutover := time.Date(2026, 10, 14, 0, 10, 0, 0, time.UTC)
	began := cutover.Add(-time.Minute)
	input := []transcripts.SourceWindow{{Start: began, End: cutover, Began: began}, {Start: began, End: cutover.Add(time.Minute), Began: began}, {Start: cutover.Add(2 * time.Minute), End: cutover.Add(3 * time.Minute), Began: cutover.Add(2 * time.Minute)}}
	got := clipOwnedAfterCutover(input, cutover)
	if len(got) != 2 || !got[0].Start.Equal(cutover) || !got[0].Began.Equal(began) || !input[1].Start.Equal(began) {
		t.Fatalf("clipping changed provenance or overlap: %+v", got)
	}
}

func TestLegacyPrefixAuthorityUsesPinnedLegacyFallbacks(t *testing.T) {
	for _, mode := range []string{"working", "first"} {
		t.Run(mode, func(t *testing.T) {
			_, ownerRoot, cardPath, _ := legacySpanningFixture(t, mode)
			repo, err := tracker.NewRepository(t.Context(), ownerRoot, "origin")
			if err != nil {
				t.Fatal(err)
			}
			snap, err := repo.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			head := strings.TrimSpace(testfix.Git(t, ownerRoot, "rev-parse", "HEAD"))
			prefix, _, err := legacyActualPrefixAuthority(t.Context(), ownerRoot, head, snap.Ref(), cardPath, hostClaimant(t, ownerRoot).Repository, "")
			want := "2026-10-14T00:01:00Z"
			if mode == "first" {
				want = "2026-10-14T00:00:00Z"
			}
			if err != nil || prefix == nil || prefix.Since != want {
				t.Fatalf("legacy %s anchor missing: %+v %v", mode, prefix, err)
			}
		})
	}
}

func TestLegacyPrefixOnlyRefusesUntilPostCutoverWork(t *testing.T) {
	_, ownerRoot, _, _ := legacySpanningFixture(t)
	cutover := strings.TrimSpace(testfix.Git(t, ownerRoot, "log", "-1", "--diff-filter=A", "--format=%H", "--", tracker.CutoverMarkerPath))
	testfix.Git(t, ownerRoot, "checkout", "-q", "--detach", cutover)
	got := computeActualForPurpose(t.Context(), ownerRoot, "", "335", "", "workshop/plans", actualCloseCandidate)
	if got.Status != actualError || !strings.Contains(got.Detail, "no post-cutover owned interval yet") || !strings.Contains(got.Detail, "before closing") {
		t.Fatalf("prefix-only preview should explain refusal, not adopt: %+v", got)
	}
}

func TestActualAcrossLegacyCutoverQueueRetainsHybridMethod(t *testing.T) {
	r, ownerRoot, cardPath, _ := legacySpanningFixture(t)
	inputs, err := resolveOwnedActualInputsFor(t.Context(), ownerRoot, "", "335", "workshop/plans", actualCloseCandidate)
	if err != nil {
		t.Fatal(err)
	}
	queued := time.Date(2026, 10, 14, 0, 13, 0, 0, time.UTC)
	acquired := queued.Add(time.Minute)
	queue := 60.0
	a := SuiteAttempt{Version: 1, ID: "aabb", Repo: "xianxu/ariadne", SHA: inputs.Snapshot.CodeHead, Check: "suite", Slot: "ariadne:1", QueuedAt: &queued, AcquiredAt: &acquired, StartedAt: acquired, FinishedAt: acquired.Add(time.Minute), QueueSeconds: &queue, LogSHA256: strings.Repeat("b", 64)}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	common := strings.TrimSpace(testfix.Capture(t, r.root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(r.root, common)
	}
	dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "aabb.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	// Private migration fixtures use a file: repository, while persisted suite
	// records require owner/repo names. Supply that scope at the queue seam.
	repo, err := tracker.NewRepository(t.Context(), r.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if err := attachOwnedQueueEvidence(t.Context(), ownerRoot, repo, cardPath, "github.com/xianxu/ariadne", 335, inputs, actualCloseCandidate); err != nil {
		t.Fatal(err)
	}
	measured := computeOwnedActualFor(t.Context(), ownerRoot, "335", "", inputs)
	if measured.Status != actualMeasured || measured.Inputs == nil {
		t.Fatalf("hybrid queued measurement failed: %+v", measured)
	}
	saved := measured.Inputs
	if saved.Version != gatestate.MeasurementCompactBranchInputsVersion || saved.Method != gatestate.MeasurementMethodBranchHybridQueueV1 || saved.LegacyPrefix == nil || len(saved.QueueWaits) != 1 {
		t.Fatalf("queue promotion lost hybrid boundary: %+v", saved)
	}
	if err := saved.Validate(); err != nil {
		t.Fatal(err)
	}
}
