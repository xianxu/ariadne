package main

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeCouch is one Couch behind both seams fleet resume uses: --actors
// (who is live) and the resume verbs. Resuming a slot makes it live, so a
// rerun sees it running, as the real Couch would show it.
type fakeCouch struct {
	live       map[string]bool
	actorsErr  error
	candidates map[string]ResumeCandidates
	candErr    map[string]error
	resumeErr  map[string]error
	resumed    []string
	candReads  int
}

func (f *fakeCouch) Actors() (map[string]Actor, error) {
	if f.actorsErr != nil {
		return nil, f.actorsErr
	}
	out := map[string]Actor{}
	for a := range maps.Keys(f.live) {
		out[a] = Actor{Agent: "claude"}
	}
	return out, nil
}

func (f *fakeCouch) Candidates(addr string) (ResumeCandidates, error) {
	f.candReads++
	return f.candidates[addr], f.candErr[addr]
}

func (f *fakeCouch) Resume(addr string) error {
	if err := f.resumeErr[addr]; err != nil {
		return err
	}
	if f.live[addr] {
		return fmt.Errorf("refused: %s is already running", addr)
	}
	f.live[addr] = true
	f.resumed = append(f.resumed, addr)
	return nil
}

// fleetFixture: ops (this hub) claims pair and ariadne; ops-echo claims echo
// and the exact slot pair:7; brain is nobody's.
func fleetFixture(t *testing.T) (deps, *fakeCouch, string) {
	t.Helper()
	ws := t.TempDir()
	for dir, claims := range map[string]string{"ops": "pair\nariadne\ntools\n", "ops-echo": "echo\npair:7\ntools\n"} {
		gitInit(t, filepath.Join(ws, dir))
		if err := os.WriteFile(filepath.Join(ws, dir, ".hub"), []byte(claims), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var slots []FleetSlot
	for _, a := range []string{"ops:0", "ops-echo:0", "pair:0", "pair:1", "pair:2", "pair:3", "pair:4", "pair:5", "pair:6", "pair:7", "pair:8", "ariadne:0", "ariadne:1", "brain:1", "echo:1", "tools:1"} {
		at := strings.LastIndexByte(a, ':')
		repo, n := a[:at], 0
		fmt.Sscanf(a[at+1:], "%d", &n)
		slots = append(slots, FleetSlot{Address: a, Repo: repo, Slot: n, HostPath: "/w/" + repo})
	}
	// Claims as the inventory places them on this machine's checkouts.
	claims := map[string][]string{"pair:2": {"pair#702"}, "pair:3": {"pair#703"}, "pair:5": {"pair#705"}, "pair:6": {"pair#706"}, "pair:8": {"pair#708"}}
	for i := range slots {
		slots[i].Claims, slots[i].ClaimsKnown = claims[slots[i].Address], true
		if slots[i].Address == "ariadne:1" {
			slots[i].ClaimsKnown, slots[i].ClaimsWhy = false, "claims unknown"
		}
	}
	s := &fakeSDLC{fleet: Fleet{Slots: slots}}
	c := &fakeCouch{
		live: map[string]bool{"ops:0": true, "pair:0": true, "pair:1": true, "ariadne:0": true},
		candidates: map[string]ResumeCandidates{
			"pair:2": {Proven: true}, "pair:6": {Proven: true}, "pair:8": {Proven: true},
			"pair:3": {Incomplete: true},
		},
		candErr:   map[string]error{"pair:5": errors.New("couch busy")},
		resumeErr: map[string]error{"pair:6": errors.New("refused: wrapper is stale"), "pair:8": fmt.Errorf("%w: no answer", errUncertain)},
	}
	return deps{sdlc: s, actors: c, resume: c}, c, filepath.Join(ws, "ops")
}

func TestFleetResumeDryRun(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	var out bytes.Buffer
	if err := fleetResume(&out, d, cwd, false); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		invisibleLayers,
		"resume    pair:2", "resume    pair:6", "resume    pair:8",
		"ambiguous pair:3  pair#703; no proven conversation",
		"ambiguous brain:1  unclaimed",
		"ambiguous tools:1  claimed by ops and ops-echo", // a tie is never this hub's to resume
		"unknown   pair:5  pair#705; candidates unreadable: couch busy",
		"unknown   ariadne:1  claims unreadable: claims unknown", // a failed claims read is never idle
		"running   pair:1", "idle      pair:4",
		"(3 slot(s) belong to other hubs)", // ops-echo:0, pair:7 (exact claim), echo:1
		"dry run",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	for _, never := range []string{"ops:0", "pair:7", "echo:1"} {
		if strings.Contains(got, never) {
			t.Errorf("listed %s, which is this hub or another's:\n%s", never, got)
		}
	}
	if len(c.resumed) != 0 {
		t.Fatalf("a dry run resumed %v", c.resumed)
	}
	if c.candReads != 5 { // only down slots with work: pair:2,3,5,6,8
		t.Errorf("candidates read %d times, want 5", c.candReads)
	}
}

func TestFleetResumeApplyThenRerun(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	var out bytes.Buffer
	if err := fleetResume(&out, d, cwd, true); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"resuming pair:2, pair:6, pair:8",
		"resumed   pair:2",
		"refused   pair:6: refused: wrapper is stale", // a refusal never stops the rest
		"uncertain pair:8",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	if strings.Join(c.resumed, ",") != "pair:2" {
		t.Fatalf("resumed %v", c.resumed)
	}
	// A rerun re-reads liveness: pair:2 is running now and never resumed twice.
	out.Reset()
	if err := fleetResume(&out, d, cwd, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "running   pair:2") || strings.Join(c.resumed, ",") != "pair:2" {
		t.Fatalf("rerun resumed again: %v\n%s", c.resumed, out.String())
	}
}

// A failed --actors read is unknown for every slot: nothing is down, nothing
// is resumed.
func TestFleetResumeActorsUnreadable(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	c.actorsErr = errors.New("socket denied")
	var out bytes.Buffer
	if err := fleetResume(&out, d, cwd, true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "resume    ") || len(c.resumed) != 0 || !strings.Contains(out.String(), "unknown   pair:2  couch --actors unreadable") {
		t.Fatalf("resumed on an unknown read:\n%s", out.String())
	}
}

// Run from a checkout without .hub, fleet resume refuses: only a hub acts.
func TestFleetResumeNeedsAHub(t *testing.T) {
	d, _, _ := fleetFixture(t)
	plain := filepath.Join(t.TempDir(), "pair")
	gitInit(t, plain)
	if err := fleetResume(&bytes.Buffer{}, d, plain, false); err == nil || !strings.Contains(err.Error(), "no .hub") {
		t.Fatalf("err %v", err)
	}
}

func TestDecodeCandidates(t *testing.T) {
	// Trimmed from a live `couch --resume-candidates parley.nvim:4 --json`, 2026-10-19.
	raw := `{"address":{"repo_scope":"3d","tag":"1-parley-nvim-62"},"revision":5,"candidates":[],"proven":{"key":"","label":"Resume from disk (confirmed)","agent":"claude","native_id":"94eb0fa3"},"fresh":{"label":"Start fresh"},"incomplete":true,"reasons":["native metadata listing incomplete"],"automatic":false}`
	c, err := decodeCandidates([]byte(raw))
	if err != nil || !c.Proven || !c.Incomplete || len(c.Reasons) != 1 {
		t.Fatalf("%+v %v", c, err)
	}
	c, err = decodeCandidates([]byte(`{"proven":{"native_id":""},"candidates":[{"key":"k"}]}`))
	if err != nil || c.Proven {
		t.Fatalf("an empty native id is not proven: %+v %v", c, err)
	}
}

// Live: the real Couch's --resume-candidates reply decodes (TL_LEAD_LIVE=1,
// couch on PATH; read-only). TL_LEAD_SLOT names the slot, default ariadne:0.
func TestResumeCandidatesLive(t *testing.T) {
	if os.Getenv("TL_LEAD_LIVE") != "1" {
		t.Skip("set TL_LEAD_LIVE=1 to read the real couch")
	}
	slot := os.Getenv("TL_LEAD_SLOT")
	if slot == "" {
		slot = "ariadne:0"
	}
	if _, err := (execResume{bin: "couch"}).Candidates(slot); err != nil {
		t.Fatal(err)
	}
}

// The inventory's rows carry each checkout's claims on this machine; a slot
// takes its host row's non-terminal claims, and an unread or missing row
// leaves its claims unknown, never empty (trimmed from a live
// `sdlc fleet inventory --json`, 2026-10-19).
func TestDecodeFleetClaims(t *testing.T) {
	raw := `{"schema_version":1,"machine":{"state":"present"},
"rows":[
 {"tree_path":"/w/ariadne-slot1","claims_state":"present","claims":[
   {"ref":"ariadne#000471","status":"working","claimant":{"machine":"4716","workspace":"ariadne:1"}},
   {"ref":"ariadne#000400","status":"done"}]},
 {"tree_path":"/w/ariadne-slot2","claims_state":"absent","claims":[]},
 {"tree_path":"/w/ariadne-slot3","claims_state":"unknown","claims_error":"tracker fetch failed","claims":[]},
 {"tree_path":"/w/ariadne-slot2/pair","claims_state":"present","claims":[{"ref":"pair#000710","status":"working"}]}],
"slots":[
 {"address":"ariadne:1","repo":"ariadne","slot":1,"members":[{"role":"host","path":"/w/ariadne-slot1"}]},
 {"address":"ariadne:2","repo":"ariadne","slot":2,"members":[{"role":"host","path":"/w/ariadne-slot2"},{"role":"dependency","path":"/w/ariadne-slot2/pair"}]},
 {"address":"ariadne:3","repo":"ariadne","slot":3,"members":[{"role":"host","path":"/w/ariadne-slot3"}]},
 {"address":"ariadne:4","repo":"ariadne","slot":4,"members":[{"role":"host","path":"/w/ariadne-slot4"}]},
 {"address":"ariadne:5","repo":"ariadne","slot":5,"members":[{"role":"host","path":"/w/ariadne-slot1"},{"role":"dependency","path":"/w/ariadne-slot5/pair","verdict":"missing"}]}]}`
	f, err := decodeFleet([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]FleetSlot{}
	for _, s := range f.Slots {
		by[s.Address] = s
	}
	if s := by["ariadne:1"]; !s.ClaimsKnown || strings.Join(s.Claims, ",") != "ariadne#000471" {
		t.Errorf("ariadne:1 %+v", s)
	}
	// A claim on the slot's dependency clone is the slot's work too.
	if s := by["ariadne:2"]; !s.ClaimsKnown || strings.Join(s.Claims, ",") != "pair#000710" {
		t.Errorf("ariadne:2 (host none, dependency working) %+v", s)
	}
	if s := by["ariadne:3"]; s.ClaimsKnown || !strings.Contains(s.ClaimsWhy, "unknown: tracker fetch failed") {
		t.Errorf("ariadne:3 (unread) %+v", s)
	}
	if s := by["ariadne:4"]; s.ClaimsKnown || !strings.Contains(s.ClaimsWhy, "no inventory row") {
		t.Errorf("ariadne:4 (no row) %+v", s)
	}
	// A dependency sdlc reports missing (declared, not checked out) holds no
	// claims; it does not make the slot's claims unknown.
	if s := by["ariadne:5"]; !s.ClaimsKnown || strings.Join(s.Claims, ",") != "ariadne#000471" {
		t.Errorf("ariadne:5 (missing dependency) %+v", s)
	}
}

// The couch calls are bounded and say what happened: a hung --resume-candidates
// times out; a resume whose couch is missing did not run (nothing attempted),
// distinct from Couch refusing; a hung --resume is uncertain, never failed.
func TestExecResumeBoundsAndCauses(t *testing.T) {
	hang := filepath.Join(t.TempDir(), "couch")
	if err := os.WriteFile(hang, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := execResume{bin: hang, candidates: 200 * time.Millisecond, resume: 200 * time.Millisecond}
	if _, err := c.Candidates("pair:3"); err == nil || !strings.Contains(err.Error(), "no answer within") {
		t.Errorf("hung candidates: %v", err)
	}
	if err := c.Resume("pair:3"); !errors.Is(err, errUncertain) {
		t.Errorf("hung resume: %v", err)
	}
	missing := execResume{bin: filepath.Join(t.TempDir(), "no-couch")}
	if err := missing.Resume("pair:3"); err == nil || !strings.Contains(err.Error(), "did not run") {
		t.Errorf("missing couch: %v", err)
	}
	refuse := filepath.Join(t.TempDir(), "couch")
	if err := os.WriteFile(refuse, []byte("#!/bin/sh\necho 'slot busy' >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := (execResume{bin: refuse}).Resume("pair:3"); err == nil || !strings.Contains(err.Error(), "refused") || !strings.Contains(err.Error(), "slot busy") {
		t.Errorf("refusal: %v", err)
	}
}
