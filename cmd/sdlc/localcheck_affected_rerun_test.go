package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

// #586: a known-flake child of an affected run also runs every selected
// group; the suite's batch, and a selection without groups, are unchanged.
func TestAffectedRerunBatch(t *testing.T) {
	rerun := [][]string{{"go", "test", "-run", "^TestA$", "./a"}}
	if got := affectedRerunBatch(rerun, nil); !reflect.DeepEqual(got, rerun) {
		t.Fatalf("suite: %v", got)
	}
	if got := affectedRerunBatch(rerun, &affectedRerun{Command: []string{"./t"}, GoPkgs: []string{"p"}}); !reflect.DeepEqual(got, rerun) {
		t.Fatalf("no groups: %v", got)
	}
	got := affectedRerunBatch(rerun, &affectedRerun{Command: []string{"./t", "x"}, GoPkgs: []string{"p"}, Groups: []string{"a", "b"}})
	want := append(append([][]string(nil), rerun...), []string{"env", "SDLC_AFFECTED_GO_PKGS=", "SDLC_AFFECTED_GROUPS=a b", "./t", "x"})
	if !reflect.DeepEqual(got, want) || len(rerun) != 1 {
		t.Fatalf("groups: %v (input %v)", got, rerun)
	}
	// One env builder: the run's command and the rerun's agree on the contract.
	c := affectedLocalCheck(affected.Config{Command: []string{"./t", "x"}}, affected.Selection{GoPkgs: []string{"p"}, Groups: []string{"a", "b"}}, localCheck{})
	if !reflect.DeepEqual(c.Command, affectedEnvCommand([]string{"./t", "x"}, []string{"p"}, []string{"a", "b"})) || c.Rerun == nil || !reflect.DeepEqual(c.Rerun.Groups, []string{"a", "b"}) {
		t.Fatalf("affected check %+v", c)
	}
}

func TestRerunEnvelope(t *testing.T) {
	c := localCheck{TimeoutMinutes: 30, Weight: 2}
	if w, d := rerunEnvelope(c, time.Minute); w != 1 || d != knownFlakeRerunTimeout {
		t.Fatalf("tests only: %d %v", w, d)
	}
	c.Rerun = &affectedRerun{Groups: []string{"a"}}
	if w, d := rerunEnvelope(c, time.Minute); w != 2 || d != 30*time.Minute {
		t.Fatalf("with groups: %d %v", w, d)
	}
}
