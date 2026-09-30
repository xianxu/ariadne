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
