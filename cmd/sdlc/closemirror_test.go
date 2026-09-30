package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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

// #275 (pair#358's archive half): a checkout archive projects the done card
// into the archived details, keeping their body.
func TestCheckoutArchiveMirrorsTheDoneCard(t *testing.T) {
	r, cardPath, detailPath := landedDone(t, 343)
	landed := r.git("show", "HEAD:"+detailPath)
	var stderr bytes.Buffer
	if _, err := archiveDoneIssues(context.Background(), &stderr, "", "workshop/issues", "workshop/history", "workshop/plans"); err != nil {
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
