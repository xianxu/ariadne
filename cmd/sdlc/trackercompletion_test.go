package main

import (
	"bytes"
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
	raw, _ := os.ReadFile(filepath.Join(r.root, detailPath))
	if !strings.Contains(string(raw), "status: done") {
		t.Fatal("details mirror not refreshed before archive")
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
