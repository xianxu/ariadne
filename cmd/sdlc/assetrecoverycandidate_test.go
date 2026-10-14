package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Construct an old receipt through real handoff publication, but deliberately
// prepare main through the old unvalidated callback instead of TransferOp.Prepare.
func oldPreparedAssetTransfer(t *testing.T, state string) (*trackerRepo, string) {
	t.Helper()
	r := spinOffOnCodeBranch(t)
	appendDetail(t, r, spinOffDetails, "![missing](assets/000008/"+strings.Repeat("a", 64)+".png)")
	r.git("commit", "-qam", "old body-only asset generation")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snap.Require("000008")
	if err != nil {
		t.Fatal(err)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(r.root, spinOffDetails))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := gitx.WriteBlob(env.ctx, r.root, raw)
	if err != nil {
		t.Fatal(err)
	}
	base, err := env.git("merge-base", "HEAD", view.Ref())
	if err != nil {
		t.Fatal(err)
	}
	spec := tracker.ReceiptSpec{Token: "old-asset-candidate", Repository: env.target.Repository, IssueID: "000008", CardPath: card.Path, SourcePath: spinOffDetails, DestinationPath: spinOffDetails, SourceBranch: env.branchRef(), SourceBase: base, SourceHEAD: env.head, SourceBlob: blob, CardOID: card.BlobOID, TrackerBase: snap.Ref(), MainBase: view.Ref(), Source: tracker.LocalSource}
	transfer, err := tracker.NewTransfer(spec)
	if err != nil {
		t.Fatal(err)
	}
	var effects []tracker.Effect
	step := func(event tracker.Event) {
		t.Helper()
		transfer, effects, err = tracker.StepTransfer(transfer, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	step(tracker.Event{Kind: tracker.EventBegin, Binding: transfer.Binding()})
	op := newIssueTransferOp(env)
	event, err := op.Prepare(effects[0], transfer.Receipt())
	if err != nil {
		t.Fatal(err)
	}
	step(event)
	step(tracker.Event{Kind: tracker.EventReceiptSaved, Binding: transfer.Binding()})
	event, err = op.Apply(effects[0], transfer.Receipt())
	if err != nil {
		t.Fatal(err)
	}
	step(event)
	candidate, err := env.main.PrepareCandidate(fmt.Sprintf("old initial details\n\nTracker-Operation: %s", spec.Token), func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{spinOffDetails: raw}, ExactBytes: true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	step(tracker.Event{Kind: tracker.EventCandidatePrepared, Binding: transfer.Binding(), CandidateOID: candidate.OID, CandidateBaseOID: candidate.Base})
	if state == "uncertain" || state == "landed-uncertain" {
		step(tracker.Event{Kind: tracker.EventReceiptSaved, Binding: transfer.Binding()})
		step(tracker.Event{Kind: tracker.EventUnknown, Binding: transfer.Binding()})
	}
	if state == "landed" || state == "landed-uncertain" || state == "confirmed" {
		if outcome, err := env.main.PushCandidate(candidate); err != nil || outcome != gitx.PushAccepted {
			t.Fatalf("land fixture: %v %v", outcome, err)
		}
	}
	if state == "confirmed" {
		step(tracker.Event{Kind: tracker.EventReceiptSaved, Binding: transfer.Binding()})
		step(tracker.Event{Kind: tracker.EventConfirmed, Binding: transfer.Binding(), CandidateOID: candidate.OID})
	}
	receipts, err := env.receipts()
	if err != nil {
		t.Fatal(err)
	}
	if err := receipts.Save(transfer.Receipt()); err != nil {
		t.Fatal(err)
	}
	return r, candidate.OID
}

func TestAssetRecoveryRefusesOldPreparedMissingBundle(t *testing.T) {
	for _, state := range []string{"prepared", "uncertain", "landed", "landed-uncertain", "confirmed"} {
		t.Run(state, func(t *testing.T) {
			r, candidate := oldPreparedAssetTransfer(t, state)
			main, head := r.originMain(), r.git("rev-parse", "HEAD")
			err := runRecoveryReconcile(context.Background(), io.Discard, io.Discard, 8)
			if err == nil || !strings.Contains(err.Error(), "missing linked asset") {
				t.Fatalf("recovery accepted invalid %s generation: %v", state, err)
			}
			for _, text := range []string{r.root, "refs/sdlc/recovery/old-asset-candidate:document", candidate, "sdlc issue recovery list", "ariadne#352"} {
				if !strings.Contains(err.Error(), text) {
					t.Fatalf("refusal omits repair context %q: %v", text, err)
				}
			}
			if r.originMain() != main || r.git("rev-parse", "HEAD") != head {
				t.Fatal("invalid recovery changed main or source")
			}
			if _, err := os.Stat(filepath.Join(r.root, spinOffDetails)); err != nil {
				t.Fatal("invalid recovery relinquished source")
			}
		})
	}
}
