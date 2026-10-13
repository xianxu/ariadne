package project

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func guardDoc(t *testing.T, fm, trains, log string) *Doc {
	t.Helper()
	d, err := ParseDoc("---\ntype: project\nname: demo\nstatus: executing\n" + fm + "---\n## Charter\nWhy.\n## Trains\n" + trains + "\n## Breakdown\n- [ ]\n## Log\n" + log + "\n## Decisions\n")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// Every guard the vocabulary names is registered, and nothing else is: a
// retired guard left in the registry would be dead code, a missing one a
// runtime refusal.
func TestProjectGuardsMatchVocabulary(t *testing.T) {
	named := map[string]bool{}
	for _, tr := range vocab.Project().Lifecycle {
		for _, g := range tr.Guards {
			named[g] = true
		}
	}
	reg := Guards()
	for g := range named {
		if reg[g] == nil {
			t.Errorf("vocabulary names guard %q but the registry lacks it", g)
		}
	}
	for g := range reg {
		if !named[g] {
			t.Errorf("registry holds guard %q that no transition names", g)
		}
	}
}

func TestTrainsPresentGuard(t *testing.T) {
	g := Guards()["trains-present"]
	if err := g(guardDoc(t, "", "| Train | Cars, in order | Slot |\n|---|---|---|", ""), GuardCtx{}); err == nil {
		t.Error("empty trains table passed")
	}
	if err := g(guardDoc(t, "", "| Train | Cars, in order | Slot |\n|---|---|---|\n| A | ariadne#1 | :1 |", ""), GuardCtx{}); err != nil {
		t.Errorf("one train failed: %v", err)
	}
}

func TestMVPLandedGuard(t *testing.T) {
	g := Guards()["mvp-landed"]
	terminal := map[string]bool{"ariadne#1": true, "pair#2": false}
	ctx := GuardCtx{IssueTerminal: func(ref string) (bool, error) {
		done, ok := terminal[ref]
		if !ok {
			return false, errors.New("no such issue")
		}
		return done, nil
	}}
	if err := g(guardDoc(t, "", "", ""), ctx); err == nil || !strings.Contains(err.Error(), "mvp_scope is empty") {
		t.Errorf("empty mvp_scope = %v", err)
	}
	if err := g(guardDoc(t, "mvp_scope: [ariadne#1, pair#2]\n", "", ""), ctx); err == nil || !strings.Contains(err.Error(), "pair#2") || strings.Contains(err.Error(), "ariadne#1") {
		t.Errorf("open MVP issue = %v, want only pair#2 named", err)
	}
	terminal["pair#2"] = true
	if err := g(guardDoc(t, "mvp_scope: [ariadne#1, pair#2]\n", "", ""), ctx); err != nil {
		t.Errorf("all landed = %v", err)
	}
	if err := g(guardDoc(t, "mvp_scope: [ariadne#1, gone#3]\n", "", ""), ctx); err == nil || !strings.Contains(err.Error(), "gone#3") {
		t.Errorf("lookup error = %v, want it surfaced", err)
	}
	if err := g(guardDoc(t, "mvp_scope: [ariadne#1]\n", "", ""), GuardCtx{}); err == nil {
		t.Error("no lookup passed")
	}
}

func TestRetroRecordedGuard(t *testing.T) {
	g := Guards()["retro-recorded"]
	if err := g(guardDoc(t, "", "", "\n### 2026-07-16 — retro\nLearned.\n"), GuardCtx{}); err != nil {
		t.Errorf("retro failed: %v", err)
	}
	if err := g(guardDoc(t, "", "", "\n### 2026-07-16\nNot retro.\n"), GuardCtx{}); err == nil {
		t.Error("non-retro heading passed")
	}
}
