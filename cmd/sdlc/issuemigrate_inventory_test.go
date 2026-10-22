package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestMigrationInventoryRefusesFailedRemoteObservation(t *testing.T) {
	r := emptyInitRepo(t)
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	main, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	r.git("remote", "set-url", "origin", filepath.Join(t.TempDir(), "unreachable.git"))
	if _, err := migrationInventory(env, main); err == nil {
		t.Fatal("failed publication-remote observation authorized inventory")
	}
	if got := testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker"); got != "" {
		t.Fatal("failed observation published tracker")
	}
}

func TestMigrationInventoryRefusesMovedMain(t *testing.T) {
	r := emptyInitRepo(t)
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	main, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// Main advances after its snapshot but before the branch-inventory snapshot.
	clone := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", r.origin, clone)
	writeRepoFile(t, clone, "workshop/issues/000001-new.md", "---\nid: 000001\nstatus: open\n---\n\n# Work\n")
	testfix.Git(t, clone, "add", ".")
	testfix.Git(t, clone, "commit", "-qm", "main moved")
	testfix.Git(t, clone, "push", "-q")
	if _, err := migrationInventory(env, main); err == nil || !strings.Contains(err.Error(), "main moved") {
		t.Fatalf("mixed main snapshots authorized inventory: %v", err)
	}
	if got := testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker"); got != "" {
		t.Fatal("mixed snapshot published tracker")
	}
}

func TestMigrationInventoryPinsIssueAndAnchorReads(t *testing.T) {
	r := emptyInitRepo(t)
	const name = "refs/heads/legacy-work"
	const issuePath = "workshop/issues/000001-work.md"
	r.git("switch", "-qc", "legacy-work")
	closed := "---\nid: 000001\nstatus: codecomplete\nactual_hours: 1\ncreated: 2026-10-21\n---\n\n# Work\n\n## Problem\n\nWork.\n"
	writeRepoFile(t, r.root, issuePath, closed)
	r.git("add", ".")
	r.git("commit", "-qm", "#1: close\n\nReview-Verdict: SHIP")
	anchor := r.git("rev-parse", "HEAD")
	r.git("switch", "-q", "main")
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	main, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	refs, err := publicationBranchTips(env, main.Ref())
	if err != nil {
		t.Fatal(err)
	}
	var pinned migrationBranchRef
	for _, ref := range refs {
		if ref.Name == name {
			pinned = ref
		}
	}
	if pinned.OID != anchor {
		t.Fatalf("wrong pin: %+v", pinned)
	}
	// A name changes after the snapshot. Both evidence consumers must still
	// observe the same commit, rather than silently assembling mixed evidence.
	r.git("update-ref", name, main.Ref())
	files, err := branchIssueEdits(env, pinned, main.Ref(), "workshop/issues")
	if err != nil || len(files) != 1 || string(files[0].Raw) != closed || files[0].Branch != "legacy-work" {
		t.Fatalf("issue read followed moving ref: %+v %v", files, err)
	}
	anchors, err := legacyCloseAnchors(env, main.Ref(), []migrationBranchRef{pinned}, issuePath)
	if err != nil || len(anchors) != 1 || anchors[0].Anchor != anchor || anchors[0].Ref != name || anchors[0].CodeAfter {
		t.Fatalf("anchor read followed moving ref: %+v %v", anchors, err)
	}
}

func TestMigrationInventoryFetchOnlyOwnHeadsMapping(t *testing.T) {
	r := emptyInitRepo(t)
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	main, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	r.git("tag", "retained")
	r.git("update-ref", "refs/remotes/origin/deleted", main.Ref())
	r.git("update-ref", "refs/remotes/unrelated/keep", main.Ref())
	// User preferences must not expand this explicit observation's write set.
	r.git("config", "fetch.pruneTags", "true")
	r.git("config", "--replace-all", "remote.origin.fetch", "+refs/heads/*:refs/heads/foreign/*")
	fetchHead := filepath.Join(r.root, ".git", "FETCH_HEAD")
	sentinel := []byte("preserve existing FETCH_HEAD\n")
	if err := os.WriteFile(fetchHead, sentinel, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := publicationBranchTips(env, main.Ref()); err != nil {
		t.Fatal(err)
	}
	if r.git("rev-parse", "refs/tags/retained") != main.Ref() || r.git("rev-parse", "refs/remotes/unrelated/keep") != main.Ref() {
		t.Fatal("explicit heads refresh changed tags or another remote")
	}
	if r.git("for-each-ref", "--format=%(refname)", "refs/remotes/origin/deleted", "refs/heads/foreign") != "" {
		t.Fatal("heads refresh retained deleted cache or used configured write mapping")
	}
	raw, err := os.ReadFile(fetchHead)
	if err != nil || string(raw) != string(sentinel) {
		t.Fatalf("FETCH_HEAD changed: %q %v", raw, err)
	}
}
