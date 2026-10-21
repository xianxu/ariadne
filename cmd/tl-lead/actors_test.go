package main

import (
	"fmt"
	"maps"
	"testing"
)

// Resting is a Git observation, not prompt readiness. In particular it can
// coexist with a permission picker or an active turn (ariadne#533).
func TestActorLabelDoesNotInferIdleFromRestingBranch(t *testing.T) {
	for _, agent := range []string{"claude", "codex", "qoder"} {
		for _, resting := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/resting=%t", agent, resting), func(t *testing.T) {
				raw := fmt.Sprintf(`{"Code":"ok","Actors":[{"Binding":{"Slot":"repo:1","Agent":%q},"Resting":%t,"Known":true}]}`, agent, resting)
				actors, err := decodeActors([]byte(raw))
				if err != nil {
					t.Fatal(err)
				}
				if got := agentLabel(actors, nil, "repo:1"); got != agent {
					t.Fatalf("branch observation fabricated runtime readiness: got %q, want %q", got, agent)
				}
			})
		}
	}
}

// fakeActors stands in for `couch --actors --json`.
type fakeActors struct {
	actors map[string]Actor
	err    error
}

func (f fakeActors) Actors() (map[string]Actor, error) { return maps.Clone(f.actors), f.err }

// decodeActors reads the fields tl-lead uses from a reply shaped like couch's
// (trimmed from a live `couch --actors --json`, 2026-10-16).
func TestDecodeActors(t *testing.T) {
	raw := `{"Code":"ok","Error":"","Receipt":null,"Actors":[
{"Binding":{"Slot":"pair:4","Repository":"/w/pair/.git","Agent":"qoder","Version":"1.1.71","PID":65513},"Resting":false,"Known":true,"Pending":false},
{"Binding":{"Slot":"ariadne:1","Repository":"/w/ariadne/.git","Agent":"codex","PID":74433},"Resting":true,"Known":true}]}`
	got, err := decodeActors([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Actor{"pair:4": {Agent: "qoder"}, "ariadne:1": {Agent: "codex", Resting: true}}
	if !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if _, err := decodeActors([]byte(`{"Code":"refused","Error":"no couch"}`)); err == nil {
		t.Error("a refused reply decoded")
	}
}
