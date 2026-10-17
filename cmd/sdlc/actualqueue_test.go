package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func TestSelectOwnedQueueAttempts(t *testing.T) {
	start := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	q := 60.0
	attempt := SuiteAttempt{Version: 1, ID: "aabb", Repo: "xianxu/ariadne", Check: "suite", Slot: "ariadne:1", SHA: strings.Repeat("a", 40), StartedAt: start.Add(65 * time.Second), QueueSeconds: &q}
	owner := issue.Claimant{Workspace: "ariadne:1", Worktree: "/repo/slot1"}
	engagements := []tracker.EngagementWindow{{Start: start, End: start.Add(3 * time.Minute), Owners: []issue.Claimant{owner}}}
	waits, notes, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute))
	if err != nil || len(waits) != 0 || len(notes) != 1 || !strings.Contains(notes[0], "legacy") {
		t.Fatalf("legacy wait must remain unknown: waits=%+v notes=%+v err=%v", waits, notes, err)
	}
	queued := start
	acquired := start.Add(time.Minute)
	attempt.QueuedAt, attempt.AcquiredAt = &queued, &acquired
	waits, _, err = selectOwnedQueueAttempts([]SuiteAttempt{attempt}, engagements, "github.com/xianxu/ariadne", start.Add(10*time.Second), start.Add(3*time.Minute))
	if err != nil || len(waits) != 1 || !waits[0].Exact || waits[0].QueueSeconds != 50 || waits[0].RawSeconds != 60 {
		t.Fatalf("exact clipped wait: %+v %v", waits, err)
	}
	foreign := attempt
	foreign.Slot = "ariadne:2"
	if foreignWaits, _, err := selectOwnedQueueAttempts([]SuiteAttempt{foreign}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute)); err != nil || len(foreignWaits) != 0 {
		t.Fatalf("foreign slot was not excluded: %+v %v", foreignWaits, err)
	}
	released := []tracker.EngagementWindow{{Start: start, End: start.Add(30 * time.Second), Owners: []issue.Claimant{owner}}, {Start: start.Add(40 * time.Second), End: start.Add(3 * time.Minute), Owners: []issue.Claimant{owner}}}
	if _, _, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, released, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute)); err == nil {
		t.Fatal("queue crossing release/reclaim admitted")
	}
	if _, _, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt, attempt}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute)); err == nil {
		t.Fatal("duplicate attempt admitted")
	}
	attempt.QueueSeconds = nil
	waits, warnings, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute))
	if err != nil || len(waits) != 0 || len(warnings) == 0 {
		t.Fatalf("missing queue evidence: %+v %+v %v", waits, warnings, err)
	}
	// Without a duration, an acquisition after the measurement could still
	// have queued across its end; its irrelevance is not established.
	attempt.StartedAt = start.Add(4 * time.Minute)
	_, warnings, err = selectOwnedQueueAttempts([]SuiteAttempt{attempt}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute))
	if err != nil || len(warnings) == 0 {
		t.Fatalf("unknown queue start after window was treated as irrelevant: %+v %v", warnings, err)
	}
}

func TestMixedQueueBlockersNeverSuggestLegacyExclusion(t *testing.T) {
	start := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	q := 60.0
	owner := issue.Claimant{Workspace: "ariadne:1", Worktree: "/repo/slot1"}
	engagements := []tracker.EngagementWindow{{Start: start, End: start.Add(3 * time.Minute), Owners: []issue.Claimant{owner}}}
	queued, acquired := start, start.Add(time.Minute)
	legacy := SuiteAttempt{ID: "aabb", Repo: "xianxu/ariadne", Check: "suite", Slot: owner.Workspace, StartedAt: start.Add(65 * time.Second), QueueSeconds: &q}
	unknownQueue := SuiteAttempt{ID: "ccdd", Repo: legacy.Repo, Check: "suite", Slot: owner.Workspace, StartedAt: legacy.StartedAt}
	unknownSlot := SuiteAttempt{ID: "eeff", Repo: legacy.Repo, Check: "suite", StartedAt: legacy.StartedAt, QueueSeconds: &q, QueuedAt: &queued, AcquiredAt: &acquired}
	for _, other := range []struct {
		name    string
		attempt SuiteAttempt
	}{
		{"unknown queue", unknownQueue},
		{"unknown slot", unknownSlot},
	} {
		for _, order := range []struct {
			name     string
			attempts []SuiteAttempt
		}{
			{"legacy first", []SuiteAttempt{legacy, other.attempt}},
			{"other first", []SuiteAttempt{other.attempt, legacy}},
		} {
			t.Run(other.name+"/"+order.name, func(t *testing.T) {
				_, blockers, err := selectOwnedQueueAttempts(order.attempts, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute))
				if err != nil || len(blockers) != 2 {
					t.Fatalf("expected both blockers: %v %v", blockers, err)
				}
				res := actualErrorResult("428", &queueEvidenceError{Blockers: blockers})
				for _, render := range []struct {
					name string
					fn   func(*bytes.Buffer)
				}{
					{"preview", func(out *bytes.Buffer) { printActual(out, res, "") }},
					{"milestone preview", func(out *bytes.Buffer) { printActual(out, res, "M2") }},
					{"close", func(out *bytes.Buffer) { explainActual(out, "428", "issue", "", res) }},
					{"milestone close", func(out *bytes.Buffer) { explainActual(out, "428", "milestone", "M2", res) }},
				} {
					var out bytes.Buffer
					render.fn(&out)
					got := out.String()
					if !strings.Contains(got, "ask the TL/hub to restore") || strings.Contains(got, "Exclude this actual from calibration") || !strings.Contains(got, legacy.ID) || !strings.Contains(got, other.attempt.ID) {
						t.Fatalf("%s lost mixed blocker recovery: %s", render.name, got)
					}
				}
			})
		}
	}
}

func TestAllLegacyQueueBlockersNameEveryAttempt(t *testing.T) {
	blockers := []string{
		"legacy suite attempt aabb lacks exact queue endpoints; queue idle is unknown",
		"legacy suite attempt ccdd lacks exact queue endpoints; queue idle is unknown",
	}
	res := actualErrorResult("428", fmt.Errorf("owned queue: %w", &queueEvidenceError{Blockers: blockers}))
	var out bytes.Buffer
	printActual(&out, res, "M2")
	got := out.String()
	if !strings.Contains(got, "aabb, ccdd") || !strings.Contains(got, "sdlc milestone-close --issue 428 --milestone M2 --no-actual") || strings.Contains(got, "ask the TL/hub to restore") {
		t.Fatalf("all-legacy refusal lost complete exclusion guidance: %s", got)
	}
}

func TestLegacyQueueStartCannotProveNonOverlapAcrossBoundaries(t *testing.T) {
	start := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	oneSecond := 1.0
	owner := issue.Claimant{Workspace: "ariadne:1", Worktree: "/repo/slot1"}
	attempt := SuiteAttempt{Version: 1, ID: "legacy", Repo: "xianxu/ariadne", Check: "suite", Slot: owner.Workspace, StartedAt: start.Add(65 * time.Second), QueueSeconds: &oneSecond}
	for _, tc := range []struct {
		name       string
		engagement tracker.EngagementWindow
		until      time.Time
	}{
		{"release boundary", tracker.EngagementWindow{Start: start, End: start.Add(time.Minute), Owners: []issue.Claimant{owner}}, start.Add(2 * time.Minute)},
		{"measurement boundary", tracker.EngagementWindow{Start: start, End: start.Add(2 * time.Minute), Owners: []issue.Claimant{owner}}, start.Add(time.Minute)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			waits, warnings, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, []tracker.EngagementWindow{tc.engagement}, "github.com/xianxu/ariadne", start, tc.until)
			if err != nil || len(waits) != 0 || len(warnings) != 1 || !strings.Contains(warnings[0], "legacy") {
				t.Fatalf("late legacy start concealed an unknown wait: waits=%+v warnings=%+v err=%v", waits, warnings, err)
			}
		})
	}
	attempt.StartedAt = start.Add(-time.Second)
	_, warnings, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, []tracker.EngagementWindow{{Start: start, End: start.Add(time.Minute), Owners: []issue.Claimant{owner}}}, "github.com/xianxu/ariadne", start, start.Add(time.Minute))
	if err != nil || len(warnings) != 0 {
		t.Fatalf("attempt acquired before both starts cannot overlap: warnings=%+v err=%v", warnings, err)
	}
}
