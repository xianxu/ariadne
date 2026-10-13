package vocab

import (
	"strings"
	"testing"
)

func TestProjectDiscovery(t *testing.T) {
	d := Project().Discovery()
	if d.Home != "workshop/projects" || d.Glob != "*.md" || d.Archive != "workshop/history" {
		t.Fatalf("discovery = %+v", d)
	}
	if d.Plans != "" {
		t.Fatalf("projects have no plan sidecars; Plans = %q", d.Plans)
	}
}

func TestProjectInitialStatus(t *testing.T) {
	if got := Project().InitialStatus(); got != "defined" {
		t.Fatalf("initial = %q", got)
	}
}

func TestProjectAllStatuses(t *testing.T) {
	want := []string{"defined", "executing", "closing", "closed", "dropped"}
	got := Project().AllStatuses()
	if len(got) != len(want) {
		t.Fatalf("statuses = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("statuses = %v, want %v", got, want)
		}
	}
}

func TestProjectPredicates(t *testing.T) {
	m := Project()
	cases := []struct {
		s                            string
		forming, executing, terminal bool
	}{
		{"defined", true, false, false},
		{"executing", false, true, false},
		{"closing", false, true, false},
		{"closed", false, false, true},
		{"dropped", false, false, true},
		// retired by #329: no longer statuses at all
		{"ideation", false, false, false},
		{"committed", false, false, false},
		{"paused", false, false, false},
		{"done", false, false, false},
	}
	for _, c := range cases {
		if m.IsForming(c.s) != c.forming || m.IsExecuting(c.s) != c.executing || m.IsTerminal(c.s) != c.terminal {
			t.Errorf("%s: IsForming=%v IsExecuting=%v IsTerminal=%v, want %v/%v/%v",
				c.s, m.IsForming(c.s), m.IsExecuting(c.s), m.IsTerminal(c.s), c.forming, c.executing, c.terminal)
		}
	}
}

// TestProjectTransitions pins the #329 funnel: each edge's event and guards.
func TestProjectTransitions(t *testing.T) {
	m := Project()
	cases := []struct {
		from, to, event string
		guards          []string
	}{
		{"defined", "executing", "execute", []string{"trains-present"}},
		{"executing", "closing", "wrap", []string{"mvp-landed"}},
		{"closing", "executing", "reopen", nil},
		{"closing", "closed", "close", []string{"retro-recorded"}},
		{"defined", "dropped", "drop", nil},
		{"executing", "dropped", "drop", []string{"retro-recorded"}},
		{"closing", "dropped", "drop", []string{"retro-recorded"}},
	}
	for _, c := range cases {
		tr := m.TransitionFor(c.from, c.to)
		if tr == nil || tr.Event != c.event || strings.Join(tr.Guards, ",") != strings.Join(c.guards, ",") {
			t.Errorf("%s→%s = %+v, want event %q guards %v", c.from, c.to, tr, c.event, c.guards)
		}
	}
	if len(m.Lifecycle) != len(cases) {
		t.Errorf("lifecycle has %d edges, want %d", len(m.Lifecycle), len(cases))
	}
	// close only from closing: a project wraps (every MVP issue landed) first.
	if m.CanTransition("executing", "closed") {
		t.Error("executing→closed must not exist; wrap to closing first")
	}
}

func TestProjectTransitionForEventDerivesCloseAndDrop(t *testing.T) {
	m := Project()
	closeTr := m.FirstTransitionForEvent("close")
	if closeTr == nil || closeTr.From != "closing" || m.TransitionForEvent(closeTr.From, "close") == nil {
		t.Fatalf("close event transition = %+v", closeTr)
	}
	if tr := m.TransitionForEvent("closing", "drop"); tr == nil || tr.Event != "drop" {
		t.Fatalf("closing drop transition = %+v", tr)
	}
}

func TestProjectSections(t *testing.T) {
	secs := Project().Sections()
	want := []string{"Charter", "Trains", "Breakdown", "Log", "Decisions"}
	if len(secs) != len(want) {
		t.Fatalf("sections = %+v", secs)
	}
	for i, s := range secs {
		if s.Name != want[i] {
			t.Fatalf("sections[%d] = %+v, want name %q", i, s, want[i])
		}
	}
	if secs[2].Seed != "- [ ]" {
		t.Fatalf("Breakdown seed = %q", secs[2].Seed)
	}
	if !strings.HasPrefix(secs[1].Seed, "| Train | Cars, in order | Slot |") {
		t.Fatalf("Trains seed = %q", secs[1].Seed)
	}
}

func TestProjectRenderLifecycleHelp(t *testing.T) {
	help := Project().RenderLifecycleHelp()
	for _, frag := range []string{"STATUSES", "LEGAL TRANSITIONS", "closing", "done_when audit"} {
		if !strings.Contains(help, frag) {
			t.Fatalf("lifecycle help missing %q:\n%s", frag, help)
		}
	}
}
