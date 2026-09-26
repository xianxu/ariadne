package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
