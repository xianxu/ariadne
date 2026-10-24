package issue

import (
	"strings"
	"testing"
)

func testHandoff() Handoff {
	oid := strings.Repeat("a", 40)
	return Handoff{Token: "move-1", Repository: "file:/r", SourceBranch: "refs/heads/x", SourceBase: oid, SourceHEAD: oid,
		SourcePath: "workshop/issues/000001-x.md", SourceBlob: oid, Destination: "workshop/issues/000001-x.md"}
}

func TestHandoffRecordIsWrittenOnceAndCompletedByItsOperation(t *testing.T) {
	h := testHandoff()
	card, err := SetCardHandoff([]byte(setterCard), h)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(card), strings.TrimSuffix(strings.SplitN(setterCard, "---\n\n", 2)[0], "")) {
		t.Fatalf("owned fields moved:\n%s", card)
	}
	got, ok, err := CardHandoff(card)
	if err != nil || !ok || got != h {
		t.Fatalf("read back %+v %v %v", got, ok, err)
	}
	if parsed, err := ParseCard(card); err != nil || parsed.Title != "Old title" {
		t.Fatalf("card with envelope invalid: %v", err)
	}
	done := h
	done.MainCommit = strings.Repeat("b", 40)
	completed, err := SetCardHandoff(card, done)
	if err != nil {
		t.Fatal(err)
	}
	if got, _, _ := CardHandoff(completed); got != done {
		t.Fatalf("completion lost: %+v", got)
	}
	other := h
	other.Token = "move-2"
	if _, err := SetCardHandoff(card, other); err == nil {
		t.Fatal("a second operation rewrote the initial handoff")
	}
	changed := done
	changed.SourceBlob = strings.Repeat("c", 40)
	if _, err := SetCardHandoff(card, changed); err == nil {
		t.Fatal("completion changed recorded provenance")
	}
	bad := h
	bad.SourceHEAD = "HEAD"
	if _, err := SetCardHandoff([]byte(setterCard), bad); err == nil {
		t.Fatal("accepted a symbolic source head")
	}
}

func TestDetailsFromCardMirrorsExactlyAndCarriesTheTemplate(t *testing.T) {
	card, _ := SetCardHandoff([]byte(setterCard), testHandoff())
	details, err := DetailsFromCard(card, "sha1", "2026-09-25")
	if err != nil {
		t.Fatal(err)
	}
	text := string(details)
	for _, want := range []string{"# Old title", "## Problem\nReport.", "## Spec", "## Done when", "## Plan", "## Log", MirrorField, "deps: []"} {
		if !strings.Contains(text, want) {
			t.Errorf("details missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "handoff") || strings.Contains(text, "tracker:") {
		t.Errorf("transaction envelope leaked into details:\n%s", text)
	}
	// The mirror names this exact card: a refresh against it is a no-op.
	if again, err := RefreshMirror(details, card, card); err != nil || string(again) != text {
		t.Fatalf("details are not an untouched projection of the card: %v", err)
	}
}

func testCompletion(token string) Completion {
	return Completion{Token: token, Repository: "file:/r", ReviewedHEAD: strings.Repeat("c", 40), EvidenceCommit: strings.Repeat("d", 40)}
}

func TestCompletionBindingCoexistsWithHandoffAndOnlyGainsItsLanding(t *testing.T) {
	card, err := SetCardHandoff([]byte(setterCard), testHandoff())
	if err != nil {
		t.Fatal(err)
	}
	c := testCompletion("close-1")
	card, err = SetCardCompletion(card, c)
	if err != nil {
		t.Fatal(err)
	}
	if h, ok, _ := CardHandoff(card); !ok || h != testHandoff() {
		t.Fatal("writing completion dropped the handoff record")
	}
	if got, ok, _ := CardCompletion(card); !ok || got != c {
		t.Fatalf("completion read back %+v", got)
	}
	landed := c
	landed.LandedCommit = strings.Repeat("e", 40)
	card, err = SetCardCompletion(card, landed)
	if err != nil {
		t.Fatal(err)
	}
	moved := landed
	moved.LandedCommit = strings.Repeat("f", 40)
	if _, err := SetCardCompletion(card, moved); err == nil {
		t.Fatal("a landed binding moved to another commit")
	}
	rewritten := c
	rewritten.EvidenceCommit = strings.Repeat("a", 40)
	if _, err := SetCardCompletion(card, rewritten); err == nil {
		t.Fatal("a close's evidence was rewritten under the same token")
	}
	reclose := testCompletion("close-2")
	if card, err = SetCardCompletion(card, reclose); err != nil {
		t.Fatalf("a new close generation: %v", err)
	}
	if got, _, _ := CardCompletion(card); got != reclose {
		t.Fatal("re-close did not replace the older generation")
	}
	if _, ok, _ := CardHandoff(card); !ok {
		t.Fatal("re-close dropped the handoff record")
	}
}
