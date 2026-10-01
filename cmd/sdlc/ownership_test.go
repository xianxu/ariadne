package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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

// #277: the owner moves its own work to another worktree on this machine.
// Without sdlc move's record, a branch merely missing from the owner's worktree
// is no evidence (BR-9: that is also the state right after a claim) — a claim
// elsewhere is refused. With the record, a failed re-stamp warns and keeps it,
// the gate names the repair, a repeat claim at the destination relocates, and
// is then a no-op. Another machine never relocates.
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
	// The slot leaves the branch and the destination takes it — by hand, with no
	// move record: a takeover attempt, refused (BR-9).
	testfix.Git(t, slot, "switch", "-q", "--detach")
	testfix.Git(t, r.root, "switch", "-q", "000371-relocate")
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err == nil || !strings.Contains(err.Error(), "claimed by") {
		t.Fatalf("a claim without a move record took the card: %v", err)
	}
	if refusal, _ := ownershipGate(t, r, "start-plan", 371, detailPath); !strings.Contains(refusal, "owned by") || strings.Contains(refusal, "moved here") {
		t.Fatalf("gate offered a repair without a move record: %q", refusal)
	}
	if got, _ := ownerOf(t, r, cardPath); got != source {
		t.Fatal("the owner changed without a move record")
	}
	// sdlc move records the relocation before switching.
	if err := writeRelocation(slot, "000371", issue.Relocation{From: canonRoot(slot), To: canonRoot(r.root)}); err != nil {
		t.Fatal(err)
	}
	prev := moveRelocation
	moveRelocation = func(context.Context, string, string) error { return errors.New("tracker unreachable") }
	var warn bytes.Buffer
	relocateAfterMove(context.Background(), r.root, "000371-relocate", &warn)
	moveRelocation = prev
	if !strings.Contains(warn.String(), "sdlc claim --issue 371") {
		t.Fatalf("failed re-stamp named no repair: %s", warn.String())
	}
	if got, _ := ownerOf(t, r, cardPath); got != source {
		t.Fatal("a failed re-stamp changed the owner")
	}
	if rec, err := readRelocation(r.root, "000371"); err != nil || rec == nil {
		t.Fatalf("a failed re-stamp dropped the move record: %v", err)
	}
	if refusal, _ := ownershipGate(t, r, "start-plan", 371, detailPath); !strings.Contains(refusal, "moved here by `sdlc move`") {
		t.Fatalf("gate did not name the repair: %q", refusal)
	}
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(371)); err != nil {
		t.Fatalf("repair claim: %v\n%s", err, errs.String())
	}
	moved, _ := ownerOf(t, r, cardPath)
	if moved.Worktree != canonRoot(r.root) || moved.Machine != source.Machine {
		t.Fatalf("not relocated: %+v", moved)
	}
	if rec, _ := readRelocation(r.root, "000371"); rec != nil {
		t.Fatal("the repair left the move record behind")
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
	if err := writeRelocation(slot, "000371", issue.Relocation{From: moved.Worktree, To: canonRoot(slot)}); err != nil {
		t.Fatal(err)
	}
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

// moveTrackerFixture is the slot layout (primary :0, slots :1/:2 on remote
// upstream) as a tracker repository, with #1's card working and owned by
// :1, which holds its issue branch one commit ahead. Identities share a machine
// and differ by worktree, like slots on one machine.
func moveTrackerFixture(t *testing.T) (roots []string, cardPath string) {
	t.Helper()
	roots, _ = procedureFixture(t)
	primary := roots[0]
	cardPath, card, _, _ := seededIssue(t, "000001", "procedure")
	tf, err := gitx.NewTrunkFileContext(context.Background(), primary, "upstream", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	base := issue.Claimant{Operator: "Op", Machine: issue.MachineFingerprint("this-machine"), MachineName: "box"}
	claimantIdentity = func(env *trackerEnv) (issue.Claimant, error) {
		c := base
		c.Worktree, c.Repository = canonRoot(env.root), env.target.Repository
		return c, nil
	}
	t.Cleanup(func() { claimantIdentity = resolveClaimantIdentity })
	if _, err := tf.Bootstrap(map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes(), cardPath: []byte(card)}, "bootstrap", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	markCutoverOn(t, primary, "upstream", "main")
	slot := roots[1]
	testfix.Git(t, slot, "fetch", "-q", "upstream")
	testfix.Git(t, slot, "merge", "-q", "--ff-only", "upstream/main")
	t.Chdir(slot)
	env, err := openTrackerAt(context.Background(), slot)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	rec, _ := snap.Card("000001")
	owner, _ := claimantIdentity(env)
	next, err := issue.SetCardClaimant([]byte(strings.Replace(string(rec.Raw), "status: open", "status: working\nstarted: 2026-10-01T09:00:00-07:00", 1)), owner)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.UpdateCard(rec, next, operationToken("set"), func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	testfix.Git(t, slot, "switch", "-q", "-c", "000001-procedure")
	procedureWrite(t, slot, "feature", "selected branch\n")
	testfix.Git(t, slot, "add", "feature")
	testfix.Git(t, slot, "commit", "-qm", "feature change")
	return roots, cardPath
}

func cardOwner(t *testing.T, root, cardPath string) (issue.Claimant, bool) {
	t.Helper()
	tf, err := gitx.NewTrunkFileContext(context.Background(), root, "upstream", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tf.Read(cardPath)
	if err != nil {
		t.Fatal(err)
	}
	c, ok, err := issue.CardClaimant(raw)
	if err != nil {
		t.Fatal(err)
	}
	return c, ok
}

// #277 (BR-11): a real `sdlc move` relocates its owner after the switches and
// retires its record; an unattributed issue moves with its owner unknown, a
// warning toward --adopt, and no record left behind.
func TestMoveRelocatesItsOwner(t *testing.T) {
	roots, cardPath := moveTrackerFixture(t)
	out, err := runMoveTest(t, roots[1], ":0", false)
	if err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	if owner, _ := cardOwner(t, roots[0], cardPath); owner.Worktree != canonRoot(roots[0]) {
		t.Fatalf("owner not relocated to :0: %+v\n%s", owner, out)
	}
	if rec, err := readRelocation(roots[0], "000001"); err != nil || rec != nil {
		t.Fatalf("move left its record: %+v %v", rec, err)
	}
}

func TestMoveLeavesAnUnattributedIssueUnknown(t *testing.T) {
	roots, cardPath := moveTrackerFixture(t)
	env, err := openTrackerAt(context.Background(), roots[1])
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000001", cardPath, "legacy", operationToken("set"), func(c []byte) ([]byte, error) {
		fm, body, _ := issue.Parse(string(c))
		i := strings.Index(fm, "\nclaimant:")
		return []byte(issue.Compose(fm[:i], body)), nil
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	out, err := runMoveTest(t, roots[1], ":0", false)
	if err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	if _, ok := cardOwner(t, roots[0], cardPath); ok || !strings.Contains(out, "--adopt") {
		t.Fatalf("unattributed move: owner recorded=%v\n%s", ok, out)
	}
	if rec, _ := readRelocation(roots[0], "000001"); rec != nil {
		t.Fatal("a move whose relocation does not apply left its record")
	}
}

// #277: two clones adopting the same unattributed working card race through the
// card's compare-and-swap: exactly one records its worktree, the other is
// refused and publishes nothing.
func TestAdoptRaceHasExactlyOneWinner(t *testing.T) {
	if _, err := machineID(); err != nil {
		t.Skipf("adopt needs the host machine ID, unreadable here: %v", err)
	}
	binary := buildFleetE2EBinary(t)
	cardPath, card, detailPath, detail := seededIssue(t, "000373", "adoptrace")
	working := strings.Replace(card, "status: open", "status: working\nstarted: 2026-09-01T09:00:00-07:00", 1)
	r := newTrackerRepo(t, map[string]string{cardPath: working}, map[string]string{detailPath: detail})
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "Peer")
	git(t, peer, "config", "user.email", "peer@example.com")
	results := raceBuiltBinary(t, binary, []string{r.root, peer}, "claim", "--issue", "373", "--adopt")
	wins, winner := 0, ""
	for _, res := range results {
		if res.err == nil {
			wins++
			winner = res.dir
		} else if !strings.Contains(res.out, "changed while adopting") && !strings.Contains(res.out, "never reassigns") {
			t.Errorf("loser was not a CAS/owner refusal: %v %s", res.err, res.out)
		}
	}
	if wins != 1 {
		t.Fatalf("adopt winners=%d; want exactly one", wins)
	}
	if owner, ok := ownerOf(t, r, cardPath); !ok || owner.Worktree != canonRoot(winner) {
		t.Fatalf("owner %+v, want the winner %s", owner, winner)
	}
}
