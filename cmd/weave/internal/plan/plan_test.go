package plan

import (
	"reflect"
	"strings"
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
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "BASE"}}},
		{Name: "nous", Path: "/n", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "LAYER"}}},
		{Name: "brain", Path: "/b", Intents: []intent.Intent{
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}, ProseFragments: []layer.ProseFragment{{Visibility: intent.Export, Content: "LOCAL"}}},
	}
	got, err := Plan(layers, []string{"AGENTS.md"}) // prose fanned to AGENTS.md (one entry file here)
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
	got, err := Plan(layers, []string{"AGENTS.md"})
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
	got, err := Plan(layers, []string{"AGENTS.md"})
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
	got, err := Plan(layers, []string{"AGENTS.md"})
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
	got, err := Plan(layers, []string{"AGENTS.md"})
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
	got, err := Plan(layers, []string{"AGENTS.md"})
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
	got, err := Plan(layers, []string{"AGENTS.md"})
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Plan = %#v, want no Actions (deferred kinds skipped)", got)
	}
}

func TestPlanMergeLowering(t *testing.T) {
	// A `merge` intent lowers to a MergeSettings{Sources, Target} — the settings
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
	got, err := Plan(layers, []string{"AGENTS.md"})
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		MergeSettings{Sources: []string{"/ws/ariadne/.claude/settings.ariadne.json"}, Target: ".claude/settings.json"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanGroupsMergeRowsByTargetFoundationFirst(t *testing.T) {
	layers := []layer.Layer{
		{Name: "base", Path: "/ws/base", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.base.json", Target: ".claude/settings.json"},
		}},
		{Name: "mid", Path: "/ws/mid", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.mid.json", Target: ".claude/settings.json"},
			{Kind: intent.Merge, Source: ".gemini/settings.mid.json", Target: ".gemini/settings.json"},
		}},
		{Name: "leaf", Path: "/ws/leaf", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.leaf.json", Target: ".claude/settings.json"},
		}},
	}
	got, err := Plan(layers, []string{"AGENTS.md"})
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		MergeSettings{
			Sources: []string{
				"/ws/base/.claude/settings.base.json",
				"/ws/mid/.claude/settings.mid.json",
				"/ws/leaf/.claude/settings.leaf.json",
			},
			Target: ".claude/settings.json",
		},
		MergeSettings{
			Sources: []string{"/ws/mid/.gemini/settings.mid.json"},
			Target:  ".gemini/settings.json",
		},
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
	got, err := Plan(layers, []string{"AGENTS.md"})
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{Symlink{Src: "/up/CLAUDE.md", Dst: "CLAUDE.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v", got, want)
	}
}

func TestPlanProseFannedToEntryFiles(t *testing.T) {
	// Option B (#107): the ONE composed prose is written to EACH per-harness ENTRY
	// FILE (the Union: CLAUDE.md + AGENTS.md + GEMINI.md), byte-identical. There is
	// NO `## Skills` menu — skills lower to per-harness skill DIRS (in planActions),
	// never into the prose.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/a", ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Export, Content: "BASE PROSE"},
		}},
	}
	got, err := Plan(layers, []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"})
	if err != nil {
		t.Fatalf("Plan: unexpected error: %v", err)
	}
	want := []Action{
		WriteFile{Path: "CLAUDE.md", Content: "BASE PROSE\n"},
		WriteFile{Path: "AGENTS.md", Content: "BASE PROSE\n"},
		WriteFile{Path: "GEMINI.md", Content: "BASE PROSE\n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Plan = %#v, want %#v (one prose fanned to each entry file, no menu)", got, want)
	}
	// No menu leaked into any entry file.
	for _, a := range got {
		if strings.Contains(a.(WriteFile).Content, "## Skills") {
			t.Errorf("entry file leaked a `## Skills` menu:\n%s", a.(WriteFile).Content)
		}
	}
}

func TestPlanNoEntryFilesNoProseAction(t *testing.T) {
	// A lean target writes prose to ONLY its own entry file; passing no entry files
	// (or none selected) writes none — even with prose present.
	layers := []layer.Layer{
		{Name: "ariadne", Path: "/a", ProseFragments: []layer.ProseFragment{
			{Visibility: intent.Export, Content: "PROSE"},
		}},
	}
	got, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Plan with no entry files = %#v, want no Actions", got)
	}
}
