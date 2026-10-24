package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/vocab"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
)

func metricsFixture() benchProtocolWindow {
	since, _ := time.Parse(time.RFC3339, "2026-10-23T10:00:00Z")
	return benchProtocolWindow{Since: since, Until: since.Add(time.Minute), RunID: "run", Repository: "private-origin", Issue: "000577", WorkerRoot: "/private/work", Hub: bench.RuntimeIdentity{Root: "/private/hub", Slot: "repo:0", Session: "session", Generation: "generation"}}
}
func metricsPair(id string, route ...string) []benchAuditEvent {
	w := metricsFixture()
	failed := false
	start := benchAuditEvent{ID: id, RunID: w.RunID, Kind: "command-start", At: w.Since.Add(time.Second).Format(time.RFC3339Nano), Root: w.WorkerRoot, Command: []string{"alias", "--close"}, Route: route}
	end := start
	end.Kind = "command-end"
	end.ID = id + "-end"
	end.ParentID = id
	end.Success = &failed
	end.At = w.Since.Add(2 * time.Second).Format(time.RFC3339Nano)
	return []benchAuditEvent{start, end}
}
func TestBenchProtocolMetricsTypedPairs(t *testing.T) {
	events := append(metricsPair("gate", "close"), metricsPair("nongate", "issue", "show")...)
	w := metricsFixture()
	events = append(events, benchAuditEvent{ID: "decision", RunID: w.RunID, Kind: "decision", At: w.Since.Add(time.Second).Format(time.RFC3339Nano), Root: w.Hub.Root, Slot: w.Hub.Slot, Session: w.Hub.Session, Generation: w.Hub.Generation, DecisionKind: "replan", Reason: "scope"})
	var m bench.Measurements
	collectBenchAuditMetrics(events, w, &m)
	if m.GateRefusals == nil || *m.GateRefusals != 1 || len(m.GateRefusalIDs) != 1 || len(m.Decisions) != 1 || m.Decisions[0].Session != w.Hub.Session {
		t.Fatalf("%+v", m)
	}
}
func TestBenchProtocolMetricsUnknownPairsAndBinding(t *testing.T) {
	for _, events := range [][]benchAuditEvent{metricsPair("old"), metricsPair("unfinished", "close")[:1], func() []benchAuditEvent { e := metricsPair("mismatch", "close"); e[1].Root = "/other"; return e }()} {
		m := bench.Measurements{Validity: "invalid", ValidityIssues: []string{"profile: mismatch"}}
		collectBenchAuditMetrics(events, metricsFixture(), &m)
		if m.GateRefusals != nil || m.Validity != "invalid" || len(m.ValidityIssues) < 2 {
			t.Fatalf("%+v", m)
		}
	}
	w := metricsFixture()
	e := benchAuditEvent{ID: "decision", Kind: "decision", RunID: w.RunID, Root: w.Hub.Root, Slot: w.Hub.Slot, Session: "replacement", Generation: w.Hub.Generation, At: w.Since.Format(time.RFC3339Nano), DecisionKind: "approve", Reason: "scope"}
	var m bench.Measurements
	collectBenchAuditMetrics([]benchAuditEvent{e}, w, &m)
	if len(m.Decisions) != 0 || len(m.ValidityIssues) == 0 {
		t.Fatalf("%+v", m)
	}
}
func metricsReview(id string, attempt int) reviewusage.Record {
	w := metricsFixture()
	return reviewusage.Record{Version: 1, Intent: reviewusage.Intent{ID: id, Repository: w.Repository, IssueID: w.Issue, Recipe: "estimate-quality", Round: 1, Attempt: attempt, Agent: "codex", Root: w.WorkerRoot, ReviewedHEAD: strings.Repeat("a", 40), Slot: "repo:1", RequestedAt: w.Since.Format(time.RFC3339Nano)}, Phase: reviewusage.Started, PID: 1, StartedAt: w.Since.Add(time.Second).Format(time.RFC3339Nano)}
}
func TestBenchProtocolMetricsNativeReviews(t *testing.T) {
	w := metricsFixture()
	r := metricsReview("attached", 1)
	a := metricsReview("estimate", 1)
	b := metricsReview("estimate", 2)
	l := gatestate.Ledger{Gate: "plan-quality", IssueNum: 577, IDPrefix: "PQ", Rounds: []gatestate.Round{{N: 1, Timestamp: w.Since.Add(3 * time.Second).Format(time.RFC3339Nano), ReviewerRuns: []reviewusage.Record{r}, New: []gatestate.Finding{{ID: "PQ-1"}, {ID: "PQ-2"}}}, {N: 2, Timestamp: w.Since.Add(4 * time.Second).Format(time.RFC3339Nano), ProtocolError: "interrupted", Dispositions: []gatestate.Disposition{{ID: "PQ-1", State: "fixed"}}}, {N: 3, NoCap: true, New: []gatestate.Finding{{ID: "seed"}}}}, ReviewerRuns: []reviewusage.Record{a, b}}
	var m bench.Measurements
	collectBenchReviewMetrics([]gatestate.Ledger{l}, []reviewusage.Record{r, a, b}, w, &m)
	if m.ReviewRounds == nil || *m.ReviewRounds != 3 || m.ReviewFindings != nil || len(m.ReviewFindingIDs) != 2 {
		t.Fatalf("%+v", m)
	}
	l.ReviewerRuns = nil
	collectBenchReviewMetrics([]gatestate.Ledger{l}, []reviewusage.Record{r}, w, &m)
	if m.ReviewRounds == nil || *m.ReviewRounds != 2 || m.ReviewFindings == nil || *m.ReviewFindings != 2 {
		t.Fatalf("native findings lost: %+v", m)
	}
}
func TestBenchProtocolMetricsUnfinishedReviewFindingsUnknown(t *testing.T) {
	var m bench.Measurements
	collectBenchReviewMetrics(nil, []reviewusage.Record{metricsReview("interrupted", 1)}, metricsFixture(), &m)
	if m.ReviewRounds == nil || *m.ReviewRounds != 1 || m.ReviewFindings != nil || len(m.ValidityIssues) == 0 {
		t.Fatalf("%+v", m)
	}
}

func TestBenchProtocolMetricsRetainedNamespace(t *testing.T) {
	ctx := context.Background()
	s, err := prepareFixture(t, preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	rec, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	w := metricsFixture()
	w.RunID = rec.ID
	w.Issue = rec.Task.ID
	w.WorkerRoot = filepath.Join(s.root, "public", "work")
	w.Hub.Root = filepath.Join(s.root, "public", "hub")
	w.Hub.Agent = rec.Task.TL.Agent
	w.Hub.Model = rec.Task.TL.Model
	w.Hub.Effort = rec.Task.TL.Effort
	worker := w.Hub
	worker.Root = w.WorkerRoot
	worker.Slot = "repo:1"
	worker.Session = "worker-session"
	for i, kind := range []bench.EventKind{bench.BeginDispatch, bench.DispatchConfirmed, bench.Stop, bench.CancelConfirmed} {
		op := "dispatch"
		if i > 1 {
			op = "stop"
		}
		if _, _, err := s.apply(bench.Event{ID: fmtMetricIndex(i), RunID: rec.ID, OperationID: op, Kind: kind, Hub: w.Hub, Worker: worker, Reason: "synthetic fixture"}); err != nil {
			t.Fatal(err)
		}
	}
	runtime := benchRuntimeState{Version: 1, RunID: rec.ID, Hub: benchRuntimeActor{Identity: w.Hub}, Worker: benchRuntimeActor{Identity: worker}, Since: w.Since.Format(time.RFC3339Nano), Deadline: w.Until.Format(time.RFC3339Nano), Provision: json.RawMessage(`{}`)}
	retirement := benchRetirement{Version: 1, RunID: rec.ID, Hub: w.Hub, Worker: worker, ConfirmedBy: "fixture", ConfirmedAt: w.Until.Format(time.RFC3339Nano), Evidence: "synthetic fixture", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	if err := s.putData("runtime", runtime); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("retirement", retirement); err != nil {
		t.Fatal(err)
	}
	c, err := loadBenchContext(ctx, w.WorkerRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range metricsPair("retained", "local-check", "run") {
		e.Version = 1
		e.RunID = rec.ID
		e.Root = w.WorkerRoot
		if err := appendBenchAudit(ctx, c, e); err != nil {
			t.Fatal(err)
		}
	}
	env, err := openTrackerAt(ctx, w.WorkerRoot)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snapshot.Require(rec.Task.ID)
	if err != nil {
		t.Fatal(err)
	}
	l := gatestate.Ledger{Gate: "plan-quality", IssueNum: 1, IDPrefix: "PQ", Rounds: []gatestate.Round{{N: 1, Timestamp: w.Since.Add(3 * time.Second).Format(time.RFC3339Nano), New: []gatestate.Finding{{ID: "PQ-1", Severity: "Important", Title: "fixture"}}}}}
	path := filepath.Join(w.WorkerRoot, planGateKind.path(vocab.Issue().Discovery().Plans, filepath.Base(card.Path)))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(gatestate.Render(l, "fixture")), 0600); err != nil {
		t.Fatal(err)
	}
	var m bench.Measurements
	if err := collectBenchProtocolMetrics(ctx, s, &m); err != nil {
		t.Fatal(err)
	}
	if m.GateRefusals == nil || *m.GateRefusals != 1 || m.ReviewRounds == nil || *m.ReviewRounds != 1 || m.ReviewFindings == nil || *m.ReviewFindings != 1 {
		t.Fatalf("%+v", m)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := collectBenchProtocolMetrics(ctx, s, &m); err != nil {
		t.Fatal(err)
	}
	if m.ReviewRounds != nil || m.ReviewFindings != nil || m.Validity != "unverified" {
		t.Fatalf("%+v", m)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := collectBenchProtocolMetrics(canceled, s, &m); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
func fmtMetricIndex(i int) string {
	return []string{"metric-dispatch", "metric-dispatched", "metric-stop", "metric-stopped"}[i]
}

func TestBenchProtocolMetricsSeparateRetryIdentityUnknown(t *testing.T) {
	a, b := metricsReview("attempt-one", 1), metricsReview("attempt-two", 2)
	var m bench.Measurements
	collectBenchReviewMetrics(nil, []reviewusage.Record{a, b}, metricsFixture(), &m)
	if m.ReviewRounds != nil || len(m.ReviewIDs) != 2 || m.ReviewFindings != nil {
		t.Fatalf("unproved retry grouping: %+v", m)
	}
	b = a
	b.Intent.ReviewedHEAD = strings.Repeat("b", 40)
	collectBenchReviewMetrics(nil, []reviewusage.Record{a, b}, metricsFixture(), &m)
	if m.ReviewRounds != nil {
		t.Fatalf("conflicting native identity accepted: %+v", m)
	}
}
func TestBenchProtocolMetricsTimeWindowAndSuccess(t *testing.T) {
	w := metricsFixture()
	events := metricsPair("successful", "close")
	passed := true
	events[1].Success = &passed
	outside := metricsPair("outside", "close")
	for i := range outside {
		outside[i].At = w.Since.Add(-time.Minute).Format(time.RFC3339Nano)
	}
	events = append(events, outside...)
	var m bench.Measurements
	collectBenchAuditMetrics(events, w, &m)
	if m.GateRefusals == nil || *m.GateRefusals != 0 {
		t.Fatalf("outside or successful command counted: %+v", m)
	}
	events[1].At = w.Until.Add(time.Second).Format(time.RFC3339Nano)
	collectBenchAuditMetrics(events, w, &m)
	if m.GateRefusals != nil {
		t.Fatalf("cross-window command called complete: %+v", m)
	}
}
