package main

import (
	"fmt"
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

func TestValidateLandingEvidenceRejectsEveryCohortIdentityCollision(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	mergeA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mergeB := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	makeEvidence := func() weeklyRepoEvidence {
		return weeklyRepoEvidence{
			Events: []landingline.Event{
				{Kind: landingline.EventEnqueue, Seq: 7, Issue: 42, Branch: "issue-42", At: start.Add(time.Hour)},
				{Kind: landingline.EventLanded, Seq: 7, Issue: 42, Branch: "issue-42", PR: 9, At: start.Add(2 * time.Hour)},
				{Kind: landingline.EventEnqueue, Seq: 8, Issue: 43, Branch: "issue-43", At: start.Add(time.Hour)},
				{Kind: landingline.EventLanded, Seq: 8, Issue: 43, Branch: "issue-43", PR: 10, At: start.Add(3 * time.Hour)},
			},
			Cards: map[int]weeklyCardEvidence{
				42: {Status: "done", Completion: &issue.Completion{LandedCommit: mergeA}},
				43: {Status: "done", Completion: &issue.Completion{LandedCommit: mergeB}},
			},
			Graphs: map[string]weeklyMergeGraph{
				mergeA: {MergeSHA: mergeA, MainParent: "main", IssueParent: "head-a", PR: 9, Branch: "issue-42", BranchSHAs: map[string]bool{"head-a": true}, EvidenceReachable: true, ReachableFromMain: true},
				mergeB: {MergeSHA: mergeB, MainParent: "main", IssueParent: "head-b", PR: 10, Branch: "issue-43", BranchSHAs: map[string]bool{"head-b": true}, EvidenceReachable: true, ReachableFromMain: true},
			},
		}
	}
	if got, reasons := validateLandingEvidence(makeEvidence(), w); len(got) != 2 || len(reasons) != 0 {
		t.Fatalf("independent landings should be accepted: %v, %v", got, reasons)
	}
	for _, tc := range []struct {
		name   string
		change func(*weeklyRepoEvidence)
	}{
		{"issue", func(e *weeklyRepoEvidence) { e.Events[2].Issue, e.Events[3].Issue = 42, 42 }},
		{"ticket", func(e *weeklyRepoEvidence) { e.Events[2].Seq, e.Events[3].Seq = 7, 7 }},
		{"branch", func(e *weeklyRepoEvidence) {
			e.Events[2].Branch, e.Events[3].Branch = "issue-42", "issue-42"
			graph := e.Graphs[mergeB]
			graph.Branch = "issue-42"
			e.Graphs[mergeB] = graph
		}},
		{"PR", func(e *weeklyRepoEvidence) {
			e.Events[3].PR = 9
			graph := e.Graphs[mergeB]
			graph.PR = 9
			e.Graphs[mergeB] = graph
		}},
		{"merge", func(e *weeklyRepoEvidence) {
			e.Cards[43] = weeklyCardEvidence{Status: "done", Completion: &issue.Completion{LandedCommit: mergeA}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				e := makeEvidence()
				tc.change(&e)
				if reverse {
					for left, right := 0, len(e.Events)-1; left < right; left, right = left+1, right-1 {
						e.Events[left], e.Events[right] = e.Events[right], e.Events[left]
					}
				}
				got, reasons := validateLandingEvidence(e, w)
				if len(got) != 0 || strings.Count(strings.Join(reasons, "\n"), "cohort landing identity") != 2 {
					t.Fatalf("reverse=%t: shared %s identity must reject both claimants: %v, %v", reverse, tc.name, got, reasons)
				}
				if attribution := attributeWeeklyAttempts(nil, got, w, reasons); attribution.AttemptsPerLandedIssue != nil {
					t.Fatalf("reverse=%t: shared %s identity certified a numeric ratio: %+v", reverse, tc.name, attribution)
				}
			}
		})
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

func TestValidateLandingEvidenceChecksEveryPriorAdmission(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	merge := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	events := []landingline.Event{
		{Kind: landingline.EventEnqueue, Seq: 1, Issue: 42, Branch: "issue-42", At: start.Add(time.Hour)},
		{Kind: landingline.EventRefused, Seq: 1, Issue: 42, Branch: "issue-42", At: start.Add(90 * time.Minute)},
		{Kind: landingline.EventEnqueue, Seq: 2, Issue: 42, Branch: "issue-42", At: start.Add(2 * time.Hour)}, // terminal missing
		{Kind: landingline.EventEnqueue, Seq: 3, Issue: 42, Branch: "issue-42", At: start.Add(3 * time.Hour)},
		{Kind: landingline.EventLanded, Seq: 3, Issue: 42, Branch: "issue-42", PR: 9, At: start.Add(4 * time.Hour)},
	}
	base := weeklyRepoEvidence{
		Cards:  map[int]weeklyCardEvidence{42: {Status: "done", Completion: &issue.Completion{LandedCommit: merge}}},
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: "issue", PR: 9, Branch: "issue-42", BranchSHAs: map[string]bool{"issue": true}, EvidenceReachable: true, ReachableFromMain: true}},
	}
	for _, order := range [][]landingline.Event{events, {events[2], events[4], events[0], events[3], events[1]}} {
		base.Events = order
		if got, reasons := validateLandingEvidence(base, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "prior enqueue") {
			t.Fatalf("intermediate unrefused ticket escaped validation: %v, %v", got, reasons)
		}
	}
}

func TestLegacyWeeklyAttemptUsesPhysicalCompletionForReconciliation(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	legacy := SuiteAttempt{Legacy: true, Check: suiteCheckName, Repo: "ariadne", SHA: "old", StartedAt: start.Add(time.Hour)}
	physical := aggregateWeeklyRepo("ariadne", []SuiteAttempt{legacy}, w)
	reconciled := attributeWeeklyAttempts([]SuiteAttempt{legacy}, nil, w, nil)
	if physical.Suite.Runs != 1 || reconciled.UnassignableAttempts != 1 || len(reconciled.UnknownAttempts) != 1 {
		t.Fatalf("legacy physical/reconciliation mismatch: runs=%d attribution=%+v", physical.Suite.Runs, reconciled)
	}
}

func TestValidateLandingEvidenceRequiresRetainedWholeIssueHistory(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	merge := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	base := weeklyRepoEvidence{
		CollectedAt: start.Add(8 * 24 * time.Hour),
		Attempts:    []SuiteAttempt{{ID: "older", Check: suiteCheckName, FinishedAt: start.Add(-2 * 24 * time.Hour)}},
		Events: []landingline.Event{
			{Kind: landingline.EventEnqueue, Seq: 1, Issue: 9, Branch: "older", At: start.Add(-2 * 24 * time.Hour)},
			{Kind: landingline.EventEnqueue, Seq: 2, Issue: 42, Branch: "issue-42", At: start.Add(time.Hour)},
			{Kind: landingline.EventLanded, Seq: 2, Issue: 42, Branch: "issue-42", PR: 9, At: start.Add(2 * time.Hour)},
		},
		Cards:  map[int]weeklyCardEvidence{42: {Status: "done", CreatedAt: start, Completion: &issue.Completion{LandedCommit: merge}}},
		Graphs: map[string]weeklyMergeGraph{merge: {MergeSHA: merge, MainParent: "main", IssueParent: "issue", PR: 9, Branch: "issue-42", BranchSHAs: map[string]bool{"issue": true}, EvidenceReachable: true, ReachableFromMain: true}},
	}
	if got, reasons := validateLandingEvidence(base, w); len(got) != 1 || len(reasons) != 0 {
		t.Fatalf("fresh issue with earlier event history rejected: %v, %v", got, reasons)
	}
	expired := base
	expired.CollectedAt = start.Add(40 * 24 * time.Hour)
	if got, reasons := validateLandingEvidence(expired, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "history") {
		t.Fatalf("expired attempts/admissions certified: %v, %v", got, reasons)
	}
	truncated := base
	truncated.Events = append([]landingline.Event{}, base.Events[1:]...)
	if got, reasons := validateLandingEvidence(truncated, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "history") {
		t.Fatalf("issue older than retained event log certified: %v, %v", got, reasons)
	}
	missing := base
	missing.Cards = map[int]weeklyCardEvidence{42: {Status: "done", Completion: &issue.Completion{LandedCommit: merge}}}
	if got, reasons := validateLandingEvidence(missing, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "history") {
		t.Fatalf("missing creation date certified: %v, %v", got, reasons)
	}
	noRecorder := base
	noRecorder.Attempts = nil
	if got, reasons := validateLandingEvidence(noRecorder, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "history") {
		t.Fatalf("source with no structured attempt history certified: %v, %v", got, reasons)
	}
	lateRecorder := base
	lateRecorder.Attempts = []SuiteAttempt{{ID: "first", Check: suiteCheckName, FinishedAt: start.Add(24 * time.Hour)}}
	if got, reasons := validateLandingEvidence(lateRecorder, w); len(got) != 0 || !strings.Contains(strings.Join(reasons, "\n"), "history") {
		t.Fatalf("issue predating structured recorder certified: %v, %v", got, reasons)
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

func TestAttributeWeeklyAttemptsRejectsAmbiguousOutsideWeekCohortAttempt(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	landings := []weeklyLandingEvidence{
		{Issue: 42, EnqueuedAt: start.Add(-48 * time.Hour), LandedAt: start.Add(time.Hour), BranchSHAs: map[string]bool{"shared-head": true}},
		{Issue: 43, EnqueuedAt: start.Add(-48 * time.Hour), LandedAt: start.Add(2 * time.Hour), BranchSHAs: map[string]bool{"shared-head": true}},
	}
	attempt := SuiteAttempt{ID: "cross-week", Check: suiteCheckName, SHA: "shared-head", FinishedAt: start.Add(-time.Hour)}
	got := attributeWeeklyAttempts([]SuiteAttempt{attempt}, landings, w, nil)
	if got.AttemptsPerLandedIssue != nil || !strings.Contains(strings.Join(got.CoverageUnknown, "\n"), "ambiguous branch ancestry") {
		t.Fatalf("outside-week ambiguous cohort attempt certified: %+v", got)
	}
}

func TestAttributeWeeklyAttemptsNeverCountsOneAttemptForTwoLandings(t *testing.T) {
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	w := WeeklyWindow{Start: start, End: start.Add(7 * 24 * time.Hour)}
	for count := 1; count <= 4; count++ {
		for _, when := range []time.Time{start.Add(-time.Hour), start.Add(time.Hour)} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("claimants=%d/inside=%t/reverse=%t", count, w.Contains(when), reverse), func(t *testing.T) {
					landings := make([]weeklyLandingEvidence, count)
					for i := range landings {
						landings[i] = weeklyLandingEvidence{Issue: 42 + i, EnqueuedAt: start.Add(-24 * time.Hour), LandedAt: start.Add(2 * time.Hour), BranchSHAs: map[string]bool{"same-head": true}}
					}
					if reverse {
						for left, right := 0, len(landings)-1; left < right; left, right = left+1, right-1 {
							landings[left], landings[right] = landings[right], landings[left]
						}
					}
					got := attributeWeeklyAttempts([]SuiteAttempt{{ID: "one", Check: suiteCheckName, SHA: "same-head", FinishedAt: when}}, landings, w, nil)
					if got.AttributedAttempts > 1 {
						t.Fatalf("one attempt counted twice: %+v", got)
					}
					if count == 1 {
						if got.AttributedAttempts != 1 || got.AttemptsPerLandedIssue == nil || *got.AttemptsPerLandedIssue != 1 {
							t.Fatalf("unique attempt not counted once: %+v", got)
						}
					} else if got.AttributedAttempts != 0 || got.AttemptsPerLandedIssue != nil || !strings.Contains(strings.Join(got.CoverageUnknown, "\n"), "ambiguous branch ancestry") {
						t.Fatalf("ambiguous attempt certified: %+v", got)
					}
				})
			}
		}
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
