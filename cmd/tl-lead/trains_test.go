package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xianxu/ariadne/pkg/trains"
)

// fakeSDLC serves a fleet and per-checkout observations; a missing
// observation is an error, as a failed `sdlc issue show` would be.
type fakeSDLC struct {
	fleet     Fleet
	fleetErr  error
	obs       map[string]Observation // "<repoPath>#<n>"
	callPaths map[string]int         // per-checkout observations for scope assertions
	calls     int                    // IssueShow calls; callers observe cars concurrently (boundedEach)
	mu        sync.Mutex             // guards calls
	issues    map[string][]RepoIssue // by repo path: sdlc state's issues there
	issueErr  map[string]error
}

// RepoIssues answers from the repository's recorded issues, as `sdlc state`
// reads them from that checkout.
func (f *fakeSDLC) RepoIssues(repoRoot string) ([]RepoIssue, error) {
	if err := f.issueErr[repoRoot]; err != nil {
		return nil, err
	}
	return f.issues[repoRoot], nil
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
	f.mu.Lock()
	f.calls++
	if f.callPaths == nil {
		f.callPaths = map[string]int{}
	}
	f.callPaths[repoPath]++
	f.mu.Unlock()
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

func TestRenderSlots(t *testing.T) {
	fleet := Fleet{Slots: []FleetSlot{
		{Address: "ariadne:0", Repo: "ariadne", Slot: 0, RestingBranch: "main", HostBranch: "000999-on-zero"},
		{Address: "ariadne:1", Repo: "ariadne", Slot: 1, RestingBranch: "main-slot1", HostBranch: "main-slot1"},
		{Address: "ariadne:2", Repo: "ariadne", Slot: 2, RestingBranch: "main-slot2", HostBranch: "000350-lift-the-tl-skill"},
		{Address: "pair:2", Repo: "pair", Slot: 2, RestingBranch: "main-slot2", HostBranch: "main-slot2"},
		{Address: "pair:3", Repo: "pair", Slot: 3, RestingBranch: "main-slot3", HostBranch: "operator-experiment"},
		{Address: "pair:4", Repo: "pair", Slot: 4, RestingBranch: "main-slot4", HostBranch: ""},
	}}
	lines := slotLines(fleet)
	for i := range lines {
		lines[i].Agent = "codex"
		if lines[i].ID == 350 {
			lines[i].Phase, lines[i].Bar = "working", "▉▕ M 1/2"
		}
	}
	want := "```\n" +
		"ariadne:2      ariadne#350      working       codex        ▉▕ M 1/2  lift-the-tl-skill\n" +
		"pair:3         on operator-experiment  codex\n" +
		"pair:4         detached  codex\n" +
		"free: ariadne:1 (codex), pair:2 (codex)\n" +
		"```"
	if got := renderSlots(lines); got != want {
		t.Errorf("renderSlots:\n%s\nwant:\n%s", got, want)
	}
	if got := renderSlots(nil); !strings.Contains(got, "no numbered slots") {
		t.Errorf("empty: %q", got)
	}
}

// gatherSlots over the fakes renders a fixture fleet exactly: the phase from
// the slot's issue observation (`?` when it failed), the agent from couch's
// actors (`none` where no agent runs), the bar from the checkout's issue file,
// then the free slots with their agents.
func TestGatherSlots(t *testing.T) {
	root := t.TempDir()
	slot2 := filepath.Join(root, "ariadne-slot2", "ariadne")
	slot3 := filepath.Join(root, "ariadne-slot3", "ariadne")
	for _, p := range []string{slot2, slot3} {
		if err := os.MkdirAll(filepath.Join(p, "workshop", "issues"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(slot2, "workshop", "issues", "000350-lift.md"), []byte("## Plan\n\n- [x] one\n- [ ] two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sd := &fakeSDLC{
		fleet: Fleet{Slots: []FleetSlot{
			{Address: "ariadne:0", Repo: "ariadne", Slot: 0, RestingBranch: "main", HostBranch: "main"},
			{Address: "ariadne:1", Repo: "ariadne", Slot: 1, RestingBranch: "main-slot1", HostBranch: "main-slot1"},
			{Address: "ariadne:2", Repo: "ariadne", Slot: 2, RestingBranch: "main-slot2", HostPath: slot2, HostBranch: "000350-lift"},
			{Address: "ariadne:3", Repo: "ariadne", Slot: 3, RestingBranch: "main-slot3", HostPath: slot3, HostBranch: "000351-gone"},
			{Address: "ariadne:4", Repo: "ariadne", Slot: 4, RestingBranch: "main-slot4", HostBranch: "spike"},
		}},
		obs: map[string]Observation{slot2 + "#350": {Status: "working"}},
	}
	actors := fakeActors{actors: map[string]Actor{
		"ariadne:0": {Agent: "claude", Resting: true},
		"ariadne:1": {Agent: "codex", Resting: true},
		"ariadne:2": {Agent: "claude"},
		"ariadne:4": {Agent: "qoder", Resting: true},
	}}
	lines, err := gatherSlots(sd, actors, root)
	got := renderSlots(lines)
	if err != nil {
		t.Fatal(err)
	}
	want := "```\n" +
		"ariadne:2      ariadne#350      working       claude       ▉▕ tasks 1/2  lift\n" +
		"ariadne:3      ariadne#351      ?             none         no issue file  gone\n" +
		"ariadne:4      on spike  qoder\n" +
		"free: ariadne:1 (codex)\n" +
		"```"
	if got != want {
		t.Errorf("Slots:\n%s\nwant:\n%s", got, want)
	}

	// Couch unreadable (a sandboxed call can't reach its socket): `?`, still rendered.
	lines, err = gatherSlots(sd, fakeActors{err: fmt.Errorf("couch --actors: connect: operation not permitted")}, root)
	got = renderSlots(lines)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"working       ?            ▉▕", "free: ariadne:1 (?)"} {
		if !strings.Contains(got, w) {
			t.Errorf("degraded Slots lacks %q:\n%s", w, got)
		}
	}
}

// gatherTrains end to end over the fakes: the project's table, each car's
// observation, and the claimant checkout's bar.
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
	block, err := gatherTrains(sd, project, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "→ ariadne#202 ▉▕ M 1/2 working @ariadne:1"; !strings.Contains(block, want) {
		t.Errorf("trains block lacks %q:\n%s", want, block)
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

// `status slots` writes the rendered block through the guarded path; a failed
// fleet read writes nothing and says why.
func TestStatusSlots(t *testing.T) {
	sd := &fakeSDLC{fleet: Fleet{Slots: []FleetSlot{{Address: "ariadne:1", Repo: "ariadne", Slot: 1, RestingBranch: "main-slot1", HostBranch: "main-slot1"}}}}
	f := newFake(map[string]string{"Slots": "old"})
	file, _ := statusFile(t, "")
	if _, err := run(t, deps{ss: f, sdlc: sd}, "status", "slots", file); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.sections["Slots"], "free: ariadne:1") {
		t.Errorf("Slots = %q", f.sections["Slots"])
	}
	sd.fleetErr = fmt.Errorf("sdlc fleet inventory: timed out after 1m0s")
	_, err := run(t, deps{ss: f, sdlc: sd}, "status", "slots", file)
	if err == nil || !strings.Contains(err.Error(), "timed out") || f.writes != 1 {
		t.Errorf("fleet failure: err=%v writes=%d (want the reason and no second write)", err, f.writes)
	}
}

// A soak hold (ariadne#482) lives in the rendered Slots section: --hold sets
// it, every later render carries it, --release drops it. A held slot is never
// on the free line, so it can't be dispatched.
func TestStatusSlotsSoakHold(t *testing.T) {
	sd := &fakeSDLC{fleet: Fleet{Slots: []FleetSlot{
		{Address: "ariadne:1", Repo: "ariadne", Slot: 1, RestingBranch: "main-slot1", HostBranch: "main-slot1"},
		{Address: "ariadne:2", Repo: "ariadne", Slot: 2, RestingBranch: "main-slot2", HostBranch: "main-slot2"},
	}}}
	f := newFake(map[string]string{"Slots": "old"})
	file, _ := statusFile(t, "")
	d := deps{ss: f, sdlc: sd}
	if _, err := run(t, d, "status", "slots", file, "--hold", "ariadne:1=until the next train lands"); err != nil {
		t.Fatal(err)
	}
	got := f.sections["Slots"]
	if !strings.Contains(got, "ariadne:1") || !strings.Contains(got, "hold: soak until the next train lands") || strings.Contains(got, "free: ariadne:1") || !strings.Contains(got, "free: ariadne:2") {
		t.Fatalf("hold not rendered, or the held slot listed free:\n%s", got)
	}
	// A plain re-render (every pass) keeps the hold.
	if _, err := run(t, d, "status", "slots", file); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.sections["Slots"], "hold: soak until the next train lands") {
		t.Fatalf("a re-render dropped the hold:\n%s", f.sections["Slots"])
	}
	if _, err := run(t, d, "status", "slots", file, "--release", "ariadne:1"); err != nil {
		t.Fatal(err)
	}
	if got := f.sections["Slots"]; strings.Contains(got, "hold:") || !strings.Contains(got, "free: ariadne:1 (?), ariadne:2 (?)") {
		t.Fatalf("release left:\n%s", got)
	}
	for _, bad := range [][]string{{"--hold", "ariadne:9=until later"}, {"--hold", "ariadne:1=tomorrow"}, {"--release", "ariadne:2"}} {
		if _, err := run(t, d, append([]string{"status", "slots", file}, bad...)...); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
}
