package tracker

import (
	"fmt"
	"testing"
)

func TestCompletionRequiresEvidenceExactLandingAndDoneBeforeArchive(t *testing.T) {
	spec := operationSpec()
	spec.ReviewedHEAD = spec.SourceHEAD
	s, err := NewCompletion(spec)
	if err != nil {
		t.Fatal(err)
	}
	s, effects, err := StepCompletion(s, Event{Kind: EventBegin, Binding: s.Binding()})
	if err != nil {
		t.Fatal(err)
	}
	expected := []EffectKind{WriteEvidence, PublishCard, ObserveLanding, PublishCard, ArchiveDetails}
	for i, kind := range expected {
		if kind == ObserveLanding {
			if effects[0].Kind != ObserveLanding {
				t.Fatalf("landing %v", effects)
			}
			proof := LandingEvidence{Repository: spec.Repository, ReviewedHEAD: spec.ReviewedHEAD, EvidenceOID: s.Binding().EvidenceOID, LandedHEAD: s.Binding().EvidenceOID, IntegrationOID: fmt.Sprintf("%040x", 500)}
			wrong := proof
			wrong.ReviewedHEAD = fmt.Sprintf("%040x", 999)
			if _, _, err := StepCompletion(s, Event{Kind: EventLandingConfirmed, Binding: s.Binding(), Landing: &wrong}); err == nil {
				t.Fatal("accepted different reviewed generation")
			}
			s, effects, err = StepCompletion(s, Event{Kind: EventLandingConfirmed, Binding: s.Binding(), Landing: &proof})
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		if effects[0].Kind != PrepareCandidate {
			t.Fatalf("prepare %v", effects)
		}
		oid := fmt.Sprintf("%040x", 100+i)
		s, effects, err = StepCompletion(s, Event{Kind: EventCandidatePrepared, Binding: s.Binding(), CandidateOID: oid})
		if err != nil || effects[0].Kind != PersistReceipt {
			t.Fatalf("save %v %v", effects, err)
		}
		s, effects, err = StepCompletion(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
		if err != nil || effects[0].Kind != kind {
			t.Fatalf("apply %v %v", effects, err)
		}
		s, _, err = StepCompletion(s, Event{Kind: EventUnknown, Binding: s.Binding()})
		if err != nil {
			t.Fatal(err)
		}
		event := Event{Kind: EventConfirmed, Binding: s.Binding(), CandidateOID: oid}
		if kind == PublishCard {
			event.ResultCardOID = fmt.Sprintf("%040x", 200+i)
		}
		s, effects, err = StepCompletion(s, event)
		if err != nil {
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
