package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// mirrorsCard asserts details project exactly the given card blob (#275).
func mirrorsCard(t *testing.T, details, card, status string) {
	t.Helper()
	oid, err := issue.CardBlobOID([]byte(card), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(details, "status: "+status+"\n") || !strings.Contains(details, "card_mirror: '"+oid+"'") {
		t.Fatalf("details do not mirror the %s card %s:\n%s", status, oid, details)
	}
}

// bodyOf is a details file below its frontmatter.
func bodyOf(t *testing.T, details string) string {
	t.Helper()
	_, body, err := issue.Parse(details)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// #275: a close leaves the issue branch mirroring the codecomplete card, in a
// narrow commit after the evidence commit the card is bound to.
func TestCloseRefreshesDetailsToCodecompleteCard(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 341)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "341", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	card := r.card(cardPath)
	c, _, _ := issue.CardCompletion([]byte(card))
	if r.git("rev-parse", "HEAD^") != c.EvidenceCommit {
		t.Fatalf("the mirror commit is not directly on the evidence commit %s", c.EvidenceCommit)
	}
	if files := r.git("show", "--name-only", "--format=", "HEAD"); files != detailPath {
		t.Fatalf("mirror commit carried %q", files)
	}
	head := r.git("show", "HEAD:"+detailPath)
	mirrorsCard(t, head+"\n", card, "codecomplete")
	if bodyOf(t, head) != bodyOf(t, r.git("show", c.EvidenceCommit+":"+detailPath)) {
		t.Fatal("the mirror commit changed the details body")
	}
	if dirty := r.git("status", "--porcelain"); dirty != "" {
		t.Fatalf("close left the checkout dirty: %q", dirty)
	}
}

// #275: an uncommitted details edit survives the refresh: HEAD gets the
// committed bytes refreshed, the worktree keeps its edit plus the new mirror.
func TestCloseMirrorKeepsDirtyDetailsBody(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 342)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nrename a thing\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "342", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	writeRepoFile(t, r.root, "cmd/a.go", "package a // fixed\n")
	r.git("commit", "-qm", "#342: fix", "--", "cmd/a.go")
	abs := filepath.Join(r.root, detailPath)
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, detailPath, string(raw)+"- an uncommitted note\n")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 342); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	card := r.card(cardPath)
	head := r.git("show", "HEAD:"+detailPath)
	mirrorsCard(t, head+"\n", card, "codecomplete")
	if strings.Contains(head, "an uncommitted note") {
		t.Fatal("the mirror commit swept an uncommitted edit")
	}
	worktree, _ := os.ReadFile(abs)
	mirrorsCard(t, string(worktree), card, "codecomplete")
	if !strings.Contains(string(worktree), "- an uncommitted note\n") {
		t.Fatalf("the uncommitted edit was lost:\n%s", worktree)
	}
}

// #275: a close interrupted after codecomplete but before its mirror commit is
// finished by recovery reconcile.
func TestReconcileRetriesAnInterruptedCloseMirror(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 350)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "350", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	r.git("reset", "-q", "--hard", evidenceRev) // the crash: no mirror commit
	abs := filepath.Join(r.root, detailPath)
	raw, _ := os.ReadFile(abs)
	writeRepoFile(t, r.root, detailPath, string(raw)+"- a staged note\n")
	r.git("add", detailPath) // a staged edit stays staged
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 350); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	mirrorsCard(t, r.git("show", "HEAD:"+detailPath)+"\n", r.card(cardPath), "codecomplete")
	if staged := r.git("diff", "--cached", "--", detailPath); !strings.Contains(staged, "+- a staged note") {
		t.Fatalf("the staged edit was overwritten in the index:\n%s", staged)
	}
	before := r.git("rev-parse", "HEAD")
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 350); err != nil || r.git("rev-parse", "HEAD") != before {
		t.Fatalf("a second reconcile was not a no-op: %v", err)
	}
}

// landedDone lands #N's close on main and completes its card (the publish
// flip), leaving the details at main's HEAD for an archive to move.
func landedDone(t *testing.T, id int) (*trackerRepo, string, string) {
	t.Helper()
	r, cardPath, detailPath := closedAndLanded(t, id)
	if done, err := publishCodecompleteIssues(context.Background(), "workshop/issues"); err != nil || len(done) != 1 {
		t.Fatalf("flip: %v %v", done, err)
	}
	return r, cardPath, detailPath
}

func historyOf(detailPath string) string {
	return "workshop/history/issues/" + filepath.Base(detailPath)
}

// #275 (pair#358's archive half): push's and merge's checkout archives project
// the done card into the archived details, keeping their body.
func TestCheckoutArchiveMirrorsTheDoneCard(t *testing.T) {
	for _, c := range []struct {
		name    string
		id      int
		archive func(stderr *bytes.Buffer, root string) error
	}{
		{"push", 343, func(stderr *bytes.Buffer, _ string) error {
			_, err := archiveDoneIssues(context.Background(), stderr, "", "workshop/issues", "workshop/history", "workshop/plans")
			return err
		}},
		{"merge", 349, func(stderr *bytes.Buffer, root string) error {
			_, err := archiveDoneIssuesInDir(context.Background(), stderr, "", root, "workshop/issues", "workshop/history", "workshop/plans")
			return err
		}},
	} {
		t.Run(c.name, func(t *testing.T) { checkoutArchiveMirrors(t, c.id, c.archive) })
	}
}

func checkoutArchiveMirrors(t *testing.T, id int, archive func(*bytes.Buffer, string) error) {
	r, cardPath, detailPath := landedDone(t, id)
	landed := r.git("show", "HEAD:"+detailPath)
	var stderr bytes.Buffer
	if err := archive(&stderr, r.root); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr.String())
	}
	archived, err := os.ReadFile(filepath.Join(r.root, historyOf(detailPath)))
	if err != nil {
		t.Fatal(err)
	}
	card := r.card(cardPath)
	mirrorsCard(t, string(archived), card, "done")
	for _, want := range []string{"actual_hours: 1\n", "updated: " + issueField(t, card, "updated") + "\n"} {
		if !strings.Contains(string(archived), want) {
			t.Errorf("archived details lack %q:\n%s", want, archived)
		}
	}
	if bodyOf(t, string(archived)) != bodyOf(t, landed+"\n") {
		t.Fatal("the archive changed the details body")
	}
}

// #275: a hand-edited mirrored field cannot be refreshed; the details are
// archived as they are, with a warning — the card stays the authority.
func TestCheckoutArchiveKeepsAHandEditedMirrorWithAWarning(t *testing.T) {
	r, _, detailPath := landedDone(t, 344)
	abs := filepath.Join(r.root, detailPath)
	raw, _ := os.ReadFile(abs)
	edited := strings.Replace(string(raw), "status: codecomplete", "status: working", 1)
	writeRepoFile(t, r.root, detailPath, edited)
	var stderr bytes.Buffer
	if _, err := archiveDoneIssues(context.Background(), &stderr, "", "workshop/issues", "workshop/history", "workshop/plans"); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr.String())
	}
	archived, _ := os.ReadFile(filepath.Join(r.root, historyOf(detailPath)))
	if string(archived) != edited {
		t.Fatalf("hand-edited details were rewritten:\n%s", archived)
	}
	if !strings.Contains(stderr.String(), "mirror not refreshed") {
		t.Fatalf("no warning for the unrefreshed mirror:\n%s", stderr.String())
	}
}

// #275: an interrupted archive's recovery refreshes the moved details before
// committing them.
func TestRecoverInterruptedArchiveMirrorsTheDoneCard(t *testing.T) {
	r, cardPath, detailPath := landedDone(t, 345)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(r.root, historyOf(detailPath))), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(r.root, detailPath), filepath.Join(r.root, historyOf(detailPath))); err != nil {
		t.Fatal(err)
	}
	r.git("branch", "--set-upstream-to=origin/main")
	var stdout, stderr bytes.Buffer
	recovered, err := recoverInterruptedArchive(context.Background(), &stdout, &stderr, &pushFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history", PlansDir: "workshop/plans"})
	if err != nil || !recovered {
		t.Fatalf("recovery: %v %v\n%s", recovered, err, stderr.String())
	}
	mirrorsCard(t, r.git("show", "HEAD:"+historyOf(detailPath))+"\n", r.card(cardPath), "done")
}

func issueField(t *testing.T, doc, key string) string {
	t.Helper()
	fm, _, err := issue.Parse(doc)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := issue.GetField(fm, key)
	return v
}

// #275: the landing archive's projection is a function of its inputs alone —
// the done card mirrored in, or the details kept when they cannot take it.
func TestArchivedDetailsProjectsOrKeeps(t *testing.T) {
	full := "---\nid: 000346\nstatus: working\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# p\n\n## Problem\n\nx\n\n## Log\n"
	baseline, details, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	done := []byte(strings.Replace(strings.Replace(string(baseline), "status: working", "status: done", 1), "updated: 2026-09-01", "updated: 2026-09-30\nactual_hours: 1", 1))
	codecomplete := []byte(strings.Replace(string(baseline), "status: working", "status: codecomplete", 1))
	refreshed := archivedDetails(details, baseline, done)
	mirrorsCard(t, string(refreshed), string(done), "done")
	if !bytes.Equal(archivedDetails(details, baseline, done), refreshed) {
		t.Fatal("the projection is not deterministic")
	}
	edited := []byte(strings.Replace(string(details), "status: working", "status: done", 1))
	for name, c := range map[string]struct{ details, baseline, card []byte }{
		"hand-edited mirror":  {edited, baseline, done},
		"no baseline":         {details, nil, done},
		"card not done":       {details, baseline, codecomplete},
		"unmirrored details":  {[]byte(full), baseline, done},
		"baseline mismatched": {details, done, done},
	} {
		if got := archivedDetails(c.details, c.baseline, c.card); !bytes.Equal(got, c.details) {
			t.Errorf("%s: details rewritten:\n%s", name, got)
		}
	}
}

// #275: a card that changes after the archive (still done, same close) does not
// break the landing's retry proof — the archived mirror pins the card it used.
func TestDurableLandingProofSurvivesACardChangeAfterArchive(t *testing.T) {
	r, cardPath, _ := closeReady(t, 347)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "347", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	head, branch := r.git("rev-parse", "HEAD"), r.git("branch", "--show-current")
	base := r.git("merge-base", "HEAD", "origin/main")
	r.git("push", "-q", "origin", "HEAD:main")
	pr := landingPR{Number: 347, State: "MERGED", Repo: "test/repo", HeadRef: branch, HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: head}
	if err := completeLandingPR(context.Background(), r.root, "workshop/issues", pr); err != nil {
		t.Fatal(err)
	}
	if err := laArchive(r.root, pr); err != nil {
		t.Fatal(err)
	}
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000347", cardPath, "retitle", operationToken("set"), func(c []byte) ([]byte, error) {
		return []byte(strings.Replace(string(c), "# e2e", "# renamed", 1)), nil
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	if !strings.Contains(r.card(cardPath), "# renamed") || !strings.Contains(r.card(cardPath), "status: done") {
		t.Fatalf("fixture: card not retitled:\n%s", r.card(cardPath))
	}
	tip := r.originMain()
	if complete, err := laProof(r.root, tip, pr); err != nil || !complete {
		t.Fatalf("archive proof after a card change: %v %v", complete, err)
	}
}

// #275: a details mirror naming a card blob the tracker cannot read must not
// wedge a landing: the details are archived unchanged and the proof completes.
func TestDurableLandingArchivesAnUnreadableBaselineUnchanged(t *testing.T) {
	r, _, detailPath := closeReady(t, 348)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "348", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	abs := filepath.Join(r.root, detailPath)
	raw, _ := os.ReadFile(abs)
	oid, err := issue.MirrorBaselineOID(raw)
	if err != nil {
		t.Fatal(err)
	}
	missing := strings.Replace(string(raw), oid, strings.Repeat("0", len(oid)), 1)
	writeRepoFile(t, r.root, detailPath, missing)
	r.git("commit", "-qm", "#348: a mirror naming a missing blob", "--", detailPath)
	head, branch := r.git("rev-parse", "HEAD"), r.git("branch", "--show-current")
	base := r.git("merge-base", "HEAD", "origin/main")
	r.git("push", "-q", "origin", "HEAD:main")
	pr := landingPR{Number: 348, State: "MERGED", Repo: "test/repo", HeadRef: branch, HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: head}
	if err := completeLandingPR(context.Background(), r.root, "workshop/issues", pr); err != nil {
		t.Fatal(err)
	}
	if err := laArchive(r.root, pr); err != nil {
		t.Fatalf("an unreadable baseline wedged the archive: %v", err)
	}
	if got := testfix.Capture(t, r.origin, "show", "main:"+historyOf(detailPath)); got != missing {
		t.Fatalf("details with an unreadable baseline were rewritten:\n%s", got)
	}
	if complete, err := laProof(r.root, r.originMain(), pr); err != nil || !complete {
		t.Fatalf("archive proof: %v %v", complete, err)
	}
}
