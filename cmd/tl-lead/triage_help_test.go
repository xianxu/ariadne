package main

import (
	"strings"
	"testing"
)

// #633: the weekly triage duty states its steps in order, with the marks, the
// one-bug-slot default, the velocity sizing, themes as projects and the
// Needs-you shape.
func TestLeadHelpWeeklyTriage(t *testing.T) {
	steps := []struct {
		heading string
		wants   []string
	}{
		{"1. Read the numbers:", []string{"tl-lead triage stats", "opened and closed", "open bugs at the week's end", "close date", "approximate", "open counts only"}},
		{"2. Mark each open issue:", []string{"sdlc issue set-kind", "bug or feature", "unclassified", "ready", "waiting", "stale", "duplicate", "not on the card", "sdlc abandon"}},
		{"3. Capacity:", []string{"one bug slot per repo", "bug quota", "median", "never below one", "The rest wait"}},
		{"4. Features:", []string{"themes", "candidate project", "project datatype"}},
		{"5. Report in Needs you", []string{"before anything proceeds", "bugs this week", "close?", "feature themes, proposed as projects", "trade-off", "operator decides"}},
		{"6. The 21:00 ready-queue check", []string{"approved ready list"}},
	}
	for _, args := range [][]string{{"--help"}, {}} {
		help := strings.Join(strings.Fields(helpOutput(t, args...)), " ")
		start, end := strings.Index(help, "WEEKLY BACKLOG TRIAGE "), strings.Index(help, "DISPATCH ")
		if start < 0 || end <= start {
			t.Fatalf("tl-lead %v: WEEKLY BACKLOG TRIAGE must precede DISPATCH", args)
		}
		section := help[start:end]
		for _, want := range []string{"Once a week", "scheduled task", "Nothing proceeds until the operator approves"} {
			if !strings.Contains(section, want) {
				t.Errorf("tl-lead %v: triage preamble lacks %q", args, want)
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
					t.Fatalf("tl-lead %v: triage steps out of order after %q", args, step.heading)
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
