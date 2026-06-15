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
// tests. Path is the "owner" for the tool check.
func fullLayer() layer.Layer {
	return layer.Layer{
		Name: "base", Path: "/ws/ariadne",
		Intents: []intent.Intent{
			{Kind: intent.Symlink, Source: "Makefile", Target: "Makefile"},
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Scaffold, Source: "atlas", Target: "atlas"},
			{Kind: intent.Touch, Source: "workshop/lessons.md", Target: "workshop/lessons.md"},
			{Kind: intent.Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
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
		plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"},
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

func TestCheckCompletenessSkillCoveredByMenuOnly(t *testing.T) {
	// The codex/agy target emits NO .claude/skills symlinks; the skill intent is
	// covered by the AGENTS.md `## Skills` menu instead. A plan with the menu in
	// AGENTS.md but zero skill symlinks must still report zero under-production.
	actions := fullActions()
	var codexPlan []plan.Action
	for _, a := range actions {
		if s, ok := a.(plan.Symlink); ok && strings.HasPrefix(s.Dst, ".claude/skills/") {
			continue // codex drops the symlink backend
		}
		if w, ok := a.(plan.WriteFile); ok && w.Path == "AGENTS.md" {
			a = plan.WriteFile{Path: "AGENTS.md", Content: "composed prose\n\n## Skills\n\n- xx-fix — fix"}
		}
		codexPlan = append(codexPlan, a)
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, codexPlan)
	if len(got) != 0 {
		t.Fatalf("menu-only (codex) plan reported %d uncovered: %+v", len(got), got)
	}
}

func TestCheckCompletenessSkillUncoveredWhenNeitherBackend(t *testing.T) {
	// Neither backend present: no .claude/skills symlinks AND an AGENTS.md with no
	// `## Skills` menu → the skill intent is under-produced. (An AGENTS.md write
	// alone, for prose, must NOT count as skill coverage.)
	actions := fullActions()
	var pruned []plan.Action
	for _, a := range actions {
		if s, ok := a.(plan.Symlink); ok && strings.HasPrefix(s.Dst, ".claude/skills/") {
			continue // drop the symlink backend
		}
		pruned = append(pruned, a) // AGENTS.md kept, but it has no `## Skills`
	}
	got := CheckCompleteness([]layer.Layer{fullLayer()}, pruned)
	if len(got) != 1 || got[0].Verb != "skill" {
		t.Fatalf("neither-backend: got %+v, want one uncovered skill", got)
	}
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
