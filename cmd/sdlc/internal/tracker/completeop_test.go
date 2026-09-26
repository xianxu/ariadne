package tracker

import (
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
