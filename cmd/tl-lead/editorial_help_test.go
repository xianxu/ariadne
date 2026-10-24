package main

import (
	"strings"
	"testing"
)

func TestLeadEditorialCreationHelp(t *testing.T) {
	s := helpOutput(t, "--help")
	for _, want := range []string{
		"`@@chat@@`", "directly above `## Operator instructions`", "no review mode keys", "YYYY-MM/YYYY-MM-DD-NN-<slug>.md",
		"parent: ../../session.md", "Existing flat", "collaboration-modes.md",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("rendered help lacks %q", want)
		}
	}
	for _, retired := range []string{"collaboration-mode: {chat:", "same session.side/ folder (NN-", "--add session.side/NN-",
		"collaboration-mode: discussion", "presentation-mode: highlight", "processing-mode: free-form", "explicit default mode keys", "use processing-instruction"} {
		if strings.Contains(s, retired) {
			t.Errorf("rendered help still prescribes %q", retired)
		}
	}
}
