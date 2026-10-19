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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func TestCreditedIntervalsNeedPinnedSlotAndPositiveAllocation(t *testing.T) {
	start := time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)
	segments := []activetime.Segment{
		{Start: start, End: start.Add(10 * time.Minute), Source: "owned", Alloc: map[string]float64{"649": 10}},
		{Start: start.Add(10 * time.Minute), End: start.Add(20 * time.Minute), Source: "owned", Alloc: map[string]float64{"649": 0}},
		{Start: start.Add(20 * time.Minute), End: start.Add(30 * time.Minute), Source: "other", Alloc: map[string]float64{"649": 10}},
	}
	phases := []activetime.Phase{{Start: start, End: start.Add(30 * time.Minute), Kind: activetime.Build}}
	got, err := creditedIntervals("649", segments, phases, map[string]string{"owned": "pair:6", "other": "pair:7"})
	if err != nil || len(got) != 2 || got[0].Slot != "pair:6" || got[0].Issue != "649" || !got[0].Start.Equal(start) || got[1].Slot != "pair:7" {
		t.Fatalf("credited=%+v err=%v", got, err)
	}
	_, err = creditedIntervals("649", segments, phases, map[string]string{"other": "pair:7"})
	if err == nil || !strings.Contains(err.Error(), "owned") {
		t.Fatalf("missing source slot should refuse, got %v", err)
	}
}

func TestCreditedIntervalsProtectOnlyCountedPhases(t *testing.T) {
	start := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	phases := []activetime.Phase{
		{Start: start, End: start.Add(20 * time.Minute), Kind: activetime.Build},
		{Start: start.Add(20 * time.Minute), End: start.Add(50 * time.Minute), Kind: activetime.Converge},
		{Start: start.Add(50 * time.Minute), End: start.Add(time.Hour), Kind: activetime.Tail},
	}
	segments := []activetime.Segment{
		{Start: start.Add(45 * time.Minute), End: start.Add(55 * time.Minute), Source: "owned", Alloc: map[string]float64{"649": 10}},
		{Start: start.Add(55 * time.Minute), End: start.Add(time.Hour), Source: "owned", Alloc: map[string]float64{"649": 5}},
	}
	got, err := creditedIntervals("649", segments, phases, map[string]string{"owned": "ariadne:6"})
	if err != nil || len(got) != 1 || !got[0].Start.Equal(start.Add(45*time.Minute)) || !got[0].End.Equal(start.Add(50*time.Minute)) {
		t.Fatalf("tail was protected or counted converge lost: intervals=%+v err=%v", got, err)
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

func TestSavedAdoptionEndedBeforeLaterClaim(t *testing.T) {
	// Real #417 was adopted on Oct 17; #443 was claimed on Oct 18 in the
	// same slot. A retained claimant does not extend the earlier actual.
	claim := time.Date(2026, 10, 18, 17, 57, 47, 0, time.UTC)
	if savedMeasurementMayOverlap("2026-10-17T02:43:32.067Z", "2026-10-17T04:38:21.716Z", claim, claim.Add(time.Hour)) {
		t.Fatal("#417's Oct 17 saved measurement overlapped #443's Oct 18 claim")
	}
	if savedMeasurementMayOverlap("2026-10-17T02:43:32.067Z", claim.Format(time.RFC3339Nano), claim, claim.Add(time.Hour)) {
		t.Fatal("touching half-open measurement windows overlapped")
	}
	if !savedMeasurementMayOverlap(claim.Add(-time.Minute).Format(time.RFC3339Nano), claim.Add(time.Minute).Format(time.RFC3339Nano), claim, claim.Add(time.Hour)) {
		t.Fatal("actual overlap was discarded")
	}
}

// The candidate's parked claim spans #649's adopted work. Its own scoped
// commit would otherwise claim that entire transcript run (#636's shape).
func TestParkedClaimExcludesOtherAdoptedIssueInSameSlot(t *testing.T) {
	ownedFixtureAt(t, 0)
	t.Setenv("HOME", t.TempDir())
	aPath, aCard, aDetail, aBody := seededIssue(t, "000636", "parked")
	bPath, bCard, bDetail, bBody := seededIssue(t, "000649", "active")
	cards := map[string]string{aPath: aCard, bPath: bCard}
	details := map[string]string{aDetail: aBody, bDetail: bBody}
	otherPaths := make([]string, 8)
	for i := range otherPaths {
		id := fmt.Sprintf("%06d", 650+i)
		path, card, detailPath, detail := seededIssue(t, id, "other-slot")
		cards[path], details[detailPath], otherPaths[i] = card, detail, path
	}
	r := newTrackerRepo(t, cards, details)
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
	found := append([]transcripts.SessionEvidence(nil), current.Sessions...)
	replay, err := ownedReplaySources(*got.Inputs, found)
	if err != nil {
		t.Fatal(err)
	}
	replay.Scope.Issue = "636"
	again := computeOwnedActualFor(context.Background(), r.root, "636", "", replay)
	if again.Status != actualMeasured || again.Hours != got.Hours || len(again.Exclusions) != len(got.Exclusions) {
		t.Fatalf("historical replay moved the adopted actual: replay=%+v candidate=%+v", again, got)
	}
	// Four, then eight other-slot numeric adoptions have no retained source
	// evidence. If the loader tries to replay any one, it fails; successful
	// results at N and 2N prove that unrelated cards cause zero replays.
	for i := range otherPaths {
		id := fmt.Sprintf("%06d", 650+i)
		other := owner
		other.Workspace = "ariadne:7"
		update(id, "claim-"+id, 15+i, func(raw []byte) ([]byte, error) {
			return issue.SetCardClaimant(raw, other)
		})
		update(id, "close-"+id, 32+i, func(raw []byte) ([]byte, error) {
			raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-" + id, Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: evidence})
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
		if i == 3 || i == 7 {
			intervals, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", current)
			if err != nil || len(intervals) == 0 {
				t.Fatalf("%d unrelated adoptions caused replay or lost same-slot protection: intervals=%+v err=%v", i+1, intervals, err)
			}
		}
	}
	unbound := *current
	unbound.SourceSlots = map[string]string{}
	if _, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", &unbound); err == nil || !strings.Contains(err.Error(), "has no pinned slot") {
		t.Fatalf("relevant adopted issue accepted an unbound candidate source: %v", err)
	}
	if err := os.WriteFile(file, []byte("lost retained transcript\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// The live #417 -> #443 shape: an adopted measurement ended before the
	// next issue's claim, though the done card still retains its claimant.
	// Its unavailable old source cannot make the later issue unmeasurable.
	late := *current
	late.Snapshot.Since = "2026-10-14T00:35:00Z"
	late.Windows = map[string][]activetime.ActivityWindow{file: {{Start: time.Date(2026, 10, 14, 0, 35, 0, 0, time.UTC), End: time.Date(2026, 10, 14, 0, 40, 0, 0, time.UTC)}}}
	if intervals, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", &late); err != nil || len(intervals) != 0 {
		t.Fatalf("ended adopted measurement blocked the later claim: intervals=%+v err=%v", intervals, err)
	}
	_, err = loadAdoptedExclusions(context.Background(), r.root, "636", "", current)
	if err == nil || !strings.Contains(err.Error(), "#649") || !strings.Contains(err.Error(), "2026-10-14T00:10:00Z") {
		t.Fatalf("missing adopted evidence did not name issue and overlap: %v", err)
	}
}
