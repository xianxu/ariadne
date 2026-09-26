package tracker

import (
	"fmt"
	"testing"
)

// A close commits its evidence before publishing codecomplete, and the card
// stage is bound to that evidence commit.
func TestCompletionCommitsEvidenceBeforeCodecomplete(t *testing.T) {
	spec := operationSpec()
	spec.ReviewedHEAD = spec.SourceHEAD
	spec.EvidenceMessage = "#1: close\n\nClose-Actual: 1"
	spec.EvidencePaths = spec.SourcePath
	s, err := NewCompletion(spec)
	if err != nil {
		t.Fatal(err)
	}
	s, effects, err := StepCompletion(s, Event{Kind: EventBegin, Binding: s.Binding()})
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range []EffectKind{WriteEvidence, PublishCard} {
		if effects[0].Kind != PrepareCandidate {
			t.Fatalf("prepare %v", effects)
		}
		oid := fmt.Sprintf("%040x", 100+i)
		s, effects, err = StepCompletion(s, preparedCandidate(s.Binding(), oid))
		if err != nil || effects[0].Kind != PersistReceipt {
			t.Fatalf("save %v %v", effects, err)
		}
		s, effects, err = StepCompletion(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
		if err != nil || effects[0].Kind != kind {
			t.Fatalf("apply %v %v", effects, err)
		}
		if kind == PublishCard && effects[0].Expected.EvidenceOID != fmt.Sprintf("%040x", 100) {
			t.Fatal("codecomplete is not bound to the evidence commit")
		}
		event := Event{Kind: EventConfirmed, Binding: s.Binding(), CandidateOID: oid}
		if kind == PublishCard {
			event.ResultCardOID = fmt.Sprintf("%040x", 200+i)
		}
		if s, effects, err = StepCompletion(s, event); err != nil {
			t.Fatal(err)
		}
	}
	if effects[0].Kind != PersistReceipt {
		t.Fatal(effects)
	}
	s, effects, err = StepCompletion(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
	if err != nil || s.Outcome() != Finalized || effects[0].Kind != CleanupReceipt {
		t.Fatalf("finish %v %v", effects, err)
	}
}
