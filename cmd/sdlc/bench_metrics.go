package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type benchProtocolWindow struct {
	Since, Until                         time.Time
	RunID, Repository, Issue, WorkerRoot string
	Hub                                  bench.RuntimeIdentity
}

func (w benchProtocolWindow) contains(raw string) bool {
	at, err := time.Parse(time.RFC3339Nano, raw)
	return err == nil && !at.Before(w.Since) && !at.After(w.Until)
}
func benchMetricUnknown(m *bench.Measurements, source, reason string) {
	issue := source + ": " + reason
	for _, old := range m.ValidityIssues {
		if old == issue {
			return
		}
	}
	m.ValidityIssues = append(m.ValidityIssues, issue)
	if m.Validity != "invalid" {
		m.Validity = "unverified"
	}
}
func benchMetricBefore(a, b string) bool {
	x, xe := time.Parse(time.RFC3339Nano, a)
	y, ye := time.Parse(time.RFC3339Nano, b)
	return xe != nil || ye != nil || x.Before(y)
}
func benchMetricCount(ids []string) *int { n := len(ids); return &n }
func benchWorkflowGate(route []string) bool {
	switch strings.Join(route, " ") {
	case "claim", "start-plan", "change-code", "close", "milestone-close", "local-check run", "local-check republish", "local-check resume", "local-check reconcile", "issue publish":
		return true
	}
	return false
}

// collectBenchAuditMetrics uses Cobra's canonical route, never literal flag or
// alias guesses. A refusal is a failed, completely paired native gate command.
func collectBenchAuditMetrics(events []benchAuditEvent, w benchProtocolWindow, m *bench.Measurements) {
	m.GateRefusals = nil
	m.GateRefusalIDs = nil
	m.Decisions = nil
	starts := map[string]benchAuditEvent{}
	ends := map[string]benchAuditEvent{}
	complete := true
	for _, e := range events {
		if !w.contains(e.At) {
			continue
		}
		if e.RunID != w.RunID || (e.Root != w.WorkerRoot && e.Root != w.Hub.Root) {
			complete = false
			continue
		}
		switch e.Kind {
		case "command-start":
			if _, ok := starts[e.ID]; ok {
				complete = false
			}
			starts[e.ID] = e
		case "command-end":
			if _, ok := ends[e.ParentID]; ok {
				complete = false
			}
			ends[e.ParentID] = e
		case "decision":
			if e.Root != w.Hub.Root || e.Slot != w.Hub.Slot || e.Session != w.Hub.Session || e.Generation != w.Hub.Generation || strings.TrimSpace(e.Reason) == "" {
				benchMetricUnknown(m, "audit-decisions", "unbound TL decision omitted")
				continue
			}
			switch e.DecisionKind {
			case "approve", "replan", "ruling", "stop":
			default:
				benchMetricUnknown(m, "audit-decisions", "unknown TL decision category omitted")
				continue
			}
			m.Decisions = append(m.Decisions, bench.TLDecision{ID: "audit:" + e.ID, Kind: e.DecisionKind, Reason: e.Reason, Session: e.Session, Generation: e.Generation})
		}
	}
	for id, start := range starts {
		end, ok := ends[id]
		if !ok || len(start.Route) == 0 || len(end.Route) == 0 || !reflect.DeepEqual(start.Route, end.Route) || !reflect.DeepEqual(start.Command, end.Command) || end.Root != start.Root || end.Success == nil || end.ID != start.ID+"-end" || benchMetricBefore(end.At, start.At) {
			complete = false
			continue
		}
		if benchWorkflowGate(start.Route) && !*end.Success {
			m.GateRefusalIDs = append(m.GateRefusalIDs, "audit:"+id)
		}
	}
	for id := range ends {
		if _, ok := starts[id]; !ok {
			complete = false
		}
	}
	sort.Strings(m.GateRefusalIDs)
	if complete {
		m.GateRefusals = benchMetricCount(m.GateRefusalIDs)
	} else {
		benchMetricUnknown(m, "audit-gates", "missing canonical route or incomplete command pair")
	}
}

// Native rounds are counted even after demotion or interruption. Accounting
// receipts supplement rounds but provider attempts sharing an intent count once.
func collectBenchReviewMetrics(ledgers []gatestate.Ledger, records []reviewusage.Record, w benchProtocolWindow, m *bench.Measurements) {
	m.ReviewRounds = nil
	m.ReviewFindings = nil
	m.ReviewIDs = nil
	m.ReviewFindingIDs = nil
	attached := map[string]bool{}
	roundIDs := map[string]bool{}
	findingIDs := map[string]bool{}
	actual := map[string]bool{}
	retryIDs := map[string]bool{}
	firstIDs := map[string]bool{}
	identities := map[string]reviewusage.Intent{}
	roundsKnown, findingsKnown := true, true
	for _, l := range ledgers {
		for _, r := range l.Rounds {
			records = append(records, r.ReviewerRuns...)
			if r.NoCap {
				for _, a := range r.ReviewerRuns {
					attached[a.Intent.ID] = true
				}
				continue
			}
			if !w.contains(r.Timestamp) {
				roundsKnown = false
				findingsKnown = false
				continue
			}
			id := fmt.Sprintf("gate:%s:%d:%s", l.IDPrefix, r.N, r.Timestamp)
			if r.N < 1 || roundIDs[id] {
				roundsKnown = false
			}
			roundIDs[id] = true
			for _, a := range r.ReviewerRuns {
				attached[a.Intent.ID] = true
			}
			for _, f := range r.New {
				id := "finding:" + l.IDPrefix + ":" + f.ID
				if f.ID == "" || len(id) > 256 || findingIDs[id] {
					findingsKnown = false
					continue
				}
				findingIDs[id] = true
			}
		}
		records = append(records, l.ReviewerRuns...)
	}
	for _, r := range records {
		if reviewusage.Validate(r) != nil {
			roundsKnown = false
			findingsKnown = false
			continue
		}
		i := r.Intent
		identity := i
		identity.Attempt = 0
		if previous, ok := identities[i.ID]; ok && previous != identity {
			roundsKnown = false
			findingsKnown = false
			continue
		}
		identities[i.ID] = identity
		if i.Repository != w.Repository || i.IssueID != w.Issue {
			roundsKnown = false
			findingsKnown = false
			continue
		}
		if !w.contains(i.RequestedAt) {
			continue
		}
		if r.Phase == reviewusage.Requested {
			if !attached[i.ID] {
				roundsKnown = false
				findingsKnown = false
			}
			continue
		}
		if r.StartedAt == "" {
			continue
		} // A proved launch failure never consumed an actual review.
		if !w.contains(r.StartedAt) || (r.FinishedAt != "" && !w.contains(r.FinishedAt)) {
			roundsKnown = false
			findingsKnown = false
			continue
		}
		actual[i.ID] = true
		if i.Attempt > 1 {
			retryIDs[i.ID] = true
		} else {
			firstIDs[i.ID] = true
		}
	}
	for id := range actual {
		if !attached[id] {
			roundIDs["review:"+id] = true
			if retryIDs[id] && !firstIDs[id] { /* Separate native IDs cannot prove logical retry grouping. */
				roundsKnown = false
			}
			findingsKnown = false
		}
	}
	for id := range roundIDs {
		m.ReviewIDs = append(m.ReviewIDs, id)
	}
	sort.Strings(m.ReviewIDs)
	for id := range findingIDs {
		m.ReviewFindingIDs = append(m.ReviewFindingIDs, id)
	}
	sort.Strings(m.ReviewFindingIDs)
	if roundsKnown {
		m.ReviewRounds = benchMetricCount(m.ReviewIDs)
	} else {
		benchMetricUnknown(m, "native-reviews", "incomplete round attribution")
	}
	if findingsKnown {
		m.ReviewFindings = benchMetricCount(m.ReviewFindingIDs)
	} else {
		benchMetricUnknown(m, "native-findings", "unledgered or incomplete review findings")
	}
}

// collectBenchProtocolMetrics reads only the run's explicitly bound fresh
// namespace. Missing evidence degrades coverage; cancellation and invalid runtime
// authority are returned to the observer instead of manufacturing measurements.
func collectBenchProtocolMetrics(ctx context.Context, s *benchStore, m *bench.Measurements) error {
	if ctx == nil || s == nil || m == nil {
		return errors.New("bench metrics needs context, store and measurements")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r, run, err := s.read()
	if err != nil {
		return err
	}
	runtime, err := loadBenchRuntime(r, run)
	if err != nil {
		return err
	}
	var retirement benchRetirement
	if err := decodeBenchRequired(r.Data["retirement"], &retirement, "version", "run_id", "hub", "worker", "confirmed_by", "confirmed_at", "evidence", "processes_gone", "survivors", "reviewer_sources"); err != nil {
		return err
	}
	if err := validateBenchRetirement(runtime, retirement); err != nil {
		return err
	}
	since, _ := time.Parse(time.RFC3339Nano, runtime.Since)
	until, _ := time.Parse(time.RFC3339Nano, retirement.ConfirmedAt)
	c, err := loadBenchContext(ctx, runtime.Worker.Identity.Root)
	if err != nil {
		return err
	}
	if c == nil || c.RunID != r.ID || c.WorkRoot != runtime.Worker.Identity.Root || c.HubRoot != runtime.Hub.Identity.Root || c.OriginPath != filepath.Join(s.root, "public", "origin.git") {
		return errors.New("bench metrics public context differs from retained run")
	}
	taskNumber, err := strconv.Atoi(r.Task.ID)
	if err != nil {
		return err
	}
	w := benchProtocolWindow{Since: since, Until: until, RunID: r.ID, Repository: c.OriginIdentity, Issue: fmt.Sprintf("%06d", taskNumber), WorkerRoot: c.WorkRoot, Hub: runtime.Hub.Identity}
	events, auditErr := readBenchAudit(ctx, c)
	if auditErr != nil {
		m.GateRefusals = nil
		m.GateRefusalIDs = nil
		m.Decisions = nil
		benchMetricUnknown(m, "audit-gates", "audit unavailable or malformed")
	} else {
		collectBenchAuditMetrics(events, w, m)
	}
	var ledgers []gatestate.Ledger
	var records []reviewusage.Record
	reviewsKnown := true
	env, err := openTrackerAt(ctx, c.WorkRoot)
	if err == nil {
		snapshot, e := env.repo.Snapshot()
		err = e
		if err == nil {
			card, e := snapshot.Require(r.Task.ID)
			err = e
			if err == nil {
				for _, kind := range gateLedgerKinds {
					raw, e := readBenchStatusFile(filepath.Join(c.WorkRoot, kind.path(vocab.Issue().Discovery().Plans, filepath.Base(card.Path))))
					if errors.Is(e, os.ErrNotExist) {
						continue
					}
					if e != nil {
						reviewsKnown = false
						continue
					}
					l, e := gatestate.ParseSidecar(string(raw))
					n, _ := strconv.Atoi(r.Task.ID)
					if e != nil || l.Gate != kind.Gate || l.IDPrefix != kind.IDPrefix || l.IssueNum != n {
						reviewsKnown = false
						continue
					}
					ledgers = append(ledgers, l)
				}
			}
		}
	}
	if err != nil {
		reviewsKnown = false
	}
	store, e := reviewusage.NewStore(c.WorkRoot)
	if e == nil {
		entries, err := store.AccountingEvidence(ctx)
		if err != nil {
			reviewsKnown = false
		}
		for _, entry := range entries {
			records = append(records, entry.Record)
		}
	} else {
		reviewsKnown = false
	}
	collectBenchReviewMetrics(ledgers, records, w, m)
	if !reviewsKnown {
		m.ReviewRounds = nil
		m.ReviewFindings = nil
		benchMetricUnknown(m, "native-reviews", "retained gate or accounting evidence unavailable")
	}
	return ctx.Err()
}
