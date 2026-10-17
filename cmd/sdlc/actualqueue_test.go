package main

import (
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
	waits, _, err := selectOwnedQueueAttempts([]SuiteAttempt{attempt}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute))
	if err != nil || len(waits) != 1 || waits[0].Exact || waits[0].Start != start.Add(5*time.Second).Format(time.RFC3339Nano) || waits[0].End != attempt.StartedAt.Format(time.RFC3339Nano) {
		t.Fatalf("legacy shifted wait: %+v %v", waits, err)
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
	if _, _, err := selectOwnedQueueAttempts([]SuiteAttempt{foreign}, engagements, "github.com/xianxu/ariadne", start, start.Add(3*time.Minute)); err == nil {
		t.Fatal("foreign slot admitted")
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
}
