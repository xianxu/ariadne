package vocab

import (
	"reflect"
	"strings"
	"testing"
)

func TestIssuePredicates(t *testing.T) {
	m := Issue()
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsTerminal(done)", m.IsTerminal("done"), true},
		{"IsTerminal(working)", m.IsTerminal("working"), false},
		{"IsActive(blocked)", m.IsActive("blocked"), true},
		{"IsActive(done)", m.IsActive("done"), false},
		{"IsOpen(open)", m.IsOpen("open"), true},
		{"IsOpen(working)", m.IsOpen("working"), false},
		{"CanTransition(open,working)", m.CanTransition("open", "working"), true},
		{"CanTransition(open,done)", m.CanTransition("open", "done"), false},
		{"CanTransition(done,working)", m.CanTransition("done", "working"), true},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestAllStatuses(t *testing.T) {
	// Ordered open → active → terminal; must match the legacy validStatuses set.
	want := []string{"open", "working", "blocked", "done", "wontfix", "punt"}
	if got := Issue().AllStatuses(); !reflect.DeepEqual(got, want) {
		t.Errorf("AllStatuses() = %v, want %v", got, want)
	}
}

// TestRenderLifecycleHelp_DerivesFromModel pins #125's render: every status appears
// with its When semantics, the legal edges are present, and the output is byte-stable
// (two calls identical) — so the help text can't drift from the model.
func TestRenderLifecycleHelp_DerivesFromModel(t *testing.T) {
	m := Issue()
	out := m.RenderLifecycleHelp()
	if out != m.RenderLifecycleHelp() {
		t.Fatal("RenderLifecycleHelp is not byte-stable across calls")
	}
	for _, s := range m.AllStatuses() {
		if !strings.Contains(out, s) {
			t.Errorf("render missing status %q", s)
		}
		if w := m.When[s]; w != "" && !strings.Contains(out, w) {
			t.Errorf("render missing When semantics for %q (%q)", s, w)
		}
	}
	// A known legal edge renders (open → working).
	if !strings.Contains(out, "STATUSES") || !strings.Contains(out, "LEGAL TRANSITIONS") {
		t.Errorf("render missing a section header:\n%s", out)
	}
	if !strings.Contains(out, "working") {
		t.Errorf("render missing the open→working legal target:\n%s", out)
	}
}

func TestStatusNamesAndGloss(t *testing.T) {
	m := Issue()
	names := m.StatusNames(" | ")
	for _, s := range m.AllStatuses() {
		if !strings.Contains(names, s) {
			t.Errorf("StatusNames missing %q: %q", s, names)
		}
	}
	gloss := m.StatusGloss()
	if m.StatusGloss() != gloss {
		t.Fatal("StatusGloss not byte-stable")
	}
	for _, s := range m.AllStatuses() {
		if w := m.When[s]; w != "" && !strings.Contains(gloss, w) {
			t.Errorf("StatusGloss missing When for %q", s)
		}
	}
}
