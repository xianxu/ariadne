package main

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// loseResponses makes every single-card publication land, then report a lost
// acknowledgement — the publication outcome an agent cannot see.
func loseResponses(t *testing.T) (restore func()) {
	t.Helper()
	prev := cardPublish
	cardPublish = func(env *trackerEnv, expected tracker.Record, next []byte, token string, trailers []string, before func(string, string) error) error {
		if err := prev(env, expected, next, token, trailers, before); err != nil {
			return err
		}
		return fmt.Errorf("%w: push response lost", gitx.ErrPublicationUncertain)
	}
	restore = func() { cardPublish = prev }
	t.Cleanup(restore)
	return restore
}

// #280: claim after a lost response. The claim landed but its caller cannot
// know; the error says to rerun, and the rerun is settled by the card.
func TestClaimRerunSettlesALostResponse(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000410", "lost")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	restore := loseResponses(t)
	err := runClaim(context.Background(), &out, &errs, claimFlagsFor(410))
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command (sdlc claim --issue 410)") {
		t.Fatalf("a lost response gave no recovery action: %v", err)
	}
	landed := r.card(cardPath)
	if !strings.Contains(landed, "status: working") || !strings.Contains(landed, "claimant:") {
		t.Fatalf("the claim did not land:\n%s", landed)
	}
	out.Reset()
	errs.Reset()
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(410)); err != nil || !strings.Contains(errs.String(), "already claimed by this workspace") || r.card(cardPath) != landed {
		t.Fatalf("the rerun did not settle it: %v\n%s", err, errs.String())
	}
}

// #280: set-status after a lost response — the same contract.
func TestSetStatusRerunSettlesALostResponse(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000411", "lost")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(411)); err != nil {
		t.Fatal(err)
	}
	f := &setStatusFlags{Issue: 411, Status: "blocked", IssuesDir: "workshop/issues"}
	restore := loseResponses(t)
	err := runSetStatus(context.Background(), &out, &errs, f)
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command") {
		t.Fatalf("a lost response gave no recovery action: %v", err)
	}
	landed := r.card(cardPath)
	if !strings.Contains(landed, "status: blocked") {
		t.Fatalf("the change did not land:\n%s", landed)
	}
	errs.Reset()
	if err := runSetStatus(context.Background(), &out, &errs, f); err != nil || !strings.Contains(errs.String(), "already has that") || r.card(cardPath) != landed {
		t.Fatalf("the rerun did not settle it: %v\n%s", err, errs.String())
	}
}

// #280: close is non-repeatable — a second close after SHIP is a new
// generation: a new token and evidence commit rebinding the card, with the
// first evidence commit kept in history.
func TestCloseRerunAfterShipStartsANewGeneration(t *testing.T) {
	r, cardPath, _ := closeReady(t, 412)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	closeIt := func() issue.Completion {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand("close", "--issue", "412", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
			t.Fatalf("close: %v\n%s", err, stderr)
		}
		b, ok, err := issue.CardCompletion([]byte(r.card(cardPath)))
		if err != nil || !ok {
			t.Fatalf("no binding: %v", err)
		}
		return b
	}
	first := closeIt()
	writeRepoFile(t, r.root, "cmd/c.go", "package a\n")
	r.git("add", "cmd/c.go")
	r.git("commit", "-qm", "#412: more work after the close")
	second := closeIt()
	if second.Token == first.Token || second.EvidenceCommit == first.EvidenceCommit {
		t.Fatalf("a re-run did not start a new generation: %+v then %+v", first, second)
	}
	r.git("merge-base", "--is-ancestor", first.EvidenceCommit, second.EvidenceCommit) // fails the test if the first generation left history
}

// #280: a card reopened after its close is not completed by a later landing of
// that close — the landing's guarantee ends at the reopen.
func TestLandingLeavesAReopenedCardAlone(t *testing.T) {
	r, cardPath, _ := closedAndLanded(t, 413)
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000413", cardPath, "reopen", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardField(c, "status", "working")
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	done, err := publishCodecompleteIssues(context.Background(), "workshop/issues")
	if err != nil || len(done) != 0 {
		t.Fatalf("a landing completed a reopened card: %v %v", done, err)
	}
	if card := r.card(cardPath); !strings.Contains(card, "status: working") {
		t.Fatalf("card:\n%s", card)
	}
}

// cardPublish's callers are exactly the single-card CAS verbs.
func TestCardPublishCallers(t *testing.T) {
	allowed := map[string]bool{"runCardUpdate": true, "runClaim": true, "adoptClaim": true, "relocateClaimant": true}
	seen := map[string]bool{}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			for _, decl := range file.Decls {
				owner := ""
				var body ast.Node
				switch d := decl.(type) {
				case *ast.FuncDecl:
					owner, body = d.Name.Name, d.Body
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						if vs, ok := spec.(*ast.ValueSpec); ok && vs.Names[0].Name == "reclaimEffect" {
							owner, body = "reclaimEffect", vs
						}
					}
				}
				if body == nil {
					continue
				}
				ast.Inspect(body, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok && id.Name == "cardPublish" {
						seen[owner] = true
						if !allowed[owner] && owner != "reclaimEffect" {
							t.Errorf("%s:%s publishes a card through cardPublish; only the single-card verbs may", name, owner)
						}
					}
					return true
				})
			}
		}
	}
	for owner := range allowed {
		if !seen[owner] {
			t.Errorf("%s no longer publishes through cardPublish", owner)
		}
	}
}
