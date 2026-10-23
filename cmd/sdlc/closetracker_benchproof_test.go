package main

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func closeBenchProofFixture(t *testing.T) (*trackerRepo, *trackerEnv, tracker.Record, issue.Completion, string, string) {
	t.Helper()
	r, _, details := closeReady(t, 6578)
	candidate := r.git("rev-parse", "HEAD")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "6578", "--verified", "native close proof fixture", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snap.Require("006578")
	if err != nil {
		t.Fatal(err)
	}
	binding, ok, err := issue.CardCompletion(card.Raw)
	if err != nil || !ok {
		t.Fatal(err)
	}
	return r, env, card, binding, candidate, details
}

func TestCloseBenchProofRealCloseEvidenceAndExactMirror(t *testing.T) {
	r, env, card, binding, candidate, details := closeBenchProofFixture(t)
	for _, head := range []string{binding.EvidenceCommit, r.git("rev-parse", "HEAD")} {
		if err := verifyCloseEvidenceChain(env, card, binding, candidate, head, details, "workshop/plans"); err != nil {
			t.Fatalf("normal close refused: %v", err)
		}
	}
}

func TestCloseBenchProofRefusesExtraCommitsAndMirrorBodyDrift(t *testing.T) {
	for _, change := range []string{"extra doc commit", "mirror body"} {
		t.Run(change, func(t *testing.T) {
			r, env, card, binding, candidate, details := closeBenchProofFixture(t)
			switch change {
			case "extra doc commit":
				writeRepoFile(t, r.root, "atlas/extra.md", "not close evidence\n")
				r.git("add", "atlas/extra.md")
				r.git("commit", "-qm", "unapproved docs after close")
			case "mirror body":
				writeRepoFile(t, r.root, details, r.git("show", "HEAD:"+details)+"\nUnreviewed body edit\n")
				r.git("add", details)
				r.git("commit", "--amend", "-qm", "not an exact mirror")
			}
			if err := verifyCloseEvidenceChain(env, card, binding, candidate, r.git("rev-parse", "HEAD"), details, "workshop/plans"); err == nil {
				t.Fatal("unapproved post-review metadata accepted")
			}
		})
	}
}

func TestCloseBenchProofRefusesMixedCodeEvidenceAndUnboundToken(t *testing.T) {
	for _, change := range []string{"code", "other docs", "token", "merge"} {
		t.Run(change, func(t *testing.T) {
			r, env, card, binding, candidate, details := closeBenchProofFixture(t)
			original := binding.EvidenceCommit
			r.git("checkout", "--detach", original)
			if change == "code" {
				writeRepoFile(t, r.root, "cmd/a.go", "package a // unreviewed\n")
				r.git("add", "cmd/a.go")
			}
			if change == "other docs" {
				writeRepoFile(t, r.root, "docs/unapproved.md", "unapproved\n")
				r.git("add", "docs/unapproved.md")
			}
			message := r.git("show", "-s", "--format=%B", original)
			oldToken := binding.Token
			binding.Token = "close-forged-generation"
			message = strings.Replace(message, oldToken, binding.Token, 1)
			if change == "token" {
				message = strings.Replace(message, binding.Token, "close-another", 1)
			}
			if change == "merge" {
				tree := r.git("rev-parse", original+"^{tree}")
				base := r.git("rev-parse", candidate+"^")
				binding.EvidenceCommit = r.git("commit-tree", tree, "-p", candidate, "-p", base, "-m", message)
			} else {
				r.git("commit", "--amend", "--allow-empty", "-qm", message)
				binding.EvidenceCommit = r.git("rev-parse", "HEAD")
			}
			var err error
			card.Raw, err = issue.SetCardCompletion(card.Raw, binding)
			if err != nil {
				t.Fatal(err)
			}
			if err := verifyCloseEvidenceChain(env, card, binding, candidate, binding.EvidenceCommit, details, "workshop/plans"); err == nil {
				t.Fatal("forged evidence relationship accepted")
			}
		})
	}
}
