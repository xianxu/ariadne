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
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "BASE"}}},
		{Name: "nous", Path: "/n", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "LAYER"}}},
		{Name: "brain", Path: "/b", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "LOCAL"}}},
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

func TestPlanProseVisibilitySelection(t *testing.T) {
	// The 𝒜(R) invariant (workshop/targets/base-layer-mechanics.md): a
	// synthetic 3-layer stack — foundation with BOTH export prose AND internal
	// prose; a middle layer with export prose; a leaf with internal prose. The
	// leaf composes [foundation-export, middle-export, leaf-internal] and does
	// NOT contain the foundation's or middle's internal prose. (The middle here
	// has no internal; the foundation's internal is the exclusion proof.)
	layers := []layer.Layer{
		{Name: "foundation", Path: "/f", Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Export, Source: "AGENTS.base.md", Target: "AGENTS.base.md"},
			{Kind: intent.Prose, Visibility: intent.Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Export, Content: "FOUNDATION-EXPORT"},
			{Visibility: intent.Internal, Content: "FOUNDATION-INTERNAL"},
		}},
		{Name: "middle", Path: "/m", Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Export, Source: "AGENTS.base.md", Target: "AGENTS.base.md"},
		}, ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Export, Content: "MIDDLE-EXPORT"},
		}},
		{Name: "leaf", Path: "/l", Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Internal, Content: "LEAF-INTERNAL"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	// export-prose foundation-first (foundation, middle) then the LEAF's internal LAST.
	want := []Action{
		WriteFile{Path: "AGENTS.md", Content: "FOUNDATION-EXPORT\n\nMIDDLE-EXPORT\n\nLEAF-INTERNAL\n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
	// Belt-and-suspenders: the excluded ancestor internals are absent.
	body := got[0].(WriteFile).Content
	if strings.Contains(body, "FOUNDATION-INTERNAL") {
		t.Errorf("leaf composition leaked the foundation's INTERNAL prose:\n%s", body)
	}
}

func TestPlanLeafExportProseBeforeLeafInternal(t *testing.T) {
	// When the LEAF declares both an export and an internal prose fragment, the
	// leaf's export sits in the foundation-first export sweep and the leaf's
	// internal lands LAST (per the algebra: …∥ export(Lₙ) ∥ internal(Lₙ)).
	layers := []layer.Layer{
		{Name: "leaf", Path: "/l", ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Internal, Content: "LEAF-INTERNAL"},
			{Visibility: intent.Export, Content: "LEAF-EXPORT"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	// Export first (the foundation-first sweep), then the leaf's internal last —
	// regardless of declaration order in the manifest.
	want := []Action{WriteFile{Path: "AGENTS.md", Content: "LEAF-EXPORT\n\nLEAF-INTERNAL\n"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanAncestorInternalFileOpExcluded(t *testing.T) {
	// The export/leaf filter applies UNIFORMLY to every kind, not just prose: an
	// ANCESTOR's internal file-op (here a symlink) is excluded, while the same
	// kind exported by an ancestor, and an internal declared by the LEAF, are
	// included. (Today every real file-op is export, so this only guards the
	// filter's presence — behavior on the real manifest is unchanged.)
	layers := []layer.Layer{
		{Name: "ancestor", Path: "/a", Intents: []intent.Intent{
			{Kind: intent.Symlink, Visibility: intent.Export, Source: "shared.md", Target: "shared.md"},
			{Kind: intent.Symlink, Visibility: intent.Internal, Source: "secret.md", Target: "secret.md"},
		}},
		{Name: "leaf", Path: "/l", Intents: []intent.Intent{
			{Kind: intent.Symlink, Visibility: intent.Internal, Source: "leaf-only.md", Target: "leaf-only.md"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		Symlink{Src: "/a/shared.md", Dst: "shared.md"},       // ancestor export — included
		Symlink{Src: "/l/leaf-only.md", Dst: "leaf-only.md"}, // leaf internal — included
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v (ancestor INTERNAL symlink must be excluded)", got, want)
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

func TestPlanSeedLowering(t *testing.T) {
	// A `seed` intent lowers to a Seed{Src: upstream/Source, Dst: Target} — the
	// SAME joinPath the symlink case uses for the absolute upstream source. The
	// pure planner records only the path facts (it does NOT read the upstream
	// bytes — ARCH-PURE); applySeed reads Src + conditionally writes Dst.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/ws/ariadne", Intents: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Seed, Source: ".github/workflows/merge-check.yml", Target: ".github/workflows/merge-check.yml"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		Seed{Src: "/ws/ariadne/.github/workflows/merge-check.yml", Dst: ".github/workflows/merge-check.yml"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanDeferredKindsAreNoOps(t *testing.T) {
	// Skill lowering is still deferred (the M3 skill index feeds the menu, not a
	// file-op). It must not error or emit an Action here — just skip. (Merge now
	// lowers — see TestPlanMergeLowering.)
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
	// (optional) local off disk, runs settingsx.Merge, writes Target. The manifest
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
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "BASE PROSE"}}},
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
