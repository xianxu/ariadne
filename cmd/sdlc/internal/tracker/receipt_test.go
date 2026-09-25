package tracker

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestReceiptRejectsUntrustedShapeIdentityAndImpossibleState(t *testing.T) {
	s, err := NewTransfer(operationSpec())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := MarshalReceipt(s.Receipt())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseReceipt(raw, "another/repo"); err == nil {
		t.Fatal("wrong repository accepted")
	}
	for _, bad := range [][]byte{
		bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":2`), 1),
		bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1),
		bytes.Replace(raw, []byte(`"phase":"queued"`), []byte(`"phase":"finalized"`), 1),
		bytes.Replace(raw, []byte(`"stage":0`), []byte(`"stage":3`), 1),
		bytes.Replace(raw, []byte(`"issue_id":"000001"`), []byte(`"issue_id":"000000"`), 1),
		bytes.Replace(raw, []byte(`"stage":0,`), nil, 1),
		bytes.Replace(raw, []byte(`"version":1`), []byte(`"Version":1`), 1),
		bytes.Replace(raw, []byte(`"proofs":[]`), []byte(`"proofs":null`), 1),
		bytes.Replace(raw, []byte(`"stage":0`), []byte(`"stage":null`), 1),
		append(append([]byte(nil), raw...), []byte(` {}`)...),
		[]byte(strings.Repeat("x", MaxReceiptBytes+1)),
	} {
		if _, err := ParseReceipt(bad, operationSpec().Repository); err == nil {
			t.Fatalf("accepted forged receipt: %s", bad)
		}
	}
}

func TestReceiptRejectsMixedHashAlgorithmsAndInvalidRefComponents(t *testing.T) {
	for _, branch := range []string{"refs/heads/.hidden", "refs/heads/a.lock/b", "refs/heads/a\x01b"} {
		spec := operationSpec()
		spec.SourceBranch = branch
		if _, err := NewTransfer(spec); err == nil {
			t.Errorf("accepted invalid ref %q", branch)
		}
	}
	s, _ := NewTransfer(operationSpec())
	s, _, _ = StepTransfer(s, Event{Kind: EventBegin, Binding: s.Binding()})
	if _, _, err := StepTransfer(s, Event{Kind: EventCandidatePrepared, Binding: s.Binding(), CandidateOID: strings.Repeat("1", 64)}); err == nil {
		t.Fatal("accepted mixed-algorithm candidate")
	}
}

func TestReceiptEventTagsDoNotAcceptForeignPayload(t *testing.T) {
	s, _ := NewTransfer(operationSpec())
	if _, _, err := StepTransfer(s, Event{Kind: EventBegin, Binding: s.Binding(), CandidateOID: strings.Repeat("9", 40)}); err == nil {
		t.Fatal("untagged payload accepted")
	}
}

// This harness calls the three public transition functions, not their shared
// implementation. Its independent oracle is the required mutation order.
type transactionHarness struct {
	receipt func() Receipt
	binding func() Binding
	step    func(Event) ([]Effect, error)
	resume  func(Receipt) error
}

func transactionForTest(t *testing.T, kind string) transactionHarness {
	t.Helper()
	spec := operationSpec()
	switch kind {
	case "creation":
		s, err := NewCreation(spec)
		if err != nil {
			t.Fatal(err)
		}
		return transactionHarness{func() Receipt { return s.Receipt() }, func() Binding { return s.Binding() }, func(e Event) ([]Effect, error) {
			next, effects, err := StepCreation(s, e)
			if err == nil {
				s = next
			}
			return effects, err
		}, func(r Receipt) error {
			next, err := ResumeCreation(r)
			if err == nil {
				s = next
			}
			return err
		}}
	case "transfer", "transfer-from-card":
		if kind == "transfer-from-card" {
			spec.Source = CardSource
		}
		s, err := NewTransfer(spec)
		if err != nil {
			t.Fatal(err)
		}
		return transactionHarness{func() Receipt { return s.Receipt() }, func() Binding { return s.Binding() }, func(e Event) ([]Effect, error) {
			next, effects, err := StepTransfer(s, e)
			if err == nil {
				s = next
			}
			return effects, err
		}, func(r Receipt) error {
			next, err := ResumeTransfer(r)
			if err == nil {
				s = next
			}
			return err
		}}
	case "completion":
		spec.ReviewedHEAD = spec.SourceHEAD
		s, err := NewCompletion(spec)
		if err != nil {
			t.Fatal(err)
		}
		return transactionHarness{func() Receipt { return s.Receipt() }, func() Binding { return s.Binding() }, func(e Event) ([]Effect, error) {
			next, effects, err := StepCompletion(s, e)
			if err == nil {
				s = next
			}
			return effects, err
		}, func(r Receipt) error {
			next, err := ResumeCompletion(r)
			if err == nil {
				s = next
			}
			return err
		}}
	}
	t.Fatal("unknown harness")
	return transactionHarness{}
}

func TestReceiptGeneratedInterruptionsAtEveryDeclaredEffect(t *testing.T) {
	oracles := map[string][]EffectKind{
		"creation":           {PublishCard, MaterializeDetail},
		"transfer":           {PublishCard, PublishMain, PublishCard, RemoveSource},
		"transfer-from-card": {PublishCard, PublishMain, PublishCard},
		"completion":         {WriteEvidence, PublishCard, ObserveLanding, PublishCard, ArchiveDetails},
	}
	for kind, expected := range oracles {
		// Baseline plus an interruption at each effect in that baseline. Two
		// recovery observations model effect-applied and confirmed-not-applied.
		baselineEffects := 3*len(expected) + 2
		if kind == "completion" {
			baselineEffects -= 2
		}
		for interrupt := -1; interrupt < baselineEffects; interrupt++ {
			for _, applied := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/effect-%d/applied-%v", kind, interrupt, applied), func(t *testing.T) {
					h := transactionForTest(t, kind)
					effects, err := h.step(Event{Kind: EventBegin, Binding: h.binding()})
					if err != nil {
						t.Fatal(err)
					}
					mutationIndex := 0
					interrupted := false
					var priorBindings []Binding
					for tick := 0; len(effects) > 0 && tick < 100; tick++ {
						if len(effects) != 1 {
							t.Fatal("ambiguous concurrent effects")
						}
						effect := effects[0]
						// Old stage/attempt acknowledgments must not be accepted at a
						// new stage just because repository/card/source still match.
						for _, old := range priorBindings {
							if old == h.binding() {
								continue
							}
							if got, err := h.step(Event{Kind: EventReceiptSaved, Binding: old}); err == nil || len(got) != 0 {
								t.Fatal("accepted delayed receipt acknowledgment")
							}
						}
						priorBindings = append(priorBindings, h.binding())
						if tick == interrupt && !interrupted {
							interrupted = true
							if got, err := h.step(Event{Kind: EventUnknown, Binding: h.binding()}); err != nil || len(got) != 0 {
								t.Fatalf("interrupt %s: %v %v", effect.Kind, got, err)
							}
							raw, err := MarshalReceipt(h.receipt())
							if err != nil {
								t.Fatal(err)
							}
							r, err := ParseReceipt(raw, operationSpec().Repository)
							if err != nil {
								t.Fatal(err)
							}
							if err = h.resume(r); err != nil {
								t.Fatal(err)
							}
							effects, err = h.step(Event{Kind: EventBegin, Binding: h.binding()})
							if err != nil {
								t.Fatal(err)
							}
							continue
						}
						event := Event{Binding: h.binding()}
						switch effect.Kind {
						case PrepareCandidate:
							event.Kind = EventCandidatePrepared
							event.CandidateOID = fmt.Sprintf("%040x", 100+mutationIndex)
						case PersistReceipt:
							event.Kind = EventReceiptSaved
						case RefreshInputs:
							event.Kind = EventRevalidated
							spec := h.receipt().Spec()
							event.Replacement = &spec
						case ProbePublication:
							if got, err := h.step(Event{Kind: EventProbeUnknown, Binding: h.binding()}); err != nil || len(got) != 0 {
								t.Fatal("unknown probe released an effect")
							}
							if !applied {
								event.Kind = EventNotApplied
								event.CandidateOID = effect.CandidateOID
								break
							}
							fallthrough
						case PublishCard, PublishMain, MaterializeDetail, WriteEvidence, RemoveSource, ArchiveDetails:
							want := expected[mutationIndex]
							if effect.Kind != ProbePublication && effect.Kind != want {
								t.Fatalf("out of order: got %s want %s", effect.Kind, want)
							}
							if want == RemoveSource && h.receipt().ConfirmedStages() != 3 {
								t.Fatal("removed source before main provenance")
							}
							event.Kind = EventConfirmed
							event.CandidateOID = effect.CandidateOID
							if want == PublishCard {
								event.ResultCardOID = fmt.Sprintf("%040x", 200+mutationIndex)
							}
							mutationIndex++
						case ObserveLanding:
							if expected[mutationIndex] != ObserveLanding {
								t.Fatal("landing out of order")
							}
							b := h.binding()
							event.Kind = EventLandingConfirmed
							event.Landing = &LandingEvidence{Repository: b.Repository, ReviewedHEAD: b.ReviewedHEAD, EvidenceOID: b.EvidenceOID, LandedHEAD: b.EvidenceOID, IntegrationOID: fmt.Sprintf("%040x", 500)}
							mutationIndex++
						case CleanupReceipt:
							if h.receipt().Outcome() != Finalized || mutationIndex != len(expected) {
								t.Fatal("cleanup before finalization")
							}
							event.Kind = EventCleanupConfirmed
						default:
							t.Fatalf("unmodeled effect %s", effect.Kind)
						}
						effects, err = h.step(event)
						if err != nil {
							t.Fatalf("%s: %v", effect.Kind, err)
						}
					}
					if mutationIndex != len(expected) || len(effects) != 0 {
						t.Fatal("operation failed to finish")
					}
					if interrupt >= 0 && !interrupted {
						t.Fatal("declared interruption was not exercised")
					}
				})
			}
		}
	}
}
