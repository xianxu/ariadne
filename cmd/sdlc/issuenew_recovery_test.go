package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// rejectMainPushes makes origin refuse any update of main until the returned
// func lifts it, as a hook, an outage or a lost response would.
func rejectMainPushes(t *testing.T, r *trackerRepo) (lift func()) {
	t.Helper()
	hook := filepath.Join(r.origin, "hooks", "pre-receive")
	script := "#!/bin/sh\nwhile read old new ref; do [ \"$ref\" = refs/heads/main ] && { echo 'main is closed' >&2; exit 1; }; done; exit 0\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Remove(hook); err != nil {
			t.Fatal(err)
		}
	}
}

// stuckCreation files #8 while main refuses the details' push: the card is
// reserved, the details are not on main, and the receipt stays at creation.main.
func stuckCreation(t *testing.T, r *trackerRepo) {
	t.Helper()
	slotClaimant(t, "ariadne:0")
	lift := rejectMainPushes(t, r)
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Stuck Filing"}); err == nil {
		t.Fatal("filing succeeded although main refused the details")
	}
	lift()
	if r.card("workshop/issue-cards/000008-stuck-filing.md") == "" {
		t.Fatal("fixture: the card was not reserved")
	}
	if r.mainFile("workshop/issues/000008-stuck-filing.md") != "" {
		t.Fatal("fixture: the details reached main")
	}
}

// BR-1 (#308 M1 review): creation's stages both publish to origin, so a
// stopped filing is finished from any checkout — here another branch than the
// one that filed it — and still writes nothing local.
func TestIssueNewStoppedAtMainFinishesFromAnotherCheckout(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	stuckCreation(t, r)
	r.git("switch", "-q", "-c", "elsewhere")
	head, status := r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 8); err != nil {
		t.Fatalf("reconcile from another branch: %v\n%s", err, errs.String())
	}
	if r.mainFile("workshop/issues/000008-stuck-filing.md") == "" {
		t.Fatal("reconcile did not publish the details")
	}
	assertCheckoutUntouched(t, r, head, status)
}

// BR-2: a receipt an older sdlc left in the retired local stage (its candidate
// is the detail blob) is released by reconcile, naming `issue publish`; main
// is not touched.
func TestRecoveryLegacyCreationReceiptPointsAtPublish(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	stuckCreation(t, r)
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	receipts, err := env.receipts()
	if err != nil {
		t.Fatal(err)
	}
	left, err := receipts.List()
	if err != nil || len(left) != 1 {
		t.Fatalf("fixture: want one receipt, got %d (%v)", len(left), err)
	}
	// Rewrite it as the pre-#308 binary left it: stage 1 pinning the blob.
	raw, err := tracker.MarshalReceipt(left[0])
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["stage"] != float64(1) {
		t.Fatalf("fixture: receipt at stage %v, want 1", wire["stage"])
	}
	wire["candidate_oid"] = wire["spec"].(map[string]any)["source_blob"]
	edited, _ := json.Marshal(wire)
	legacy, err := tracker.ParseReceipt(edited, env.target.Repository)
	if err != nil {
		t.Fatal(err)
	}
	if !legacy.LegacyCreationDetail() {
		t.Fatal("fixture: the rewritten receipt is not the legacy shape")
	}
	if err := receipts.Save(legacy); err != nil {
		t.Fatal(err)
	}
	mainBefore := r.originMain()
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 8); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	if !strings.Contains(errs.String(), "sdlc issue publish --issue 8") {
		t.Errorf("no next action naming publish:\n%s", errs.String())
	}
	if r.originMain() != mainBefore {
		t.Fatal("releasing a legacy receipt published to main")
	}
	fresh, _ := openTracker(context.Background())
	again, _ := fresh.receipts()
	if rest, err := again.List(); err != nil || len(rest) != 0 {
		t.Fatalf("legacy receipt not released: %d %v", len(rest), err)
	}
}
