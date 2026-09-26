package main

import (
	"bytes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// closedAndLanded closes #N on its branch (SHIP) and lands the branch on main.
func closedAndLanded(t *testing.T, id int) (*trackerRepo, string, string) {
	t.Helper()
	r, cardPath, detailPath := closeReady(t, id)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", itoa(id), "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	branch := r.git("branch", "--show-current")
	r.git("switch", "-q", "main")
	r.git("merge", "-q", "--no-ff", "--no-edit", branch)
	r.git("push", "-q", "origin", "main")
	return r, cardPath, detailPath
}

func TestPublishFlipCompletesLandedClosesAndArchivesByCard(t *testing.T) {
	r, cardPath, detailPath := closedAndLanded(t, 311)
	landed := r.git("rev-parse", "HEAD")
	done, err := publishCodecompleteIssues("workshop/issues")
	if err != nil || len(done) != 1 || done[0] != "000311" {
		t.Fatalf("flip: %v %v", done, err)
	}
	card := r.card(cardPath)
	c, _, _ := issue.CardCompletion([]byte(card))
	if !strings.Contains(card, "status: done") || c.LandedCommit != landed {
		t.Fatalf("card not done at the landing (%s):\n%s", landed, card)
	}
	// Idempotent: a second publish finds nothing left to complete.
	if again, err := publishCodecompleteIssues("workshop/issues"); err != nil || len(again) != 0 {
		t.Fatalf("second flip: %v %v", again, err)
	}
	var stderr bytes.Buffer
	moves, err := archiveDoneIssues(&stderr, "", "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil || len(moves) == 0 {
		t.Fatalf("archive by card status: %v %v", moves, err)
	}
	if _, err := os.Stat(filepath.Join(r.root, detailPath)); !os.IsNotExist(err) {
		t.Fatal("done details still active")
	}
}

func TestPublishFlipLeavesAReclosedGenerationAlone(t *testing.T) {
	r, cardPath, _ := closedAndLanded(t, 312)
	// Before anything completes, the card is reopened and re-closed on a new
	// generation that has not landed: the old landing must not complete it.
	branch := "000312-e2e"
	r.git("switch", "-q", branch)
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "cmd/b.go")
	r.git("commit", "-qm", "#312: follow-up")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "312", "--verified", "again", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("re-close: %v\n%s", err, stderr)
	}
	r.git("switch", "-q", "main")
	done, err := publishCodecompleteIssues("workshop/issues")
	if err != nil || len(done) != 0 {
		t.Fatalf("a stale landing completed a newer close: %v %v", done, err)
	}
	if card := r.card(cardPath); !strings.Contains(card, "status: codecomplete") {
		t.Fatalf("card:\n%s", card)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

// Durable (slot) landing in a tracked repository: the PR's close is owned by its
// binding, the landing completes the card, and the archive moves the details as
// they are — verifiably, and idempotently on retry.
func TestDurableLandingArchivesTrackedCloseByBinding(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 321)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "321", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	head, branch := r.git("rev-parse", "HEAD"), r.git("branch", "--show-current")
	base := r.git("merge-base", "HEAD", "origin/main")
	r.git("push", "-q", "origin", "HEAD:main") // the PR merged as a fast-forward
	pr := landingPR{Number: 321, State: "MERGED", Repo: "test/repo", HeadRef: branch, HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: head}
	evidence, _, _ := issue.CardCompletion([]byte(r.card(cardPath)))

	selected, err := selectLandingIssues(r.root, pr, "workshop/issues")
	if err != nil || len(selected) != 1 || !selected[0].tracked || selected[0].anchor != evidence.EvidenceCommit {
		t.Fatalf("selection by binding: %+v %v", selected, err)
	}
	if err := settleLandingCompletions(r.root, "workshop/issues"); err != nil {
		t.Fatal(err)
	}
	if card := r.card(cardPath); !strings.Contains(card, "status: done") {
		t.Fatalf("landing did not complete the card:\n%s", card)
	}
	if done, err := selectLandingIssues(r.root, pr, "workshop/issues"); err != nil || len(done) != 1 {
		t.Fatalf("a done card left the archive without its owner: %+v %v", done, err)
	}
	if complete, err := laProof(r.root, head, pr); err != nil || complete {
		t.Fatalf("premature archive proof: %v %v", complete, err)
	}
	if err := laArchive(r.root, pr); err != nil {
		t.Fatal(err)
	}
	tip := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "main"))
	history := "workshop/history/issues/" + filepath.Base(detailPath)
	archived := testfix.Capture(t, r.origin, "show", "main:"+history)
	if archived != testfix.Capture(t, r.origin, "show", head+":"+detailPath) {
		t.Fatal("tracked details were rewritten by the archive")
	}
	if complete, err := laProof(r.root, tip, pr); err != nil || !complete {
		t.Fatalf("archive proof: %v %v", complete, err)
	}
	if err := laArchive(r.root, pr); err != nil || strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "main")) != tip {
		t.Fatalf("retried archive was not a no-op: %v", err)
	}
}
