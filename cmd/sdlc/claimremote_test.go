package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestClaimDecisionOnlyReservesOpenWellFormedRecords(t *testing.T) {
	for _, status := range append(vocab.Issue().AllStatuses(), "unknown", "") {
		raw := []byte("---\nid: 000031\nstatus: " + status + "\n---\nbody\n")
		got, err := claimDecision(raw, 31, "2026-09-23", "2026-09-23T12:00:00Z", nil)
		if status == "open" {
			if err != nil || !bytes.Contains(got, []byte("status: working")) {
				t.Fatalf("open: %s %v", got, err)
			}
		} else if err == nil {
			t.Errorf("accepted status %q", status)
		}
	}
	for _, raw := range []string{"", "---\nstatus: open\n---\n", "---\nid: 31\nstatus: open\nstatus: working\n---\n", "---\nid: 31\nstatus: [\n---\n"} {
		if _, err := claimDecision([]byte(raw), 31, "today", "now", nil); err == nil {
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
	if err == nil || !strings.Contains(err.Error(), "not claimable yet") || !strings.Contains(err.Error(), "move-detail --issue 9`") {
		t.Fatalf("claimed an incompletely created issue: %v", err)
	}
	if got := r.card(cardPath); got != card {
		t.Fatalf("refused claim changed the card:\n%s", got)
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, detailPath)); string(raw) != detail {
		t.Fatal("refused claim touched local details")
	}
}

func TestClaimReservesCardAndLeavesRestUntouched(t *testing.T) {
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
	// Refreshing the mirror on rest would dirty it; start-plan refreshes on the
	// issue branch instead (regression: claim → start-plan refused a dirty rest).
	if dirty := r.git("status", "--porcelain"); dirty != "" {
		t.Fatalf("claim dirtied the resting branch: %s", dirty)
	}
	// #277: the card names its owner; the owner's repeat claim is a no-op, and
	// another workspace's is refused, naming the owner — never a takeover.
	owner, ok, err := issue.CardClaimant([]byte(now))
	if err != nil || !ok || owner.Worktree != canonRoot(r.root) {
		t.Fatalf("claimant %+v %v %v:\n%s", owner, ok, err, now)
	}
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil || r.card(cardPath) != now {
		t.Fatalf("owner's repeated claim = %v (card changed: %v)", err, r.card(cardPath) != now)
	}
	elsewhere := owner
	elsewhere.Worktree = "/elsewhere/ariadne"
	withClaimant(t, elsewhere)
	err = runClaim(context.Background(), &out, &errs, claimFlagsFor(9))
	if err == nil || !strings.Contains(err.Error(), "claimed by "+owner.Operator) || r.card(cardPath) != now {
		t.Fatalf("another workspace's repeated claim = %v", err)
	}
}

func TestClaimRefreshesMirrorOnAFeatureBranch(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	r.git("switch", "-q", "-c", "000009-nine")
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	local, _ := os.ReadFile(filepath.Join(r.root, detailPath))
	if !strings.Contains(string(local), "status: working") || string(local) == detail {
		t.Fatalf("local mirror not refreshed:\n%s", local)
	}
	// #277: people reading the details see who owns the issue — a plain clone,
	// so no slot label.
	if !strings.Contains(string(local), "claimant:\n") || !strings.Contains(string(local), "worktree: "+canonRoot(r.root)) || strings.Contains(string(local), "workspace:") {
		t.Fatalf("details do not show the owner:\n%s", local)
	}
}

func TestClaimLeavesHandEditedMirrorAndWarns(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	r.git("switch", "-q", "-c", "000009-nine")
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
	if _, err := machineID(); err != nil {
		t.Skipf("claims need the host machine ID, unreadable here: %v", err)
	}
	binary := buildFleetE2EBinary(t)
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "Peer")
	git(t, peer, "config", "user.email", "peer@example.com")
	dirs := []string{r.root, peer}
	results := raceBuiltBinary(t, binary, dirs, "claim", "--issue", "9")
	wins := 0
	var winner, loser string
	for _, res := range results {
		if res.err == nil {
			wins++
			winner = res.dir
		} else if loser = res.dir; !strings.Contains(res.out, "not open") && !strings.Contains(res.out, "changed while claiming") && !strings.Contains(res.out, "claimed by") {
			t.Errorf("loser was not a status/CAS/owner refusal: %v %s", res.err, res.out)
		}
	}
	if wins != 1 {
		t.Fatalf("claim winners=%d; want exactly one", wins)
	}
	got := r.card(cardPath)
	if !strings.Contains(got, "status: working") {
		t.Fatalf("card = %s", got)
	}
	// #277: the winner's complete ownership landed in the same card write; the
	// loser published none. Identities differ by the clones' real worktrees (a
	// built binary cannot see an in-process seam).
	owner, ok, err := issue.CardClaimant([]byte(got))
	if err != nil || !ok || owner.Worktree != canonRoot(winner) || owner.Operator == "" || owner.Repository == "" || owner.MachineName == "" {
		t.Fatalf("winner's claimant %+v %v %v:\n%s", owner, ok, err, got)
	}
	if history := git(t, r.root, "log", "-p", "origin/issue-tracker", "--", cardPath); strings.Contains(history, canonRoot(loser)) {
		t.Fatal("the losing claimant's worktree reached the tracker")
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

// #277: claimDecision with an ownership identity — open is stamped; a working
// card is the owner's (no-op), another workspace's (refused, naming the owner)
// or unattributed (refused toward --adopt).
func TestClaimDecisionOwnership(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	open := []byte("---\nid: 000031\nstatus: open\ncreated: 2026-10-01\nupdated: 2026-10-01\n---\n\n# t\n\n## Problem\n\nx\n")
	claimed, err := claimDecision(open, 31, "2026-10-01", "2026-10-01T12:00:00Z", &me)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok, _ := issue.CardClaimant(claimed); !ok || got != me {
		t.Fatalf("open card not stamped:\n%s", claimed)
	}
	other := me
	other.Operator, other.Worktree = "Them", "/w/b"
	unattributed := bytes.Replace(open, []byte("status: open"), []byte("status: working"), 1)
	for name, c := range map[string]struct {
		raw  []byte
		want string
	}{
		"owner repeats":  {claimed, "already claimed by this workspace"},
		"another claims": {claimed, "claimed by Me on box at /w/a"},
		"unattributed":   {unattributed, "--adopt"},
		"codecomplete":   {bytes.Replace(claimed, []byte("status: working"), []byte("status: blocked"), 1), "not open"},
	} {
		who := me
		if name == "another claims" {
			who = other
		}
		_, err := claimDecision(c.raw, 31, "2026-10-01", "now", &who)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v (want %q)", name, err, c.want)
		}
	}
	if _, err := claimDecision(claimed, 31, "2026-10-01", "now", &me); !errors.Is(err, errAlreadyMine) {
		t.Fatalf("owner's repeat is not the no-op sentinel: %v", err)
	}
}

// #277: an owner's repeat claim under --dry-run changes nothing, not even the
// local mirror.
func TestClaimDryRunOwnerRepeatWritesNothing(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	r.git("switch", "-q", "-c", "000009-nine")
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	writeRepoFile(t, r.root, detailPath, detail) // a stale mirror a real run would refresh
	f := claimFlagsFor(9)
	f.DryRun = true
	if err := runClaim(context.Background(), &out, &errs, f); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, detailPath)); string(raw) != detail {
		t.Fatal("a dry-run repeat claim wrote the details")
	}
}

type raceResult struct {
	dir, out string
	err      error
}

// raceBuiltBinary runs the built binary with args in each dir concurrently. A
// pre-push hook barrier holds each process until all have pinned their
// candidate, so every racer reads the same tracker state before anyone pushes.
func raceBuiltBinary(t *testing.T, binary string, dirs []string, args ...string) []raceResult {
	t.Helper()
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
	ch := make(chan raceResult, len(dirs))
	for _, dir := range dirs {
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Dir = dir
		cmd.WaitDelay = 2 * time.Second
		go func() { out, err := cmd.CombinedOutput(); ch <- raceResult{dir, string(out), err} }()
	}
	var results []raceResult
	for range dirs {
		results = append(results, <-ch)
	}
	return results
}
