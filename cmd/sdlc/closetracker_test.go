package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// closeReady files #N with a complete design, claims it, prepares its branch,
// passes change-code and commits code: the state just before `sdlc close`.
func closeReady(t *testing.T, id int) (*trackerRepo, string, string) {
	t.Helper()
	pid := fmt.Sprintf("%06d", id)
	full := fmt.Sprintf("---\nid: %s\nstatus: open\ndeps: []\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# e2e\n\n"+
		"## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n", pid)
	card, detail, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	cardPath, detailPath := tracker.CardPath(pid, "e2e"), syncIssuesDir+"/"+pid+"-e2e.md"
	r := newTrackerRepo(t, map[string]string{cardPath: string(card)}, map[string]string{detailPath: string(detail)})
	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand(args...); err != nil {
			t.Fatalf("sdlc %s: %v\n%s", strings.Join(args, " "), err, stderr)
		}
	}
	idArg := fmt.Sprint(id)
	run("claim", "--issue", idArg)
	run("start-plan", "--issue", idArg)
	run("change-code", "--issue", idArg, "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
	writeRepoFile(t, r.root, "cmd/a.go", "package a\n")
	r.git("add", "cmd/a.go")
	r.git("commit", "-qm", fmt.Sprintf("#%d: implement", id))
	return r, cardPath, detailPath
}

func TestTrackerCloseCommitsEvidenceAndPublishesBoundCard(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 301)
	writeRepoFile(t, r.root, "unrelated.go", "package u\n")
	r.git("add", "unrelated.go") // staged unrelated work must stay out of the evidence
	reviewed := r.git("rev-parse", "HEAD")
	mainBefore := r.originMain()
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "301", "--verified", "e2e", "--actual", "1.5", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	evidence := r.git("rev-parse", "HEAD")
	if r.git("rev-parse", "HEAD^") != reviewed {
		t.Fatal("the evidence commit is not directly on the reviewed commit")
	}
	msg := r.git("log", "-1", "--format=%B")
	for _, want := range []string{"#301: close", "Review-Verdict: SHIP", "Close-Actual: 1.5"} {
		if !strings.Contains(msg, want) {
			t.Errorf("evidence message lacks %q:\n%s", want, msg)
		}
	}
	files := r.git("show", "--name-only", "--format=", "HEAD")
	if !strings.Contains(files, detailPath) || strings.Contains(files, "unrelated.go") || strings.Contains(files, "cmd/a.go") {
		t.Fatalf("evidence commit carried %q", files)
	}
	if staged := r.git("diff", "--cached", "--name-only"); staged != "unrelated.go" {
		t.Fatalf("unrelated staged work disturbed: %q", staged)
	}
	card := r.card(cardPath)
	for _, want := range []string{"status: codecomplete", "actual_hours: 1.5"} {
		if !strings.Contains(card, want) {
			t.Errorf("card lacks %q:\n%s", want, card)
		}
	}
	c, ok, err := issue.CardCompletion([]byte(card))
	if err != nil || !ok || c.EvidenceCommit != evidence || c.ReviewedHEAD != reviewed {
		t.Fatalf("binding %+v (evidence %s reviewed %s): %v", c, evidence, reviewed, err)
	}
	details := r.git("show", "HEAD:"+detailPath)
	if strings.Contains(details, "status: codecomplete") || strings.Contains(details, "actual_hours: 1.5") {
		t.Fatal("card-owned close fields were written into the details")
	}
	if !strings.Contains(details, "closed — e2e") {
		t.Fatal("the close Log line is not in the evidence")
	}
	if r.originMain() != mainBefore {
		t.Fatal("close published to main")
	}
}

func TestTrackerCloseFixThenShipLandsEvidenceAfterTheFixes(t *testing.T) {
	r, cardPath, _ := closeReady(t, 302)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nrename a thing\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "302", "--verified", "e2e", "--actual", "2", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	if card := r.card(cardPath); !strings.Contains(card, "status: working") {
		t.Fatalf("FIX-THEN-SHIP published before the fixes:\n%s", card)
	}
	var list bytes.Buffer
	if err := runRecoveryList(context.Background(), &list); err != nil || !strings.Contains(list.String(), "completion") {
		t.Fatalf("no pending close: %q %v", list.String(), err)
	}
	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	r.git("commit", "-qam", "#302: fix review finding")
	fix := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 302); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	if r.git("rev-parse", "HEAD^") != fix {
		t.Fatal("the evidence commit did not land after the fix")
	}
	c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if !ok || c.EvidenceCommit != r.git("rev-parse", "HEAD") || !strings.Contains(r.card(cardPath), "status: codecomplete") {
		t.Fatalf("card not bound to the post-fix evidence: %+v", c)
	}
}

// BR-25: a fix commit that sweeps every evidence file into HEAD must not wedge
// the deferred close: the evidence commit may be empty and still anchors it.
func TestTrackerCloseFixThenShipSurvivesASweepingFixCommit(t *testing.T) {
	r, cardPath, _ := closeReady(t, 303)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nrename a thing\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "303", "--verified", "e2e", "--actual", "2", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	r.git("add", "-A") // sweeps the details, ledgers and sidecar too
	r.git("commit", "-qm", "#303: fix and everything")
	fix := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 303); err != nil {
		t.Fatalf("reconcile after a sweeping fix: %v\n%s", err, errs.String())
	}
	if r.git("rev-parse", "HEAD^") != fix || r.git("rev-parse", "HEAD^{tree}") != r.git("rev-parse", fix+"^{tree}") {
		t.Fatal("expected an empty evidence commit on top of the sweeping fix")
	}
	if !strings.Contains(r.git("log", "-1", "--format=%B"), "Close-Actual: 2") {
		t.Fatal("empty evidence commit lost its trailers")
	}
	if c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath))); !ok || c.EvidenceCommit != r.git("rev-parse", "HEAD") {
		t.Fatalf("card not bound to the evidence: %+v", c)
	}
}

// BR-30: a fix commit that edits a pinned file after FIX-THEN-SHIP (here a
// Log line recording the fix) is newer than the pin: the evidence commit keeps
// HEAD's version instead of reverting it to the bytes pinned at close.
func TestTrackerCloseFixThenShipKeepsALaterEditOfAPinnedFile(t *testing.T) {
	r, _, detailPath := closeReady(t, 306)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nrename a thing\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "306", "--verified", "e2e", "--actual", "2", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	raw, err := os.ReadFile(filepath.Join(r.root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	details := string(raw)
	if !strings.Contains(details, "closed — e2e") {
		t.Fatal("fixture: close did not write its Log line")
	}
	edited := details + "- fixed the review finding\n"
	writeRepoFile(t, r.root, detailPath, edited)
	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	r.git("add", "cmd/a.go", detailPath)
	r.git("commit", "-qm", "#306: fix review finding")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 306); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	if got := r.git("show", "HEAD:"+detailPath); got != strings.TrimSpace(edited) {
		t.Fatalf("evidence reverted the later Log edit:\n%s", got)
	}
	if !strings.Contains(r.git("log", "-1", "--format=%B"), tracker.EvidenceKeptTrailer+": "+detailPath) {
		t.Fatal("the evidence commit does not name the superseded pin")
	}
	if !strings.Contains(errs.String(), "kept "+detailPath) {
		t.Fatalf("no warning for the superseded pin:\n%s", errs.String())
	}
}

// BR-21: re-closing supersedes an unstarted FIX-THEN-SHIP close instead of
// leaving two receipts that could rebind the card to the older review.
func TestTrackerReCloseSupersedesAnUnstartedClose(t *testing.T) {
	r, cardPath, _ := closeReady(t, 304)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nrename a thing\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "304", "--verified", "first", "--actual", "2", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "304", "--verified", "second", "--actual", "2", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("re-close: %v\n%s", err, stderr)
	}
	var list bytes.Buffer
	if err := runRecoveryList(context.Background(), &list); err != nil || !strings.Contains(list.String(), "no unfinished") {
		t.Fatalf("superseded close still pending: %q %v", list.String(), err)
	}
	if c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath))); !ok || c.EvidenceCommit != r.git("rev-parse", "HEAD") {
		t.Fatalf("card not bound to the re-close: %+v", c)
	}
}

// BR-22: a verdict-independent precondition refuses before the review runs.
func TestTrackerCloseRefusesPreconditionsBeforeReview(t *testing.T) {
	r, _, _ := closeReady(t, 305)
	r.git("switch", "-q", "--detach")
	calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	head := r.git("rev-parse", "HEAD")
	msg, died := expectDie(t, func() {
		_, _, _ = executeSDLCTestCommand("close", "--issue", "305", "--verified", "e2e", "--actual", "1", "--no-atlas")
	})
	if !died || !strings.Contains(msg, "check out the issue branch") {
		t.Fatalf("detached close: died=%v %q", died, msg)
	}
	if *calls != 0 || r.git("rev-parse", "HEAD") != head || r.git("status", "--porcelain") != "" {
		t.Fatal("close reviewed or wrote before refusing its precondition")
	}
}

// A legacy (unmirrored) details file in a checkout of a tracked repository
// must not take the legacy close/change-code path, which writes card fields
// into details: it refuses with the reconcile action. In a legacy repository
// the same file is the record and keeps its path.
func TestLegacyDetailsRefuseInATrackedCheckout(t *testing.T) {
	cardPath, card, detailPath, _ := seededIssue(t, "000330", "legacy")
	legacyDetails := "---\nid: 000330\nstatus: working\ncreated: 2026-09-01\n---\n\n# Seeded legacy\n\n## Problem\n\nx\n\n## Done when\n\n- y\n\n## Plan\n\n- [x] z\n\n## Log\n"
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: legacyDetails})
	abs := filepath.Join(r.root, detailPath)
	if _, err := refreshChangeCodeMirror(&changeCodeFlags{Issue: 330}, "000330-legacy", abs); !errors.Is(err, tracker.ErrLegacyDetails) || !strings.Contains(err.Error(), "migrate --reconcile") {
		t.Fatalf("change-code took the legacy path: %v", err)
	}
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	before := r.git("rev-parse", "HEAD")
	msg, died := expectDie(t, func() {
		_, _, _ = executeSDLCTestCommand("close", "--issue", "330", "--verified", "e2e", "--actual", "1", "--no-atlas")
	})
	if !died || !strings.Contains(msg, "migrate --reconcile") {
		t.Fatalf("close took the legacy path: died=%v %q", died, msg)
	}
	if r.git("rev-parse", "HEAD") != before || !strings.Contains(r.git("show", "HEAD:"+detailPath), "status: working") {
		t.Fatal("the refused close wrote something")
	}

	legacyRepo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	writeRepoFile(t, legacyRepo, detailPath, legacyDetails)
	if out, err := refreshChangeCodeMirror(&changeCodeFlags{Issue: 330}, "000330-legacy", filepath.Join(legacyRepo, detailPath)); err != nil || out != nil {
		t.Fatalf("a legacy repository's details lost the legacy path: %v", err)
	}
}

// A re-close modifies the boundary gate ledger its first close created. Its
// status entry then leads with a space (" M"), sorts first among the issue's
// plan files, and must still ride the evidence commit (#259: the trimmed
// status read cut it to "orkshop/…", leaving the ledger uncommitted).
func TestTrackerReCloseCommitsTheModifiedGateLedger(t *testing.T) {
	r, _, _ := closeReady(t, 302)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	closeOnce := func() {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand("close", "--issue", "302", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
			t.Fatalf("close: %v\n%s", err, stderr)
		}
	}
	closeOnce()
	ledger := "workshop/plans/000302-e2e-close-gate.md"
	if r.git("ls-tree", "--name-only", "HEAD", "--", ledger) == "" {
		t.Fatalf("the first close did not commit its new gate ledger:\n%s", r.git("show", "--name-only", "--format=", "HEAD"))
	}
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "cmd/b.go")
	r.git("commit", "-qm", "#302: a fix after the close")
	closeOnce()
	if !strings.Contains(r.git("show", "--name-only", "--format=", "HEAD"), ledger) {
		t.Fatalf("the re-close's evidence commit lacks the modified gate ledger:\n%s", r.git("show", "--name-only", "--format=", "HEAD"))
	}
	if dirty := r.git("status", "--porcelain", "--", "workshop/plans"); dirty != "" {
		t.Fatalf("the re-close left plan files uncommitted:\n%s", dirty)
	}
}
