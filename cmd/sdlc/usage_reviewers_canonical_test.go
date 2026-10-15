package main

import (
	"fmt"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func canonicalReviewerFixtures(t *testing.T) (reviewusage.Record, reviewusage.Record, reviewusage.Record, reviewusage.Record) {
	t.Helper()
	finished := usageAttributionRecord(t, "canonical", usageAttributionReviewerID, "/repo", 1, 1)
	requested, err := reviewusage.NewRequested(finished.Intent)
	if err != nil {
		t.Fatal(err)
	}
	started, err := reviewusage.Step(requested, reviewusage.Event{Kind: reviewusage.Start, Intent: requested.Intent, PID: finished.PID, At: finished.StartedAt})
	if err != nil {
		t.Fatal(err)
	}
	launchFailed, err := reviewusage.Step(requested, reviewusage.Event{Kind: reviewusage.Finish, Intent: requested.Intent, At: finished.FinishedAt, Outcome: "launch-failed"})
	if err != nil {
		t.Fatal(err)
	}
	return requested, started, finished, launchFailed
}

func assertReviewerCanonicalPair(t *testing.T, a, b reviewusage.Record, compatible bool) {
	t.Helper()
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			records := []reviewusage.Record{a, b}
			if reverse {
				records = []reviewusage.Record{b, a}
			}
			selected, bad := selectUsageReviewerRecords(records)
			if compatible {
				expected := a
				if reviewusage.Advance(a, b) == nil {
					expected = b
				} else if reviewusage.Advance(b, a) != nil {
					t.Fatal("fixture expected a legal confirmed prefix")
				}
				if bad[a.Intent.ID] || len(selected) != 1 || selected[0] != expected {
					t.Fatalf("legal prefix did not collapse to exact confirmed state: selected=%+v bad=%+v want=%+v", selected, bad, expected)
				}
				return
			}
			if !bad[a.Intent.ID] {
				t.Fatal("contradictory evidence accepted as canonical confirmation")
			}
			retained := map[reviewusage.Record]bool{}
			for _, record := range selected {
				retained[record] = true
			}
			if !retained[a] || !retained[b] {
				t.Fatalf("contradictory identity evidence disappeared: %+v", selected)
			}
			sessions := []transcripts.SessionEvidence{
				usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 12),
				usageAttributionSession(usageAttributionOtherID, "/repo", "", false, 12),
			}
			bindings := bindUsageReviewers(sessions, records, usageAttributionAt(30))
			responses := []usage.Response{
				{ID: "reported-a", Agent: "codex", SourceID: usageAttributionReviewerID, At: usageAttributionAt(15)},
				{ID: "reported-b", Agent: "codex", SourceID: usageAttributionOtherID, At: usageAttributionAt(15)},
			}
			usable := map[string]bool{"codex:" + usageAttributionReviewerID: true, "codex:" + usageAttributionOtherID: true}
			for _, record := range []reviewusage.Record{a, b} {
				if record.SessionID != "" {
					if binding, ok := bindings[record.Intent.Agent+":"+record.SessionID]; !ok || !binding.Conflict {
						t.Fatalf("contradictory session lost quarantine: record=%+v bindings=%+v", record, bindings)
					}
				}
			}
			coverage := reviewerUsageCoverage(records, bindings, responses, usable, usageAttributionAt(30), usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}})
			if coverage.SpendComplete() || coverage.Counts["unresolved_reviewer_usage"] < 1 {
				t.Fatalf("contradictory launch became confirmed non-spending/resolved: %+v", coverage)
			}
		})
	}
}

func TestUsageReviewerCanonicalPhaseAndOutcomePairs(t *testing.T) {
	requested, started, finished, launchFailed := canonicalReviewerFixtures(t)
	states := []reviewusage.Record{requested, started}
	for _, outcome := range []string{"returned", "exit-error", "agent-error", "canceled", "deadline"} {
		next := finished
		next.Outcome = outcome
		states = append(states, next)
	}
	states = append(states, launchFailed)
	for i, a := range states {
		for j := i; j < len(states); j++ {
			b := states[j]
			t.Run(fmt.Sprintf("%s-%s/%s-%s", a.Phase, a.Outcome, b.Phase, b.Outcome), func(t *testing.T) {
				if err := reviewusage.Validate(a); err != nil {
					t.Fatal(err)
				}
				if err := reviewusage.Validate(b); err != nil {
					t.Fatal(err)
				}
				// Every requested receipt can acquire either confirmed start or failed
				// launch. Started can acquire any launched terminal outcome. Different
				// terminal confirmations are immutable, even when individually valid.
				compatible := a == b || a.Phase == reviewusage.Requested || b.Phase == reviewusage.Requested || a.Phase == reviewusage.Started && b.Phase == reviewusage.Finished && b.Outcome != "launch-failed" || b.Phase == reviewusage.Started && a.Phase == reviewusage.Finished && a.Outcome != "launch-failed"
				assertReviewerCanonicalPair(t, a, b, compatible)
			})
		}
	}
}

func TestUsageReviewerCanonicalFieldRewritesRetainUncertainty(t *testing.T) {
	requested, started, finished, launchFailed := canonicalReviewerFixtures(t)
	type pair struct {
		name  string
		a, b  reviewusage.Record
		valid bool
	}
	var cases []pair
	add := func(name string, a, b reviewusage.Record, valid bool) { cases = append(cases, pair{name, a, b, valid}) }
	next := started
	next.PID++
	add("started-pid", started, next, true)
	next = started
	next.StartedAt = usageAttributionAt(11).Format("2006-01-02T15:04:05Z07:00")
	add("started-at", started, next, true)
	next = finished
	next.PID++
	add("terminal-pid", finished, next, true)
	add("started-versus-terminal-pid", started, next, true)
	next = finished
	next.StartedAt = usageAttributionAt(11).Format("2006-01-02T15:04:05Z07:00")
	add("terminal-started-at", finished, next, true)
	add("started-versus-terminal-started-at", started, next, true)
	next = finished
	next.FinishedAt = usageAttributionAt(21).Format("2006-01-02T15:04:05Z07:00")
	add("terminal-finished-at", finished, next, true)
	next = finished
	next.SessionID = usageAttributionOtherID
	add("terminal-session-id", finished, next, true)
	next = finished
	next.SessionID = ""
	add("terminal-missing-session-id", finished, next, true)
	next = finished
	next.Diagnostic = "different terminal diagnostic"
	add("terminal-diagnostic", finished, next, true)
	next = finished
	next.Intent.Root = "/other"
	add("intent-root", finished, next, true)
	next = finished
	next.Intent.Recipe = "other-review"
	add("intent-recipe", finished, next, true)
	next = finished
	next.Intent.IssueID = "000366"
	add("intent-issue", finished, next, true)
	next = finished
	next.PID = 0
	add("missing-terminal-pid", finished, next, false)
	next = finished
	next.StartedAt = ""
	add("missing-terminal-start", finished, next, false)
	next = finished
	next.FinishedAt = ""
	add("missing-terminal-finish", finished, next, false)
	next = requested
	next.Version = 99
	add("invalid-requested-copy", finished, next, false)
	add("invalid-copy-versus-nonspending", launchFailed, next, false)
	next = launchFailed
	next.PID = 123
	add("invalid-failed-launch-progress", launchFailed, next, false)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if valid := reviewusage.Validate(tc.b) == nil; valid != tc.valid {
				t.Fatalf("fixture validity=%t want%t", valid, tc.valid)
			}
			assertReviewerCanonicalPair(t, tc.a, tc.b, false)
		})
	}
}

func TestUsageReviewerConflictedNonSpendingCannotEraseSelectedIssueGap(t *testing.T) {
	_, started, _, launchFailed := canonicalReviewerFixtures(t)
	// The same dispatch ID cannot belong to two issues. The selected issue's
	// failed-launch copy remains contested even though its other copy is outside
	// this issue's diagnostic scope.
	started.Intent.IssueID = "000366"
	records := []reviewusage.Record{launchFailed, started}
	bindings := bindUsageReviewers(nil, records, usageAttributionAt(30))
	coverage := reviewerUsageCoverage(records, bindings, nil, nil, usageAttributionAt(30), usageReviewerScope{Repository: launchFailed.Intent.Repository, Issue: 365}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}})
	if coverage.SpendComplete() || coverage.Counts["unresolved_reviewer_usage"] != 1 {
		t.Fatalf("contested failed-launch copy erased selected issue uncertainty: %+v", coverage)
	}
}
