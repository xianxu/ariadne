package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestActualCorrectionTargetSelectorCommand(t *testing.T) {
	r, ledger, completion, inputs := actualCorrectionFixture(t)
	original, _ := os.ReadFile(ledger)
	rows, _ := estimate.ParseRowsChecked(string(original))
	second := rows[0]
	second.Model = "second-close"
	foreign := rows[0]
	foreign.Issue = "foreign#7"
	raw := string(original) + estimate.FormatRow(second) + "\n" + estimate.FormatRow(foreign) + "\n"
	mustWriteTargetLedger(t, ledger, []byte(raw))
	rows, _ = estimate.ParseRowsChecked(raw)
	calls := 0
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		calls++
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
	}
	before := r.originMain()
	beforeCard := r.card(card7Path)
	args := []string{"issue", "correct-actual", "--issue", "7", "--reason", "selected repair", "--brain-dir", r.root, "--dry-run"}
	_, _, err := executeSDLCTestCommand(args...)
	if err == nil || !strings.Contains(err.Error(), ledger) || !strings.Contains(err.Error(), completion.Token) || !strings.Contains(err.Error(), rows[1].RawID) || !strings.Contains(err.Error(), "--target-row") {
		t.Fatalf("missing concrete target recovery: %v", err)
	}
	if _, _, err := executeSDLCTestCommand(append(args, "--exclude")...); err == nil || strings.Count(err.Error(), "--exclude --dry-run") != 2 {
		t.Fatalf("ambiguity recovery lost the requested exclusion intent: %v", err)
	}
	for _, target := range []string{"short", strings.Repeat("f", 64), rows[2].RawID} {
		if _, _, err := executeSDLCTestCommand(append(args, "--target-row", target)...); err == nil {
			t.Fatalf("invalid target %q accepted", target)
		}
	}
	if calls != 0 {
		t.Fatalf("refusals measured %d times", calls)
	}
	out, stderr, err := executeSDLCTestCommand(append(args, "--target-row", rows[1].RawID)...)
	if err != nil || !strings.Contains(out, rows[1].RawID) {
		t.Fatalf("selected dry-run: %v\n%s\n%s", err, out, stderr)
	}
	after, _ := os.ReadFile(ledger)
	if string(after) != raw || before != r.originMain() || beforeCard != r.card(card7Path) {
		t.Fatal("dry-run/refusal changed ledger or main")
	}
	entries, _ := listActualCorrections(context.Background(), r.root)
	if len(entries) != 0 {
		t.Fatal("dry run made receipt")
	}
}

func TestActualCorrectionTargetSelectorRetry(t *testing.T) {
	r, ledger, _, inputs := actualCorrectionFixture(t)
	original, _ := os.ReadFile(ledger)
	rows, _ := estimate.ParseRowsChecked(string(original))
	second := rows[0]
	second.Model = "selected"
	raw := string(original) + estimate.FormatRow(second) + "\n"
	mustWriteTargetLedger(t, ledger, []byte(raw))
	rows, _ = estimate.ParseRowsChecked(raw)
	target := rows[1].RawID
	calls := 0
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		calls++
		if err := os.Rename(ledger, ledger+".held"); err != nil {
			t.Error(err)
		}
		return actualResult{Status: actualMeasured, Hours: 2.57, Inputs: &inputs}
	}
	args := []string{"issue", "correct-actual", "--issue", "7", "--reason", "repair", "--brain-dir", r.root, "--target-row", target}
	if _, _, err := executeSDLCTestCommand(args...); err == nil {
		t.Fatal("expected interrupted correction")
	}
	entries, err := listActualCorrections(context.Background(), r.root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries %v: %v", entries, err)
	}
	store, _ := gitx.NewRecoveryStoreFor(context.Background(), r.root, "actual-correction")
	entry, _ := store.Load(entries[0].Token)
	doc, err := readActualCorrection(entry)
	if err != nil || len(doc.State.Spec.TargetHashes) != 1 || doc.State.Spec.TargetHashes[0] != target {
		t.Fatalf("selector not bound: %+v %v", doc.State.Spec.TargetHashes, err)
	}
	mustWriteTargetLedger(t, ledger, []byte(raw))
	before, _ := os.ReadFile(ledger)
	beforeCard := r.card(card7Path)
	beforeMain := r.originMain()
	conflict := append([]string{}, args...)
	conflict[len(conflict)-1] = rows[0].RawID
	if _, _, err := executeSDLCTestCommand(conflict...); err == nil || !strings.Contains(err.Error(), "target") {
		t.Fatalf("pending retarget: %v", err)
	}
	after, _ := os.ReadFile(ledger)
	if !bytes.Equal(before, after) || beforeCard != r.card(card7Path) || beforeMain != r.originMain() {
		t.Fatal("conflicting retry changed ledger")
	}
	if _, _, err := executeSDLCTestCommand(args...); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if calls != 1 {
		t.Fatal("resume remeasured")
	}
	before, _ = os.ReadFile(ledger)
	beforeCard = r.card(card7Path)
	beforeMain = r.originMain()
	if _, _, err := executeSDLCTestCommand(conflict...); err == nil || !strings.Contains(err.Error(), "target") {
		t.Fatalf("completed retarget: %v", err)
	}
	after, _ = os.ReadFile(ledger)
	if !bytes.Equal(before, after) || beforeCard != r.card(card7Path) || beforeMain != r.originMain() {
		t.Fatal("completed conflicting retry changed stores")
	}
	if _, _, err := executeSDLCTestCommand(args...); err != nil {
		t.Fatalf("completed same target: %v", err)
	}
	data, _ := os.ReadFile(ledger)
	parsed, _ := estimate.ParseRowsChecked(string(data))
	last := parsed[len(parsed)-1]
	if !strings.HasPrefix(string(data), raw) {
		t.Fatal("original calibration rows changed")
	}
	if last.Correction == nil || len(last.Correction.Targets) != 1 || last.Correction.Targets[0] != target {
		t.Fatalf("wrong durable target: %+v", last.Correction)
	}
}

func TestActualCorrectionTargetSelectorExclusionAction(t *testing.T) {
	r, ledger, _, _ := actualCorrectionFixture(t)
	raw, _ := os.ReadFile(ledger)
	rows, _ := estimate.ParseRowsChecked(string(raw))
	target := rows[0].RawID
	measureActualCorrection = func(context.Context, string, string, string, string, string) actualResult {
		return actualResult{Status: actualError, Detail: "historical head mismatch"}
	}
	args := []string{"issue", "correct-actual", "--issue", "7", "--reason", "bad attribution", "--brain-dir", r.root, "--target-row", target}
	_, _, err := executeSDLCTestCommand(args...)
	if err == nil || !strings.Contains(err.Error(), "historical head mismatch") || !strings.Contains(err.Error(), "--exclude") || !strings.Contains(err.Error(), "--target-row "+target) {
		t.Fatalf("missing exclusion action: %v", err)
	}
	if _, _, err := executeSDLCTestCommand(append(args, "--exclude")...); err != nil {
		t.Fatal(err)
	}
}

func TestActualCorrectionTargetSelectorEligibility(t *testing.T) {
	base := estimate.LedgerRow{Issue: "pair#528", Actual: 0.05, Date: "2026-10-15", Model: "eligible"}
	foreignIssue := base
	foreignIssue.Issue = "pair#529"
	foreignDate := base
	foreignDate.Date = "2026-10-14"
	foreignActual := base
	foreignActual.Actual = 0.06
	text := estimate.Header() + "\n"
	for _, row := range []estimate.LedgerRow{base, foreignIssue, foreignDate, foreignActual} {
		text += estimate.FormatRow(row) + "\n"
	}
	rows, err := estimate.ParseRowsChecked(text)
	if err != nil {
		t.Fatal(err)
	}
	replacement := base
	replacement.Model = "replacement"
	replacement.Date = "2026-10-16"
	replacement.Correction = &estimate.CorrectionRecord{Version: 1, Operation: "replace-old", Generation: "close-old", Disposition: "replace", Targets: []string{rows[0].RawID}, InputDigest: strings.Repeat("a", 64)}
	text += estimate.FormatRow(replacement) + "\n"
	rows, err = estimate.ParseRowsChecked(text)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if _, err := actualCorrectionTargetForGeneration(text, base.Issue, "0.05", base.Date, "close-new", row.RawID); err == nil {
			t.Fatalf("foreign/superseded row admitted: %+v", row)
		}
	}
	got, err := actualCorrectionTargetForGeneration(text, base.Issue, "0.05", base.Date, "close-old", rows[4].RawID)
	if err != nil || got.RawID != rows[4].RawID {
		t.Fatalf("matching corrected generation rejected: %+v %v", got, err)
	}
	// Date does not override an exact corrected generation; a foreign completion
	// cannot adopt it merely because it has the same date and numeric actual.
	if _, err := actualCorrectionTargetForGeneration(text, base.Issue, "0.05", replacement.Date, "close-new", rows[4].RawID); err == nil {
		t.Fatal("date bypassed corrected generation")
	}
}

func mustWriteTargetLedger(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
