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
	work := func(owner string, n int) RepoIssue {
		return RepoIssue{ID: fmt.Sprintf("%06d", n), Status: "working", Owner: owner}
	}
	s := &fakeSDLC{
		fleet: Fleet{Slots: slots},
		issues: map[string][]RepoIssue{"/w/pair": {
			work("pair:2", 702), work("pair:3", 703), work("pair:5", 705), work("pair:6", 706), work("pair:8", 708),
			{ID: "000704", Status: "done", Owner: "pair:4"}, // finished work is not in flight
		}},
		issueErr: map[string]error{"/w/ariadne": errors.New("tracker unreadable")},
	}
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
		"ambiguous pair:3  #703; no proven conversation",
		"ambiguous brain:1  unclaimed",
		"ambiguous tools:1  claimed by ops and ops-echo", // a tie is never this hub's to resume
		"unknown   pair:5  #705; candidates unreadable: couch busy",
		"unknown   ariadne:1  claims unreadable",
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
