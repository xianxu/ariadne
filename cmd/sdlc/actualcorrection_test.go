package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/actualcorrection"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestActualCorrectionTargetRequiresUniqueGeneration(t *testing.T) {
	row := estimate.LedgerRow{Issue: "pair#454", Actual: 0.05, Date: "2026-10-11", WindowTrusted: true}
	text := estimate.Header() + "\n" + estimate.FormatRow(row) + "\n"
	target, err := actualCorrectionTarget(text, "pair#454", "0.05", "2026-10-11")
	if err != nil || target.RawID == "" {
		t.Fatalf("target %+v: %v", target, err)
	}
	if _, err := actualCorrectionTarget(text, "pair#454", "0.05", "2026-10-12"); err == nil {
		t.Fatal("foreign close date adopted")
	}
	if _, err := actualCorrectionTarget(text, "pair#454", "0.06", "2026-10-11"); err == nil {
		t.Fatal("foreign actual adopted")
	}
	row.Model = "foreign-generation"
	if _, err := actualCorrectionTarget(text+estimate.FormatRow(row)+"\n", "pair#454", "0.05", "2026-10-11"); err == nil {
		t.Fatal("ambiguous generation adopted")
	}
}

func TestActualCorrectionAppendIdempotentAndConflict(t *testing.T) {
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "calibration.tsv")
	target := estimate.LedgerRow{Issue: "pair#454", Actual: 0.05, Date: "2026-10-11", WindowTrusted: true}
	raw := estimate.Header() + "\n" + estimate.FormatRow(target) + "\n"
	if err := os.WriteFile(file, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	parsed, err := estimate.ParseRowsChecked(raw)
	if err != nil {
		t.Fatal(err)
	}
	row := target
	row.Actual = 2.57
	row.Correction = &estimate.CorrectionRecord{Version: 1, Operation: "repair-one", Generation: "close-one", Disposition: "replace", Targets: []string{parsed[0].RawID}, InputDigest: strings.Repeat("a", 64)}
	line := estimate.FormatRow(row)
	for i := 0; i < 3; i++ {
		if err := appendActualCorrectionRow(ctx, io.Discard, file, line, "repair-one", correctionDigest([]byte(line))); err != nil {
			t.Fatal(err)
		}
	}
	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(after), "repair-one") != 1 {
		t.Fatal("retry duplicated correction")
	}
	row.Actual = 3.14
	foreign := estimate.FormatRow(row)
	if err := appendActualCorrectionRow(ctx, io.Discard, file, foreign, "repair-one", correctionDigest([]byte(foreign))); err == nil {
		t.Fatal("conflicting same operation adopted")
	}
}

func actualCorrectionFixture(t *testing.T) (*trackerRepo, string, issue.Completion, gatestate.MeasurementInputs) {
	t.Helper()
	full := []byte("---\nid: 000007\nstatus: done\ncreated: 2026-09-01\nupdated: 2026-09-01\nactual_hours: 0.05\n---\n\n# Seven\n\n## Problem\nReport.\n")
	card, details, err := issue.SplitCardWithFormat(full, "sha1")
	if err != nil {
		t.Fatal(err)
	}
	r := newTrackerRepo(t, map[string]string{card7Path: string(card)}, map[string]string{"workshop/history/000007-seven.md": string(details)})
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reviewed := r.git("rev-parse", "HEAD")
	r.git("commit", "--allow-empty", "-qm", "#7: close\n\nClose-Actual: 0.05")
	evidence := r.git("rev-parse", "HEAD")
	r.git("push", "-q", "origin", "main")
	completion := issue.Completion{Token: "close-seven", Repository: env.target.Repository, ReviewedHEAD: reviewed, EvidenceCommit: evidence}
	updated, err := issue.SetCardCompletion(card, completion)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	old, err := snapshot.Require("000007")
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.UpdateCard(old, updated, "fixture-close", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	identity, err := resolveWorkspace(r.root)
	if err != nil {
		t.Fatal(err)
	}
	row := estimate.LedgerRow{Issue: identity.Repo + "#7", Estimate: 1, Actual: 0.05, Date: r.git("show", "-s", "--format=%cs", evidence), WindowTrusted: true}
	ledger := filepath.Join(t.TempDir(), "calibration.tsv")
	if err := os.WriteFile(ledger, []byte(estimate.Header()+"\n"+estimate.FormatRow(row)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_CALIB_LEDGER", ledger)
	inputs := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: env.target.Repository, CodeHead: reviewed, Since: "2026-10-14T00:00:00Z", Until: "2026-10-14T01:00:00Z", GapMinutes: 30}
	prior := measureActualCorrection
	t.Cleanup(func() { measureActualCorrection = prior })
	return r, ledger, completion, inputs
}

func TestActualCorrectionResumesConfirmedTrackerWithoutRemeasure(t *testing.T) {
	r, ledger, completion, inputs := actualCorrectionFixture(t)
	original, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		calls++
		if err := os.Rename(ledger, ledger+".interrupted"); err != nil {
			t.Error(err)
		}
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs, Warnings: []string{"attribution: recorded owner; bounded source prefixes"}}
	}
	if err := runActualCorrection(context.Background(), io.Discard, io.Discard, "workshop/issues", 7, "measured repair", r.root, false, false); err == nil {
		t.Fatal("expected interruption after tracker publication")
	}
	entries, err := listActualCorrections(context.Background(), r.root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("pending entries %v: %v", entries, err)
	}
	store, _ := gitx.NewRecoveryStoreFor(context.Background(), r.root, "actual-correction")
	entry, err := store.Load(entries[0].Token)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := readActualCorrection(entry)
	if err != nil {
		t.Fatal(err)
	}
	if doc.State.Phase != actualcorrection.CalibrationPending {
		t.Fatalf("confirmed tracker not durable: %s", doc.State.Phase)
	}
	if err := os.WriteFile(ledger, original, 0600); err != nil {
		t.Fatal(err)
	}
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		calls++
		return actualResult{Status: actualError, Detail: "must not remeasure a prepared correction"}
	}
	if err := runActualCorrection(context.Background(), io.Discard, io.Discard, "workshop/issues", 7, "retry", r.root, false, false); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("recovery remeasured")
	}
	entries, err = listActualCorrections(context.Background(), r.root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("completed receipt retained: %v %v", entries, err)
	}
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snap.Require("000007")
	if err != nil {
		t.Fatal(err)
	}
	closed, ok, err := issue.CardCompletion(card.Raw)
	if err != nil || !ok || closed != completion {
		t.Fatal("correction rewrote completion")
	}
	actual, _ := issue.GetField(card.Card.Frontmatter, "actual_hours")
	if actual != "2.57" {
		t.Fatalf("actual %s", actual)
	}
	data, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := estimate.ParseRowsChecked(string(data))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := estimate.ResolveCorrections(rows)
	if err != nil || len(resolved) != 1 || resolved[0].Actual != 2.57 {
		t.Fatalf("resolved %+v: %v", resolved, err)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := view.Read("workshop/history/000007-seven.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issue.RefreshMirror(mirror, card.Raw, card.Raw); err != nil {
		t.Fatalf("card/archive disagree: %v", err)
	}
}

func TestActualCorrectionDryRunAndUncertainRefusal(t *testing.T) {
	r, ledger, _, inputs := actualCorrectionFixture(t)
	before := r.originMain()
	data, _ := os.ReadFile(ledger)
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs, Warnings: []string{"provenance summary"}}
	}
	var out bytes.Buffer
	if err := runActualCorrection(context.Background(), &out, io.Discard, "workshop/issues", 7, "repair", r.root, true, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "0.05 -> 2.57") {
		t.Fatal(out.String())
	}
	if before != r.originMain() {
		t.Fatal("dry run published main")
	}
	after, _ := os.ReadFile(ledger)
	if !bytes.Equal(data, after) {
		t.Fatal("dry run appended calibration")
	}
	entries, err := listActualCorrections(context.Background(), r.root)
	if err != nil || len(entries) != 0 {
		t.Fatal("dry run created recovery state")
	}
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		return actualResult{Status: actualError, Detail: "historical provenance missing"}
	}
	if err := runActualCorrection(context.Background(), io.Discard, io.Discard, "workshop/issues", 7, "repair", r.root, false, false); err == nil {
		t.Fatal("uncertain measurement adopted numeric actual")
	}
	if err := runActualCorrection(context.Background(), io.Discard, io.Discard, "workshop/issues", 7, "proven invalid attribution", r.root, false, true); err != nil {
		t.Fatal(err)
	}
	env, _ := openTracker(context.Background())
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snap.Require("000007")
	if err != nil {
		t.Fatal(err)
	}
	evidence, ok, err := issue.CardActualCorrection(card.Raw)
	if err != nil || !ok || !evidence.Exclude || !strings.Contains(evidence.Reason, "historical provenance missing") {
		t.Fatalf("exclusion evidence %+v %v", evidence, err)
	}
	if evidence.NewActual != "N/A" {
		t.Fatal(fmt.Sprintf("exclusion invented %s hours", evidence.NewActual))
	}
}

func TestActualCorrectionAbsentMovedCandidatePreservesReceipt(t *testing.T) {
	r, ledger, _, inputs := actualCorrectionFixture(t)
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		if err := os.Rename(ledger, ledger+".interrupted"); err != nil {
			t.Error(err)
		}
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
	}
	if err := runActualCorrection(context.Background(), io.Discard, io.Discard, "workshop/issues", 7, "repair", r.root, false, false); err == nil {
		t.Fatal("expected calibration interruption")
	}
	ctx := context.Background()
	store, err := gitx.NewRecoveryStoreFor(ctx, r.root, "actual-correction")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := store.List()
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries %v: %v", entries, err)
	}
	entry, err := store.Load(entries[0].Token)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := readActualCorrection(entry)
	if err != nil {
		t.Fatal(err)
	}
	// Model an unpublished outstanding candidate while another writer moves
	// the tracker tip. Even though today's card has our identical hours and
	// token, recovery must prove this exact persisted candidate, not adopt a
	// different commit's otherwise matching card.
	trunk, err := gitx.NewTrunkFileContext(ctx, r.root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	pendingPath, pendingCard, _, _ := seededIssue(t, "000335", "unpublished-attempt")
	otherPath, otherCard, _, _ := seededIssue(t, "000336", "another-writer")
	missing, err := trunk.PrepareCandidate("unpublished attempt\n\nTracker-Operation: "+entry.Token, func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{pendingPath: []byte(pendingCard)}, ExactBytes: true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := trunk.UpdateMany("another writer", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{otherPath: []byte(otherCard)}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	doc.State.Phase = actualcorrection.UncertainTracker
	doc.State.TrackerCommit = ""
	doc.TrackerCandidate = missing
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	objects := append(entry.Data.Objects, missing.Base, missing.OID)
	oid, err := store.Save(entry.Token, entry.OID, gitx.RecoveryData{Document: raw, Objects: objects})
	if err != nil {
		t.Fatal(err)
	}
	env, err := openTracker(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	mainBefore := r.originMain()
	err = resumeActualCorrection(ctx, io.Discard, io.Discard, r.root, entry.Token)
	if err == nil || !strings.Contains(err.Error(), "not confirmed") {
		t.Fatalf("absent moved candidate was adopted: %v", err)
	}
	after, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if before.Ref() != after.Ref() || mainBefore != r.originMain() {
		t.Fatal("refused recovery republished")
	}
	pending, err := store.Load(entry.Token)
	if err != nil || pending.OID != oid {
		t.Fatalf("refusal discarded or changed pending evidence: %v", err)
	}
}
