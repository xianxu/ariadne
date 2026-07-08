package golden

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
)

// CheckCompleteness is the independent under-production guard: it enumerates the
// walked layers' manifest Intents (the setup.sh-equivalent managed set) and
// asserts weave's Plan covers each. These tests are PURE (synthetic layers +
// actions, no FS) — the purity proof matching the classifier's posture.

// fullLayer is a layer carrying one Intent of every verb, for the coverage
// tests. (The retired `tool` verb is gone — #95 M5.)
func fullLayer() layer.Layer {
	return layer.Layer{
		Name: "base", Path: "/ws/ariadne",
		Intents: []intent.Intent{
			{Kind: intent.Symlink, Source: "Makefile", Target: "Makefile"},
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Scaffold, Source: "atlas", Target: "atlas"},
			{Kind: intent.Touch, Source: "workshop/lessons.md", Target: "workshop/lessons.md"},
			{Kind: intent.Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
			{Kind: intent.Skill, Source: "construct/local", Target: "construct/local"},
		},
	}
}

// fullActions is a plan that covers every Intent in fullLayer.
func fullActions() []plan.Action {
	return []plan.Action{
		plan.WriteFile{Path: "AGENTS.md", Content: "composed prose"}, // prose body
		plan.Symlink{Src: "/ws/ariadne/Makefile", Dst: "Makefile"},
		plan.Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		plan.Mkdir{Path: "atlas"},
		plan.Touch{Path: "workshop/lessons.md"},
		plan.MergeSettings{Sources: []string{"/ws/ariadne/.claude/settings.ariadne.json"}, Target: ".claude/settings.json"},
		plan.Symlink{Src: "/ws/ariadne/construct/local/fix", Dst: ".claude/skills/xx-fix"}, // claude skill backend
	}
}

func TestCheckCompletenessZeroWhenPlanCoversAll(t *testing.T) {
	got := CheckCompleteness([]layer.Layer{fullLayer()}, fullActions())
	if len(got) != 0 {
		t.Fatalf("complete plan reported %d uncovered: %+v", len(got), got)
	}
}

func TestCheckCompletenessCatchesDroppedSeed(t *testing.T) {
	// THE bug this guard exists for: a `seed` lowering that emits NO Action (the
	// pre-M5 TODO). The seed Intent is in the manifest, but no plan.Seed covers
	// it → under-produced. (golden.go could not see this; it classifies only
	// planned actions.)
	actions := fullActions()
	// Drop the Seed action, simulating the no-op lowering.
	var pruned []plan.Action
	for _, a := range actions {
		if _, isSeed := a.(plan.Seed); isSeed {
			continue
		}
		pruned = append(pruned, a)
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, pruned)
	if len(got) != 1 || got[0].Verb != "seed" || got[0].Target != "bootstrap.sh" {
		t.Fatalf("dropped seed: got %+v, want one uncovered seed bootstrap.sh", got)
	}
}

func TestCheckCompletenessCatchesDroppedSymlinkAndMerge(t *testing.T) {
	// A plan missing a symlink AND a merge target → two under-produced rows.
	var pruned []plan.Action
	for _, a := range fullActions() {
		switch act := a.(type) {
		case plan.Symlink:
			if act.Dst == "Makefile" {
				continue // drop the Makefile symlink
			}
		case plan.MergeSettings:
			continue // drop the merge
		}
		pruned = append(pruned, a)
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, pruned)
	if len(got) != 2 {
		t.Fatalf("dropped symlink+merge: got %d uncovered, want 2: %+v", len(got), got)
	}
	// Sorted by verb: merge before symlink.
	if got[0].Verb != "merge" || got[1].Verb != "symlink" {
		t.Fatalf("uncovered verbs = [%s %s], want [merge symlink]", got[0].Verb, got[1].Verb)
	}
}

func TestCheckCompletenessCatchesDroppedMergeSource(t *testing.T) {
	layers := []layer.Layer{
		{Name: "base", Path: "/ws/base", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.base.json", Target: ".claude/settings.json"},
		}},
		{Name: "mid", Path: "/ws/mid", Intents: []intent.Intent{
			{Kind: intent.Merge, Source: ".claude/settings.mid.json", Target: ".claude/settings.json"},
		}},
	}
	actions := []plan.Action{
		plan.MergeSettings{
			Sources: []string{"/ws/base/.claude/settings.base.json"},
			Target:  ".claude/settings.json",
		},
	}
	got := CheckCompleteness(layers, actions)
	if len(got) != 1 || got[0].Verb != "merge" || got[0].Source != ".claude/settings.mid.json" {
		t.Fatalf("dropped merge source: got %+v, want one uncovered middle merge source", got)
	}
}

func TestCheckCompletenessSkillCoveredByAgentsSkills(t *testing.T) {
	// Option B: a codex/gemini target emits .agents/skills symlinks (NOT
	// .claude/skills). The skill intent is still covered — underSkills counts BOTH
	// per-harness dirs. No menu.
	actions := fullActions()
	var codexPlan []plan.Action
	for _, a := range actions {
		if s, ok := a.(plan.Symlink); ok && strings.HasPrefix(s.Dst, ".claude/skills/") {
			a = plan.Symlink{Src: s.Src, Dst: ".agents/skills/xx-fix"} // codex lowers here instead
		}
		codexPlan = append(codexPlan, a)
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, codexPlan)
	if len(got) != 0 {
		t.Fatalf("codex (.agents/skills) plan reported %d uncovered: %+v", len(got), got)
	}
}

func TestCheckCompletenessSkillUncoveredWhenNoSkillDir(t *testing.T) {
	// No skill-dir symlinks at all (.claude/skills AND .agents/skills) → the skill
	// intent is under-produced. (An entry-file write alone, for prose, must NOT count
	// as skill coverage.)
	actions := fullActions()
	var pruned []plan.Action
	for _, a := range actions {
		if s, ok := a.(plan.Symlink); ok && underAnySkillDir(s.Dst) {
			continue // drop ALL skill symlinks
		}
		pruned = append(pruned, a)
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, pruned)
	if len(got) != 1 || got[0].Verb != "skill" {
		t.Fatalf("no-skill-dir: got %+v, want one uncovered skill", got)
	}
}

func underAnySkillDir(dst string) bool {
	return strings.HasPrefix(dst, ".claude/skills/") || strings.HasPrefix(dst, ".agents/skills/")
}

func TestCheckCompletenessDedupsAcrossLayers(t *testing.T) {
	// The same verb+target declared in two layers (foundation + self) is one
	// managed path — it dedups to a single check, so a covered path reports zero
	// even when repeated.
	l := fullLayer()
	got := CheckCompleteness([]layer.Layer{l, l}, fullActions())
	if len(got) != 0 {
		t.Fatalf("dup layers over a complete plan reported %d uncovered: %+v", len(got), got)
	}
}

func TestCheckCompletenessExcludesAncestorInternal(t *testing.T) {
	// An ANCESTOR's `internal` intent is correctly EXCLUDED from 𝒜(R) by the
	// planner (it never reaches a derivative), so the completeness guard must NOT
	// flag it as under-produced — it was deliberately not planned, not dropped
	// (#99). Foundation declares an internal prose (AGENTS.local.md) the leaf does
	// not; the plan composes only the foundation's export + the leaf's internal.
	foundation := layer.Layer{
		Name: "foundation", Path: "/ws/ariadne",
		Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Export, Source: "AGENTS.base.md", Target: "AGENTS.base.md"},
			{Kind: intent.Prose, Visibility: intent.Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		},
	}
	leaf := layer.Layer{
		Name: "leaf", Path: "/ws/leaf",
		Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		},
	}
	// The plan: ONE AGENTS.md (foundation-export + leaf-internal). No Action is
	// keyed to the foundation's internal AGENTS.local.md — and that's correct.
	actions := []plan.Action{
		plan.WriteFile{Path: "AGENTS.md", Content: "FOUNDATION-EXPORT\n\nLEAF-INTERNAL\n"},
	}
	got := CheckCompleteness([]layer.Layer{foundation, leaf}, actions)
	if len(got) != 0 {
		t.Fatalf("ancestor-internal must NOT be flagged under-produced, got %+v", got)
	}
}

func TestCheckCompletenessLeafInternalStillChecked(t *testing.T) {
	// The LEAF's internal IS in 𝒜(R), so a plan that fails to compose any
	// AGENTS.md (the prose body) when the leaf declares an internal prose IS an
	// under-production gap — the filter must not silence the leaf's own internal.
	leaf := layer.Layer{
		Name: "leaf", Path: "/ws/leaf",
		Intents: []intent.Intent{
			{Kind: intent.Prose, Visibility: intent.Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		},
	}
	got := CheckCompleteness([]layer.Layer{leaf}, nil /* no AGENTS.md planned */)
	if len(got) != 1 || got[0].Verb != "prose" {
		t.Fatalf("leaf internal prose with no AGENTS.md must be under-produced, got %+v", got)
	}
}

func TestRenderCompletenessVerdict(t *testing.T) {
	clean := RenderCompleteness("/ws/ariadne", nil)
	if !strings.Contains(clean, "0 setup.sh-produced path(s) NOT planned") {
		t.Fatalf("clean verdict missing:\n%s", clean)
	}
	dirty := RenderCompleteness("/ws/nous", []Uncovered{
		{Verb: "seed", Source: "bootstrap.sh", Target: "bootstrap.sh", Reason: "no plan.Seed"},
	})
	if !strings.Contains(dirty, "UNDER-PRODUCED seed") || !strings.Contains(dirty, "1 setup.sh-produced path(s) NOT planned") {
		t.Fatalf("dirty verdict missing under-produced line:\n%s", dirty)
	}
}
