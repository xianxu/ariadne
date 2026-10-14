package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestAssetDestinationPolicyRefusesStaleLocalAllowance(t *testing.T) {
	for _, mode := range []string{"republish", "first", "legacy", "abandon"} {
		t.Run(mode, func(t *testing.T) {
			var r *trackerRepo
			id, detail := "000009", "workshop/issues/000009-s09.md"
			switch mode {
			case "first":
				r = spinOffOnCodeBranch(t)
				id, detail = "000008", spinOffDetails
			case "abandon":
				r, _, _ = startedHere(t)
				detail = handoffDetail
			default:
				r, _ = claimSetRepo(t)
			}
			asset := assetFixture(t, r, id, detail)
			if mode != "republish" {
				r.git("add", "workshop/issues")
				r.git("commit", "-qm", "new evidence")
			}
			peerAdd(t, r, assetPolicyPath, `{"version":1,"policy":"text-only"}`)
			before := r.originMain()
			var err error
			switch mode {
			case "republish":
				_, err = publish(t, 9)
			case "first":
				_, err = publish(t, 8)
			case "legacy":
				err = publishIssueCommit(context.Background(), io.Discard, io.Discard, r.root, "HEAD", "workshop/issues", "workshop/history")
			case "abandon":
				err = abandon9("punt", "later")
			}
			if err == nil || !strings.Contains(err.Error(), "text-only") {
				t.Fatalf("stale local allowance authorized %s: %v", mode, err)
			}
			if r.originMain() != before {
				t.Fatal("refusal changed main")
			}
			if _, err := os.Stat(filepath.Join(r.root, asset)); err != nil {
				t.Fatalf("source lost: %v", err)
			}
			if mode == "abandon" && remoteRef(t, r, s09Archive) != "" {
				t.Fatal("private evidence reached archive ref")
			}
		})
	}
}

func TestAssetDestinationPolicyRecheckedAfterPublicationRace(t *testing.T) {
	r, _ := claimSetRepo(t)
	assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	previous := mainPublish
	t.Cleanup(func() { mainPublish = previous })
	var destination string
	mainPublish = func(env *trackerEnv, msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error), before func(string, string) error) error {
		return previous(env, msg, prepare, func(base, candidate string) error {
			if destination == "" {
				peerAdd(t, r, assetPolicyPath, `{"version":1,"policy":"text-only"}`)
				destination = r.originMain()
			}
			return before(base, candidate)
		})
	}
	_, err := publish(t, 9)
	if err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("retry ignored destination policy: %v", err)
	}
	if destination == "" || r.originMain() != destination {
		t.Fatal("retry published evidence after destination tightened policy")
	}
}

func TestAssetDestinationPolicyRecoveryRechecksPinnedCandidate(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	asset := assetFixture(t, r, "000008", spinOffDetails)
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "private evidence")
	hooks := t.TempDir()
	hook := filepath.Join(hooks, "pre-push")
	script := "#!/bin/sh\nwhile read local oid remote old; do\n case $remote in refs/heads/main) exit 1;; esac\ndone\n"
	if err := os.WriteFile(hook, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	if _, err := publish(t, 8); err == nil {
		t.Fatal("fixture did not block publication")
	}
	peerAdd(t, r, assetPolicyPath, `{"version":1,"policy":"text-only"}`)
	before := r.originMain()
	if err := os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	err := runRecoveryReconcile(context.Background(), io.Discard, io.Discard, 8)
	if err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("recovery ignored destination policy: %v", err)
	}
	if r.originMain() != before {
		t.Fatal("recovery published private bytes")
	}
	if _, err := os.Stat(filepath.Join(r.root, asset)); err != nil {
		t.Fatalf("recovery removed source: %v", err)
	}
}

func TestAssetDestinationPolicyIntersectsLocalRestrictions(t *testing.T) {
	for _, mode := range []string{"destination-missing", "local-text-only", "invalid-destination"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			local := `{"version":1,"policy":"allowed"}`
			switch mode {
			case "local-text-only":
				peerAdd(t, r, assetPolicyPath, local)
				local = `{"version":1,"policy":"text-only"}`
			case "invalid-destination":
				peerAdd(t, r, assetPolicyPath, `{"version":1,"policy":"allowd"}`)
			}
			writeRepoFile(t, r.root, assetPolicyPath, local)
			tf, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "main")
			if err != nil {
				t.Fatal(err)
			}
			view, err := tf.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if err := requireAssetPublicationPolicy(context.Background(), r.root, view); err == nil {
				t.Fatal("local or unknown destination restriction bypassed")
			}
		})
	}
}

func TestAssetDestinationPolicyRejectsNonregularAndOversizedFiles(t *testing.T) {
	for _, mode := range []string{"directory", "symlink", "ancestor-symlink", "oversized"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			writeRepoFile(t, r.root, assetPolicyPath, `{"version":1,"policy":"allowed"}`)
			peer := testfix.Repo(t, testfix.InitialCommit())
			switch mode {
			case "directory":
				writeRepoFile(t, peer, assetPolicyPath+"/value", `{"version":1,"policy":"allowed"}`)
			case "oversized":
				writeRepoFile(t, peer, assetPolicyPath, strings.Repeat(" ", 4097))
			case "symlink":
				if err := os.MkdirAll(filepath.Join(peer, "workshop"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("elsewhere", filepath.Join(peer, assetPolicyPath)); err != nil {
					t.Fatal(err)
				}
			case "ancestor-symlink":
				if err := os.Symlink("elsewhere", filepath.Join(peer, "workshop")); err != nil {
					t.Fatal(err)
				}
			}
			git(t, peer, "add", "workshop")
			git(t, peer, "commit", "-qm", "invalid destination policy")
			r.git("fetch", "-q", peer, "HEAD")
			tf, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "main")
			if err != nil {
				t.Fatal(err)
			}
			view := tf.ViewOf(r.git("rev-parse", "FETCH_HEAD"))
			err = requireAssetPublicationPolicy(context.Background(), r.root, view)
			if err == nil || !strings.Contains(err.Error(), "operator") {
				t.Fatalf("invalid destination policy accepted or lacks remedy: %v", err)
			}
		})
	}
}

func TestAssetDestinationPolicyCannotBeRelaxedByCandidate(t *testing.T) {
	r, _ := claimSetRepo(t)
	detail := "workshop/issues/000009-s09.md"
	asset := assetFixture(t, r, "000009", detail)
	writeRepoFile(t, r.root, assetPolicyPath, `{"version":1,"policy":"text-only"}`)
	r.git("add", assetPolicyPath)
	r.git("commit", "-qm", "destination tightened policy", "--only", "--", assetPolicyPath)
	base := r.git("rev-parse", "HEAD")
	writeRepoFile(t, r.root, assetPolicyPath, `{"version":1,"policy":"allowed"}`)
	r.git("add", "workshop")
	r.git("commit", "-qm", "candidate relaxes policy and adds evidence")
	next := r.git("rev-parse", "HEAD")
	tf, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	err = validateAssetPublication(context.Background(), r.root, tf.ViewOf(base), tf.ViewOf(next), []string{detail, asset, assetPolicyPath}, "workshop/issues", "workshop/history")
	if err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("candidate authorized its own assets: %v", err)
	}
}
