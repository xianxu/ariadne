package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// fleetClaims collects the inventory of the fleet containing root, counting
// claims reads, and round-trips it through the strict JSON contract.
func fleetClaims(t *testing.T, root string) (fleet.Inventory, map[string]int) {
	t.Helper()
	reads := map[string]int{}
	inv, err := fleet.CollectInventory(context.Background(), filepath.Dir(root), fleet.InventoryOptions{
		Git:     execGitRunner{},
		Machine: fleetMachine,
		LookupClaims: func(repoRoot string) fleet.RepoClaims {
			reads[repoRoot]++
			return fleet.LookupRepoClaims(context.Background(), repoRoot)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	var back fleet.Inventory
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("inventory is not valid contract JSON: %v\n%s", err, raw)
	}
	return back, reads
}

func rowAt(t *testing.T, inv fleet.Inventory, path string) fleet.TreeRow {
	t.Helper()
	for _, r := range inv.Rows {
		if r.TreePath == canonRoot(path) {
			return r
		}
	}
	t.Fatalf("no row for %s in %+v", path, inv.Rows)
	return fleet.TreeRow{}
}

// #288: a claim made in a slot is reported on that slot's row, against this
// machine's identity as claim recorded it, from one claims read per
// repository; another machine's claim, an ownerless working card and an
// unclaimed open card are not local state (an owned open card is, #283); once the slot is removed the claim is dangling,
// never dropped.
func TestFleetInventoryPlacesClaims(t *testing.T) {
	cards, details, paths := map[string]string{}, map[string]string{}, map[int]string{}
	for _, n := range []int{471, 473, 474, 475} {
		cp, c, dp, d := seededIssue(t, itoa6(n), "c"+itoa(n))
		cards[cp], details[dp], paths[n] = c, d, cp
	}
	r := newTrackerRepo(t, cards, details)
	claim := func(n int) {
		t.Helper()
		var out, errs bytes.Buffer
		if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(n)); err != nil {
			t.Fatalf("claim #%d: %v\n%s", n, err, errs.String())
		}
		invalidateIssueRecords(context.Background())
	}
	slot := filepath.Join(t.TempDir(), "slot")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", "000471-c471", slot)
	t.Chdir(slot)
	claim(471)
	t.Chdir(r.root)
	claim(473)
	claim(474)
	other, _ := ownerOf(t, r, paths[473])
	other.Machine = issue.MachineFingerprint("another-machine")
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000473", paths[473], "owner", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardClaimant(c, other)
	}); err != nil {
		t.Fatal(err)
	}
	dropClaimant(t, r, "000474", paths[474])
	cardPath := paths[471]
	third := filepath.Join(t.TempDir(), "third") // a second worktree: reads stay one per repository
	testfix.Git(t, r.root, "worktree", "add", "-q", "--detach", third)
	owner, _ := ownerOf(t, r, cardPath)

	inv, reads := fleetClaims(t, r.root)
	if inv.Machine.State != fleet.MachinePresent || inv.Machine.Fingerprint != owner.Machine {
		t.Fatalf("machine %+v is not the identity claim recorded (%s)", inv.Machine, owner.Machine)
	}
	if len(reads) != 1 || reads[canonRoot(r.root)] != 1 {
		t.Fatalf("claims reads %v, want one for the one repository", reads)
	}
	s := rowAt(t, inv, slot)
	if s.ClaimsState != fleet.ClaimsPresent || len(s.Claims) != 1 || s.Claims[0].Ref == "" || s.Claims[0].Claimant.Worktree != canonRoot(slot) || s.Claims[0].Revision == "" {
		t.Fatalf("slot row: %+v", s)
	}
	if p := rowAt(t, inv, r.root); p.ClaimsState != fleet.ClaimsPresent || len(p.Claims) != 0 {
		t.Fatalf("primary row: %+v", p)
	}
	if len(inv.DanglingClaims) != 0 {
		t.Fatalf("dangling: %+v", inv.DanglingClaims)
	}

	testfix.Git(t, r.root, "worktree", "remove", "--force", slot)
	inv, _ = fleetClaims(t, r.root)
	if len(inv.DanglingClaims) != 1 || inv.DanglingClaims[0].Claimant.Worktree != canonRoot(slot) || inv.DanglingClaims[0].RepoRoot != canonRoot(r.root) {
		t.Fatalf("a removed slot's claim must be dangling: %+v", inv.DanglingClaims)
	}
}

// #288: with the tracker unreachable the claims answer from the last fetch,
// marked stale with the reason — the claim is still listed, never "none".
func TestFleetInventoryStaleClaimsSaySo(t *testing.T) {
	r, slot, _, _ := reclaimFixture(t, 472)
	if err := os.Rename(r.origin, r.origin+".gone"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(r.origin+".gone", r.origin) })
	inv, _ := fleetClaims(t, r.root)
	s := rowAt(t, inv, slot)
	if s.ClaimsState != fleet.ClaimsStale || s.ClaimsError == "" || len(s.Claims) != 1 {
		t.Fatalf("stale slot row: %+v", s)
	}
}

// #288 BR-11: a checkout that does not use the tracker (no cutover marker, no
// fetched tracker) is absent without contacting its remote — here one that
// does not exist, which a probe would report as unknown.
func TestFleetClaimsSkipUntrackedRemotes(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, root, "remote", "add", "origin", filepath.Join(t.TempDir(), "never-created.git"))
	testfix.Git(t, root, "config", "branch.main.remote", "origin") // a real clone's publication target
	testfix.Git(t, root, "config", "branch.main.merge", "refs/heads/main")
	if got := fleet.LookupRepoClaims(context.Background(), root); got.State != fleet.ClaimsAbsent || got.Error != "" {
		t.Fatalf("untracked checkout: %+v, want absent with no remote probe", got)
	}
}
