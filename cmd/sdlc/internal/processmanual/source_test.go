package processmanual

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

func TestRenderManual_AbsoluteAndEmptyLinks(t *testing.T) {
	sources := []InjectionSource{
		{Kind: KindMemory, Title: "MEMORY.md", Link: "/home/u/.claude/projects/x/memory/MEMORY.md"},
		{Kind: KindMemory, Title: "(none)", Link: ""},
	}
	out := renderManual(sources, "../")
	// Absolute (outside-repo) links must NOT get the relative linkPrefix.
	if !strings.Contains(out, "(/home/u/.claude/projects/x/memory/MEMORY.md)") || strings.Contains(out, "../home/u") {
		t.Errorf("absolute link should be untouched by linkPrefix:\n%s", out)
	}
	// An empty link renders as a plain heading, not `[(none)]()`.
	if strings.Contains(out, "[(none)]()") || !strings.Contains(out, "### (none)") {
		t.Errorf("empty link should render as a plain heading:\n%s", out)
	}
}

func TestRenderManual_FencesHeadingBearingBodies(t *testing.T) {
	sources := []InjectionSource{
		{Kind: KindSDLCPrompt, Title: "dry", Link: "prompts.go", Body: "You are a reviewer.\n\n# ARCH-DRY\n\nprinciple text"},
		{Kind: KindLessons, Title: "lessons.md", Link: "workshop/lessons.md", Body: "Prefer X over Y."},
	}
	out := renderManual(sources, "")
	// A body containing `# ARCH-DRY` must be fenced so the heading renders as
	// literal text (inside the fence) rather than hijacking the manual structure.
	if !strings.Contains(out, "~~~\nYou are a reviewer.") {
		t.Errorf("heading-bearing body should open with a ~~~ fence:\n%s", out)
	}
	// The embedded `# ARCH-DRY` line must sit AFTER a fence open (i.e. inside a
	// fenced block), not stand alone as a live header.
	fenceOpen := strings.Index(out, "~~~")
	heading := strings.Index(out, "\n# ARCH-DRY\n")
	if fenceOpen < 0 || heading < fenceOpen {
		t.Errorf("`# ARCH-DRY` should be inside the fence:\n%s", out)
	}
	// A prose-only body is left unfenced.
	if strings.Contains(out, "~~~\nPrefer X over Y") {
		t.Errorf("prose-only body should not be fenced:\n%s", out)
	}
}
