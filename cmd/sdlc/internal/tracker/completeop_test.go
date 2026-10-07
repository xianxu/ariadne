package tracker

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestCodecompleteCardBindsTheCloseAndRefusesDone(t *testing.T) {
	spec := operationSpec()
	spec.ReviewedHEAD = spec.SourceHEAD
	spec.EvidenceMessage = "#252: close\n\nReview-Verdict: SHIP\nClose-Actual: 2.5"
	working := strings.Replace(testCard, "status: open", "status: working", 1)
	evidence := strings.Repeat("e", 40)
	next, err := CodecompleteCard([]byte(working), spec, evidence, "2026-09-25")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"status: codecomplete", "actual_hours: 2.5", "updated: 2026-09-25"} {
		if !strings.Contains(string(next), want) {
			t.Errorf("card lacks %q:\n%s", want, next)
		}
	}
	c, ok, err := issue.CardCompletion(next)
	if err != nil || !ok || c.EvidenceCommit != evidence || c.Token != spec.Token || c.ReviewedHEAD != spec.ReviewedHEAD {
		t.Fatalf("binding %+v %v %v", c, ok, err)
	}
	// A re-close of codecomplete replaces the binding; done and open refuse.
	if _, err := CodecompleteCard(next, spec, evidence, "2026-09-26"); err != nil {
		t.Fatalf("re-close: %v", err)
	}
	for _, status := range []string{"open", "blocked"} {
		if _, err := CodecompleteCard([]byte(strings.Replace(testCard, "status: open", "status: "+status, 1)), spec, evidence, "d"); err == nil {
			t.Errorf("closed a %s card", status)
		}
	}
	done := strings.Replace(string(next), "status: codecomplete", "status: done", 1)
	if _, err := CodecompleteCard([]byte(done), spec, evidence, "d"); err == nil {
		t.Error("closed a done card")
	}
	for _, msg := range []string{"#1: close", "#1: close\n\nClose-Actual: 1\nClose-Actual: 2"} {
		bad := spec
		bad.EvidenceMessage = msg
		if _, err := CodecompleteCard([]byte(working), bad, evidence, "d"); err == nil {
			t.Errorf("accepted evidence message %q", msg)
		}
	}
}

func TestNewestCloseRefusesAnOlderReviewThanTheCardsClose(t *testing.T) {
	spec := operationSpec()
	spec.ReviewedHEAD = spec.SourceHEAD
	other := issue.Completion{Token: "close-newer", Repository: spec.Repository, ReviewedHEAD: strings.Repeat("9", 40), EvidenceCommit: strings.Repeat("8", 40)}
	card, err := issue.SetCardCompletion([]byte(strings.Replace(testCard, "status: open", "status: working", 1)), other)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		older bool
		want  error
	}{{true, nil}, {false, ErrSupersededClose}} {
		op := &CompletionOp{ancestor: func(a, b string) (bool, error) { return c.older, nil }}
		if err := op.newestClose(spec, card); !errors.Is(err, c.want) && !(c.want == nil && err == nil) {
			t.Errorf("card review older=%v: %v", c.older, err)
		}
	}
	same := other
	same.Token = spec.Token
	card, _ = issue.SetCardCompletion([]byte(strings.Replace(testCard, "status: open", "status: working", 1)), same)
	op := &CompletionOp{ancestor: func(a, b string) (bool, error) { t.Fatal("own binding needs no ancestry"); return false, nil }}
	if err := op.newestClose(spec, card); err != nil {
		t.Fatal(err)
	}
}

// #283: a rebase rewrites the reviewed commit of the card's close away. That
// close no longer precedes anything, so ancestry alone kept it "newer" forever
// and no later close of the issue could land. A binding the branch no longer
// contains is superseded by a receipt whose review the branch does contain; a
// stale receipt (its own review rewritten away) is still refused.
func TestNewestCloseAfterARebase(t *testing.T) {
	branch := "refs/heads/000001-x"
	onBranch, rewritten, other := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
	// ancestor over a tiny graph: the branch contains onBranch only; the two
	// reviews are unrelated (a rewritten history).
	graph := func(a, b string) (bool, error) {
		return a == b || (b == branch && a == onBranch), nil
	}
	bind := func(reviewed string) []byte {
		card, err := issue.SetCardCompletion([]byte(strings.Replace(testCard, "status: open", "status: working", 1)),
			issue.Completion{Token: "close-card", Repository: "r", ReviewedHEAD: reviewed, EvidenceCommit: strings.Repeat("8", 40)})
		if err != nil {
			t.Fatal(err)
		}
		return card
	}
	for _, c := range []struct {
		name          string
		card, receipt string
		branch        string
		want          error
	}{
		{"card's review rewritten away, receipt on the branch", rewritten, onBranch, branch, nil},
		{"stale receipt from before the rebase", onBranch, rewritten, branch, ErrSupersededClose},
		{"neither on the branch", rewritten, other, branch, ErrSupersededClose},
		{"detached: ancestry only", rewritten, onBranch, "", ErrSupersededClose},
	} {
		spec := operationSpec()
		spec.ReviewedHEAD = c.receipt
		op := &CompletionOp{branch: c.branch, ancestor: graph}
		if err := op.newestClose(spec, bind(c.card)); !errors.Is(err, c.want) && !(c.want == nil && err == nil) {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
