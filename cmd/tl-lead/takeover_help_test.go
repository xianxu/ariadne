package main

import (
	"strings"
	"testing"
)

func TestLeadHelpTakeoverSequence(t *testing.T) {
	steps := []struct {
		heading string
		wants   []string
	}{
		{"1. Read the contract.", []string{"tl-lead --help", "follow it"}},
		{"2. Read the session.", []string{"session.md", "session.side/", "Threads", "_updated YYYY-MM-DD HH:MM_", "ephemeral scratch pad", "deletes handled items", "not a complete history", "tolerate some loss"}},
		{"3. Fill only the gap.", []string{"If the gap since that stamp matters", "previous agent's terminal transcript", "from that time onward", "couch --peek <hub> --previous --all", "transcript on disk", "only for the gap"}},
		{"4. Re-establish.", []string{"Peek working slots", "sdlc landing status", "sdlc queue status", "open project files", "refresh Threads", "tl-lead status", "fresh stamp"}},
		{"5. Give the takeover notice.", []string{"operator", "one line", "takeover is done", "may have been lost"}},
	}
	for _, args := range [][]string{{"--help"}, {}} {
		help := strings.Join(strings.Fields(helpOutput(t, args...)), " ")
		start, end := strings.Index(help, "TAKEOVER ("), strings.Index(help, "ROLES ")
		if start < 0 || end <= start {
			t.Fatalf("tl-lead %v: TAKEOVER must precede ROLES", args)
		}
		section := help[start:end]
		for _, want := range []string{"forced hub agent switch", "same-slot force switch", "fresh hub session"} {
			if !strings.Contains(section, want) {
				t.Errorf("tl-lead %v: takeover scope lacks %q", args, want)
			}
		}
		for i, step := range steps {
			pos := strings.Index(section, step.heading)
			if pos < 0 {
				t.Fatalf("tl-lead %v: missing step %q", args, step.heading)
			}
			next := len(section)
			if i+1 < len(steps) {
				next = strings.Index(section, steps[i+1].heading)
				if next <= pos {
					t.Fatalf("tl-lead %v: takeover steps out of order after %q", args, step.heading)
				}
			}
			for _, want := range step.wants {
				if !strings.Contains(section[pos:next], want) {
					t.Errorf("tl-lead %v: %s lacks %q", args, step.heading, want)
				}
			}
		}
	}
}
