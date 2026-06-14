package plan

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
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
	got, err := Plan(layers, nil) // no skills ⇒ no `## Skills` section appended
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
	got, err := Plan(layers, nil)
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
	// Skill lowering is still deferred (the M3 skill index feeds the menu, not a
	// file-op). It must not error or emit an Action here — just skip. (Tool and
	// Merge now lower — see TestPlanToolLowering / TestPlanMergeLowering.)
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/up", Intents: []intent.Intent{
			{Kind: intent.Skill, Source: "construct/skills/x", Target: "construct/skills/x"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Plan = %#v, want no Actions (deferred kinds skipped)", got)
	}
}

func TestPlanMergeLowering(t *testing.T) {
	// A `merge` intent lowers to a MergeSettings{Source, Target} — the settings
	// cascade (ported from setup.sh's `merge` case + merge-settings.sh). Source is
	// the layer's base settings (settings.ariadne.json), Target the composed
	// settings.json. The pure planner records the path facts; Apply reads base +
	// (optional) local off disk, runs mergeSettings, writes Target. The manifest
	// row is `merge .claude/settings.ariadne.json .claude/settings.json`.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/ws/ariadne", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanToolLowering(t *testing.T) {
	// A `tool` intent lowers to a ToolDep{Owner, Path}: Owner is the absolute
	// path of the layer that DECLARES the tool (its owner); Path is the tool's
	// path within that owner module. The pure planner records both facts; the
	// IO seam (Apply) decides derivative-vs-owner by comparing Owner to the
	// compiling repo root. Ported from setup.sh:ensure_go_tool_dependency, where
	// `upstream` = owner path, `source` = tool path.
	//
	// Two layers, foundation-first: ariadne (the owner, declares `tool cmd/sdlc`)
	// then a derivative repo as root (self, last). Both layers may carry the tool
	// intent (the derivative's own base.manifest also lists `tool cmd/sdlc`), and
	// each lowers to a ToolDep with its OWN layer Path as Owner — Apply then
	// classifies each by Owner==root.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/ws/ariadne", Intents: []intent.Intent{
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
		}},
		{Name: "derivative", Path: "/ws/derivative", Intents: []intent.Intent{
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"},
		ToolDep{Owner: "/ws/derivative", Path: "cmd/sdlc"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanProseOmittedWhenNoFragments(t *testing.T) {
	// No prose anywhere ⇒ no AGENTS.md WriteFile (don't emit an empty file).
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/up", Intents: []intent.Intent{
			{Kind: intent.Symlink, Source: "CLAUDE.md", Target: "CLAUDE.md"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{Symlink{Src: "/up/CLAUDE.md", Dst: "CLAUDE.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanAppendsSkillMenuToAGENTS(t *testing.T) {
	// With prose AND skills, the composed AGENTS.md ends with a `## Skills`
	// section: a note pointing at `weave skill <name>` for the body, then one
	// `name — description` line per skill, in menu order. The menu is always-on
	// discovery; the bodies are served on demand.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/a", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []string{"BASE PROSE"}},
	}
	menu := []skill.MenuItem{
		{Name: "xx-sdlc", Description: "SDLC checkpoint gates"},
		{Name: "superpowers-brainstorming", Description: "Brainstorm before building"},
	}
	got, err := Plan(layers, menu)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Plan = %#v, want one AGENTS.md WriteFile", got)
	}
	wf, ok := got[0].(WriteFile)
	if !ok || wf.Path != "AGENTS.md" {
		t.Fatalf("Plan[0] = %#v, want WriteFile{AGENTS.md}", got[0])
	}
	body := wf.Content
	if !strings.HasPrefix(body, "BASE PROSE\n") {
		t.Errorf("body should start with the prose, got:\n%s", body)
	}
	if !strings.Contains(body, "## Skills") {
		t.Errorf("body missing `## Skills` section:\n%s", body)
	}
	if !strings.Contains(body, "weave skill <name>") {
		t.Errorf("body missing the `weave skill <name>` note:\n%s", body)
	}
	if !strings.Contains(body, "xx-sdlc — SDLC checkpoint gates") {
		t.Errorf("body missing the xx-sdlc menu line:\n%s", body)
	}
	if !strings.Contains(body, "superpowers-brainstorming — Brainstorm before building") {
		t.Errorf("body missing the brainstorming menu line:\n%s", body)
	}
	// Menu order preserved: sdlc line before brainstorming line.
	if strings.Index(body, "xx-sdlc —") > strings.Index(body, "superpowers-brainstorming —") {
		t.Errorf("menu lines out of order:\n%s", body)
	}
}

func TestPlanSkillMenuWithoutProseStillWritesAGENTS(t *testing.T) {
	// Skills are always-on discovery: even with NO prose, a non-empty menu must
	// land in AGENTS.md (it's the floor's home), so the section still appears.
	menu := []skill.MenuItem{{Name: "xx-fix", Description: "fix markers"}}
	got, err := Plan(nil, menu)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Plan = %#v, want one AGENTS.md WriteFile", got)
	}
	wf := got[0].(WriteFile)
	if wf.Path != "AGENTS.md" || !strings.Contains(wf.Content, "xx-fix — fix markers") {
		t.Errorf("AGENTS.md missing the skill menu:\n%#v", wf)
	}
}
