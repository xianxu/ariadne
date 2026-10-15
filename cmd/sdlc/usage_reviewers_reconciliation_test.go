package main

import (
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func TestReviewerUsageReconcilesEveryRelevantAttempt(t *testing.T) {
	for _, mode := range []string{"missing-identity", "missing-exact-source", "metadata-only", "invalid-source", "outside-attempt-response", "provisional-response", "usable-zero", "usable-positive", "requested", "started"} {
		t.Run(mode, func(t *testing.T) {
			record := usageAttributionRecord(t, "reconcile", usageAttributionReviewerID, "/repo", 1, 1)
			sessions := []transcripts.SessionEvidence{usageAttributionSession(record.SessionID, "/repo", "", false, 12)}
			response := usage.Response{ID: "reported", SourceID: record.SessionID, Agent: "codex", At: usageAttributionAt(15)}
			responses := []usage.Response{response}
			usable := map[string]bool{"codex:" + record.SessionID: true}
			switch mode {
			case "missing-identity":
				record.SessionID = ""
				sessions = nil
				responses = nil
			case "missing-exact-source":
				sessions = nil
				responses = nil
			case "metadata-only":
				responses = nil
			case "invalid-source":
				usable["codex:"+record.SessionID] = false
			case "outside-attempt-response":
				responses[0].At = usageAttributionAt(25)
			case "provisional-response":
				responses[0].Provisional = true
			case "usable-positive":
				responses[0].Tokens = usage.Counters{Input: 17, Output: 3}
			case "requested":
				var err error
				record, err = reviewusage.NewRequested(record.Intent)
				if err != nil {
					t.Fatal(err)
				}
				sessions = nil
				responses = nil
			case "started":
				record.Phase = reviewusage.Started
				record.SessionID = ""
				record.FinishedAt = ""
				record.Outcome = ""
				sessions = nil
				responses = nil
			}
			if err := reviewusage.Validate(record); err != nil {
				t.Fatal(err)
			}
			bindings := bindUsageReviewers(sessions, []reviewusage.Record{record}, usageAttributionAt(30))
			coverage := reviewerUsageCoverage([]reviewusage.Record{record}, bindings, responses, usable, usageAttributionAt(30), usageReviewerScope{Repository: record.Intent.Repository, Issue: 365}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}})
			resolved := mode == "usable-zero" || mode == "usable-positive"
			if resolved {
				if !coverage.Complete() || coverage.Counts["unresolved_reviewer_usage"] != 0 {
					t.Fatalf("usable reported evidence unresolved: %+v", coverage)
				}
			} else if coverage.Complete() || coverage.SpendComplete() || coverage.Counts["unresolved_reviewer_usage"] != 1 {
				t.Fatalf("known possible spend silently complete: %+v", coverage)
			}
		})
	}
}

func TestReviewerUsageConfirmedLaunchFailureIsNonSpending(t *testing.T) {
	record := usageAttributionRecord(t, "not-launched", "", "/repo", 1, 1)
	record.PID = 0
	record.StartedAt = ""
	record.Outcome = "launch-failed"
	record.Intent.DisposableRoot = true
	if err := reviewusage.Validate(record); err != nil {
		t.Fatal(err)
	}
	sessions := []transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 12)}
	bindings := bindUsageReviewers(sessions, []reviewusage.Record{record}, usageAttributionAt(30))
	if len(bindings) != 0 {
		t.Fatalf("failed launch claimed a worker session: %+v", bindings)
	}
	coverage := reviewerUsageCoverage([]reviewusage.Record{record}, bindings, nil, nil, usageAttributionAt(30), usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}})
	if !coverage.Complete() {
		t.Fatalf("confirmed non-spending launch treated as missing spend: %+v", coverage)
	}
}

func TestReviewerUsageCanonicalCopiesResolveButInvalidCopiesStayUncertain(t *testing.T) {
	record := usageAttributionRecord(t, "prefix", usageAttributionReviewerID, "/repo", 1, 1)
	requested, err := reviewusage.NewRequested(record.Intent)
	if err != nil {
		t.Fatal(err)
	}
	session := usageAttributionSession(record.SessionID, "/repo", "", false, 12)
	response := usage.Response{ID: "reported", Agent: "codex", SourceID: record.SessionID, At: usageAttributionAt(15)}
	usable := map[string]bool{"codex:" + record.SessionID: true}
	invalid := requested
	invalid.Version = 99
	conflict := record
	conflict.SessionID = usageAttributionOtherID
	for _, tc := range []struct {
		name    string
		records []reviewusage.Record
		gap     bool
	}{
		{"phase-prefix", []reviewusage.Record{requested, record, requested, record}, false},
		{"invalid-copy", []reviewusage.Record{requested, record, invalid}, true},
		{"conflicting-copy", []reviewusage.Record{requested, record, conflict}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bindings := bindUsageReviewers([]transcripts.SessionEvidence{session}, tc.records, usageAttributionAt(30))
			coverage := reviewerUsageCoverage(tc.records, bindings, []usage.Response{response}, usable, usageAttributionAt(30), usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}})
			if tc.gap {
				if coverage.SpendComplete() || coverage.Counts["unresolved_reviewer_usage"] != 1 {
					t.Fatalf("invalid/conflicting copy lost uncertainty: %+v", coverage)
				}
			} else if !coverage.Complete() || coverage.Counts["unfinished_reviewer_bindings"] != 0 {
				t.Fatalf("phase-prefix copies made a completed attempt unfinished: %+v", coverage)
			}
		})
	}
}

func TestReviewerUsageGapScopeUsesRepositoryIssueAgentAndWindowUnion(t *testing.T) {
	record := usageAttributionRecord(t, "scope", usageAttributionReviewerID, "/repo", 1, 1)
	for _, tc := range []struct {
		name    string
		scope   usageReviewerScope
		windows []usage.Window
		gap     bool
	}{
		{"selected-issue", usageReviewerScope{Repository: record.Intent.Repository, Issue: 365}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}}, true},
		{"other-repository", usageReviewerScope{Repository: "github.com/example/other", Issue: 365}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}}, false},
		{"other-issue", usageReviewerScope{Repository: record.Intent.Repository, Issue: 366}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}}, false},
		{"other-agent", usageReviewerScope{Agent: "claude"}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}}, false},
		{"burn-gap", usageReviewerScope{Agent: "codex"}, []usage.Window{{Since: usageAttributionAt(25), Until: usageAttributionAt(30)}}, false},
		{"meter-window-union", usageReviewerScope{Agent: "codex"}, []usage.Window{{Since: usageAttributionAt(25), Until: usageAttributionAt(30)}, {Since: usageAttributionAt(5), Until: usageAttributionAt(18)}}, true},
		{"between-windows", usageReviewerScope{Agent: "codex"}, []usage.Window{{Since: usageAttributionAt(25), Until: usageAttributionAt(30)}, {Since: usageAttributionAt(0), Until: usageAttributionAt(5)}}, false},
		{"right-edge", usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(0), Until: usageAttributionAt(10)}}, false},
		{"left-edge", usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(20), Until: usageAttributionAt(30)}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coverage := reviewerUsageCoverage([]reviewusage.Record{record}, nil, nil, nil, usageAttributionAt(30), tc.scope, tc.windows)
			if gap := coverage.Counts["unresolved_reviewer_usage"] == 1; gap != tc.gap {
				t.Fatalf("scope/window mismatch: %+v", coverage)
			}
		})
	}
	// A malformed timestamp cannot prove the dispatch happened outside the window.
	record.Intent.RequestedAt = "invalid"
	coverage := reviewerUsageCoverage([]reviewusage.Record{record}, nil, nil, nil, usageAttributionAt(30), usageReviewerScope{}, []usage.Window{{Since: usageAttributionAt(25), Until: usageAttributionAt(30)}})
	if coverage.SpendComplete() {
		t.Fatalf("invalid timestamp erased uncertain spend: %+v", coverage)
	}
}
