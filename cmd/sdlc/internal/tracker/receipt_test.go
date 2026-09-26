package tracker

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func preparedCandidate(binding Binding, oid string) Event {
	event := Event{Kind: EventCandidatePrepared, Binding: binding, CandidateOID: oid}
	switch binding.Stage {
	case "creation.reserve", "transfer.handoff", "transfer.record", "completion.codecomplete":
		event.CandidateBaseOID = binding.TrackerBase
	case "transfer.main":
		event.CandidateBaseOID = binding.MainBase
	}
	return event
}

func TestCandidatePreparationBindsObservedParentAcrossStages(t *testing.T) {
	for _, kind := range []string{"creation", "transfer", "completion"} {
		t.Run(kind, func(t *testing.T) {
			h := transactionForTest(t, kind)
			effects, err := h.step(Event{Kind: EventBegin, Binding: h.binding()})
			if err != nil {
				t.Fatal(err)
			}
			for tick := 0; len(effects) > 0 && tick < 40; tick++ {
				effect := effects[0]
				event := Event{Binding: h.binding()}
				switch effect.Kind {
				case PrepareCandidate:
					before := h.receipt().Spec()
					event = preparedCandidate(h.binding(), fmt.Sprintf("%040x", 100+tick))
					remote := event.CandidateBaseOID != ""
					if remote {
						event.CandidateBaseOID = fmt.Sprintf("%040x", 700+tick)
					}
					effects, err = h.step(event)
					if err != nil {
						t.Fatal(err)
					}
					after := h.receipt().Spec()
					want := before
					if remote {
						if effect.Stage == "transfer.main" {
							want.MainBase = event.CandidateBaseOID
						} else {
							want.TrackerBase = event.CandidateBaseOID
						}
					}
					if after != want {
						t.Fatalf("preparation changed source/identity or wrong destination:\n got %+v\nwant %+v", after, want)
					}
					if effects[0].Kind != PersistReceipt || effects[0].Expected != h.binding() {
						t.Fatal("candidate parent not pinned before persistence")
					}
					if remote {
						raw, err := MarshalReceipt(h.receipt())
						if err != nil {
							t.Fatal(err)
						}
						r, err := ParseReceipt(raw, before.Repository)
						if err != nil {
							t.Fatal(err)
						}
						resumed := transactionForTest(t, kind)
						if err = resumed.resume(r); err != nil {
							t.Fatal(err)
						}
						if resumed.receipt().Spec() != want || resumed.binding() != h.binding() {
							t.Fatal("resume lost observed parent")
						}
						probe, err := resumed.step(Event{Kind: EventBegin, Binding: resumed.binding()})
						if err != nil || probe[0].Kind != ProbePublication {
							t.Fatal("resume did not probe pinned candidate")
						}
					}
					continue
				case PersistReceipt:
					event.Kind = EventReceiptSaved
				case PublishCard, PublishMain, MaterializeDetail, WriteEvidence, RemoveSource:
					event.Kind = EventConfirmed
					event.CandidateOID = effect.CandidateOID
					if effect.Kind == PublishCard {
						event.ResultCardOID = fmt.Sprintf("%040x", 400+tick)
					}
				case CleanupReceipt:
					event.Kind = EventCleanupConfirmed
				default:
					t.Fatalf("unexpected effect %s", effect.Kind)
				}
				effects, err = h.step(event)
				if err != nil {
					t.Fatal(err)
				}
			}
			if len(effects) != 0 || h.receipt().Outcome() != Finalized {
				t.Fatal("operation did not finish")
			}
		})
	}
}

func TestCandidatePreparationCannotRefreshStaleCardOrSource(t *testing.T) {
	h := transactionForTest(t, "transfer")
	_, _ = h.step(Event{Kind: EventBegin, Binding: h.binding()})
	before, _ := MarshalReceipt(h.receipt())
	for _, mutate := range []func(*Binding){func(b *Binding) { b.CardOID = strings.Repeat("9", 40) }, func(b *Binding) { b.SourceHEAD = strings.Repeat("9", 40) }, func(b *Binding) { b.SourceBlob = strings.Repeat("9", 40) }, func(b *Binding) { b.Repository = "other/repo" }, func(b *Binding) { b.Token = "other" }, func(b *Binding) { b.IssueID = "000002" }} {
		event := preparedCandidate(h.binding(), strings.Repeat("7", 40))
		event.CandidateBaseOID = strings.Repeat("8", 40)
		mutate(&event.Binding)
		if effects, err := h.step(event); err == nil || len(effects) != 0 {
			t.Fatal("observed parent bypassed stale generation refusal")
		}
		after, _ := MarshalReceipt(h.receipt())
		if !bytes.Equal(before, after) {
			t.Fatal("refused event changed receipt")
		}
	}
	for _, base := range []string{"", "bad", strings.Repeat("0", 40), strings.Repeat("a", 64)} {
		event := preparedCandidate(h.binding(), strings.Repeat("7", 40))
		event.CandidateBaseOID = base
		if _, err := h.step(event); err == nil {
			t.Fatalf("accepted invalid parent %q", base)
		}
	}
}

func TestCandidatePreparationParentDoesNotSpendPublicationAttempt(t *testing.T) {
	h := transactionForTest(t, "creation")
	_, _ = h.step(Event{Kind: EventBegin, Binding: h.binding()})
	for attempt := 0; attempt < 3; attempt++ {
		event := preparedCandidate(h.binding(), fmt.Sprintf("%040x", 100+attempt))
		event.CandidateBaseOID = fmt.Sprintf("%040x", 700+attempt)
		if _, err := h.step(event); err != nil {
			t.Fatal(err)
		}
		if _, err := h.step(Event{Kind: EventReceiptSaved, Binding: h.binding()}); err != nil {
			t.Fatal(err)
		}
		effects, err := h.step(Event{Kind: EventRefRace, Binding: h.binding()})
		if attempt == 2 {
			if err == nil || len(effects) != 0 {
				t.Fatal("granted fourth publication attempt")
			}
			return
		}
		if err != nil || effects[0].Kind != RefreshInputs {
			t.Fatalf("observed parent consumed retry %d: %v", attempt, err)
		}
		spec := h.receipt().Spec()
		if _, err := h.step(Event{Kind: EventRevalidated, Binding: h.binding(), Replacement: &spec}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCandidatePreparationParentIsRemotePreparationOnly(t *testing.T) {
	h := transactionForTest(t, "completion")
	_, _ = h.step(Event{Kind: EventBegin, Binding: h.binding()})
	before, _ := MarshalReceipt(h.receipt())
	event := preparedCandidate(h.binding(), strings.Repeat("7", 40))
	event.CandidateBaseOID = strings.Repeat("8", 40)
	if effects, err := h.step(event); err == nil || len(effects) != 0 {
		t.Fatal("local evidence preparation refreshed remote parent")
	}
	event = Event{Kind: EventBegin, Binding: h.binding(), CandidateBaseOID: strings.Repeat("8", 40)}
	if effects, err := h.step(event); err == nil || len(effects) != 0 {
		t.Fatal("another event variant refreshed remote parent")
	}
	after, _ := MarshalReceipt(h.receipt())
	if !bytes.Equal(before, after) {
		t.Fatal("refusal modified source or intent")
	}
}

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
	if _, _, err := StepTransfer(s, preparedCandidate(s.Binding(), strings.Repeat("1", 64))); err == nil {
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
		spec.EvidenceMessage = "#1: close\n\nClose-Actual: 1"
		spec.EvidencePaths = spec.SourcePath
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
		"completion":         {WriteEvidence, PublishCard},
	}
	for kind, expected := range oracles {
		// Baseline plus an interruption at each effect in that baseline. Two
		// recovery observations model effect-applied and confirmed-not-applied.
		baselineEffects := 3*len(expected) + 2
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
							event = preparedCandidate(h.binding(), fmt.Sprintf("%040x", 100+mutationIndex))
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
						case PublishCard, PublishMain, MaterializeDetail, WriteEvidence, RemoveSource:
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
