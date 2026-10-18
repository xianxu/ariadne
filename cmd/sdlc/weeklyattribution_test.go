package main

import (
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestWeeklyHeadlineShowsObservedCoverageAndTopReasons(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	raw := weeklyRepoEvidence{
		Events: []landingline.Event{
			{Kind: landingline.EventLanded, Issue: 42, At: start.Add(time.Hour)},
			{Kind: landingline.EventLanded, Issue: 43, At: start.Add(time.Hour)},
		},
		Cards:            map[int]weeklyCardEvidence{44: {Completion: &issue.Completion{LandedCommit: "merge"}}},
		RecentMainMerges: map[string]time.Time{"merge": start.Add(time.Hour)},
	}
	attribution := weeklyAttribution{LandedIssues: 1, AttributedAttempts: 2, CoverageUnknown: []string{
		"issue 42 ticket 1: no done completion binding",
		"issue 43 ticket 2: no done completion binding",
		"issue 44: main landing has no landed event",
	}}
	got := weeklyHeadline(raw, w, attribution)
	if got.ObservedLandedIssues != 3 || got.CoveredLandedIssues != 0 || got.UnknownLandedIssues != 3 || got.AttemptsPerLandedIssue != nil {
		t.Fatalf("headline coverage = %+v", got)
	}
	if len(got.TopUnknownReasons) != 2 || got.TopUnknownReasons[0].Reason != "no done completion binding" || got.TopUnknownReasons[0].Count != 2 {
		t.Fatalf("top reasons = %+v", got.TopUnknownReasons)
	}
	if !strings.Contains(got.Text, "0 of 3 observed landed issues") || !strings.Contains(got.Text, "3 unknown") {
		t.Fatalf("headline not glanceable: %q", got.Text)
	}
	ratio := 2.0
	attribution.CoverageUnknown = nil
	attribution.AttemptsPerLandedIssue = &ratio
	got = weeklyHeadline(raw, w, attribution)
	if got.CoveredLandedIssues != 1 || got.UnknownLandedIssues != 2 || got.AttemptsPerLandedIssue == nil {
		t.Fatalf("numeric headline coverage = %+v", got)
	}
}

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
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: head, PR: 9, Branch: "issue-42", BranchSHAs: map[string]bool{head: true}, EvidenceReachable: true, ReachableFromMain: true}},
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
	wrongPR := base
	wrongPR.Graphs = map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: head, PR: 10, Branch: "issue-42", BranchSHAs: map[string]bool{head: true}, EvidenceReachable: true, ReachableFromMain: true}}
	if got, reasons := validateLandingEvidence(wrongPR, w); len(got) != 0 || len(reasons) == 0 {
		t.Fatalf("wrong PR binding must be unknown: %v, %v", got, reasons)
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
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: "issue", PR: 9, Branch: "issue-42", BranchSHAs: map[string]bool{"issue": true}, EvidenceReachable: true, ReachableFromMain: true}},
	}
	landings, reasons := validateLandingEvidence(e, w)
	if len(reasons) != 0 || len(landings) != 1 || !landings[0].EnqueuedAt.Equal(start.Add(-time.Hour)) {
		t.Fatalf("first enqueue lost: %v, %v", landings, reasons)
	}
	incomplete := e
	incomplete.Events = append([]landingline.Event{}, e.Events[:1]...)
	incomplete.Events = append(incomplete.Events, e.Events[2:]...)
	if got, reasons := validateLandingEvidence(incomplete, w); len(got) != 0 || len(reasons) == 0 {
		t.Fatalf("unpaired earlier enqueue must be unknown: %v, %v", got, reasons)
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
	legacy := SuiteAttempt{Legacy: true, Check: suiteCheckName, SHA: "issue-head", StartedAt: start.Add(time.Hour)}
	partial := attributeWeeklyAttempts(append(rows[:3], legacy), landings, w, nil)
	if partial.AttemptsPerLandedIssue != nil || len(partial.CoverageUnknown) == 0 {
		t.Fatalf("legacy attempt on cohort head must make coverage unknown: %+v", partial)
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
