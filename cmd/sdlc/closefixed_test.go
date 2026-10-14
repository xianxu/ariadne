package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestFixedThenShipRefusal(t *testing.T) {
	inter := reviewWindow{Kind: windowInterdiff, Base: "s", HumanBase: "h"}
	codeConflict := inter
	codeConflict.Conflicted = []string{"workshop/lessons.md", "cmd/sdlc/close.go"}
	docConflict := inter
	docConflict.Conflicted = []string{"workshop/lessons.md"}
	cases := []struct {
		name      string
		w         reviewWindow
		upgraded  bool
		suite     string
		wantEmpty bool
		wantHave  string
	}{
		{"no window: names the reason and the full close", reviewWindow{Kind: windowNone, Note: "no finalized close review to fix against"}, false, "", false, "sdlc close --issue 7 --verified"},
		{"quick shell crossed: full review due", inter, true, "", false, "quick-flow shell"},
		{"a code conflict needs the full suite", codeConflict, false, "", false, "--full-suite"},
		{"a code conflict with suite evidence proceeds", codeConflict, false, "make test: ok", true, ""},
		{"a non-code conflict proceeds", docConflict, false, "", true, ""},
		{"plain fixes proceed", inter, false, "", true, ""},
	}
	for _, c := range cases {
		got := fixedThenShipRefusal("7", c.w, c.upgraded, c.suite)
		if c.wantEmpty != (got == "") || !strings.Contains(got, c.wantHave) {
			t.Errorf("%s: refusal = %q", c.name, got)
		}
	}
	if got := fixedThenShipRefusal("7", codeConflict, false, ""); !strings.Contains(got, "cmd/sdlc/close.go") || strings.Contains(got, "lessons.md") {
		t.Errorf("the code-conflict refusal must name only the code paths: %q", got)
	}
}

func TestFixedThenShipNotesAndTag(t *testing.T) {
	if n := fixedThenShipNotes(reviewWindow{Kind: windowInterdiff}); n != nil {
		t.Errorf("no conflicts, no notes: %v", n)
	}
	if n := fixedThenShipNotes(reviewWindow{Conflicted: []string{"a.go", "b.md"}}); len(n) != 1 || !strings.Contains(n[0], "a.go, b.md") {
		t.Errorf("conflict note = %v", n)
	}
	if got := fixedThenShipLogTag("  "); got != " (fixed-then-ship)" {
		t.Errorf("tag = %q", got)
	}
	if got := fixedThenShipLogTag("make test ok"); got != " (fixed-then-ship; full suite: make test ok)" {
		t.Errorf("tag = %q", got)
	}
}

func TestDoneWhenDecision(t *testing.T) {
	for _, c := range []struct {
		stale, review, waived bool
		want                  doneWhenAction
	}{
		{false, false, false, doneWhenFresh},
		{false, true, true, doneWhenFresh},
		{true, true, false, doneWhenDefer},
		{true, true, true, doneWhenDefer}, // the review runs, so it judges; nothing to waive
		{true, false, false, doneWhenRefuse},
		{true, false, true, doneWhenWaived},
	} {
		if got := doneWhenDecision(c.stale, c.review, c.waived); got != c.want {
			t.Errorf("doneWhenDecision(stale=%v, review=%v, waived=%v) = %d, want %d", c.stale, c.review, c.waived, got, c.want)
		}
	}
}

// fixedReview is a reviewer response that finalizes and carries a findings fence, so
// its round stamps the reviewed head the next --fixed-then-ship anchors on.
func fixedReview(verdict string) string {
	if verdict == "REWORK" {
		return "VERDICT: REWORK\n" + reworkBoundaryFindings
	}
	return "VERDICT: " + verdict + " (confidence: high)\n\nfine\n\n```findings\nfindings: []\n```\n"
}

func closeArgs(id string, extra ...string) []string {
	return append([]string{"close", "--issue", id, "--verified", "e2e", "--actual", "2", "--no-atlas"}, extra...)
}

// The A6 shape (#183): a SHIP close, then a post-close code fix. The publish gate
// refuses and names --fixed-then-ship; that close reviews only the fix with the
// fix-review recipe, rebinds the card to a new evidence commit, and the publish passes.
func TestFixedThenShip_PostCloseFixPublishes(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 401)
	stubJudge(t, fixedReview("SHIP"))
	if _, stderr, err := executeSDLCTestCommand(closeArgs("401")...); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	first, _, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	reviewed := r.git("rev-parse", first.EvidenceCommit+"^")

	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "-A")
	r.git("commit", "-qm", "#401: fix review finding")
	base := r.git("merge-base", "origin/main", "HEAD")
	err := runPublishGate(context.Background(), base, "workshop/issues", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "--fixed-then-ship") {
		t.Fatalf("publish after a post-close fix: %v; want a refusal naming --fixed-then-ship", err)
	}

	calls, prompt := stubJudge(t, fixedReview("SHIP"))
	if _, stderr, err := executeSDLCTestCommand(closeArgs("401", "--fixed-then-ship")...); err != nil {
		t.Fatalf("close --fixed-then-ship: %v\n%s", err, stderr)
	}
	if *calls != 1 || !strings.Contains(*prompt, "## Fix review") {
		t.Fatalf("calls=%d; want one fix-review dispatch:\n%s", *calls, *prompt)
	}
	l, err := readBoundaryGateLedger("workshop/plans", filepath.Base(detailPath), 401)
	if err != nil || len(l.Rounds) == 0 {
		t.Fatalf("ledger: %v", err)
	}
	last := l.Rounds[len(l.Rounds)-1]
	if last.Recipe != string(judge.FixReview) {
		t.Errorf("the fix round must stamp the fix-review recipe: %+v", last)
	}
	if want := r.git("rev-parse", evidenceRev+"^"); last.Reviewed != want {
		t.Errorf("the fix round reviewed %q, want the fix head %q", last.Reviewed, want)
	}
	second, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if !ok || second.EvidenceCommit == first.EvidenceCommit || second.EvidenceCommit != r.git("rev-parse", evidenceRev) {
		t.Fatalf("card not rebound to the fix close's evidence: first=%+v second=%+v", first, second)
	}
	if !strings.Contains(r.git("log", "-1", "--format=%B", evidenceRev), "Review-Window: "+reviewed[:7]) {
		t.Errorf("the trailer must name the reviewed close head as the window base:\n%s", r.git("log", "-1", "--format=%B", evidenceRev))
	}
	if err := runPublishGate(context.Background(), base, "workshop/issues", io.Discard); err != nil {
		t.Fatalf("publish after --fixed-then-ship: %v", err)
	}
}

// After a FIX-THEN-SHIP close (receipt stored unstarted), --fixed-then-ship reviews the
// fixes and supersedes that receipt; REWORK publishes nothing.
func TestFixedThenShip_AfterFixThenShipVerdict(t *testing.T) {
	r, cardPath, _ := closeReady(t, 402)
	stubJudge(t, fixedReview("FIX-THEN-SHIP"))
	if _, stderr, err := executeSDLCTestCommand(closeArgs("402")...); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	r.git("commit", "-qam", "#402: fix review finding")

	stubJudge(t, fixedReview("REWORK"))
	if _, _, err := executeSDLCTestCommand(closeArgs("402", "--fixed-then-ship")...); err == nil {
		t.Fatal("a REWORK fix review must not finalize")
	}
	if card := r.card(cardPath); strings.Contains(card, "status: codecomplete") {
		t.Fatalf("REWORK published the card:\n%s", card)
	}

	stubJudge(t, "VERDICT: SHIP\n"+addressedBoundaryFindings)
	if _, stderr, err := executeSDLCTestCommand(closeArgs("402", "--fixed-then-ship")...); err != nil {
		t.Fatalf("close --fixed-then-ship: %v\n%s", err, stderr)
	}
	c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if !ok || c.EvidenceCommit != r.git("rev-parse", evidenceRev) || !strings.Contains(r.card(cardPath), "status: codecomplete") {
		t.Fatalf("card not codecomplete on the fix close's evidence: %+v", c)
	}
}

// With no finalized close review there is nothing to fix against: refuse, name the
// full close, run no review.
func TestFixedThenShip_RefusesWithoutACloseReview(t *testing.T) {
	closeReady(t, 403)
	calls, _ := stubJudge(t, fixedReview("SHIP"))
	msg, died := expectDie(t, func() { executeSDLCTestCommand(closeArgs("403", "--fixed-then-ship")...) })
	if !died || !strings.Contains(msg, "no finalized close review") || !strings.Contains(msg, "sdlc close --issue 403") || *calls != 0 {
		t.Fatalf("died=%v calls=%d msg=%q", died, *calls, msg)
	}
	for _, extra := range [][]string{{"--fixed-then-ship", "--no-judge"}, {"--full-suite", "ok"}} {
		if msg, died := expectDie(t, func() { executeSDLCTestCommand(closeArgs("403", extra...)...) }); !died || *calls != 0 {
			t.Errorf("%v: died=%v calls=%d msg=%q, want a refusal before any review", extra, died, *calls, msg)
		}
	}
}

func TestContractChangeNotes(t *testing.T) {
	body := func(spec, rev, done string) string {
		return "# x\n\n## Spec\n\n" + spec + "\n\n## Done when\n\n" + done + "\n\n## Plan\n\n- [x] a\n\n## Log\n\n- l\n" + rev
	}
	before := body("A thing.", "", "- it works")
	if n := contractChangeNotes(before, body("A   thing.", "", "- it works")+"- another log line\n"); n != nil {
		t.Errorf("re-wrapping and Log lines are not contract changes: %v", n)
	}
	n := contractChangeNotes(before, body("A bigger thing.", "\n## Revisions\n\n- 10-10: scope grew\n", "- it works"))
	if len(n) != 2 || !strings.Contains(n[0], "`## Spec` changed") || !strings.Contains(n[0], "A bigger thing.") ||
		!strings.Contains(n[0], "A thing.") || !strings.Contains(n[1], "`## Revisions` changed") || !strings.Contains(n[1], "(empty)") {
		t.Errorf("Spec and Revisions changes must each carry before/after: %q", n)
	}
	if got := clipNote(strings.Repeat("x", maxContractNoteSection-1) + "é tail"); !utf8.ValidString(got) {
		t.Errorf("clipping split a multi-byte rune: %q", got[len(got)-40:])
	}
	long := strings.Repeat("x", maxContractNoteSection+50)
	if got := contractChangeNotes(before, body(long, "", "- it works")); len(got) != 1 || !strings.Contains(got[0], "clipped") {
		t.Errorf("a long section must be clipped: %d notes", len(got))
	}
}

// BR-1: the fix review's window excludes the issues directory, so a Spec change made
// after the close review reaches the reviewer only as a review note.
func TestFixedThenShip_ScopeChangeReachesTheReviewer(t *testing.T) {
	r, _, detailPath := closeReady(t, 404)
	stubJudge(t, fixedReview("SHIP"))
	if _, stderr, err := executeSDLCTestCommand(closeArgs("404")...); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	raw, err := os.ReadFile(filepath.Join(r.root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, detailPath, strings.Replace(string(raw), "A thing.", "A much bigger thing.", 1))
	writeRepoFile(t, r.root, "cmd/a.go", "package a // more\n")
	r.git("add", "-A")
	r.git("commit", "-qm", "#404: grow the scope")
	_, prompt := stubJudge(t, fixedReview("REWORK"))
	executeSDLCTestCommand(closeArgs("404", "--fixed-then-ship")...)
	if !strings.Contains(*prompt, "`## Spec` changed since the close review") || !strings.Contains(*prompt, "A much bigger thing.") {
		t.Fatalf("the scope change did not reach the fix review:\n%s", *prompt)
	}
}
