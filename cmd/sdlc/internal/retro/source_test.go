package retro

import (
	"strings"
	"testing"
)

func TestRenderManual_GroupsByKindWithLinks(t *testing.T) {
	sources := []InjectionSource{
		{Kind: KindHelpText, Title: "close", When: "printed by `sdlc close --help`", Link: "cmd/sdlc/helptext/close.md", Body: "Close gate…"},
		{Kind: KindSDLCPrompt, Title: "milestone-review", When: "boundary review at `sdlc close`", Link: "cmd/sdlc/internal/judge/prompts.go", Body: "You are conducting a fresh-context review…"},
	}
	out := renderManual(sources, "")

	// Grouped section headers appear, sdlc-prompt group before help-text (stable order).
	if i, j := strings.Index(out, "## sdlc-injected prompts"), strings.Index(out, "## Help text"); i < 0 || j < 0 || i > j {
		t.Fatalf("sections missing or misordered:\n%s", out)
	}
	// Each source renders a linked title, its When, and its Body.
	for _, want := range []string{
		"[milestone-review](cmd/sdlc/internal/judge/prompts.go)",
		"boundary review at `sdlc close`",
		"You are conducting a fresh-context review",
		"[close](cmd/sdlc/helptext/close.md)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("manual missing %q:\n%s", want, out)
		}
	}
}

func TestRenderManual_LinkPrefixApplied(t *testing.T) {
	out := renderManual([]InjectionSource{{Kind: KindLessons, Title: "lessons.md", Link: "workshop/lessons.md"}}, "../")
	if !strings.Contains(out, "(../workshop/lessons.md)") {
		t.Errorf("linkPrefix not applied:\n%s", out)
	}
}
