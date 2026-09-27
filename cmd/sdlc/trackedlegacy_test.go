package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// In a repository cut over to the issue tracker, `issue sync` stays a local
// checkpoint on the issue branch; its legacy behaviors (a resting-branch
// commit, --push to main) and `issue publish` refuse (#252).
func TestLegacyIssueEntrypointsInATrackedRepository(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	mainBefore := r.originMain()
	writeRepoFile(t, r.root, detailPath, detail+"\n## Log\n- a design note\n")

	for name, args := range map[string][]string{
		"sync on rest": {"issue", "sync", "--issue", "9"},
		"sync --push":  {"issue", "sync", "--issue", "9", "--push"},
	} {
		head := r.git("rev-parse", "HEAD")
		msg, died := expectDie(t, func() { _, _, _ = executeSDLCTestCommand(args...) })
		if !died || !strings.Contains(msg, "issue tracker") {
			t.Fatalf("%s: died=%v %q", name, died, msg)
		}
		if r.git("rev-parse", "HEAD") != head {
			t.Fatalf("%s committed", name)
		}
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "publish", "--commit", r.git("rev-parse", "HEAD")); err == nil || !strings.Contains(err.Error()+stderr, "move-detail") {
		t.Fatalf("issue publish ran in a tracked repository: %v", err)
	}

	r.git("switch", "-q", "-c", "000009-nine")
	if _, stderr, err := executeSDLCTestCommand("issue", "sync", "--issue", "9"); err != nil {
		t.Fatalf("checkpoint on the issue branch: %v\n%s", err, stderr)
	}
	if !strings.Contains(r.git("show", "HEAD:"+detailPath), "a design note") || !strings.HasPrefix(r.git("log", "-1", "--format=%s"), "#9: issue-sync") {
		t.Fatal("the checkpoint did not commit the details on the issue branch")
	}
	if r.originMain() != mainBefore {
		t.Fatal("the checkpoint published to main")
	}
}

// The workflow Makefile's shell fallbacks allocate IDs and read/write status in
// details; in a tracked repository they refuse and ask for the binary.
func TestMakefileFallbacksRefuseInATrackedRepository(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not installed")
	}
	root := realRepoRoot()
	consumer := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, consumer, "workshop/issues/.keep", "")
	if err := os.Symlink(filepath.Join(root, "Makefile.workflow"), filepath.Join(consumer, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	run := func(target ...string) (string, error) {
		cmd := exec.Command("make", append([]string{"-f", "Makefile.workflow", "WF_ISSUES_DIR=workshop/issues", "WF_HISTORY_DIR=workshop/history"}, target...)...)
		cmd.Dir = consumer
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	targets := [][]string{{"fetch", "FETCH_NUM=1"}, {"push"}, {"merge"}, {"close-issue", "ISSUE=1"}, {"worktree"}}
	refuseAll := func(state string) {
		for _, target := range targets {
			out, err := run(target...)
			if err == nil || !strings.Contains(out, "uses the issue tracker") {
				t.Errorf("%s: make %s ran its fallback: %v\n%s", state, target[0], err, out)
			}
		}
	}
	// A pre-cutover branch: no marker yet, but the clone has fetched a tracker.
	testfix.Git(t, consumer, "update-ref", "refs/remotes/origin/issue-tracker", "HEAD")
	refuseAll("fetched tracker")
	testfix.Git(t, consumer, "update-ref", "-d", "refs/remotes/origin/issue-tracker")
	writeRepoFile(t, consumer, tracker.CutoverMarkerPath, string(tracker.CutoverMarkerBytes(strings.Repeat("a", 40))))
	refuseAll("marker")
}

func TestDetailsWithoutCardsIsPure(t *testing.T) {
	cards := map[string]string{"000009": tracker.CardPath("000009", "nine"), "000010": tracker.CardPath("000010", "ten")}
	base := map[int][]string{7: {"workshop/issues/000007-seven.md"}}
	head := map[int][]string{
		7:  {"workshop/issues/000007-seven.md"},
		8:  {"workshop/issues/000008-hand-made.md"},
		9:  {"workshop/issues/000009-nine.md"},
		10: {"workshop/issues/000010-renamed.md"},
		11: {"workshop/history/issues/000011-archived.md"},
	}
	added := addedDetails(base, head, "workshop/issues")
	got := detailsWithoutCards(added, cards)
	if strings.Join(got, ",") != "workshop/issues/000008-hand-made.md,workshop/issues/000010-renamed.md" {
		t.Fatalf("cardless: %v (added %v)", got, added)
	}
}

// In a tracker repository, CI's id check refuses details a range adds without
// a card (a hand-made file would collide with the next allocated id) and
// accepts details that belong to their card (#252 M4).
func TestLintIDsRefusesCardlessDetailsInATrackerRepository(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "sdlc")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build sdlc: %v\n%s", err, out)
	}
	cardPath, card, _, _ := seededIssue(t, "000009", "nine")
	_, _, detail9Path, detail9 := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7, cardPath: card}, nil)
	lint := func() (int, string) {
		cmd := exec.Command(bin, "issue", "lint-ids", "--base", "main", "--head", "HEAD")
		cmd.Dir = r.root
		cmd.Env = envWithTMPDIR(t)
		out, err := cmd.CombinedOutput()
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), string(out)
		} else if err != nil {
			t.Fatalf("run lint-ids: %v\n%s", err, out)
		}
		return 0, string(out)
	}
	r.git("switch", "-q", "-c", "000009-nine")
	writeRepoFile(t, r.root, detail9Path, detail9)
	r.git("add", detail9Path)
	r.git("commit", "-qm", "#9: details for its card")
	if code, out := lint(); code != 0 {
		t.Fatalf("carded details refused (%d):\n%s", code, out)
	}
	writeRepoFile(t, r.root, "workshop/issues/000010-hand-made.md", "---\nid: 000010\nstatus: open\n---\n\n# Hand made\n\n## Problem\nx\n")
	r.git("add", "workshop/issues/000010-hand-made.md")
	r.git("commit", "-qm", "a hand-made issue")
	if code, out := lint(); code != 1 || !strings.Contains(out, "000010-hand-made.md") || !strings.Contains(out, "no matching issue card") {
		t.Fatalf("cardless details not refused (%d):\n%s", code, out)
	}
}
