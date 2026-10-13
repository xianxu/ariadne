package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// landedFixture is #239's shape (#311): issue 330's card is codecomplete with
// no close binding, as a card migrated from before the tracker is, and its
// work was merged from a branch whose name is not the card's slug, then the
// branch was deleted. It returns the merge commit and a commit on main that
// brought in no #330 work.
func landedFixture(t *testing.T, status string) (r *trackerRepo, cardPath, detailPath, merge, unrelated string) {
	t.Helper()
	full := fmt.Sprintf("---\nid: 000330\nstatus: %s\ndeps: []\ncreated: 2026-09-01\nupdated: 2026-09-01\nactual_hours: 1\n---\n\n# e2e\n\n## Problem\n\nA gap.\n\n## Log\n", status)
	card, detail, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	cardPath, detailPath = tracker.CardPath("000330", "e2e"), syncIssuesDir+"/000330-e2e.md"
	r = newTrackerRepo(t, map[string]string{cardPath: string(card)}, map[string]string{detailPath: string(detail)})
	peer := t.TempDir()
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	testfix.Git(t, peer, "config", "user.name", "web")
	testfix.Git(t, peer, "config", "user.email", "w@w")
	commit := func(file, msg string) {
		writeRepoFile(t, peer, file, "x\n")
		testfix.Git(t, peer, "add", file)
		testfix.Git(t, peer, "commit", "-qm", msg)
	}
	testfix.Git(t, peer, "switch", "-qc", "standalone-restart") // not 000330-e2e
	commit("cmd/a.go", "#330: implement")
	testfix.Git(t, peer, "switch", "-q", "main")
	commit("cmd/b.go", "#3300: another issue")
	unrelated = strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD"))
	testfix.Git(t, peer, "merge", "-q", "--no-ff", "--no-edit", "standalone-restart")
	merge = strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD"))
	testfix.Git(t, peer, "push", "-q", "origin", "main")
	r.git("fetch", "-q", "origin")
	return r, cardPath, detailPath, merge, unrelated
}

func stateIssueIDs(t *testing.T) string {
	t.Helper()
	out, stderr, err := executeSDLCTestCommand("state", "--json")
	if err != nil {
		t.Fatalf("state: %v\n%s", err, stderr)
	}
	var s State
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatalf("state json: %v\n%s", err, out)
	}
	var ids []string
	for _, i := range s.Issues {
		ids = append(ids, i.ID)
	}
	return strings.Join(ids, ",")
}

// #311: reconcile --landed completes an unbound codecomplete card from the
// commit that merged its renamed, deleted branch: card done, details archived
// on main, gone from state; a rerun changes nothing.
func TestReconcileLandedRepairsAnUnboundCard(t *testing.T) {
	r, cardPath, detailPath, merge, _ := landedFixture(t, "codecomplete")
	if !strings.Contains(stateIssueIDs(t), "000330") {
		t.Fatal("fixture: state does not list #330")
	}
	var out, errs bytes.Buffer
	if err := runRecoveryReconcileLanded(context.Background(), &out, &errs, 330, merge[:10]); err != nil {
		t.Fatalf("reconcile --landed: %v\n%s", err, errs.String())
	}
	if c := r.card(cardPath); !strings.Contains(c, "status: done") {
		t.Fatalf("card:\n%s", c)
	}
	r.git("fetch", "-q", "origin")
	archived := "workshop/history/issues/" + filepath.Base(detailPath)
	if tree := r.git("ls-tree", "-r", "--name-only", "origin/main"); strings.Contains(tree, detailPath) || !strings.Contains(tree, archived) {
		t.Fatalf("not archived on main:\n%s", tree)
	}
	r.git("pull", "-q", "--ff-only") // the archive is on main
	if ids := stateIssueIDs(t); strings.Contains(ids, "000330") {
		t.Fatalf("state still lists #330: %s", ids)
	}
	before, card := r.originMain(), r.card(cardPath)
	errs.Reset()
	if err := runRecoveryReconcileLanded(context.Background(), &out, &errs, 330, merge); err != nil || r.originMain() != before || r.card(cardPath) != card {
		t.Fatalf("rerun: %v (changed: %v)\n%s", err, r.originMain() != before || r.card(cardPath) != card, errs.String())
	}
}

// #311: --landed refuses, changing nothing, a commit that is not on main or
// brought in no work citing the issue, and a card that is not closed.
func TestReconcileLandedRefuses(t *testing.T) {
	cases := map[string]struct {
		status string
		commit func(r *trackerRepo, merge, unrelated string) string
		want   string
	}{
		"not on main": {"codecomplete", func(r *trackerRepo, _, _ string) string {
			r.git("switch", "-qc", "local")
			writeRepoFile(t, r.root, "cmd/c.go", "x\n")
			r.git("add", "cmd/c.go")
			r.git("commit", "-qm", "#330: unpushed")
			return r.git("rev-parse", "HEAD")
		}, "is not on main"},
		"only bookkeeping": {"codecomplete", func(r *trackerRepo, _, _ string) string {
			peer := t.TempDir()
			testfix.Git(t, "", "clone", "-q", r.origin, peer)
			testfix.Git(t, peer, "config", "user.name", "web")
			testfix.Git(t, peer, "config", "user.email", "w@w")
			testfix.Git(t, peer, "commit", "-q", "--allow-empty", "-m", "#330: close")
			testfix.Git(t, peer, "push", "-q", "origin", "main")
			r.git("fetch", "-q", "origin")
			return strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD"))
		}, "brought in no work commit citing #330"},
		"no issue commit": {"codecomplete", func(_ *trackerRepo, _, unrelated string) string { return unrelated }, "brought in no work commit citing #330"},
		"unknown commit":  {"codecomplete", func(*trackerRepo, string, string) string { return "deadbeef" }, "no such commit"},
		"not closed":      {"working", func(_ *trackerRepo, merge, _ string) string { return merge }, "not codecomplete"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r, cardPath, _, merge, unrelated := landedFixture(t, c.status)
			commit := c.commit(r, merge, unrelated)
			before, card := r.originMain(), r.card(cardPath)
			var out, errs bytes.Buffer
			err := runRecoveryReconcileLanded(context.Background(), &out, &errs, 330, commit)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want %q\n%s", err, c.want, errs.String())
			}
			if r.originMain() != before || r.card(cardPath) != card {
				t.Fatal("a refused repair changed the remote")
			}
		})
	}
}

// A bound card is left to the settle its binding drives.
func TestDoneUnboundCardRefusesABoundCard(t *testing.T) {
	done := []byte("---\nid: '000330'\nstatus: done\nactual_hours: 1\ntracker:\n    version: 1\n    completion:\n        token: close-0123456789ab\n        repository: github.com/test/repo\n        reviewed_head: " + strings.Repeat("a", 40) + "\n        evidence_commit: " + strings.Repeat("b", 40) + "\n---\n\n# x\n\n## Problem\n\nA gap.\n")
	if _, err := doneUnboundCard(done, "2026-10-12"); !errors.Is(err, tracker.ErrNoChange) {
		t.Fatalf("a done bound card: got %v, want no change", err)
	}
	card := []byte("---\nid: '000330'\nstatus: codecomplete\nactual_hours: 1\ntracker:\n    version: 1\n    completion:\n        token: close-0123456789ab\n        repository: github.com/test/repo\n        reviewed_head: " + strings.Repeat("a", 40) + "\n        evidence_commit: " + strings.Repeat("b", 40) + "\n---\n\n# x\n\n## Problem\n\nA gap.\n")
	if _, err := doneUnboundCard(card, "2026-10-12"); err == nil || !strings.Contains(err.Error(), "close binding") {
		t.Fatalf("got %v", err)
	}
}

func TestCitesIssue(t *testing.T) {
	for subjects, want := range map[string]bool{
		"#330: x":                 true,
		"other\n#330 M1: x":       true,
		"#3300: x":                false,
		"fix #330":                false,
		"Merge pull request #330": false,
		"#330: close":             false,
		"sdlc: #330 implement":    true,
	} {
		if got := citesIssue(subjects, "000330"); got != want {
			t.Errorf("%q: got %v", subjects, got)
		}
	}
}
