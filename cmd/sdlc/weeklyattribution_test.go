package main

import (
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestValidateLandingEvidenceRejectsConflictingIdentity(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	merge := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	head := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	base := weeklyRepoEvidence{
		Events: []landingline.Event{
			{Kind: landingline.EventEnqueue, Seq: 7, Issue: 42, Branch: "issue-42", At: start.Add(time.Hour)},
			{Kind: landingline.EventLanded, Seq: 7, Issue: 42, Branch: "issue-42", PR: 9, At: start.Add(2 * time.Hour)},
		},
		Cards:  map[int]weeklyCardEvidence{42: {Status: "done", Completion: &issue.Completion{LandedCommit: merge}}},
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: head, BranchSHAs: map[string]bool{head: true}, EvidenceReachable: true, ReachableFromMain: true}},
	}
	if got, reasons := validateLandingEvidence(base, w); len(got) != 1 || len(reasons) != 0 {
		t.Fatalf("valid landing = %v, %v", got, reasons)
	}
	bad := base
	bad.Events = append(append([]landingline.Event{}, base.Events...), landingline.Event{Kind: landingline.EventLanded, Seq: 7, Issue: 42, Branch: "other", PR: 9, At: start.Add(2 * time.Hour)})
	if got, reasons := validateLandingEvidence(bad, w); len(got) != 0 || len(reasons) == 0 {
		t.Fatalf("conflicting ticket must be unknown: %v, %v", got, reasons)
	}
	missingCard := base
	missingCard.Cards = nil
	if got, reasons := validateLandingEvidence(missingCard, w); len(got) != 0 || len(reasons) == 0 {
		t.Fatalf("missing card must be unknown: %v, %v", got, reasons)
	}
	missingGraph := base
	missingGraph.Graphs = nil
	if got, reasons := validateLandingEvidence(missingGraph, w); len(got) != 0 || len(reasons) == 0 {
		t.Fatalf("missing graph must be unknown: %v, %v", got, reasons)
	}
}

func TestValidateLandingEvidenceStartsAtFirstEnqueueAfterRefusal(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	merge := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	e := weeklyRepoEvidence{
		Events: []landingline.Event{
			{Kind: landingline.EventEnqueue, Seq: 3, Issue: 42, Branch: "issue-42", At: start.Add(-time.Hour)},
			{Kind: landingline.EventRefused, Seq: 3, Issue: 42, Branch: "issue-42", At: start.Add(-30 * time.Minute)},
			{Kind: landingline.EventEnqueue, Seq: 4, Issue: 42, Branch: "issue-42", At: start.Add(time.Hour)},
			{Kind: landingline.EventLanded, Seq: 4, Issue: 42, Branch: "issue-42", PR: 9, At: start.Add(2 * time.Hour)},
		},
		Cards:  map[int]weeklyCardEvidence{42: {Status: "done", Completion: &issue.Completion{LandedCommit: merge}}},
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: "issue", BranchSHAs: map[string]bool{"issue": true}, EvidenceReachable: true, ReachableFromMain: true}},
	}
	landings, reasons := validateLandingEvidence(e, w)
	if len(reasons) != 0 || len(landings) != 1 || !landings[0].EnqueuedAt.Equal(start.Add(-time.Hour)) {
		t.Fatalf("first enqueue lost: %v, %v", landings, reasons)
	}
}

func TestAttributeWeeklyAttemptsCountsCrossBoundaryAndDuplicateUnknown(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	landings := []weeklyLandingEvidence{{Issue: 42, EnqueuedAt: start.Add(-time.Hour), LandedAt: start.Add(2 * time.Hour), BranchSHAs: map[string]bool{"issue-head": true}}}
	rows := []SuiteAttempt{
		{ID: "failed", Check: suiteCheckName, SHA: "issue-head", FinishedAt: start.Add(-30 * time.Minute), Exit: 1},
		{ID: "passed", Check: suiteCheckName, SHA: "issue-head", FinishedAt: start.Add(time.Hour)},
		{ID: "elsewhere", Check: suiteCheckName, SHA: "other", FinishedAt: start.Add(time.Hour)},
	}
	got := attributeWeeklyAttempts(rows, landings, w, nil)
	if got.LandedIssues != 1 || got.AttributedAttempts != 2 || got.AttributedFailed != 1 || got.AttributedPassed != 1 || got.AttributedRepeats != 1 || got.OutsideWeekAttributed != 1 || got.UnassignableAttempts != 1 || got.AttemptsPerLandedIssue == nil || *got.AttemptsPerLandedIssue != 2 {
		t.Fatalf("cross-boundary attribution = %+v", got)
	}
	rows = append(rows, rows[1])
	got = attributeWeeklyAttempts(rows, landings, w, nil)
	if got.AttemptsPerLandedIssue != nil || len(got.CoverageUnknown) == 0 {
		t.Fatalf("duplicate attempt must make coverage unknown: %+v", got)
	}
	if empty := attributeWeeklyAttempts(nil, nil, w, nil); empty.AttemptsPerLandedIssue != nil || empty.LandedIssues != 0 {
		t.Fatalf("zero landings must have null ratio: %+v", empty)
	}
}

func TestValidateLandingEvidenceMarksMissingEventForWeeklyMergeUnknown(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	merge := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	e := weeklyRepoEvidence{
		Cards:            map[int]weeklyCardEvidence{42: {Status: "done", Completion: &issue.Completion{LandedCommit: merge}}},
		RecentMainMerges: map[string]time.Time{merge: start.Add(time.Hour)},
	}
	if landings, unknown := validateLandingEvidence(e, w); len(landings) != 0 || len(unknown) == 0 {
		t.Fatalf("missing event for weekly merge accepted: %v, %v", landings, unknown)
	}
}
