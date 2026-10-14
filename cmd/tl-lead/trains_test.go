package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/trains"
)

// fakeSDLC serves a fleet and per-checkout observations; a missing
// observation is an error, as a failed `sdlc issue show` would be.
type fakeSDLC struct {
	fleet    Fleet
	fleetErr error
	obs   map[string]Observation // "<repoPath>#<n>"
	calls int
}

func (f *fakeSDLC) FleetInventory(string) (Fleet, error) { return f.fleet, f.fleetErr }

// RepoRoot resolves a repository to its :0 checkout, as `sdlc workspace` does.
func (f *fakeSDLC) RepoRoot(_, repo string) (string, error) {
	for _, sl := range f.fleet.Slots {
		if sl.Repo == repo && sl.Slot == 0 {
			return sl.HostPath, nil
		}
	}
	return "", fmt.Errorf("sdlc workspace %s: no such repository", repo)
}

func (f *fakeSDLC) IssueShow(repoPath string, n int) (Observation, error) {
	f.calls++
	o, ok := f.obs[fmt.Sprintf("%s#%d", repoPath, n)]
	if !ok {
		return Observation{}, fmt.Errorf("sdlc issue show %d: exit status 1: no such issue", n)
	}
	return o, nil
}

func TestDecodeFixtures(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "issue-show.json"))
	if err != nil {
		t.Fatal(err)
	}
	o, err := decodeObservation(raw)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != "working" || o.Claimant != "ariadne:2" || !strings.HasSuffix(o.ClaimantPath, "ariadne-slot2/ariadne") {
		t.Errorf("observation = %+v", o)
	}
	raw, err = os.ReadFile(filepath.Join("testdata", "fleet-inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	f, err := decodeFleet(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Slots) < 6 || f.Slots[2].Address != "ariadne:2" || !strings.HasPrefix(f.Slots[2].HostBranch, "000350-") {
		t.Errorf("fleet = %+v", f.Slots)
	}
	for _, s := range f.Slots {
		if s.Address == "pair:0" && s.HostBranch != "" {
			t.Errorf("pair:0 is detached in the fixture; got branch %q", s.HostBranch)
		}
	}
}

func TestDecodeRefuses(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"other schema", `{"schema_version":2}`, "schema_version 2"},
		{"unread card", `{"schema_version":1,"card":{"state":"unknown","error":"tracker unreachable"}}`, "card unknown: tracker unreachable"},
		{"not json", `{`, "unreadable JSON"},
	} {
		if _, err := decodeObservation([]byte(tc.raw)); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", tc.name, err, tc.want)
		}
	}
	if _, err := decodeFleet([]byte(`{"schema_version":3}`)); err == nil || !strings.Contains(err.Error(), "schema_version 3") {
		t.Errorf("fleet: err = %v", err)
	}
}

func ref(repo string, id int) trains.IssueRef { return trains.IssueRef{Repo: repo, ID: id} }

func TestRenderTrainsGolden(t *testing.T) {
	plan := trains.TrainPlan{Trains: []trains.Train{
		{Name: "Review quality", Slot: "ariadne:2 (from #202)", Cars: []trains.IssueRef{ref("ariadne", 189), ref("ariadne", 202), ref("ariadne", 204)}},
		{Name: "Merge and CI", Slot: "ariadne:4, after", Cars: []trains.IssueRef{ref("pair", 324), ref("ariadne", 309), ref("ariadne", 228)}},
		{Name: "A very long train name over the limit", Slot: "—", Cars: []trains.IssueRef{ref("ariadne", 1)}},
	}}
	cars := map[trains.IssueRef]CarState{
		ref("ariadne", 189): {Status: "done"},
		ref("ariadne", 202): {Status: "working", Claimant: "ariadne:2", Bar: "▉▕ M 1/2"},
		ref("ariadne", 204): {Status: "open"},
		ref("pair", 324):    {Status: "codecomplete", Claimant: "pair:3"},
		ref("ariadne", 309): {Err: "sdlc issue show 309: timed out"},
		ref("ariadne", 228): {Status: "open", Claimant: "ariadne:4"},
		ref("ariadne", 1):   {Status: "open"},
	}
	got := renderTrains(plan, cars)
	want := "```\n" +
		"Review quality        ariadne:2  ✓◐○          ariadne#189 ✓ → ariadne#202 ▉▕ M 1/2 working @ariadne:2 ariadne#204\n" +
		"Merge and CI          ariadne:4  ◑?○          pair#324 ◑ ariadne#309 ? ariadne#228\n" +
		"A very long train nam —          ○            ariadne#1\n" +
		"```"
	if got != want {
		t.Errorf("renderTrains:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderOffTrain(t *testing.T) {
	fleet := Fleet{Slots: []FleetSlot{
		{Address: "ariadne:0", Repo: "ariadne", Slot: 0, RestingBranch: "main", HostBranch: "000999-on-zero"},
		{Address: "ariadne:1", Repo: "ariadne", Slot: 1, RestingBranch: "main-slot1", HostBranch: "000202-on-a-train"},
		{Address: "ariadne:2", Repo: "ariadne", Slot: 2, RestingBranch: "main-slot2", HostBranch: "000350-lift-the-tl-skill"},
		{Address: "pair:2", Repo: "pair", Slot: 2, RestingBranch: "main-slot2", HostBranch: "main-slot2"},
		{Address: "pair:3", Repo: "pair", Slot: 3, RestingBranch: "main-slot3", HostBranch: "operator-experiment"},
	}}
	listed := map[trains.IssueRef]bool{ref("ariadne", 202): true}
	got := renderOffTrain(offTrainSlots(fleet, listed, map[string]string{"ariadne:2": "▉▕ M 1/2"}))
	want := "```\n" +
		"ariadne:2      ariadne#350      ▉▕ M 1/2  lift-the-tl-skill\n" +
		"pair:2         idle\n" +
		"```"
	if got != want {
		t.Errorf("renderOffTrain:\n%s\nwant:\n%s", got, want)
	}
	if got := renderOffTrain(nil); got != "_None._" {
		t.Errorf("empty: %q", got)
	}
}

// gatherTrains end to end over the fakes: the project's table, each car's
// observation, the claimant checkout's bar, and the off-train slot.
func TestGatherTrains(t *testing.T) {
	root := t.TempDir()
	zero := filepath.Join(root, "ariadne")
	slot2 := filepath.Join(root, "worktree", "ariadne-slot2", "ariadne")
	issues := filepath.Join(slot2, "workshop", "issues")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(p, s string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(issues, "000202-judge.md"), "---\nid: 000202\n---\n## Plan\n\n- [x] M1 — a\n- [ ] M2 — b\n")
	write(filepath.Join(issues, "000350-lift.md"), "## Plan\n\n- [x] one\n- [ ] two\n")
	project := filepath.Join(root, "project.md")
	write(project, "# p\n\n## Trains\n\n| Train | Cars, in order | Slot |\n|---|---|---|\n| Review | ariadne#189 → #202 → #204 | ariadne:2 |\n\n## Log\n")

	sd := &fakeSDLC{
		fleet: Fleet{Slots: []FleetSlot{
			{Address: "ariadne:0", Repo: "ariadne", Slot: 0, RestingBranch: "main", HostPath: zero, HostBranch: "main"},
			{Address: "ariadne:2", Repo: "ariadne", Slot: 2, RestingBranch: "main-slot2", HostPath: slot2, HostBranch: "000350-lift"},
		}},
		obs: map[string]Observation{
			zero + "#189": {Status: "done"},
			zero + "#202": {Status: "working", Claimant: "ariadne:1", ClaimantPath: slot2},
			zero + "#204": {Status: "open"},
		},
	}
	r, err := gatherTrains(sd, project, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "→ ariadne#202 ▉▕ M 1/2 working @ariadne:1"; !strings.Contains(r.Trains, want) {
		t.Errorf("Trains lacks %q:\n%s", want, r.Trains)
	}
	if want := "ariadne:2      ariadne#350      ▉▕ tasks 1/2  lift"; !strings.Contains(r.Slots, want) {
		t.Errorf("Slots lacks %q:\n%s", want, r.Slots)
	}
	if sd.calls != 3 {
		t.Errorf("observed %d cars, want 3 (each car once)", sd.calls)
	}
}

func TestGatherTrainsRefusesNoTable(t *testing.T) {
	p := filepath.Join(t.TempDir(), "project.md")
	if err := os.WriteFile(p, []byte("# p\n\n## Log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := gatherTrains(&fakeSDLC{}, p, ""); err == nil || !strings.Contains(err.Error(), "no `## Trains` section") {
		t.Errorf("err = %v", err)
	}
}

// `status trains` writes both rendered blocks through the guarded path.
func TestStatusTrainsWritesBoth(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project.md")
	if err := os.WriteFile(project, []byte("## Trains\n\n| Train | Cars, in order | Slot |\n|---|---|---|\n| T | ariadne#1 | ariadne:1 |\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sd := &fakeSDLC{
		fleet: Fleet{Slots: []FleetSlot{{Address: "ariadne:0", Repo: "ariadne", HostPath: root, HostBranch: "main", RestingBranch: "main"}}},
		obs:   map[string]Observation{root + "#1": {Status: "done"}},
	}
	f := newFake(map[string]string{"Trains": "old", "Slots (off-train)": "old"})
	file, _ := statusFile(t, "")
	if _, err := run(t, deps{ss: f, sdlc: sd}, "status", "trains", file, project); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.sections["Trains"], "ariadne#1 ✓") || f.sections["Slots (off-train)"] != "_None._" {
		t.Errorf("sections = %q", f.sections)
	}
}

// A failed fleet inventory leaves the Trains block intact and marks only the
// off-train block unread.
func TestGatherTrainsFleetFailureDegrades(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project.md")
	if err := os.WriteFile(project, []byte("## Trains\n\n| Train | Cars, in order | Slot |\n|---|---|---|\n| T | ariadne#1 | ariadne:1 |\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sd := &fakeSDLC{
		fleet:    Fleet{Slots: []FleetSlot{{Address: "ariadne:0", Repo: "ariadne", HostPath: root}}},
		fleetErr: fmt.Errorf("sdlc fleet inventory: timed out after 1m0s"),
		obs:      map[string]Observation{root + "#1": {Status: "done"}},
	}
	r, err := gatherTrains(sd, project, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Trains, "ariadne#1 ✓") {
		t.Errorf("Trains lost on a fleet failure:\n%s", r.Trains)
	}
	if !strings.Contains(r.Slots, "? off-train slots unread: sdlc fleet inventory: timed out") {
		t.Errorf("Slots = %q", r.Slots)
	}
}
