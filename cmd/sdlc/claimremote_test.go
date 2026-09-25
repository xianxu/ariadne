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

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestClaimDecisionOnlyReservesOpenWellFormedRecords(t *testing.T) {
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
}

func claimFlagsFor(id int) *claimFlags {
	return &claimFlags{Issue: id, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
}

func TestClaimRefusesUntilDetailsLandOnMain(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	// Card-only, and details present only locally: neither completes creation.
	writeRepoFile(t, r.root, detailPath, detail)
	var out, errs bytes.Buffer
	err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9))
	if err == nil || !strings.Contains(err.Error(), "not claimable yet") || !strings.Contains(err.Error(), "move-detail --issue 000009") {
		t.Fatalf("claimed an incompletely created issue: %v", err)
	}
	if got := r.card(cardPath); got != card {
		t.Fatalf("refused claim changed the card:\n%s", got)
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, detailPath)); string(raw) != detail {
		t.Fatal("refused claim touched local details")
	}
}

func TestClaimReservesCardAndRefreshesLocalMirror(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	mainBefore, headBefore := r.originMain(), r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	now := r.card(cardPath)
	if !strings.Contains(now, "status: working") || !strings.Contains(now, "started:") {
		t.Fatalf("card not reserved:\n%s", now)
	}
	if r.originMain() != mainBefore || r.git("rev-parse", "HEAD") != headBefore {
		t.Fatal("claim moved main or committed locally")
	}
	local, _ := os.ReadFile(filepath.Join(r.root, detailPath))
	if !strings.Contains(string(local), "status: working") || string(local) == detail {
		t.Fatalf("local mirror not refreshed:\n%s", local)
	}
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err == nil || !strings.Contains(err.Error(), "not open") {
		t.Fatalf("repeated claim = %v", err)
	}
}

func TestClaimLeavesHandEditedMirrorAndWarns(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	edited := strings.Replace(detail, "status: open", "status: blocked", 1)
	writeRepoFile(t, r.root, detailPath, edited)
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, detailPath)); string(raw) != edited {
		t.Fatal("refresh overwrote a hand-edited mirrored field")
	}
	if !strings.Contains(errs.String(), "owned by the card") {
		t.Fatalf("no ownership warning:\n%s", errs.String())
	}
}

func TestClaimRaceHasExactlyOneWinner(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "Peer")
	git(t, peer, "config", "user.email", "peer@example.com")
	dirs := []string{r.root, peer}
	barrier := t.TempDir()
	hooks := t.TempDir()
	// Each claimant reaches pre-push only after pinning its candidate on the
	// open card; neither pushes until both have.
	for i, dir := range dirs {
		hookDir := filepath.Join(hooks, fmt.Sprint(i))
		if err := os.MkdirAll(hookDir, 0o755); err != nil {
			t.Fatal(err)
		}
		hook := fmt.Sprintf("#!/bin/sh\ntouch '%s/ready%d'\nn=0\nwhile [ ! -f '%s/ready%d' ]; do n=$((n+1)); [ \"$n\" -lt 1000 ] || exit 1; sleep 0.01; done\n", barrier, i, barrier, 1-i)
		if err := os.WriteFile(filepath.Join(hookDir, "pre-push"), []byte(hook), 0o755); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "config", "core.hooksPath", hookDir)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	type result struct {
		out string
		err error
	}
	results := make(chan result, 2)
	for _, dir := range dirs {
		cmd := exec.CommandContext(ctx, binary, "claim", "--issue", "9")
		cmd.Dir = dir
		cmd.WaitDelay = 2 * time.Second
		go func() { out, err := cmd.CombinedOutput(); results <- result{string(out), err} }()
	}
	wins := 0
	for range 2 {
		res := <-results
		if res.err == nil {
			wins++
		} else if !strings.Contains(res.out, "not open") && !strings.Contains(res.out, "changed while claiming") {
			t.Errorf("loser was not a status/CAS refusal: %v %s", res.err, res.out)
		}
	}
	if wins != 1 {
		t.Fatalf("claim winners=%d; want exactly one", wins)
	}
	if got := r.card(cardPath); !strings.Contains(got, "status: working") {
		t.Fatalf("card = %s", got)
	}
}

func TestIssueNewRaceAllocatesDistinctIDs(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "Peer")
	git(t, peer, "config", "user.email", "peer@example.com")
	dirs := []string{r.root, peer}
	barrier, hooks := t.TempDir(), t.TempDir()
	for i, dir := range dirs {
		hookDir := filepath.Join(hooks, fmt.Sprint(i))
		if err := os.MkdirAll(hookDir, 0o755); err != nil {
			t.Fatal(err)
		}
		hook := fmt.Sprintf("#!/bin/sh\ntouch '%s/ready%d'\nn=0\nwhile [ ! -f '%s/ready%d' ]; do n=$((n+1)); [ \"$n\" -lt 1000 ] || exit 1; sleep 0.01; done\n", barrier, i, barrier, 1-i)
		if err := os.WriteFile(filepath.Join(hookDir, "pre-push"), []byte(hook), 0o755); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "config", "core.hooksPath", hookDir)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	type result struct {
		out string
		err error
	}
	results := make(chan result, 2)
	for _, dir := range dirs {
		cmd := exec.CommandContext(ctx, binary, "issue", "new", "Identical creation")
		cmd.Dir = dir
		cmd.WaitDelay = 2 * time.Second
		go func() { out, err := cmd.CombinedOutput(); results <- result{string(out), err} }()
	}
	for range 2 {
		if res := <-results; res.err != nil {
			t.Errorf("creation failed: %v %s", res.err, res.out)
		}
	}
	for _, id := range []string{"000008", "000009"} {
		if got := r.card("workshop/issue-cards/" + id + "-identical-creation.md"); !strings.Contains(got, "id: "+id) {
			t.Fatalf("reservation %s missing: %q", id, got)
		}
	}
}

func TestClaimOfflineRefusesWithoutLocalMutation(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	git(t, r.root, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err == nil {
		t.Fatal("offline claim succeeded")
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, detailPath)); string(raw) != detail {
		t.Fatal("offline claim edited local details")
	}
}
