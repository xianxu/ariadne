package vocab

import (
	"reflect"
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
