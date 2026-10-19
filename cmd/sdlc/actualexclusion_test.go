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
	if err != nil || len(got) != 2 || got[0].Slot.String() != "pair:6" || got[0].Issue != "649" || !got[0].Start.Equal(start) || got[1].Slot.String() != "pair:7" {
		t.Fatalf("credited=%+v err=%v", got, err)
	}
	_, err = creditedIntervals("649", segments, phases, map[string]string{"other": "pair:7"})
	if err == nil || !strings.Contains(err.Error(), "owned") {
		t.Fatalf("missing source slot should refuse, got %v", err)
	}
	for _, bad := range []string{"ariadne:06", "ariadne:-1", "ariadne", ":6"} {
		_, err = creditedIntervals("649", segments, phases, map[string]string{"owned": bad, "other": "pair:7"})
		if err == nil || !strings.Contains(err.Error(), "owned") {
			t.Fatalf("retained source slot %q must refuse before comparison: %v", bad, err)
		}
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

func TestProtectedEngineBoundaryRequiresCanonicalSlots(t *testing.T) {
	start := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	inputs := &ownedActualInputs{
		Files: []string{"owner"}, SourceSlots: map[string]string{"owner": "ariadne:6"},
		Protected: []activetime.CreditedInterval{{Issue: "649", Slot: "ariadne:6", Start: start, End: start.Add(time.Minute)}},
	}
	if err := validateProtectedSlotIdentities(inputs); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", "ariadne:06", "ariadne", ":6"} {
		inputs.SourceSlots["owner"] = bad
		if err := validateProtectedSlotIdentities(inputs); err == nil {
			t.Errorf("source slot %q reached protected comparison", bad)
		}
	}
	inputs.SourceSlots["owner"] = "ariadne:6"
	inputs.Protected[0].Slot = "ariadne:06"
	if err := validateProtectedSlotIdentities(inputs); err == nil {
		t.Fatal("noncanonical protected slot reached comparison")
	}
	inputs.Protected = nil // unprotected legacy replay permits absent pins
	inputs.SourceSlots = nil
	if err := validateProtectedSlotIdentities(inputs); err != nil {
		t.Fatalf("unprotected legacy replay refused: %v", err)
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

func TestAdoptedRelevanceAllWindowAndSlotStates(t *testing.T) {
	cases := []struct {
		window adoptedWindowState
		slot   adoptedSlotState
		want   adoptedDecision
	}{
		{windowDisjoint, slotDifferent, adoptionSkip},
		{windowDisjoint, slotSame, adoptionSkip},
		{windowDisjoint, slotUnknown, adoptionSkip},
		{windowOverlap, slotDifferent, adoptionSkip},
		{windowOverlap, slotSame, adoptionReplay},
		{windowOverlap, slotUnknown, adoptionRefuse},
		{windowMissing, slotDifferent, adoptionSkip},
		{windowMissing, slotSame, adoptionRefuse},
		{windowMissing, slotUnknown, adoptionRefuse},
	}
	for _, tc := range cases {
		if got := classifyAdopted(tc.window, tc.slot); got != tc.want {
			t.Errorf("window=%v slot=%v: got %v want %v", tc.window, tc.slot, got, tc.want)
		}
	}
}

func TestAdoptedSlotStateRequiresProof(t *testing.T) {
	start := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	candidate := []actualSlotWindow{{slot: parseAdoptedSlot("ariadne:6"), start: start, end: start.Add(time.Hour)}}
	owner := func(slot string) issue.Claimant { return issue.Claimant{Workspace: slot} }
	cases := []struct {
		name        string
		engagements []tracker.EngagementWindow
		windows     []actualSlotWindow
		want        adoptedSlotState
	}{
		{"same", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:6")}}}, candidate, slotSame},
		{"same slot with disjoint source windows", []tracker.EngagementWindow{{Start: start.Add(-time.Hour), End: start, Owners: []issue.Claimant{owner("ariadne:6")}}}, candidate, slotSame},
		{"different", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:7")}}}, candidate, slotDifferent},
		{"empty owner workspace", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("")}}}, candidate, slotUnknown},
		{"absent engagements", nil, candidate, slotUnknown},
		{"absent owners", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour)}}, candidate, slotUnknown},
		{"mixed different and unknown", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:7"), owner("")}}}, candidate, slotUnknown},
		{"mixed same and unknown", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:6"), owner("")}}}, candidate, slotUnknown},
		{"candidate missing slot", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:7")}}}, []actualSlotWindow{{start: start, end: start.Add(time.Hour)}}, slotUnknown},
		{"noncanonical owner slot", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:06")}}}, candidate, slotUnknown},
		{"malformed owner slot", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:-1")}}}, candidate, slotUnknown},
		{"incomplete owner slot", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne")}}}, candidate, slotUnknown},
		{"missing owner repository", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner(":7")}}}, candidate, slotUnknown},
		{"mixed valid and noncanonical", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:7"), owner("ariadne:06")}}}, candidate, slotUnknown},
		{"noncanonical candidate slot", []tracker.EngagementWindow{{Start: start, End: start.Add(time.Hour), Owners: []issue.Claimant{owner("ariadne:7")}}}, []actualSlotWindow{{slot: parseAdoptedSlot("ariadne:06"), start: start, end: start.Add(time.Hour)}}, slotUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, _, _ := adoptedSlotRelationship(tc.engagements, tc.windows); got != tc.want {
				t.Fatalf("slot relationship=%v want %v", got, tc.want)
			}
		})
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
	unknownPath, unknownCard, unknownDetail, unknownBody := seededIssue(t, "000658", "unknown-slot")
	cards[unknownPath], details[unknownDetail] = unknownCard, unknownBody
	legacyPath, legacyCard, legacyDetail, legacyBody := seededIssue(t, "000659", "legacy-no-completion")
	cards[legacyPath], details[legacyDetail] = legacyCard, legacyBody
	malformedPath, malformedCard, malformedDetail, malformedBody := seededIssue(t, "000660", "malformed-slot")
	cards[malformedPath], details[malformedDetail] = malformedCard, malformedBody
	badSourcePath, badSourceCard, badSourceDetail, badSourceBody := seededIssue(t, "000661", "malformed-source-slot")
	cards[badSourcePath], details[badSourceDetail] = badSourceCard, badSourceBody
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
	// A numeric legacy card without completion has no saved window. Pinned
	// ownership in another slot is still sufficient to reject it before
	// demanding completion evidence.
	legacyOwner := owner
	legacyOwner.Workspace = "ariadne:7"
	update("000659", "claim-659", 15, func(raw []byte) ([]byte, error) {
		return issue.SetCardClaimant(raw, legacyOwner)
	})
	update("000659", "legacy-actual-659", 42, func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardField(raw, "actual_hours", "0.17")
		if err != nil {
			return nil, err
		}
		raw, err = issue.SetCardField(raw, "updated", "2026-10-14")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", "codecomplete")
	})
	if intervals, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", current); err != nil || len(intervals) == 0 {
		t.Fatalf("other-slot legacy card without completion blocked the parked actual: intervals=%+v err=%v", intervals, err)
	}
	// The saved #649-style source pin is noncanonical while the claimant and
	// candidate are both ariadne:6. It must refuse, never disappear through
	// a string mismatch after replay (#661 reproduces the review finding).
	update("000661", "claim-661", 15, func(raw []byte) ([]byte, error) {
		return issue.SetCardClaimant(raw, owner)
	})
	badSource := saved
	badSource.Sources = append([]gatestate.MeasurementSource(nil), saved.Sources...)
	if len(badSource.Sources) == 0 {
		t.Fatal("fixture has no retained source")
	}
	badSource.Sources[0].Slot = "ariadne:06"
	badSourceGate := gatestate.Ledger{Gate: "boundary-review", IssueNum: 661, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: saved.Until, Dispatched: saved.Until, Reviewed: saved.CodeHead, ActualInputs: &badSource}}}
	writeRepoFile(t, r.root, "workshop/plans/000661-malformed-source-slot-close-gate.md", gatestate.Render(badSourceGate, "fixture"))
	testfix.Git(t, r.root, "add", "workshop/plans/000661-malformed-source-slot-close-gate.md")
	ownedFixtureAt(t, 43)
	testfix.Git(t, r.root, "commit", "-qm", "#661: close evidence")
	badSourceEvidence := r.git("rev-parse", "HEAD")
	update("000661", "close-661", 44, func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-661", Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: badSourceEvidence})
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
	if _, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", current); err == nil || !strings.Contains(err.Error(), "#661") || !strings.Contains(err.Error(), "source") {
		t.Fatalf("noncanonical retained source silently double-counted #661 with #636: %v", err)
	}
	update("000661", "retire-661", 45, func(raw []byte) ([]byte, error) {
		return issue.SetCardField(raw, "actual_hours", "N/A")
	})
	// BR-5's #649/#636 shape with adopted cards whose workspaces are
	// noncanonical or empty. Both are unknown, never proof of a different slot.
	addUnknownAdoption := func(id, slug, slot string, at int) {
		t.Helper()
		unknown := owner
		unknown.Workspace = slot
		update(id, "claim-"+id, 15, func(raw []byte) ([]byte, error) {
			return issue.SetCardClaimant(raw, unknown)
		})
		number := 0
		fmt.Sscanf(id, "%d", &number)
		gatePath := fmt.Sprintf("workshop/plans/%s-%s-close-gate.md", id, slug)
		unknownGate := gatestate.Ledger{Gate: "boundary-review", IssueNum: number, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Timestamp: saved.Until, Dispatched: saved.Until, Reviewed: saved.CodeHead, ActualInputs: &saved}}}
		writeRepoFile(t, r.root, gatePath, gatestate.Render(unknownGate, "fixture"))
		testfix.Git(t, r.root, "add", gatePath)
		ownedFixtureAt(t, at)
		testfix.Git(t, r.root, "commit", "-qm", "#"+id+": close evidence")
		unknownEvidence := r.git("rev-parse", "HEAD")
		update(id, "close-"+id, at+1, func(raw []byte) ([]byte, error) {
			raw, err := issue.SetCardCompletion(raw, issue.Completion{Token: "close-" + id, Repository: saved.Repository, ReviewedHEAD: saved.CodeHead, EvidenceCommit: unknownEvidence})
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
		if _, err := loadAdoptedExclusions(context.Background(), r.root, "636", "", current); err == nil || !strings.Contains(err.Error(), "#"+strings.TrimLeft(id, "0")) || !strings.Contains(err.Error(), "slot") {
			t.Fatalf("unknown claimant slot %q silently double-counted #%s with #636: %v", slot, id, err)
		}
	}
	addUnknownAdoption("000660", "malformed-slot", "ariadne:06", 41)
	addUnknownAdoption("000658", "unknown-slot", "", 43)
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
