package main

import (
	"strings"
	"testing"
)

// Embedded help is the delivered procedure. Check both entry points and keep
// ownership inside each ordered step, so swapping drivers cannot pass by
// leaving the right words elsewhere in the help.
func TestLeadHelpShapingSequence(t *testing.T) {
	steps := []struct {
		heading string
		wants   []string
	}{
		{"1. Users and moments.", []string{"Operator drives", "decision"}},
		{"2. Flows.", []string{"Operator drives", "TL spots gaps", "step by step"}},
		{"3. Clickable mock", []string{"TL builds", "operator sketches", "before components or issues"}},
		{"4. Visual design", []string{"Operator supplies style references", "TL proposes two or three directions", "same wireframe", "tokens.css", "component patterns"}},
		{"5. Components and boundaries.", []string{"TL proposes", "operator smell-checks", "repos", "contracts"}},
		{"6. Slices.", []string{"TL drives", "operator reviews", "dependencies", "acceptance", "operator design review", "failure-sensitive"}},
	}
	for _, args := range [][]string{{"--help"}, {}} {
		help := strings.Join(strings.Fields(helpOutput(t, args...)), " ")
		start, end := strings.Index(help, "SHAPING ("), strings.Index(help, "RUNNING A PROJECT (")
		if start < 0 || end <= start {
			t.Fatalf("tl-lead %v: shaping must precede the execution playbook", args)
		}
		section := help[start:end]
		for _, want := range []string{"before filing implementation issues", "team's ops repo at :0", "Operator reviews each step before the next", "project.md", "design.md", "mocks/", "straight to main"} {
			if !strings.Contains(section, want) {
				t.Errorf("tl-lead %v: shaping boundary lacks %q", args, want)
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
					t.Fatalf("tl-lead %v: shaping steps out of order after %q", args, step.heading)
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

func TestLeadHelpShapingMockAndShippingBoundary(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {}} {
		help := strings.Join(strings.Fields(helpOutput(t, args...)), " ")
		start, end := strings.Index(help, "SHAPING ("), strings.Index(help, "RUNNING A PROJECT (")
		if start < 0 || end <= start {
			t.Fatalf("tl-lead %v: shaping section missing", args)
		}
		section := help[start:end]
		for _, want := range []string{
			"HTML, CSS and a little JS", "no framework", "couch mock serve <project>/mocks", "live reload",
			"state list", "screen", "actions", "empty", "error", "long-list", "fixtures.json",
			"description", "blank canvas", "mock-1", "mock-2", "current version",
			"version and screen", "new version starts clean", "reverting", "sketches reappear",
			"screen by screen", "same fixtures", "API response shape", "contract tests",
			"development-only", "per-screen", "MOCK badge", "real", "mock-backed",
			"production build refuses", "not yet", "never fixture-backed mock screens", "screens real",
		} {
			if !strings.Contains(section, want) {
				t.Errorf("tl-lead %v: shaping mock contract lacks %q", args, want)
			}
		}
	}
}
