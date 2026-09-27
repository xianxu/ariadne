package main

import (
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
	writeRepoFile(t, consumer, tracker.CutoverMarkerPath, string(tracker.CutoverMarkerBytes(strings.Repeat("a", 40))))
	for _, target := range [][]string{{"fetch", "FETCH_NUM=1"}, {"push"}, {"merge"}, {"close-issue", "ISSUE=1"}, {"worktree"}} {
		out, err := run(target...)
		if err == nil || !strings.Contains(out, "uses the issue tracker") {
			t.Errorf("make %s ran its fallback in a tracked repository: %v\n%s", target[0], err, out)
		}
	}
}
