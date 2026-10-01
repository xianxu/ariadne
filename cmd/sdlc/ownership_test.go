package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// ownerOf reads #id's recorded claimant from the tracker.
func ownerOf(t *testing.T, r *trackerRepo, cardPath string) (issue.Claimant, bool) {
	t.Helper()
	c, ok, err := issue.CardClaimant([]byte(r.card(cardPath)))
	if err != nil {
		t.Fatal(err)
	}
	return c, ok
}

// dropClaimant rewrites #id's card without a claimant — a working card claimed
// before #277.
func dropClaimant(t *testing.T, r *trackerRepo, id, cardPath string) {
	t.Helper()
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	err = env.repo.ChangeCard(id, cardPath, "legacy", operationToken("set"), func(c []byte) ([]byte, error) {
		fm, body, err := issue.Parse(string(c))
		if err != nil {
			return nil, err
		}
		i := strings.Index(fm, "\nclaimant:")
		if i < 0 {
			return nil, errors.New("fixture: no claimant to drop")
		}
		rest := fm[i+1:]
		end := len(rest)
		for _, line := range strings.SplitAfter(rest, "\n")[1:] {
			if !strings.HasPrefix(line, " ") {
				end = strings.Index(rest, line)
				break
			}
		}
		return []byte(issue.Compose(fm[:i+1]+rest[end:], body)), nil
	})
	invalidateIssueRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ownerOf(t, r, cardPath); ok {
		t.Fatal("fixture: claimant still present")
	}
}

// ownershipGate runs one continuation gate in-process, reporting its refusal (die or
// error) and whether the review judge was dispatched.
func ownershipGate(t *testing.T, r *trackerRepo, name string, id int, detailPath string) (refusal string, judged bool) {
	t.Helper()
	calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	idArg := itoa(id)
	run := func(args ...string) string {
		var out string
		msg, died := expectDie(t, func() {
			_, stderr, err := executeSDLCTestCommand(args...)
			if err != nil {
				out = err.Error() + "\n" + stderr
			}
		})
		if died {
			return msg
		}
		return out
	}
	switch name {
	case "start-plan":
		refusal = run("start-plan", "--issue", idArg)
	case "change-code":
		stem := strings.TrimSuffix(filepath.Base(detailPath), ".md")
		if _, err := refreshChangeCodeMirror(&changeCodeFlags{Issue: id}, stem, filepath.Join(r.root, detailPath)); err != nil {
			refusal = err.Error()
		}
	case "close":
		refusal = run("close", "--issue", idArg, "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger")
	case "milestone-close":
		refusal = run("milestone-close", "--issue", idArg, "--milestone", "M1", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
	}
	return refusal, *calls > 0
}

// #277: every continuation gate refuses another workspace's issue (naming the
// owner) and an unattributed one (toward --adopt), before any review runs; the
// owner passes. Adoption then restores the owner's continuation.
func TestOwnershipGatesRefuseForeignAndUnknown(t *testing.T) {
	gates := []string{"start-plan", "change-code", "close", "milestone-close"}
	for i, name := range gates {
		t.Run(name, func(t *testing.T) {
			id := 360 + i
			r, cardPath, detailPath := closeReady(t, id)
			owner, ok := ownerOf(t, r, cardPath)
			if !ok {
				t.Fatal("closeReady's claim recorded no owner")
			}
			elsewhere := owner
			elsewhere.Worktree = "/elsewhere/ariadne"
			withClaimant(t, elsewhere)
			refusal, judged := ownershipGate(t, r, name, id, detailPath)
			if !strings.Contains(refusal, "owned by "+owner.Operator) || judged {
				t.Fatalf("foreign: refusal %q, judged %v", refusal, judged)
			}
			withClaimant(t, owner)
			dropClaimant(t, r, itoa6(id), cardPath)
			refusal, judged = ownershipGate(t, r, name, id, detailPath)
			if !strings.Contains(refusal, "--adopt") || judged {
				t.Fatalf("unknown: refusal %q, judged %v", refusal, judged)
			}
			var out, errs bytes.Buffer
			if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issue: id, IssuesDir: "workshop/issues", HistoryDir: "workshop/history", Adopt: true}); err != nil {
				t.Fatalf("adopt: %v\n%s", err, errs.String())
			}
			if got, _ := ownerOf(t, r, cardPath); got != owner {
				t.Fatalf("adopted owner %+v", got)
			}
			if name == "start-plan" || name == "change-code" {
				if refusal, _ := ownershipGate(t, r, name, id, detailPath); refusal != "" {
					t.Fatalf("owner refused after adopt: %s", refusal)
				}
			}
		})
	}
}

func itoa6(id int) string { return strings.Repeat("0", 6-len(itoa(id))) + itoa(id) }

// #277: --adopt records an owner only on an unattributed working card; it never
// reassigns an owned card and does not claim an open one.
func TestAdoptOnlyUnattributedWork(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000370", "adopt")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	adopt := func() error {
		var out, errs bytes.Buffer
		return runClaim(context.Background(), &out, &errs, &claimFlags{Issue: 370, IssuesDir: "workshop/issues", HistoryDir: "workshop/history", Adopt: true})
	}
	if err := adopt(); err == nil || !strings.Contains(err.Error(), "plain `sdlc claim") {
		t.Fatalf("adopt on an open card = %v", err)
	}
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(370)); err != nil {
		t.Fatal(err)
	}
	owner, _ := ownerOf(t, r, cardPath)
	if err := adopt(); err != nil {
		t.Fatalf("owner's adopt is not a no-op: %v", err)
	}
	other := owner
	other.Worktree = "/elsewhere/ariadne"
	withClaimant(t, other)
	before := r.card(cardPath)
	if err := adopt(); err == nil || !strings.Contains(err.Error(), "never reassigns") || r.card(cardPath) != before {
		t.Fatalf("adopt took an owned card: %v", err)
	}
}

// #277: the owner moves its own work to another worktree on this machine. The
// relocation runs after the move (here: a failed re-stamp leaves the claimant on
// the source and warns); a repeat claim at the destination repairs it, and is
// then a no-op. Another worktree still holding the branch blocks relocation.
func TestRelocationAfterMoveAndRepair(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000371", "relocate")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	slot := filepath.Join(t.TempDir(), "slot")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000371-relocate", slot)
	t.Chdir(slot)
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err != nil {
		t.Fatalf("claim in the slot: %v\n%s", err, errs.String())
	}
	source, _ := ownerOf(t, r, cardPath)
	if source.Worktree != canonRoot(slot) {
		t.Fatalf("owner %+v", source)
	}
	t.Chdir(r.root)
	// The old worktree still holds the branch: not a relocation.
	if err := moveRelocation(context.Background(), r.root, "000371"); err == nil {
		t.Fatal("relocated while the source worktree still holds the branch")
	}
	// sdlc move's switches: the slot leaves the branch, the destination takes it.
	testfix.Git(t, slot, "switch", "-q", "--detach")
	testfix.Git(t, r.root, "switch", "-q", "000371-relocate")
	prev := moveRelocation
	moveRelocation = func(context.Context, string, string) error { return errors.New("tracker unreachable") }
	var warn bytes.Buffer
	relocateAfterMove(r.root, "000371-relocate", &warn)
	moveRelocation = prev
	if !strings.Contains(warn.String(), "sdlc claim --issue 371") {
		t.Fatalf("failed re-stamp named no repair: %s", warn.String())
	}
	if got, _ := ownerOf(t, r, cardPath); got != source {
		t.Fatal("a failed re-stamp changed the owner")
	}
	if refusal, _ := ownershipGate(t, r, "start-plan", 371, detailPath); !strings.Contains(refusal, "relocated work") {
		t.Fatalf("gate did not name the repair: %q", refusal)
	}
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err != nil {
		t.Fatalf("repair claim: %v\n%s", err, errs.String())
	}
	moved, _ := ownerOf(t, r, cardPath)
	if moved.Worktree != canonRoot(r.root) || moved.Machine != source.Machine {
		t.Fatalf("not relocated: %+v", moved)
	}
	before := r.card(cardPath)
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err != nil || r.card(cardPath) != before {
		t.Fatalf("repeat after relocation was not a no-op: %v", err)
	}
	// Another machine never relocates, even with the branch here.
	away := moved
	away.Worktree, away.Machine = canonRoot(slot), issue.MachineFingerprint("another-machine")
	withClaimant(t, away)
	t.Chdir(slot)
	testfix.Git(t, r.root, "switch", "-q", "--detach")
	testfix.Git(t, slot, "switch", "-q", "000371-relocate")
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err == nil || r.card(cardPath) != before {
		t.Fatalf("another machine relocated the owner: %v", err)
	}
}

// #277: ownership survives the claiming process: a fresh process (the built
// binary) continues from the claiming workspace, and the same path on another
// machine is refused.
func TestOwnershipSurvivesRestart(t *testing.T) {
	if _, err := machineID(); err != nil {
		t.Skipf("claims need the host machine ID, unreadable here: %v", err)
	}
	binary := buildFleetE2EBinary(t)
	cardPath, card, detailPath, detail := seededIssue(t, "000372", "restart")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	for _, verb := range []string{"claim", "start-plan"} {
		cmd := exec.Command(binary, verb, "--issue", "372")
		cmd.Dir = r.root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s from a fresh process: %v\n%s", verb, err, out)
		}
	}
	owner, _ := ownerOf(t, r, cardPath)
	away := owner
	away.Machine = issue.MachineFingerprint("another-machine")
	withClaimant(t, away)
	if refusal, _ := ownershipGate(t, r, "start-plan", 372, detailPath); !strings.Contains(refusal, "owned by") {
		t.Fatalf("same path on another machine continued: %q", refusal)
	}
}
