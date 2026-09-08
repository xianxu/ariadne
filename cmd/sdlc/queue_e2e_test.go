package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// End-to-end over REAL git: the verb, the pure core, and gitx.TrunkFile wired
// together against a real bare origin.
//
// The unit tests above use a fake trunk, which proves the verb's logic but not
// the wiring — and the wiring is where this feature's risk actually lives, since
// it spans a package boundary and a plumbing sequence. Nothing exercised the
// assembled path until this test.
func TestQueueE2E_AgainstRealGit(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, repo, "remote", "add", "origin", origin)
	testfix.Git(t, repo, "push", "-q", "-u", "origin", "main")
	// Work from a feature branch with NO worktree on main anywhere — the
	// condition the whole design exists for.
	testfix.Git(t, repo, "checkout", "-q", "-b", "feature")

	store, err := gitx.NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}

	add := func(ref, why, tag string) {
		t.Helper()
		var out, errOut bytes.Buffer
		if err := runQueueEdit(&out, &errOut, store, queue.Intent{
			Op: queue.OpAdd, Ref: ref, WhyNow: why, Tag: tag}); err != nil {
			t.Fatalf("add %s: %v\n%s", ref, err, errOut.String())
		}
	}

	add("ariadne#207", "after #206 merges, same dispatch", "sdlc")
	add("ariadne#188", "supersede: only the retry bullet survives", "sdlc")
	add("pair#171", "floor under attention", "couch")

	// The file landed on the TRUNK, not in the working tree.
	onTrunk := testfix.Capture(t, origin, "show", "main:"+queuePath)
	for _, want := range []string{"ariadne#207", "ariadne#188", "pair#171", "[couch]"} {
		if !strings.Contains(onTrunk, want) {
			t.Errorf("trunk missing %q:\n%s", want, onTrunk)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, queuePath)); !os.IsNotExist(err) {
		t.Error("the working tree must not have been written")
	}
	if dirty := testfix.Capture(t, repo, "status", "--porcelain"); dirty != "" {
		t.Errorf("working tree dirty: %q", dirty)
	}

	// Reorder, then read it back through the verb.
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, store, queue.Intent{
		Op: queue.OpMove, Ref: "pair#171", Anchor: "ariadne#207"}); err != nil {
		t.Fatalf("move: %v\n%s", err, errOut.String())
	}
	out.Reset()
	errOut.Reset()
	if err := runQueueList(&out, &errOut, store); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 || !strings.Contains(lines[0], "pair#171") {
		t.Errorf("move did not take effect through the real store:\n%s", out.String())
	}

	// Remove converges, and the refusal path works over real git too.
	out.Reset()
	errOut.Reset()
	if err := runQueueEdit(&out, &errOut, store, queue.Intent{
		Op: queue.OpMove, Ref: "ariadne#207", Anchor: "never#0"}); err == nil {
		t.Error("a missing anchor must refuse over real git as well")
	} else if !strings.Contains(errOut.String(), "ariadne#188") {
		t.Errorf("the refusal must print the real trunk state:\n%s", errOut.String())
	}
}
