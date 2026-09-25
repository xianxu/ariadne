package tracker

import (
	"fmt"
	"strings"
	"testing"
)

func TestTransferGeneratedInterruptionSequences(t *testing.T) {
	// Each bit interrupts one mutation after its receipt is saved. Enumerate all
	// combinations, including interruption at every declared mutation in one run.
	for mask := 0; mask < 16; mask++ {
		t.Run(fmt.Sprint(mask), func(t *testing.T) {
			s, err := NewTransfer(operationSpec())
			if err != nil {
				t.Fatal(err)
			}
			s, effects, err := StepTransfer(s, Event{Kind: EventBegin, Binding: s.Binding()})
			if err != nil {
				t.Fatal(err)
			}
			expected := []EffectKind{PublishCard, PublishMain, PublishCard, RemoveSource}
			for i, kind := range expected {
				if len(effects) != 1 || effects[0].Kind != PrepareCandidate {
					t.Fatalf("prepare %d: %v", i, effects)
				}
				oid := fmt.Sprintf("%040x", 100+i)
				s, effects, err = StepTransfer(s, preparedCandidate(s.Binding(), oid))
				if err != nil || effects[0].Kind != PersistReceipt {
					t.Fatalf("receipt %v %v", effects, err)
				}
				// A restart from the durable pre-mutation receipt must probe, never blindly
				// repeat the mutation, even when the process died before it was attempted.
				raw, err := MarshalReceipt(s.Receipt())
				if err != nil {
					t.Fatal(err)
				}
				parsed, err := ParseReceipt(raw, operationSpec().Repository)
				if err != nil {
					t.Fatal(err)
				}
				recovered, err := ResumeTransfer(parsed)
				if err != nil {
					t.Fatal(err)
				}
				_, probe, err := StepTransfer(recovered, Event{Kind: EventBegin, Binding: recovered.Binding()})
				if err != nil || probe[0].Kind != ProbePublication {
					t.Fatalf("unsafe replay %v %v", probe, err)
				}
				s, effects, err = StepTransfer(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
				if err != nil || effects[0].Kind != kind {
					t.Fatalf("mutation %d: %v %v", i, effects, err)
				}
				if kind == RemoveSource && s.Receipt().ConfirmedStages() != 3 {
					t.Fatal("removal before provenance publication")
				}
				if mask&(1<<i) != 0 {
					s, effects, err = StepTransfer(s, Event{Kind: EventUnknown, Binding: s.Binding()})
					if err != nil || len(effects) != 0 || s.Outcome() != Unconfirmed {
						t.Fatalf("unknown %v %v", effects, err)
					}
					s, effects, err = StepTransfer(s, Event{Kind: EventBegin, Binding: s.Binding()})
					if err != nil || effects[0].Kind != ProbePublication {
						t.Fatalf("probe %v %v", effects, err)
					}
					s, effects, err = StepTransfer(s, Event{Kind: EventProbeUnknown, Binding: s.Binding()})
					if err != nil || len(effects) != 0 {
						t.Fatalf("failed probe %v %v", effects, err)
					}
				}
				event := Event{Kind: EventConfirmed, Binding: s.Binding(), CandidateOID: oid}
				if kind == PublishCard {
					event.ResultCardOID = fmt.Sprintf("%040x", 200+i)
				}
				s, effects, err = StepTransfer(s, event)
				if err != nil {
					t.Fatal(err)
				}
			}
			if effects[0].Kind != PersistReceipt {
				t.Fatalf("final receipt %v", effects)
			}
			s, effects, err = StepTransfer(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
			if err != nil || s.Outcome() != Finalized || effects[0].Kind != CleanupReceipt {
				t.Fatalf("finalize %v %v", effects, err)
			}
		})
	}
}

func TestTransferRejectsStaleIdentityAtEveryEvent(t *testing.T) {
	s, _ := NewTransfer(operationSpec())
	for _, mutate := range []func(*Binding){func(b *Binding) { b.CardOID = strings.Repeat("9", 40) }, func(b *Binding) { b.SourceBlob = strings.Repeat("9", 40) }, func(b *Binding) { b.SourceHEAD = strings.Repeat("9", 40) }, func(b *Binding) { b.Repository = "another/repo" }, func(b *Binding) { b.Token = "other" }} {
		binding := s.Binding()
		mutate(&binding)
		if _, effects, err := StepTransfer(s, Event{Kind: EventBegin, Binding: binding}); err == nil || len(effects) != 0 {
			t.Fatal("accepted stale binding")
		}
	}
}

func TestTransferRestFinalizationCanUsePublishedMainCommit(t *testing.T) {
	s, _ := NewTransfer(operationSpec())
	s, _, _ = StepTransfer(s, Event{Kind: EventBegin, Binding: s.Binding()})
	for i := 0; i < 3; i++ {
		oid := fmt.Sprintf("%040x", 100+i)
		var err error
		s, _, err = StepTransfer(s, preparedCandidate(s.Binding(), oid))
		if err != nil {
			t.Fatal(err)
		}
		s, _, err = StepTransfer(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
		if err != nil {
			t.Fatal(err)
		}
		event := Event{Kind: EventConfirmed, Binding: s.Binding(), CandidateOID: oid}
		if i != 1 {
			event.ResultCardOID = fmt.Sprintf("%040x", 200+i)
		}
		s, _, err = StepTransfer(s, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	mainOID := fmt.Sprintf("%040x", 101)
	s, effects, err := StepTransfer(s, preparedCandidate(s.Binding(), mainOID))
	if err != nil || effects[0].Kind != PersistReceipt {
		t.Fatalf("rest finalization candidate rejected: %v %v", effects, err)
	}
	if _, err := MarshalReceipt(s.Receipt()); err != nil {
		t.Fatal(err)
	}
}
