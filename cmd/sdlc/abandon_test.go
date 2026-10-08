package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const (
	s09Archive = "refs/ariadne/abandoned/000009"
	s09History = "workshop/history/issues/000009-s09.md"
)

func abandon9(as, reason string) error {
	var out, errs bytes.Buffer
	err := runAbandon(context.Background(), &out, &errs, &abandonFlags{Issue: 9, As: as, Reason: reason,
		IssuesDir: "workshop/issues", PlansDir: "workshop/plans", HistoryDir: "workshop/history"})
	if err != nil {
		return errors.Join(err, errors.New(errs.String()))
	}
	return nil
}

func remoteRef(t *testing.T, r *trackerRepo, ref string) string {
	t.Helper()
	if f := strings.Fields(r.git("ls-remote", "origin", ref)); len(f) > 0 {
		return f[0]
	}
	return ""
}

// assertAbandoned checks abandon's whole end state for #9's started work.
func assertAbandoned(t *testing.T, r *trackerRepo, cardPath, as, tipBefore string) {
	t.Helper()
	if r.git("for-each-ref", "refs/heads/"+s09Branch) != "" || remoteTip(t, r, s09Branch) != "" {
		t.Fatal("the issue branch survived")
	}
	if on := r.git("branch", "--show-current"); on != "main" {
		t.Fatalf("the checkout is on %s, not its resting branch", on)
	}
	kept := remoteRef(t, r, s09Archive)
	r.git("fetch", "-q", "origin", s09Archive)
	if kept == "" || !gitSucceeds(r.root, "merge-base", "--is-ancestor", tipBefore, kept) ||
		!strings.Contains(r.git("log", "-1", "--format=%s", kept), "log: abandon ("+as+")") {
		t.Fatalf("the archive ref does not keep the work plus its note: %q", kept)
	}
	card := r.card(cardPath)
	rec, ok, err := issue.CardAbandoned([]byte(card))
	if err != nil || !ok || rec.Ref != s09Archive || rec.Head != kept || rec.Branch != s09Branch || !strings.Contains(card, "status: "+as) || !strings.Contains(card, "claimant:") {
		t.Fatalf("card: %+v %v %v\n%s", rec, ok, err, card)
	}
	r.git("fetch", "-q", "origin")
	tree := r.git("ls-tree", "-r", "--name-only", "origin/main")
	if strings.Contains(tree, handoffDetail) || !strings.Contains(tree, s09History) {
		t.Fatalf("details not archived on main:\n%s", tree)
	}
	archived := r.git("show", "origin/main:"+s09History)
	if !strings.Contains(archived, "status: "+as) || !strings.Contains(archived, "abandoned ("+as+")") {
		t.Fatalf("archived details do not mirror the card or carry the note:\n%s", archived)
	}
}

// #286: abandoning started work keeps it under the archive ref, ends the card,
// archives the details on main and deletes the branch everywhere.
func TestAbandonStartedWork(t *testing.T) {
	r, paths, _ := startedHere(t)
	tip := r.git("rev-parse", "HEAD")
	if err := abandon9("punt", "after the freeze"); err != nil {
		t.Fatal(err)
	}
	assertAbandoned(t, r, paths["000009"], "punt", tip)
	// A rerun finds everything done.
	before := r.git("rev-parse", "origin/main")
	if err := abandon9("punt", "after the freeze"); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	r.git("fetch", "-q", "origin")
	if r.git("rev-parse", "origin/main") != before {
		t.Fatal("a rerun committed to main again")
	}
}

// An open issue has no branch: the card ends with an empty record and the
// details, with the note, are archived.
func TestAbandonOpenIssue(t *testing.T) {
	r, paths := claimSetRepo(t)
	claimFor(t, 9)
	if err := abandon9("wontfix", "out of scope"); err != nil {
		t.Fatal(err)
	}
	rec, ok, _ := issue.CardAbandoned([]byte(r.card(paths["000009"])))
	if !ok || rec.Started() || !strings.Contains(r.card(paths["000009"]), "status: wontfix") {
		t.Fatalf("card: %+v %v", rec, ok)
	}
	r.git("fetch", "-q", "origin")
	if got := r.git("show", "origin/main:"+s09History); !strings.Contains(got, "abandoned (wontfix): out of scope") {
		t.Fatalf("archived details:\n%s", got)
	}
}

// Each refusal names its own check, before any effect.
func TestAbandonRefusals(t *testing.T) {
	for _, c := range []struct {
		name, as, reason, want string
		shape                  func(t *testing.T, r *trackerRepo)
	}{
		{"bad --as", "done", "x", "--as must be", nil},
		{"no reason", "punt", " ", "--reason is required", nil},
		{"dirty tree", "punt", "x", "clean tree", func(t *testing.T, r *trackerRepo) { writeRepoFile(t, r.root, "cmd/nine.go", "package x\n") }},
		{"from rest", "punt", "x", "from its issue branch", func(t *testing.T, r *trackerRepo) { r.git("switch", "-q", "main") }},
		{"not owner", "punt", "x", "is owned by", func(t *testing.T, r *trackerRepo) { withClaimant(t, otherSlot) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, _, _ := startedHere(t)
			if c.shape != nil {
				c.shape(t, r)
			}
			head := r.git("rev-parse", "HEAD")
			err := abandon9(c.as, c.reason)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
			if r.git("rev-parse", "HEAD") != head || remoteRef(t, r, s09Archive) != "" {
				t.Fatal("a refused abandon had an effect")
			}
		})
	}
	t.Run("already terminal without the record", func(t *testing.T) {
		r, paths := claimSetRepo(t)
		claimFor(t, 9)
		env, err := openTracker(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if err := env.repo.ChangeCard("000009", paths["000009"], "fixture", operationToken("set"), func(c []byte) ([]byte, error) {
			return issue.SetCardField(c, "status", "wontfix")
		}); err != nil {
			t.Fatal(err)
		}
		_ = r
		if err := abandon9("punt", "x"); err == nil || !strings.Contains(err.Error(), "already wontfix") {
			t.Fatalf("got %v", err)
		}
	})
}

// An interrupted abandon is finished by its rerun, from wherever it stopped
// — including from the resting branch once the checkout has left the issue
// branch (the rerun is recognised from the card before the branch checks).
func TestAbandonRerunResumes(t *testing.T) {
	failCard := func() func() {
		prev := cardPublish
		cardPublish = func(*trackerEnv, tracker.Record, []byte, string, []string, func(string, string) error) error {
			return errors.New("interrupted")
		}
		return func() { cardPublish = prev }
	}
	failMain := func() func() {
		prev := mainPublish
		mainPublish = func(*trackerEnv, string, func(*gitx.TrunkView) (gitx.TrunkWrite, error), func(string, string) error) error {
			return errors.New("interrupted")
		}
		return func() { mainPublish = prev }
	}
	for _, c := range []struct {
		name  string
		stub  func() (undo func())
		after func(r *trackerRepo) // the rest of the interrupted state
	}{
		{"after the archive ref, before the card", failCard, nil},
		{"after the card, before main's archive", failMain, nil},
		{"remote branch deleted, still on the branch", failMain, func(r *trackerRepo) {
			r.git("push", "-q", "origin", "--delete", s09Branch)
		}},
		{"back on rest, local branch still present", failMain, func(r *trackerRepo) {
			r.git("push", "-q", "origin", "--delete", s09Branch)
			r.git("switch", "-q", "main")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, paths, _ := startedHere(t)
			if err := pushIssueBranch(boundaryEnv(t), s09Branch); err != nil {
				t.Fatal(err)
			}
			tip := r.git("rev-parse", "HEAD")
			undo := c.stub()
			err := abandon9("punt", "r")
			undo()
			if err == nil {
				t.Fatal("fixture did not interrupt")
			}
			if c.after != nil {
				c.after(r)
			}
			if err := abandon9("punt", "r"); err != nil {
				t.Fatalf("rerun: %v", err)
			}
			assertAbandoned(t, r, paths["000009"], "punt", tip)
			if n := strings.Count(r.git("log", "--format=%s", remoteRef(t, r, s09Archive)), "log: abandon"); n != 1 {
				t.Fatalf("%d abandon notes on the kept tip", n)
			}
		})
	}
}

// abandonDecision follows the lifecycle model: every non-terminal status the
// model lets abandon/defer leave ends there; anything else refuses.
func TestAbandonDecision(t *testing.T) {
	_, card, _, _ := seededIssue(t, "000009", "s09")
	for _, status := range vocab.Issue().AllStatuses() {
		for _, as := range []string{"wontfix", "punt"} {
			in, err := issue.SetCardField([]byte(card), "status", status)
			if err != nil {
				continue // a status a card can't simply be put into (done needs hours)
			}
			out, err := abandonDecision(in, as, "2026-10-08", issue.Abandoned{})
			edge := vocab.Issue().TransitionForEvent(status, abandonEvent[as])
			if (edge != nil && edge.To == as) != (err == nil) {
				t.Fatalf("%s as %s: edge %v, err %v", status, as, edge, err)
			}
			if err == nil && (!strings.Contains(string(out), "status: "+as) || !strings.Contains(string(out), "abandoned:")) {
				t.Fatalf("%s as %s:\n%s", status, as, out)
			}
		}
	}
}
