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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func TestCreditedIntervalsNeedPinnedSlotAndPositiveAllocation(t *testing.T) {
	start := time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)
	segments := []activetime.Segment{
		{Start: start, End: start.Add(10 * time.Minute), Source: "owned", Alloc: map[string]float64{"649": 10}},
		{Start: start.Add(10 * time.Minute), End: start.Add(20 * time.Minute), Source: "owned", Alloc: map[string]float64{"649": 0}},
		{Start: start.Add(20 * time.Minute), End: start.Add(30 * time.Minute), Source: "other", Alloc: map[string]float64{"649": 10}},
	}
	got, err := creditedIntervals("649", segments, map[string]string{"owned": "pair:6", "other": "pair:7"})
	if err != nil || len(got) != 2 || got[0].Slot != "pair:6" || got[0].Issue != "649" || !got[0].Start.Equal(start) || got[1].Slot != "pair:7" {
		t.Fatalf("credited=%+v err=%v", got, err)
	}
	_, err = creditedIntervals("649", segments, map[string]string{"other": "pair:7"})
	if err == nil || !strings.Contains(err.Error(), "owned") {
		t.Fatalf("missing source slot should refuse, got %v", err)
	}
}

func TestParkedClaimWhollyExcludedNamesNoActualAction(t *testing.T) {
	start := time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)
	result := actualResult{Status: actualEmptyWindow, Issue: "636", Exclusions: []activetime.Exclusion{{Issue: "649", Slot: "pair:6", Start: start, End: start.Add(10 * time.Minute), Minutes: 10}}}
	for _, render := range []func(*bytes.Buffer){
		func(out *bytes.Buffer) { printActual(out, result, "") },
		func(out *bytes.Buffer) { explainActual(out, "636", "issue", "", result) },
	} {
		var out bytes.Buffer
		render(&out)
		for _, needle := range []string{"#649", "pair:6", "2026-10-18T01:00:00Z", "sdlc close --issue 636 --no-actual"} {
			if !strings.Contains(out.String(), needle) {
				t.Fatalf("missing %q in %s", needle, out.String())
			}
		}
	}
}

func TestAdoptedCardDateFilterKeepsTimezoneEdges(t *testing.T) {
	since := time.Date(2026, 10, 18, 4, 0, 0, 0, time.UTC) // Oct 17 in Los Angeles
	if !cardDateMayOverlap("2026-10-17", "2026-10-17", since, since.Add(time.Hour)) {
		t.Fatal("local-date update at UTC day boundary must remain a candidate")
	}
	if cardDateMayOverlap("2026-10-01", "2026-10-01", since, since.Add(time.Hour)) {
		t.Fatal("unrelated old adopted issue was not filtered")
	}
}

// The candidate's parked claim spans #649's adopted work. Its own scoped
// commit would otherwise claim that entire transcript run (#636's shape).
func TestParkedClaimExcludesOtherAdoptedIssueInSameSlot(t *testing.T) {
	ownedFixtureAt(t, 0)
	t.Setenv("HOME", t.TempDir())
	aPath, aCard, aDetail, aBody := seededIssue(t, "000636", "parked")
	bPath, bCard, bDetail, bBody := seededIssue(t, "000649", "active")
	r := newTrackerRepo(t, map[string]string{aPath: aCard, bPath: bCard}, map[string]string{aDetail: aBody, bDetail: bBody})
	ownerRoot := filepath.Join(t.TempDir(), "owner")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000636-parked", ownerRoot)
	owner := hostClaimant(t, ownerRoot)
	owner.Workspace = "ariadne:6"
	update := func(id, token string, at int, edit func([]byte) ([]byte, error)) {
		t.Helper()
		ownedFixtureAt(t, at)
		repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err := repo.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		card, ok := snapshot.Card(id)
		if !ok {
			t.Fatalf("missing card %s", id)
		}
		raw, err := edit(card.Raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.UpdateCard(card, raw, token, func(string, string) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []struct {
		id string
		at int
	}{{"000636", 1}, {"000649", 10}} {
		update(item.id, "claim-"+item.id, item.at, func(raw []byte) ([]byte, error) {
			raw, err := issue.SetCardClaimant(raw, owner)
			if err != nil {
				return nil, err
			}
			return issue.SetCardField(raw, "started", fmt.Sprintf("2026-10-14T00:%02d:00Z", item.at))
		})
	}
	file := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", "owner.jsonl")
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	data := fmt.Sprintf("{\"type\":\"session_meta\",\"timestamp\":\"2026-10-14T00:00:00Z\",\"payload\":{\"id\":\"owner\",\"cwd\":%q}}\n", owner.Worktree)
	for _, event := range []struct {
		minute int
		issue  string
	}{{20, "649"}, {30, "649"}} {
		data += fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#%s\"}}\n", event.minute, event.issue)
	}
	if err := os.WriteFile(file, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	ownedFixtureAt(t, 30)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#649: active work")
	prior, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "649", "", actualCloseCandidate)
	if err != nil {
		t.Fatal(err)
	}
	measured := computeOwnedActualFor(context.Background(), r.root, "649", "", prior)
	if measured.Status != actualMeasured || measured.Hours <= 0 {
		t.Fatalf("prior actual: %+v", measured)
	}
	saved := *measured.Inputs
	gate := gatestate.Ledger{Gate: "boundary-review", IssueNum: 649, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: saved.Until, Dispatched: saved.Until, Reviewed: saved.CodeHead, ActualInputs: &saved}}}
	writeRepoFile(t, r.root, "workshop/plans/000649-active-close-gate.md", gatestate.Render(gate, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000649-active-close-gate.md")
	ownedFixtureAt(t, 31)
	testfix.Git(t, r.root, "commit", "-qm", "#649: close evidence")
	evidence := r.git("rev-parse", "HEAD")
	update("000649", "close-649", 32, func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-649", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: evidence})
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "actual_hours", "0.17")
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "updated", "2026-10-14")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", "codecomplete")
	})
	appendFile, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	for _, minute := range []int{35, 40} {
		_, err = fmt.Fprintf(appendFile, "{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":\"#636\"}}\n", minute)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := appendFile.Close(); err != nil {
		t.Fatal(err)
	}
	ownedFixtureAt(t, 40)
	testfix.Git(t, r.root, "commit", "--allow-empty", "-qm", "#636: resume after parking")
	current, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "636", "", actualCloseCandidate)
	if err != nil {
		t.Fatal(err)
	}
	baseline := computeOwnedActualFor(context.Background(), r.root, "636", "", current)
	got := computeActualForPurpose(context.Background(), r.root, "", "636", "", "", actualCloseCandidate)
	allWarnings := strings.Join(got.Warnings, "\n")
	if got.Status != actualMeasured || !(got.Hours < baseline.Hours) || !strings.Contains(allWarnings, "issue #649") || !strings.Contains(allWarnings, "ariadne:6") || !strings.Contains(allWarnings, "2026-10-14T00:20:00Z") {
		t.Fatalf("parked actual=%+v baseline=%+v", got, baseline)
	}
	if err := os.WriteFile(file, []byte("lost retained transcript\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = loadAdoptedExclusions(context.Background(), r.root, "", "636", "", current)
	if err == nil || !strings.Contains(err.Error(), "#649") || !strings.Contains(err.Error(), "2026-10-14T00:10:00Z") {
		t.Fatalf("missing adopted evidence did not name issue and overlap: %v", err)
	}
}
