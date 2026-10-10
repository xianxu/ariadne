package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestPinReviewed(t *testing.T) {
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	head := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))
	if w := pinReviewed("000304", "M1", head); w != "" {
		t.Fatal(w)
	}
	if w := pinReviewed("000304", "", head); w != "" {
		t.Fatal(w)
	}
	if w := pinReviewed("000305", "M1", head); w != "" {
		t.Fatal(w)
	}
	testfix.Git(t, dir, "commit", "-q", "--allow-empty", "-m", "next")
	next := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))
	if w := pinReviewed("000304", "M1", next); w != "" {
		t.Fatal(w)
	}
	if got := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "refs/sdlc/reviewed/000304/M1")); got != next {
		t.Fatalf("re-pin did not advance: %s", got)
	}
	if got := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "refs/sdlc/reviewed/000304/close")); got != head {
		t.Fatalf("whole-issue pin = %s", got)
	}
	if w := pinReviewed("000304", "M2", "HEAD"); w == "" {
		t.Fatal("an unresolved commit must not be pinned")
	}

	if w := sweepReviewedPins(func(id string) bool { return id == "000304" }); w != "" {
		t.Fatal(w)
	}
	refs := testfix.Capture(t, dir, "for-each-ref", "--format=%(refname)", "refs/sdlc/reviewed/")
	if strings.Contains(refs, "000305") || !strings.Contains(refs, "000304/M1") {
		t.Fatalf("sweep must drop only non-live ids:\n%s", refs)
	}
	if w := unpinReviewed("000304"); w != "" {
		t.Fatal(w)
	}
	if refs := strings.TrimSpace(testfix.Capture(t, dir, "for-each-ref", "refs/sdlc/reviewed/")); refs != "" {
		t.Fatalf("unpin left refs:\n%s", refs)
	}
	if w := unpinReviewed("000304"); w != "" {
		t.Fatalf("unpinning nothing must be a no-op: %s", w)
	}
}
