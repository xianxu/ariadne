package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// closeReady files #N with a complete design, claims it, prepares its branch,
// passes change-code and commits code: the state just before `sdlc close`.
// evidenceRev is a close's evidence commit: the close's mirror commit (#275)
// sits directly on it.
const evidenceRev = "HEAD^"

func closeReady(t *testing.T, id int) (*trackerRepo, string, string) {
	t.Helper()
	return closeReadyWithPlan(t, id, "- [x] do it\n")
}

// closeReadyWithPlan is closeReady with the given ## Plan body (milestone rows, say).
func closeReadyWithPlan(t *testing.T, id int, plan string) (*trackerRepo, string, string) {
	t.Helper()
	pid := fmt.Sprintf("%06d", id)
	full := fmt.Sprintf("---\nid: %s\nstatus: open\ndeps: []\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# e2e\n\n"+
		"## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n%s\n## Log\n", pid, plan)
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
	ccArgs := []string{"change-code", "--issue", idArg, "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon"}
	if strings.Contains(plan, "M1") {
		// Milestone rows infer the full flow; this fixture has no durable plan.
		ccArgs = append(ccArgs, "--force", "fixture: milestone plan without a durable plan")
	}
	run(ccArgs...)
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
	evidence := r.git("rev-parse", evidenceRev)
	if r.git("rev-parse", evidenceRev+"^") != reviewed {
		t.Fatal("the evidence commit is not directly on the reviewed commit")
	}
	msg := r.git("log", "-1", "--format=%B", evidenceRev)
	for _, want := range []string{"#301: close", "Review-Verdict: SHIP", "Close-Actual: 1.5"} {
		if !strings.Contains(msg, want) {
			t.Errorf("evidence message lacks %q:\n%s", want, msg)
		}
	}
	files := r.git("show", "--name-only", "--format=", evidenceRev)
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
	details := r.git("show", evidenceRev+":"+detailPath)
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
	if r.git("rev-parse", evidenceRev+"^") != fix {
		t.Fatal("the evidence commit did not land after the fix")
	}
	c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if !ok || c.EvidenceCommit != r.git("rev-parse", evidenceRev) || !strings.Contains(r.card(cardPath), "status: codecomplete") {
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
	if r.git("rev-parse", evidenceRev+"^") != fix || r.git("rev-parse", evidenceRev+"^{tree}") != r.git("rev-parse", fix+"^{tree}") {
		t.Fatal("expected an empty evidence commit on top of the sweeping fix")
	}
	if !strings.Contains(r.git("log", "-1", "--format=%B", evidenceRev), "Close-Actual: 2") {
		t.Fatal("empty evidence commit lost its trailers")
	}
	if c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath))); !ok || c.EvidenceCommit != r.git("rev-parse", evidenceRev) {
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
	if got := r.git("show", evidenceRev+":"+detailPath); got != strings.TrimSpace(edited) {
		t.Fatalf("evidence reverted the later Log edit:\n%s", got)
	}
	if !strings.Contains(r.git("log", "-1", "--format=%B", evidenceRev), tracker.EvidenceKeptTrailer+": "+detailPath) {
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
	if c, ok, _ := issue.CardCompletion([]byte(r.card(cardPath))); !ok || c.EvidenceCommit != r.git("rev-parse", evidenceRev) {
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
	if r.git("ls-tree", "--name-only", evidenceRev, "--", ledger) == "" {
		t.Fatalf("the first close did not commit its new gate ledger:\n%s", r.git("show", "--name-only", "--format=", evidenceRev))
	}
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "cmd/b.go")
	r.git("commit", "-qm", "#302: a fix after the close")
	closeOnce()
	if !strings.Contains(r.git("show", "--name-only", "--format=", evidenceRev), ledger) {
		t.Fatalf("the re-close's evidence commit lacks the modified gate ledger:\n%s", r.git("show", "--name-only", "--format=", evidenceRev))
	}
	if dirty := r.git("status", "--porcelain", "--", "workshop/plans"); dirty != "" {
		t.Fatalf("the re-close left plan files uncommitted:\n%s", dirty)
	}
}

// #301: a close survives a rebase, through both callers of the close-generation
// rule. The first close binds the card to a reviewed commit; rebasing onto an
// advanced main rewrites it off the branch; a reopen and a second close must
// land their completion, not be released against the rewritten one. "pruned"
// also drops the old commits from the clone (a rebase done elsewhere), which is
// where each caller's closeAncestorOf wiring matters: a SHIP close judges the
// generation itself, while a FIX-THEN-SHIP close defers it to reconcile.
func TestTrackerCloseSurvivesARebase(t *testing.T) {
	for i, tc := range []struct {
		name    string
		prune   bool
		verdict string
	}{
		{"close, old review still in the clone", false, "SHIP"},
		{"close, old review pruned", true, "SHIP"},
		{"reconcile, old review pruned", true, "FIX-THEN-SHIP"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := 310 + i
			idArg := fmt.Sprint(id)
			r, cardPath, _ := closeReady(t, id)
			stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
			if _, stderr, err := executeSDLCTestCommand("close", "--issue", idArg, "--verified", "first", "--actual", "1", "--no-atlas", "--no-ledger"); err != nil {
				t.Fatalf("first close: %v\n%s", err, stderr)
			}
			first, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
			if !ok {
				t.Fatal("first close bound nothing")
			}

			peerCommit(t, r, "other.go")
			r.git("fetch", "-q", "origin")
			r.git("rebase", "-q", "origin/main")
			if gitSucceeds(r.root, "merge-base", "--is-ancestor", first.ReviewedHEAD, "HEAD") {
				t.Fatal("fixture: the rebase did not rewrite the reviewed commit")
			}
			if tc.prune {
				// The first close pushed the branch (#286); the owner's leased
				// push of the rebase is what leaves the old commits unreferenced.
				r.git("push", "-q", "--force-with-lease", "origin", r.git("branch", "--show-current"))
				r.git("reflog", "expire", "--expire=now", "--all")
				r.git("gc", "-q", "--prune=now")
				if gitSucceeds(r.root, "cat-file", "-e", first.ReviewedHEAD+"^{commit}") {
					t.Fatal("fixture: the old reviewed commit survived the prune")
				}
			}

			if _, stderr, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", idArg); err != nil {
				t.Fatalf("reopen: %v\n%s", err, stderr)
			}
			stubJudge(t, "VERDICT: "+tc.verdict+" (confidence: high)\n\nfine\n")
			if _, stderr, err := executeSDLCTestCommand("close", "--issue", idArg, "--verified", "second", "--actual", "1", "--no-atlas", "--no-ledger"); err != nil {
				t.Fatalf("second close: %v\n%s", err, stderr)
			}
			if tc.verdict == "FIX-THEN-SHIP" {
				writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
				r.git("commit", "-qam", "#"+idArg+": fix review finding")
				var out, errs bytes.Buffer
				if err := runRecoveryReconcile(context.Background(), &out, &errs, id); err != nil {
					t.Fatalf("reconcile: %v\n%s", err, errs.String())
				}
			}
			second, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
			if !ok || second.Token == first.Token || second.EvidenceCommit != r.git("rev-parse", evidenceRev) || !strings.Contains(r.card(cardPath), "status: codecomplete") {
				t.Fatalf("the close after the rebase did not land (first %s, now %+v):\n%s", first.Token, second, r.card(cardPath))
			}
		})
	}
}

// gitSucceeds reports whether a git command in dir exits 0.
func gitSucceeds(dir string, args ...string) bool {
	return exec.Command("git", append([]string{"-C", dir}, args...)...).Run() == nil
}

// #304/#197: a tracker-era milestone close commits its OWN evidence — subject
// `#N Mx: close`, the verdict trailers, only the issue's files — pins it, and pushes it,
// so the whole-issue close's milestone-verdict gate passes with nothing pasted.
func TestMilestoneClose_CommitsItsOwnEvidence(t *testing.T) {
	r, _, detailPath := closeReadyWithPlan(t, 304, "- [ ] M1 — part one\n")
	writeRepoFile(t, r.root, "unrelated.go", "package u\n")
	r.git("add", "unrelated.go") // staged unrelated work must stay staged and uncommitted
	reviewed := r.git("rev-parse", "HEAD")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n\n```findings\nfindings: []\n```\n")
	var stderr string
	var err error
	if msg, died := expectDie(t, func() {
		_, stderr, err = executeSDLCTestCommand("milestone-close", "--issue", "304", "--milestone", "M1", "--verified", "e2e", "--actual", "0.5", "--no-atlas")
	}); died || err != nil {
		t.Fatalf("milestone-close: died=%v %q err=%v\n%s", died, msg, err, stderr)
	}
	head := r.git("rev-parse", "HEAD")
	if r.git("rev-parse", "HEAD^") != reviewed {
		t.Fatal("the milestone evidence commit is not directly on the reviewed commit")
	}
	msg := r.git("log", "-1", "--format=%B")
	for _, want := range []string{"#304 M1: close", "Review-Verdict: SHIP"} {
		if !strings.Contains(msg, want) {
			t.Errorf("evidence message lacks %q:\n%s", want, msg)
		}
	}
	for _, f := range strings.Fields(r.git("show", "--name-only", "--format=", "HEAD")) {
		if f != detailPath && !strings.HasPrefix(f, "workshop/plans/000304-") {
			t.Errorf("milestone evidence carried a foreign file %q", f)
		}
	}
	if staged := r.git("diff", "--cached", "--name-only"); staged != "unrelated.go" {
		t.Fatalf("unrelated staged work disturbed: %q", staged)
	}
	if pin := r.git("rev-parse", "refs/sdlc/reviewed/000304/M1"); pin != head {
		t.Fatalf("pin = %s, want the evidence commit %s", pin, head)
	}
	if tip := remoteTip(t, r, r.git("branch", "--show-current")); tip != head {
		t.Fatalf("pushed tip %s is not the evidence commit %s", tip, head)
	}
	ok, err := milestoneHasVerdictCommit("304", "M1", detailPath)
	if err != nil || !ok {
		t.Fatalf("the milestone-verdict gate must find the binary's own commit: %v %v", ok, err)
	}
}

// applyEvidenceCommit is a compare-and-swap: a branch that moved after the evidence
// commit was built on its old tip is left alone, and the swap refuses.
func TestApplyEvidenceCommit_RefusesAMovedBranch(t *testing.T) {
	r := newTrackerRepo(t, nil, nil)
	r.git("switch", "-q", "-c", "issue-1")
	env := boundaryEnv(t)
	head := r.git("rev-parse", "HEAD")
	evidence := r.git("commit-tree", r.git("rev-parse", "HEAD^{tree}"), "-p", head, "-m", "#1 M1: close")
	r.git("commit", "-q", "--allow-empty", "-m", "moved meanwhile")
	moved := r.git("rev-parse", "HEAD")
	if err := applyEvidenceCommit(env, "refs/heads/issue-1", evidence, nil); err == nil {
		t.Fatal("the swap must refuse a branch that moved off the evidence commit's parent")
	}
	if got := r.git("rev-parse", "HEAD"); got != moved {
		t.Fatalf("the moved branch was overwritten: %s", got)
	}
}
