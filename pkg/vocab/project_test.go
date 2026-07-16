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
	if got := Project().InitialStatus(); got != "ideation" {
		t.Fatalf("initial = %q", got)
	}
}

func TestProjectAllStatuses(t *testing.T) {
	want := []string{"ideation", "defined", "committed", "executing", "paused", "done", "dropped"}
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
		{"ideation", true, false, false},
		{"defined", true, false, false},
		{"committed", false, false, false},
		{"executing", false, true, false},
		{"paused", false, true, false},
		{"done", false, false, true},
		{"dropped", false, false, true},
	}
	for _, c := range cases {
		if m.IsForming(c.s) != c.forming || m.IsExecuting(c.s) != c.executing || m.IsTerminal(c.s) != c.terminal {
			t.Errorf("%s: IsForming=%v IsExecuting=%v IsTerminal=%v, want %v/%v/%v",
				c.s, m.IsForming(c.s), m.IsExecuting(c.s), m.IsTerminal(c.s), c.forming, c.executing, c.terminal)
		}
	}
}

func TestProjectTransitionFor(t *testing.T) {
	tr := Project().TransitionFor("executing", "done")
	if tr == nil || tr.Event != "close" {
		t.Fatalf("executing→done = %+v", tr)
	}
	if len(tr.Guards) != 2 || tr.Guards[0] != "retro-recorded" || tr.Guards[1] != "fog-factor-recorded" {
		t.Fatalf("close guards = %v", tr.Guards)
	}
}

func TestProjectTransitionForEventDerivesCloseAndDrop(t *testing.T) {
	m := Project()
	closeTr := m.FirstTransitionForEvent("close")
	if closeTr == nil || m.TransitionForEvent(closeTr.From, "close") == nil {
		t.Fatal("close event transition not derivable")
	}
	if tr := m.TransitionForEvent("paused", "drop"); tr == nil || tr.Event != "drop" {
		t.Fatalf("paused drop transition = %+v", tr)
	}
}

func TestProjectSections(t *testing.T) {
	secs := Project().Sections()
	want := []string{"PRD", "Estimate", "Breakdown", "Log"}
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
}

func TestProjectRenderLifecycleHelp(t *testing.T) {
	help := Project().RenderLifecycleHelp()
	for _, frag := range []string{"STATUSES", "LEGAL TRANSITIONS", "ideation", "baseline set"} {
		if !strings.Contains(help, frag) {
			t.Fatalf("lifecycle help missing %q:\n%s", frag, help)
		}
	}
}
