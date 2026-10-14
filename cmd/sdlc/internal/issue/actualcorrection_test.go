package issue

import (
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
}
