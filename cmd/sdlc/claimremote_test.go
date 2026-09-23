package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestClaimRemoteStatusOnly(t *testing.T) {
	repo, origin := syncRepo(t)
	remote := "---\nid: 206\nstatus: open\n---\nremote body\n"
	writeSyncIssue(t, repo, filepath.Base(issuePath206), remote)
	git(t, repo, "add", "--", issuePath206)
	git(t, repo, "commit", "-m", "open issue")
	git(t, repo, "push", "origin", "main")
	local := strings.Replace(remote, "remote body", "unpublished local design", 1)
	writeSyncIssue(t, repo, filepath.Base(issuePath206), local)
	head := git(t, repo, "rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runClaim(&out, &errs, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}); err != nil {
		t.Fatal(err)
	}
	got := git(t, origin, "show", "main:"+issuePath206)
	if strings.Contains(got, "unpublished") || !strings.Contains(got, "status: working") {
		t.Fatalf("remote claim swept local body: %s", got)
	}
	if got := git(t, repo, "rev-parse", "HEAD"); got != head {
		t.Fatalf("claim moved local branch: %s != %s", got, head)
	}
	raw, err := os.ReadFile(filepath.Join(repo, issuePath206))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "unpublished local design") || !strings.Contains(string(raw), "status: working") {
		t.Fatalf("local reconciliation: %s", raw)
	}
	if err := runClaim(&out, &errs, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}); err == nil || !strings.Contains(err.Error(), "working") {
		t.Fatalf("repeated claim = %v", err)
	}
}

func TestCreationOccupiedID(t *testing.T) {
	const path = "workshop/issues/000206-same.md"
	v, paths := decideCollision(206, path, map[int][]string{206: {path}}, nil, true)
	if v != verdictReallocate || len(paths) != 1 {
		t.Fatalf("occupied same slug: %v %v", v, paths)
	}
}

func TestCreationRemoteRace(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	repo, origin := syncRepo(t)
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", origin, peer)
	for _, dir := range []string{repo, peer} {
		git(t, dir, "config", "user.name", "Identical")
		git(t, dir, "config", "user.email", "same@example.com")
	}
	barrier := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	type result struct {
		out string
		err error
	}
	results := make(chan result, 2)
	for i, dir := range []string{repo, peer} {
		// Both reservations are constructed against the same free remote ID.
		// Bound the hook itself so a failing peer never leaves a waiting child.
		hook := fmt.Sprintf("#!/bin/sh\ntouch '%s/ready%d'\nn=0\nwhile [ ! -f '%s/ready%d' ]; do n=$((n+1)); [ \"$n\" -lt 1000 ] || exit 1; sleep 0.01; done\n", barrier, i, barrier, 1-i)
		if err := os.WriteFile(filepath.Join(dir, ".git/hooks/pre-push"), []byte(hook), 0755); err != nil {
			t.Fatal(err)
		}
		cmd := exec.CommandContext(ctx, binary, "issue", "new", "Identical creation")
		cmd.Dir = dir
		cmd.WaitDelay = 2 * time.Second
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-09-23T12:00:00Z", "GIT_COMMITTER_DATE=2026-09-23T12:00:00Z")
		go func() { out, err := cmd.CombinedOutput(); results <- result{string(out), err} }()
	}
	for range 2 {
		r := <-results
		if r.err != nil || strings.Contains(r.out, "not broadcast") || strings.Contains(r.out, "uncertain") {
			t.Errorf("creation failed: %v %s", r.err, r.out)
		}
	}
	for _, id := range []int{207, 208} {
		p := fmt.Sprintf("workshop/issues/%06d-identical-creation.md", id)
		if got := git(t, origin, "show", "main:"+p); !strings.Contains(got, fmt.Sprintf("id: %06d", id)) {
			t.Fatalf("reservation %d: %s", id, got)
		}
	}
}

func TestClaimNonOpenAndUncertain(t *testing.T) {
	for _, status := range append(vocab.Issue().AllStatuses(), "unknown", "") {
		raw := []byte("---\nid: 000031\nstatus: " + status + "\n---\nbody\n")
		got, err := claimDecision(raw, 31, "2026-09-23", "2026-09-23T12:00:00Z")
		if status == "open" {
			if err != nil || !bytes.Contains(got, []byte("status: working")) {
				t.Fatalf("open: %s %v", got, err)
			}
		} else if err == nil {
			t.Errorf("accepted status %q", status)
		}
	}
	for _, raw := range []string{"", "---\nstatus: open\n---\n", "---\nid: 31\nstatus: open\nstatus: working\n---\n", "---\nid: 31\nstatus: [\n---\n"} {
		if _, err := claimDecision([]byte(raw), 31, "today", "now"); err == nil {
			t.Errorf("accepted malformed record: %q", raw)
		}
	}
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 206\nstatus: open\n---\nbody\n")
	git(t, repo, "add", "--", issuePath206)
	git(t, repo, "commit", "-m", "open")
	before, _ := os.ReadFile(filepath.Join(repo, issuePath206))
	saved := newTrunkPublisher
	t.Cleanup(func() { newTrunkPublisher = saved })
	newTrunkPublisher = func(string) (trunkPublisher, error) {
		return &fakePublisher{view: viewFor(t, repo), err: gitx.ErrPublicationUncertain}, nil
	}
	var out, errs bytes.Buffer
	if err := runClaim(&out, &errs, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}); err == nil {
		t.Fatal("uncertain claim succeeded")
	}
	after, _ := os.ReadFile(filepath.Join(repo, issuePath206))
	if !bytes.Equal(before, after) {
		t.Fatal("uncertain claim reconciled local status")
	}
}

func TestClaimRemoteIdentityMismatch(t *testing.T) {
	repo, origin := syncRepo(t)
	writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 206\nstatus: open\n---\nbody\n")
	git(t, repo, "add", "--", issuePath206)
	git(t, repo, "commit", "-m", "open")
	git(t, repo, "push", "origin", "main")
	other := filepath.Join(repo, syncIssuesDir, "000206-other.md")
	if err := os.Rename(filepath.Join(repo, issuePath206), other); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if err := runClaim(&out, &errs, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}); err == nil || !strings.Contains(err.Error(), "different local and remote names") {
		t.Fatalf("mismatched identity: %v", err)
	}
	if err := os.Remove(other); err != nil {
		t.Fatal(err)
	}
	if err := runClaim(&out, &errs, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}); err != nil {
		t.Fatalf("remote-only claim: %v", err)
	}
	if got := git(t, origin, "show", "main:"+issuePath206); !strings.Contains(got, "status: working") {
		t.Fatal(got)
	}
}

func TestClaimRemoteStatusRace(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	for _, linked := range []bool{false, true} {
		t.Run(fmt.Sprintf("linked=%v", linked), func(t *testing.T) {
			repo, origin := syncRepo(t)
			writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 206\nstatus: open\n---\nremote body\n")
			git(t, repo, "add", "--", issuePath206)
			git(t, repo, "commit", "-m", "open")
			git(t, repo, "push", "origin", "main")
			dirs := []string{repo, filepath.Join(t.TempDir(), "peer")}
			if linked {
				git(t, repo, "worktree", "add", "-b", "peer", dirs[1])
			} else {
				git(t, "", "clone", origin, dirs[1])
				git(t, dirs[1], "config", "user.name", "Test")
				git(t, dirs[1], "config", "user.email", "test@example.com")
			}
			barrier := t.TempDir()
			if !linked {
				// pre-push is reached only after this command read remote open and
				// constructed its claim. Both clones must reach it before either pushes.
				for i, dir := range dirs {
					hook := fmt.Sprintf("#!/bin/sh\ntouch '%s/ready%d'\nwhile [ ! -f '%s/release' ]; do sleep 0.02; done\n", barrier, i, barrier)
					if err := os.WriteFile(filepath.Join(dir, ".git/hooks/pre-push"), []byte(hook), 0755); err != nil {
						t.Fatal(err)
					}
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			type result struct {
				out string
				err error
			}
			results := make(chan result, 2)
			pending := 0
			defer func() {
				_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0644)
				cancel()
				for pending > 0 {
					select {
					case <-results:
						pending--
					case <-time.After(5 * time.Second):
						t.Errorf("claim subprocess did not reap")
						return
					}
				}
			}()
			for i, dir := range dirs {
				writeSyncIssue(t, dir, filepath.Base(issuePath206), fmt.Sprintf("---\nid: 206\nstatus: open\n---\nlocal draft %d\n", i))
				var cmd *exec.Cmd
				if linked {
					// Rendezvous before launching the production CLI, hence before its
					// common-directory lock; never bypass the lock for the test.
					script := "touch \"$1/ready$2\"; while [ ! -f \"$1/release\" ]; do sleep 0.02; done; exec \"$3\" claim --issue 206"
					cmd = exec.CommandContext(ctx, "sh", "-c", script, "sh", barrier, fmt.Sprint(i), binary)
				} else {
					cmd = exec.CommandContext(ctx, binary, "claim", "--issue", "206")
				}
				cmd.Dir = dir
				cmd.WaitDelay = 2 * time.Second
				pending++
				go func() { out, err := cmd.CombinedOutput(); results <- result{string(out), err} }()
			}
			deadline := time.Now().Add(20 * time.Second)
			for {
				_, a := os.Stat(filepath.Join(barrier, "ready0"))
				_, b := os.Stat(filepath.Join(barrier, "ready1"))
				if a == nil && b == nil {
					break
				}
				if time.Now().After(deadline) {
					cancel()
					t.Fatal("claimants did not rendezvous")
				}
				time.Sleep(10 * time.Millisecond)
			}
			if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0644); err != nil {
				t.Fatal(err)
			}
			wins := 0
			for range 2 {
				r := <-results
				pending--
				if r.err == nil {
					wins++
				} else if !strings.Contains(r.out, "not open") {
					t.Errorf("loser was not status refusal: %v %s", r.err, r.out)
				}
			}
			if wins != 1 {
				t.Fatalf("claim winners=%d; want exactly one", wins)
			}
			remote := git(t, origin, "show", "main:"+issuePath206)
			if !strings.Contains(remote, "status: working") || strings.Contains(remote, "local draft") {
				t.Fatalf("remote = %s", remote)
			}
			for i, dir := range dirs {
				raw, err := os.ReadFile(filepath.Join(dir, issuePath206))
				if err != nil || !strings.Contains(string(raw), fmt.Sprintf("local draft %d", i)) {
					t.Fatalf("local body lost: %s %v", raw, err)
				}
			}
		})
	}
}
