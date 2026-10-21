package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scopeFixture(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for name, claims := range map[string]string{"ops": "own\ntied\n", "other": "foreign\ntied\n"} {
		root := filepath.Join(ws, name)
		gitInit(t, root)
		if err := os.WriteFile(filepath.Join(root, ".hub"), []byte(claims), 0600); err != nil {
			t.Fatal(err)
		}
	}
	nested := filepath.Join(ws, "ops", "teams")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}
	return nested
}

func TestGatherSlotsScopeBeforeIssueIO(t *testing.T) {
	for _, n := range []int{3, 6} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			vantage := scopeFixture(t)
			sd := &fakeSDLC{}
			for _, repo := range []string{"own", "foreign", "tied", "unclaimed"} {
				for i := 1; i <= n; i++ {
					sd.fleet.Slots = append(sd.fleet.Slots, FleetSlot{Repo: repo, Slot: i, Address: fmt.Sprintf("%s:%d", repo, i), HostPath: filepath.Join(vantage, fmt.Sprintf("%s-slot%d", repo, i)), HostBranch: "000001-task"})
				}
			}
			lines, err := gatherSlots(sd, fakeActors{}, vantage)
			if err != nil {
				t.Fatal(err)
			}
			if len(lines) != n || sd.calls != n {
				t.Fatalf("lines=%d issue reads=%d; want %d selected lines and reads", len(lines), sd.calls, n)
			}
			for repo, reads := range sd.callPaths {
				if !strings.HasPrefix(repo, filepath.Join(vantage, "own-slot")) || reads != 1 {
					t.Fatalf("excluded or duplicate observations: %v", sd.callPaths)
				}
			}
			for _, l := range lines {
				if l.Slot.Repo != "own" {
					t.Fatalf("excluded row: %+v", l)
				}
			}
		})
	}
}

func TestStatusSlotsScopeAndHiddenHold(t *testing.T) {
	vantage := scopeFixture(t)
	sd := &fakeSDLC{fleet: Fleet{Slots: []FleetSlot{
		{Repo: "own", Slot: 1, Address: "own:1", HostBranch: "main-slot1", RestingBranch: "main-slot1"},
		{Repo: "foreign", Slot: 1, Address: "foreign:1", HostBranch: "main-slot1", RestingBranch: "main-slot1"},
		{Repo: "tied", Slot: 1, Address: "tied:1", HostBranch: "main-slot1", RestingBranch: "main-slot1"},
		{Repo: "unclaimed", Slot: 1, Address: "unclaimed:1", HostBranch: "main-slot1", RestingBranch: "main-slot1"},
	}}}
	f := newFake(map[string]string{"Slots": "foreign:1 agent hold: soak until tomorrow"})
	file := filepath.Join(vantage, "status.md")
	d := deps{ss: f, sdlc: sd}
	_, err := run(t, d, "status", "slots", file)
	if err == nil || !strings.Contains(err.Error(), "--all") || f.writes != 0 {
		t.Fatalf("hidden hold err=%v writes=%d", err, f.writes)
	}
	if _, err = run(t, d, "status", "slots", file, "--all"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.sections["Slots"], "hold: soak until tomorrow") {
		t.Fatal("--all lost hold")
	}
	if _, err = run(t, d, "status", "slots", file, "--release", "foreign:1"); err != nil {
		t.Fatal(err)
	}
	body := f.sections["Slots"]
	for _, want := range []string{"free: own:1", "1 unclaimed", "1 foreign", "1 tied", "tied:1", "ownership"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "free: foreign") || strings.Contains(body, "free: tied") {
		t.Fatal(body)
	}
}

func TestGatherSlotsMalformedScopeRefuses(t *testing.T) {
	for _, name := range []string{"ops", "other"} {
		t.Run(name, func(t *testing.T) {
			vantage := scopeFixture(t)
			root := filepath.Join(filepath.Dir(filepath.Dir(vantage)), name)
			if err := os.WriteFile(filepath.Join(root, ".hub"), []byte("../invalid"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := gatherSlots(&fakeSDLC{}, fakeActors{}, vantage); err == nil {
				t.Fatal("malformed hub silently widened scope")
			}
		})
	}
}

func TestHubScopeOwnershipAndClaims(t *testing.T) {
	scope := hubScope{enabled: true, self: "ops", hubs: []hub{
		{Name: "ops", Claims: []hubClaim{{Repo: "shared", Slot: -1}, {Repo: "exact", Slot: 2}, {Repo: "tied", Slot: 1}}},
		{Name: "other", Claims: []hubClaim{{Repo: "shared", Slot: 2}, {Repo: "tied", Slot: 1}}},
	}}
	for _, tc := range []struct {
		repo string
		slot int
		own  bool
	}{
		{"shared", 1, true}, {"shared", 2, false}, {"exact", 2, true}, {"exact", 3, false}, {"tied", 1, false}, {"absent", 1, false}, {"ops", 0, true},
	} {
		slot := FleetSlot{Repo: tc.repo, Slot: tc.slot}
		if got := scope.owns(slot); got != tc.own {
			t.Errorf("owns(%+v)=%v want%v", slot, got, tc.own)
		}
	}
	if got := strings.Join(scope.repoNames(), ","); got != "exact,shared,tied" {
		t.Fatalf("repository claims=%q", got)
	}
	if !(hubScope{}).owns(FleetSlot{Repo: "anything", Slot: 1}) {
		t.Fatal("zero scope lost legacy mode")
	}
}

func TestHubScopeAbsentEmptyAndUnreadable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ops")
	gitInit(t, root)
	scope, err := readHubScope(root)
	if err != nil || scope.enabled {
		t.Fatalf("absent hub: %+v %v", scope, err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hub"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	scope, err = readHubScope(root)
	if err != nil || !scope.enabled || scope.owns(FleetSlot{Repo: "any", Slot: 1}) {
		t.Fatalf("empty hub: %+v %v", scope, err)
	}
	if err := os.Remove(filepath.Join(root, ".hub")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".hub"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := readHubScope(root); err == nil {
		t.Fatal("unreadable local .hub silently widened scope")
	}
}

func TestStatusSlotsActorFailureReportedOnce(t *testing.T) {
	vantage := scopeFixture(t)
	sd := &fakeSDLC{}
	for i := 1; i <= 3; i++ {
		sd.fleet.Slots = append(sd.fleet.Slots, FleetSlot{Repo: "own", Slot: i, Address: fmt.Sprintf("own:%d", i), HostBranch: "000001-task"})
	}
	f := newFake(map[string]string{"Slots": "old"})
	_, err := run(t, deps{ss: f, sdlc: sd, actors: fakeActors{err: fmt.Errorf("socket denied")}}, "status", "slots", filepath.Join(vantage, "status.md"))
	if err != nil {
		t.Fatal(err)
	}
	body := f.sections["Slots"]
	if strings.Count(body, "socket denied") != 1 || strings.Count(body, "?") != 6 {
		t.Fatalf("expected one diagnostic and unknown phase+actor per row: %s", body)
	}
}

func TestHubScopeDanglingPeerClaimRefuses(t *testing.T) {
	vantage := scopeFixture(t)
	peer := filepath.Join(filepath.Dir(filepath.Dir(vantage)), "other", ".hub")
	if err := os.Remove(peer); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-claims", peer); err != nil {
		t.Fatal(err)
	}
	if _, err := readHubScope(vantage); err == nil {
		t.Fatal("unreadable peer claim was treated as absent")
	}
}

func TestHubScopeGitFailureDoesNotIgnoreLocalClaim(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".hub"), []byte("own\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readHubScope(root); err == nil {
		t.Fatal("git failure silently widened local scope")
	}
}

func TestStatusSlotsHoldGuardRechecksAfterStale(t *testing.T) {
	vantage := scopeFixture(t)
	sd := &fakeSDLC{fleet: Fleet{Slots: []FleetSlot{{Repo: "own", Slot: 1, Address: "own:1", HostBranch: "main-slot1", RestingBranch: "main-slot1"}}}}
	f := newFake(map[string]string{"Slots": "old"})
	hold := "foreign:1 agent hold: soak until tomorrow"
	f.editBefore = map[int]map[string]string{1: {"Slots": hold}}
	_, err := run(t, deps{ss: f, sdlc: sd}, "status", "slots", filepath.Join(vantage, "status.md"))
	if err == nil || !strings.Contains(err.Error(), "foreign:1") || f.writes != 0 || f.writeCalls != 1 || f.sections["Slots"] != hold {
		t.Fatalf("latest hold lost: err=%v calls=%d writes=%d body=%q", err, f.writeCalls, f.writes, f.sections["Slots"])
	}
}
