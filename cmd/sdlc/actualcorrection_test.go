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

func TestActualCorrectionAppendUpgradesLegacyHeaderWithoutChangingSamples(t *testing.T) {
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "calibration.tsv")
	oldHeader, _, _ := strings.Cut(estimate.Header(), "\tcorrection") // the header before the correction column
	target := estimate.LedgerRow{Issue: "pair#454", Actual: 0.05, Date: "2026-10-11", WindowTrusted: true}
	other := estimate.LedgerRow{Issue: "pair#515", Actual: 0.15, Date: "2026-10-12"}
	samples := estimate.FormatRow(target) + "\r\n" + estimate.FormatRow(other)
	raw := "# retained comment\r\n" + oldHeader + "\r\n" + samples
	if err := os.WriteFile(file, []byte(raw), 0640); err != nil {
		t.Fatal(err)
	}
	parsed, err := estimate.ParseRowsChecked(raw)
	if err != nil || len(parsed) != 2 {
		t.Fatalf("rows %+v: %v", parsed, err)
	}
	row := target
	row.Actual = 2.57
	row.Correction = &estimate.CorrectionRecord{Version: 1, Operation: "header-repair", Generation: "close-one", Disposition: "replace", Targets: []string{strings.Repeat("f", 64)}, InputDigest: strings.Repeat("a", 64)}
	invalid := estimate.FormatRow(row)
	if err := appendActualCorrectionRow(ctx, io.Discard, file, invalid, "header-repair", correctionDigest([]byte(invalid))); err == nil {
		t.Fatal("missing target accepted")
	}
	unchanged, err := os.ReadFile(file)
	if err != nil || string(unchanged) != raw {
		t.Fatalf("invalid targets changed old header/sample bytes: %v", err)
	}
	row.Correction.Targets = []string{parsed[0].RawID}
	line := estimate.FormatRow(row)
	for i := 0; i < 3; i++ {
		if err := appendActualCorrectionRow(ctx, io.Discard, file, line, "header-repair", correctionDigest([]byte(line))); err != nil {
			t.Fatal(err)
		}
	}
	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "\n"+estimate.Header()+"\n") {
		t.Fatalf("correction column header not upgraded: %s", after)
	}
	if !strings.Contains(string(after), samples+"\n"+line+"\n") {
		t.Fatal("existing raw sample bytes changed")
	}
	if strings.Count(string(after), "header-repair") != 1 {
		t.Fatal("header retry duplicated operation row")
	}
	// Recovery also repairs an older header when its exact row was already
	// appended, without adding a second row or changing sample identities.
	oldAgain := strings.Replace(string(after), estimate.Header(), oldHeader, 1)
	if err := os.WriteFile(file, []byte(oldAgain), 0640); err != nil {
		t.Fatal(err)
	}
	if err := appendActualCorrectionRow(ctx, io.Discard, file, line, "header-repair", correctionDigest([]byte(line))); err != nil {
		t.Fatal(err)
	}
	after, err = os.ReadFile(file)
	if err != nil || !strings.Contains(string(after), "\n"+estimate.Header()+"\n") || strings.Count(string(after), "header-repair") != 1 {
		t.Fatalf("existing-row header recovery failed: %v", err)
	}
	updated, err := estimate.ParseRowsChecked(string(after))
	if err != nil || len(updated) != 3 {
		t.Fatalf("rows %+v: %v", updated, err)
	}
	for i := range parsed {
		if parsed[i].RawID != updated[i].RawID {
			t.Fatal("header migration changed sample identity")
		}
	}
	info, err := os.Stat(file)
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("ledger permissions changed: %v %v", info, err)
	}
	leftover, err := filepath.Glob(filepath.Join(filepath.Dir(file), ".actual-correction-*"))
	if err != nil || len(leftover) != 0 {
		t.Fatalf("temporary files leaked: %v %v", leftover, err)
	}
}

func actualCorrectionFixture(t *testing.T) (*trackerRepo, string, issue.Completion, gatestate.MeasurementInputs) {
	t.Helper()
	return actualCorrectionFixtureAt(t, "workshop/history")
}

func actualCorrectionFixtureAt(t *testing.T, archiveDir string) (*trackerRepo, string, issue.Completion, gatestate.MeasurementInputs) {
	t.Helper()
	t.Setenv("WF_HISTORY_DIR", "workshop/history")
	full := []byte("---\nid: 000007\nstatus: done\ncreated: 2026-09-01\nupdated: 2026-09-01\nactual_hours: 0.05\n---\n\n# Seven\n\n## Problem\nReport.\n")
	card, details, err := issue.SplitCardWithFormat(full, "sha1")
	if err != nil {
		t.Fatal(err)
	}
	r := newTrackerRepo(t, map[string]string{card7Path: string(card)}, map[string]string{filepath.ToSlash(filepath.Join(archiveDir, "000007-seven.md")): string(details)})
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

func TestActualCorrectionStandardArchiveMirror(t *testing.T) {
	r, _, _, inputs := actualCorrectionFixtureAt(t, "workshop/history/issues")
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
	}
	ctx := context.Background()
	if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, "repair standard archive", r.root, true, false); err != nil {
		t.Fatal(err)
	}
	if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, "repair standard archive", r.root, false, false); err != nil {
		t.Fatal(err)
	}
	env, err := openTracker(ctx)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snapshot.Require("000007")
	if err != nil {
		t.Fatal(err)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := view.Read("workshop/history/issues/000007-seven.md")
	if err != nil || len(mirror) == 0 {
		t.Fatalf("standard archive missing: %v", err)
	}
	if _, err := issue.RefreshMirror(mirror, card.Raw, card.Raw); err != nil {
		t.Fatalf("standard card/archive disagree: %v", err)
	}
	if exists, err := view.Exists("workshop/history/000007-seven.md"); err != nil || exists {
		t.Fatalf("created legacy copy: %v", err)
	}
}

func TestActualCorrectionCompleteRepeatDoesNotRemeasure(t *testing.T) {
	for _, excluded := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude=%v", excluded), func(t *testing.T) {
			r, ledger, _, inputs := actualCorrectionFixtureAt(t, "workshop/history/issues")
			ctx := context.Background()
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				if excluded {
					return actualResult{Status: actualError, Detail: "historical provenance missing"}
				}
				return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
			}
			const reason = "repeat measured repair"
			if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, reason, r.root, false, excluded); err != nil {
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
			ledgerBefore, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			measureCalls := 0
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				measureCalls++
				return actualResult{Status: actualError, Detail: "complete repeat must not measure"}
			}
			for i := 0; i < 2; i++ {
				if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, reason, r.root, i == 1, excluded); err != nil {
					t.Fatal(err)
				}
			}
			if measureCalls != 0 {
				t.Fatal("completed correction remeasured")
			}
			after, err := env.repo.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			ledgerAfter, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			if before.Ref() != after.Ref() || mainBefore != r.originMain() || !bytes.Equal(ledgerBefore, ledgerAfter) {
				t.Fatal("completed repeat changed authoritative evidence")
			}
			entries, err := listActualCorrections(ctx, r.root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("repeat created recovery refs: %v %v", entries, err)
			}
		})
	}
}

func TestActualCorrectionCompleteRepeatRefusesIncompleteProof(t *testing.T) {
	for _, broken := range []string{"missing ledger", "foreign operation", "foreign generation", "foreign digest", "foreign issue", "stale mirror", "newer ordinary row"} {
		t.Run(broken, func(t *testing.T) {
			r, ledger, _, inputs := actualCorrectionFixtureAt(t, "workshop/history/issues")
			ctx := context.Background()
			env, err := openTracker(ctx)
			if err != nil {
				t.Fatal(err)
			}
			view, err := env.main.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			const mirrorPath = "workshop/history/issues/000007-seven.md"
			originalMirror, err := view.Read(mirrorPath)
			if err != nil {
				t.Fatal(err)
			}
			originalLedger, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
			}
			const reason = "repair proof test"
			if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, reason, r.root, false, false); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := estimate.ParseRowsChecked(string(data))
			if err != nil || len(rows) != 2 {
				t.Fatalf("rows %+v: %v", rows, err)
			}
			switch broken {
			case "missing ledger":
				data = originalLedger
			case "stale mirror":
				if err := env.main.UpdateMany("fixture: restore stale mirror", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
					return gitx.TrunkWrite{Write: map[string][]byte{mirrorPath: originalMirror}, ExactBytes: true}, nil
				}); err != nil {
					t.Fatal(err)
				}
			case "newer ordinary row":
				later := rows[0]
				later.Correction = nil
				later.Date = "2099-01-01"
				later.Model = "later-close"
				data = append(data, []byte(estimate.FormatRow(later)+"\n")...)
			default:
				correction := rows[1]
				switch broken {
				case "foreign operation":
					correction.Correction.Operation = "foreign-operation"
				case "foreign generation":
					correction.Correction.Generation = "foreign-generation"
				case "foreign digest":
					correction.Correction.InputDigest = strings.Repeat("f", 64)
				case "foreign issue":
					rows[0].Issue = "foreign#7"
					correction.Issue = "foreign#7"
					correction.Correction.Targets = []string{correctionDigest([]byte(estimate.FormatRow(rows[0])))}
				}
				data = []byte(estimate.Header() + "\n" + estimate.FormatRow(rows[0]) + "\n" + estimate.FormatRow(correction) + "\n")
			}
			if err := os.WriteFile(ledger, data, 0600); err != nil {
				t.Fatal(err)
			}
			// Foreign identities above are otherwise valid resolved ledgers, so
			// each refusal is owned by its completion-proof guard.
			parsed, err := estimate.ParseRowsChecked(string(data))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := estimate.ResolveCorrections(parsed); err != nil {
				t.Fatal(err)
			}
			before, err := env.repo.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			mainBefore := r.originMain()
			measureCalls := 0
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				measureCalls++
				return actualResult{Status: actualError, Detail: "must not remeasure incomplete own operation"}
			}
			err = runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, reason, r.root, false, false)
			if err == nil || !strings.Contains(err.Error(), "not proved complete") || !strings.Contains(err.Error(), "reconcile") {
				t.Fatalf("missing proof was not actionable refusal: %v", err)
			}
			if measureCalls != 0 {
				t.Fatal("incomplete prior operation remeasured")
			}
			after, err := env.repo.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			ledgerAfter, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			if before.Ref() != after.Ref() || mainBefore != r.originMain() || !bytes.Equal(data, ledgerAfter) {
				t.Fatal("refusal wrote authoritative stores")
			}
			entries, err := listActualCorrections(ctx, r.root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("refusal created refs: %v %v", entries, err)
			}
		})
	}
}

func TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust(t *testing.T) {
	for _, tc := range []struct {
		name, method string
		trusted      bool
	}{{"legacy", gatestate.MeasurementMethodLegacyV1, false}, {"engagement", gatestate.MeasurementMethodEngagementV1, true}, {"queue", gatestate.MeasurementMethodQueueV1, true}, {"hybrid", gatestate.MeasurementMethodHybridV1, true}, {"hybrid-queue", gatestate.MeasurementMethodHybridQueueV1, true}} {
		t.Run(tc.name, func(t *testing.T) {
			r, ledger, _, inputs := actualCorrectionFixture(t)
			data, err := os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := estimate.ParseRowsChecked(string(data))
			if err != nil || len(rows) != 1 {
				t.Fatalf("rows %+v: %v", rows, err)
			}
			rows[0].WindowTrusted = false
			if err := os.WriteFile(ledger, []byte(estimate.Header()+"\n"+estimate.FormatRow(rows[0])+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if tc.trusted {
				env, err := openTracker(ctx)
				if err != nil {
					t.Fatal(err)
				}
				snapshot, err := env.repo.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				inputs.Method = tc.method
				inputs.TrackerHead = snapshot.Ref()
				inputs.Engagements = []gatestate.MeasurementEngagement{{Start: inputs.Since, End: inputs.Until, Began: inputs.Since, Roots: []string{r.root}}}
				if gatestate.IsHybridMethod(tc.method) {
					prefix := inputs
					prefix.Method = gatestate.MeasurementMethodLegacyV1
					prefix.TrackerHead = ""
					prefix.Engagements = nil
					prefix.Until = "2026-10-14T00:00:30Z"
					inputs.Version = gatestate.MeasurementHybridInputsVersion
					inputs.LegacyPrefix = &prefix
					inputs.LegacyCutover = inputs.CodeHead
					inputs.Engagements[0].Start = prefix.Until
				}
				if gatestate.IsQueueMethod(tc.method) {
					inputs.QueueWaits = []gatestate.MeasurementQueueWait{{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: inputs.Engagements[0].Start, End: "2026-10-14T00:01:30Z", QueueSeconds: 90, RawSeconds: 90, Exact: true}}
					if gatestate.IsHybridMethod(tc.method) {
						inputs.QueueWaits[0].QueueSeconds = 60
						inputs.QueueWaits[0].RawSeconds = 60
					}
				}
			}
			if err := inputs.Validate(); err != nil {
				t.Fatal(err)
			}
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
			}
			if err := runActualCorrection(ctx, io.Discard, io.Discard, "workshop/issues", 7, "repair ownership", r.root, false, false); err != nil {
				t.Fatal(err)
			}
			data, err = os.ReadFile(ledger)
			if err != nil {
				t.Fatal(err)
			}
			rows, err = estimate.ParseRowsChecked(string(data))
			if err != nil {
				t.Fatal(err)
			}
			resolved, err := estimate.ResolveCorrections(rows)
			if err != nil || len(resolved) != 1 {
				t.Fatalf("resolved %+v: %v", resolved, err)
			}
			if resolved[0].WindowTrusted != tc.trusted {
				t.Fatalf("correction trust %v, want %v", resolved[0].WindowTrusted, tc.trusted)
			}
		})
	}
}

func TestActualCorrectionNumericDiagnostics(t *testing.T) {
	for _, dryRun := range []bool{true, false} {
		t.Run(fmt.Sprintf("dry=%v", dryRun), func(t *testing.T) {
			r, _, _, inputs := actualCorrectionFixture(t)
			measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
				return actualResult{Status: actualMeasured, Issue: "7", Hours: 2.57, Window: "2026-10-14T00:00:00Z → 2026-10-14T01:00:00Z", Inputs: &inputs, Warnings: []string{"historical reconstruction from close evidence", "provenance: owner roots; excluded brain: 2 sessions, 3.00 minutes"}}
			}
			var stdout, stderr bytes.Buffer
			if err := runActualCorrection(context.Background(), &stdout, &stderr, "workshop/issues", 7, "repair with diagnostics", r.root, dryRun, false); err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"window 2026-10-14T00:00:00Z → 2026-10-14T01:00:00Z", "historical reconstruction from close evidence", "provenance: owner roots; excluded brain: 2 sessions, 3.00 minutes"} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("missing diagnostic %q: stdout=%s stderr=%s", want, &stdout, &stderr)
				}
			}
		})
	}
}
