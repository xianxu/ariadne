package plan

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
)

// Plan lowers a foundation-first []Layer into []Action. Pure: tested by
// asserting the Actions computed from in-memory Layers (no IO, ARCH-PURE).
// Layers arrive in resolved (foundation-first) order; root is last and self.

func TestPlanProseAcrossLayersToOneAGENTS(t *testing.T) {
	// Three layers, each contributing a prose fragment; foundation-first.
	// All Prose intents lower to ONE WriteFile{AGENTS.md, composed body} with
	// the fragments concatenated in layer order (the @AGENTS.local.md fix).
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/a", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []string{"BASE"}},
		{Name: "nous", Path: "/n", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []string{"LAYER"}},
		{Name: "brain", Path: "/b", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []string{"LOCAL"}},
	}
	got, err := Plan(layers)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		WriteFile{Path: "AGENTS.md", Content: "BASE\n\nLAYER\n\nLOCAL\n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanSymlinkAndScaffold(t *testing.T) {
	// File-op intents lower near-identity to their Actions (ported from
	// walk_manifest's case). symlink → Symlink{upstream/src, target},
	// scaffold → Mkdir{target}, touch → Touch{target} (create-if-missing, NOT a
	// content-clobbering WriteFile).
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/up", Intents: []intent.Intent{
			{Kind: intent.Symlink, Source: "AGENTS.md", Target: "AGENTS.md"},
			{Kind: intent.Scaffold, Source: ".claude/skills", Target: ".claude/skills"},
			{Kind: intent.Touch, Source: "workshop/lessons.md", Target: "workshop/lessons.md"},
		}},
	}
	got, err := Plan(layers)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		Symlink{Src: "/up/AGENTS.md", Dst: "AGENTS.md"},
		Mkdir{Path: ".claude/skills"},
		Touch{Path: "workshop/lessons.md"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanDeferredKindsAreNoOps(t *testing.T) {
	// Merge/Tool/Skill lowering is deferred (M4 settings, M3 skill, exec
	// seam). They must not error or emit Actions in this unit — just skip.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/up", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: "a.json", Target: "b.json"},
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
			{Kind: intent.Skill, Source: "construct/skills/x", Target: "construct/skills/x"},
		}},
	}
	got, err := Plan(layers)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Plan = %#v, want no Actions (deferred kinds skipped)", got)
	}
}

func TestPlanProseOmittedWhenNoFragments(t *testing.T) {
	// No prose anywhere ⇒ no AGENTS.md WriteFile (don't emit an empty file).
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/up", Intents: []intent.Intent{
			{Kind: intent.Symlink, Source: "CLAUDE.md", Target: "CLAUDE.md"},
		}},
	}
	got, err := Plan(layers)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{Symlink{Src: "/up/CLAUDE.md", Dst: "CLAUDE.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}
