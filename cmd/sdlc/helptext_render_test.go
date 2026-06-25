package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// TestRenderLong_SetStatusDerivesLifecycle pins #125: the set-status help renders the
// lifecycle FACTS from the model (every status + its When + a LEGAL TRANSITIONS
// section), carries no surviving placeholder, and the hand-maintained "all other
// transitions allowed" claim is gone (the drift that motivated the issue).
func TestRenderLong_SetStatusDerivesLifecycle(t *testing.T) {
	long := renderLong("set-status")
	if strings.Contains(long, "{{") {
		t.Errorf("set-status Long has an unsubstituted placeholder:\n%s", long)
	}
	if strings.Contains(long, "all other transitions allowed") {
		t.Error("set-status still carries the hand-maintained 'all other transitions allowed' claim (#125 Done-when)")
	}
	if !strings.Contains(long, "LEGAL TRANSITIONS") {
		t.Error("set-status should render the derived LEGAL TRANSITIONS section")
	}
	m := vocab.Issue()
	for _, s := range m.AllStatuses() {
		if !strings.Contains(long, s) {
			t.Errorf("set-status help missing model status %q", s)
		}
		if w := m.When[s]; w != "" && !strings.Contains(long, w) {
			t.Errorf("set-status help missing When semantics for %q (%q)", s, w)
		}
	}
}

// TestRenderLong_IssueDerivesStatuses pins issue.md's status block derives too — the
// biggest restatement per the plan-quality shadow-sweep.
func TestRenderLong_IssueDerivesStatuses(t *testing.T) {
	long := renderLong("issue")
	if strings.Contains(long, "{{") {
		t.Errorf("issue Long has an unsubstituted placeholder:\n%s", long)
	}
	for _, s := range vocab.Issue().AllStatuses() {
		if !strings.Contains(long, s) {
			t.Errorf("issue help missing model status %q", s)
		}
	}
}

// TestNoCommandLongHasSurvivingPlaceholder is the wiring guard: every command's Long
// in the REAL tree is loaded via renderLong, so no help text ships a literal {{…}}.
// This catches a future helptext placeholder loaded through a forgotten bare
// helptext.MustGet site — exactly the Critical plan-quality finding (set-status's
// issue.go:46 + alias load sites bypass main.go's add()), as a regression guard.
func TestNoCommandLongHasSurvivingPlaceholder(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if strings.Contains(c.Long, "{{") {
			t.Errorf("command %q ships an unsubstituted help placeholder (a Long load that skipped renderLong):\n%s", c.CommandPath(), c.Long)
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(buildRoot())
}
