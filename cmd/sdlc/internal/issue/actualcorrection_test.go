package issue

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestActualCorrectionPreservesCompletion(t *testing.T) {
	c := testCompletion("close-one")
	card, err := SetCardCompletion([]byte(setterCard), c)
	if err != nil {
		t.Fatal(err)
	}
	record := ActualCorrection{Version: 1, Token: "repair-one", CompletionToken: c.Token, Repository: c.Repository, EvidenceCommit: c.EvidenceCommit, ReviewedHEAD: c.ReviewedHEAD, PreviousActual: "0.05", NewActual: "2.57", Reason: "measured correction", MeasurementDigest: strings.Repeat("a", 64), Inputs: `{"version":1,"repository":"file:/r","code_head":"` + c.ReviewedHEAD + `"}`}
	record.MeasurementDigest = fmt.Sprintf("%x", sha256.Sum256([]byte(record.Inputs)))
	updated, err := SetCardActualCorrection(card, record)
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := CardActualCorrection(updated)
	if err != nil || !ok || got != record {
		t.Fatalf("got %+v %v %v", got, ok, err)
	}
	after, ok, err := CardCompletion(updated)
	if err != nil || !ok || after != c {
		t.Fatal("completion changed")
	}
	if _, err := SetCardActualCorrection(updated, record); err != nil {
		t.Fatal(err)
	}
	changed := record
	changed.Reason = "another reason"
	if _, err := SetCardActualCorrection(updated, changed); err == nil {
		t.Fatal("rewrote same operation")
	}
	changed = record
	changed.CompletionToken = "foreign"
	if _, err := SetCardActualCorrection(card, changed); err == nil {
		t.Fatal("foreign completion adopted")
	}
	changed = record
	changed.Inputs = "not-json"
	if _, err := SetCardActualCorrection(card, changed); err == nil {
		t.Fatal("invalid snapshot accepted")
	}
	changed = record
	changed.MeasurementDigest = strings.Repeat("b", 64)
	if _, err := SetCardActualCorrection(card, changed); err == nil {
		t.Fatal("mismatched snapshot digest accepted")
	}
	changed = record
	changed.IntentReason = "original intent"
	if _, err := SetCardActualCorrection(card, changed); err != nil {
		t.Fatal(err)
	}
	changed.IntentReason = strings.Repeat("x", 4097)
	if _, err := SetCardActualCorrection(card, changed); err == nil {
		t.Fatal("unbounded intent reason accepted")
	}
	changed.IntentReason = "invalid\x00intent"
	if _, err := SetCardActualCorrection(card, changed); err == nil {
		t.Fatal("NUL intent reason accepted")
	}
}

func TestCompletionGenerationRetiresOnlyItsPriorActualCorrection(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude=%v", exclude), func(t *testing.T) {
			completion := testCompletion("close-old")
			card, err := SetCardCompletion([]byte(setterCard), completion)
			if err != nil {
				t.Fatal(err)
			}
			inputs := `{"version":1,"repository":"file:/r","code_head":"` + completion.ReviewedHEAD + `"}`
			record := ActualCorrection{Version: 1, Token: "repair-old", CompletionToken: completion.Token, Repository: completion.Repository, EvidenceCommit: completion.EvidenceCommit, ReviewedHEAD: completion.ReviewedHEAD, PreviousActual: "0.05", NewActual: "2.57", Reason: "repair old close", Inputs: inputs, MeasurementDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(inputs))), Exclude: exclude}
			if exclude {
				record.NewActual = "N/A"
			}
			corrected, err := SetCardActualCorrection(card, record)
			if err != nil {
				t.Fatal(err)
			}
			oldEvidence := append([]byte(nil), corrected...)
			unchanged, err := SetCardCompletion(corrected, completion)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok, err := CardActualCorrection(unchanged); err != nil || !ok || got != record {
				t.Fatalf("same-generation repeat lost correction: %+v %v %v", got, ok, err)
			}
			landed := completion
			landed.LandedCommit = strings.Repeat("e", 40)
			landedCard, err := SetCardCompletion(corrected, landed)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok, err := CardActualCorrection(landedCard); err != nil || !ok || got != record {
				t.Fatalf("same-generation landing lost correction: %+v %v %v", got, ok, err)
			}
			if got, ok, err := CardCompletion(landedCard); err != nil || !ok || got != landed {
				t.Fatal("landing binding changed")
			}
			rewritten := landed
			rewritten.EvidenceCommit = strings.Repeat("f", 40)
			beforeRejected := append([]byte(nil), landedCard...)
			if out, err := SetCardCompletion(landedCard, rewritten); err == nil || out != nil {
				t.Fatal("same-token rewrite emitted a replacement card")
			}
			if !bytes.Equal(landedCard, beforeRejected) {
				t.Fatal("rejected rewrite mutated its input")
			}
			if got, ok, err := CardActualCorrection(landedCard); err != nil || !ok || got != record {
				t.Fatal("rejected rewrite retired correction")
			}
			newCompletion := testCompletion("close-new")
			newCompletion.ReviewedHEAD = strings.Repeat("a", 40)
			newCompletion.EvidenceCommit = strings.Repeat("b", 40)
			newCard, err := SetCardCompletion(landedCard, newCompletion)
			if err != nil {
				t.Fatal(err)
			}
			if got, ok, err := CardActualCorrection(newCard); err != nil || ok {
				t.Fatalf("new completion retained prior-generation correction: %+v %v %v", got, ok, err)
			}
			if got, ok, err := CardCompletion(newCard); err != nil || !ok || got != newCompletion {
				t.Fatal("new completion did not replace old generation")
			}
			if !bytes.Equal(corrected, oldEvidence) {
				t.Fatal("new generation rewrote immutable old evidence")
			}
			if got, ok, err := CardActualCorrection(oldEvidence); err != nil || !ok || got != record {
				t.Fatal("old evidence no longer retains old correction")
			}
		})
	}
}
