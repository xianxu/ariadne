package main

import (
	"os"
	"path/filepath"
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

// #304 Task 6(d): a legacy repository (no tracker) keeps the paste protocol — the
// milestone close makes NO commit — but its ledger round is still stamped and the
// reviewed head is pinned, so the next window and a rebase both work there too.
func TestMilestoneClose_LegacyStampsWithoutCommitting(t *testing.T) {
	issues := closeRepo(t, 69)
	plans := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plans, 0o755); err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(testfix.Capture(t, "", "rev-parse", "HEAD"))
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n\n```findings\nfindings: []\n```\n")
	var stdout, stderr string
	var err error
	if msg, died := expectDie(t, func() {
		stdout, stderr, err = executeSDLCTestCommand("milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1",
			"--verified", "e2e", "--no-atlas", "--issues-dir", issues, "--plans-dir", plans, "--brain-dir", "../nonexistent-brain")
	}); died || err != nil {
		t.Fatalf("milestone-close: died=%v %q err=%v\n%s", died, msg, err, stderr)
	}
	if now := strings.TrimSpace(testfix.Capture(t, "", "rev-parse", "HEAD")); now != head {
		t.Fatalf("a legacy milestone close must not commit (HEAD %s → %s)", head, now)
	}
	if !strings.Contains(stdout, "paste into commit message") {
		t.Fatalf("legacy keeps the paste block:\n%s", stdout)
	}
	l, lerr := readBoundaryGateLedger(plans, "000069-x.md", 69)
	if lerr != nil || len(l.Rounds) == 0 || l.Rounds[len(l.Rounds)-1].Reviewed != head {
		t.Fatalf("the round must record the reviewed head: %+v %v", l.Rounds, lerr)
	}
	if pin := strings.TrimSpace(testfix.Capture(t, "", "rev-parse", "refs/sdlc/reviewed/000069/M1")); pin != head {
		t.Fatalf("pin = %q, want %s", pin, head)
	}
}
