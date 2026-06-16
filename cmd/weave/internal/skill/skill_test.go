package skill

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
)

func names(es []Entry) []string {
	out := make([]string, 0, len(es))
	for _, e := range es {
		out = append(out, e.Name)
	}
	return out
}

func TestEntryCarriesVisibilityAndLayer(t *testing.T) {
	e := Entry{Name: "xx-fix", Description: "d", BodyPath: "/a/fix/SKILL.md",
		Visibility: intent.Internal, LayerIndex: 2}
	if e.Visibility != intent.Internal || e.LayerIndex != 2 {
		t.Fatalf("entry did not carry visibility/layer: %+v", e)
	}
}

func TestSelectVisible(t *testing.T) {
	// 𝒜(R): every layer's exports + the leaf's internals. The ONLY exclusion is an
	// ancestor's internal. Foundation-first order preserved.
	leaf := 2
	in := []Entry{
		{Name: "base-export", Visibility: intent.Export, LayerIndex: 0},
		{Name: "ancestor-internal", Visibility: intent.Internal, LayerIndex: 0}, // DROP
		{Name: "leaf-internal", Visibility: intent.Internal, LayerIndex: 2},     // KEEP
		{Name: "leaf-export", Visibility: intent.Export, LayerIndex: 2},
	}
	got := names(SelectVisible(in, leaf))
	want := []string{"base-export", "leaf-internal", "leaf-export"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectVisible = %v, want %v", got, want)
	}
}

// SkillIndex is PURE: it takes already-parsed entries (gathered foundation-first
// by the IO seam) and produces an ordered, collision-free menu + a name→body
// lookup. These tests run with no IO — that's the purity proof.

func TestBuild_OrderIsFoundationFirstByAppearance(t *testing.T) {
	// Entries arrive foundation-first (the walk emits base layer's skills before
	// the consuming repo's). The menu preserves first-appearance order.
	entries := []Entry{
		{Name: "superpowers-brainstorming", Description: "brainstorm", BodyPath: "/base/construct/adapted/superpowers-brainstorming/SKILL.md"},
		{Name: "xx-sdlc", Description: "sdlc gates", BodyPath: "/base/construct/local/sdlc/SKILL.md"},
		{Name: "xx-issues", Description: "issue files", BodyPath: "/repo/construct/local/issues/SKILL.md"},
	}
	idx := Build(entries)

	wantMenu := []MenuItem{
		{Name: "superpowers-brainstorming", Description: "brainstorm"},
		{Name: "xx-sdlc", Description: "sdlc gates"},
		{Name: "xx-issues", Description: "issue files"},
	}
	if !reflect.DeepEqual(idx.Menu(), wantMenu) {
		t.Errorf("menu mismatch\n got: %#v\nwant: %#v", idx.Menu(), wantMenu)
	}
}

func TestBuild_DownstreamOverridesFoundation(t *testing.T) {
	// A downstream layer (later in the slice) re-declaring a name overrides the
	// foundation's entry — the cascade. The menu keeps the name in its ORIGINAL
	// position (first appearance) but with the overriding description + body, so
	// a derivative can customize a base skill in place without reordering.
	entries := []Entry{
		{Name: "xx-sdlc", Description: "base sdlc", BodyPath: "/base/construct/local/sdlc/SKILL.md"},
		{Name: "superpowers-brainstorming", Description: "brainstorm", BodyPath: "/base/construct/adapted/superpowers-brainstorming/SKILL.md"},
		{Name: "xx-sdlc", Description: "repo sdlc", BodyPath: "/repo/construct/local/sdlc/SKILL.md"},
	}
	idx := Build(entries)

	wantMenu := []MenuItem{
		{Name: "xx-sdlc", Description: "repo sdlc"}, // overridden description, original slot
		{Name: "superpowers-brainstorming", Description: "brainstorm"},
	}
	if !reflect.DeepEqual(idx.Menu(), wantMenu) {
		t.Errorf("menu mismatch\n got: %#v\nwant: %#v", idx.Menu(), wantMenu)
	}
	// The body lookup resolves to the overriding (downstream) body.
	got, ok := idx.BodyPath("xx-sdlc")
	if !ok {
		t.Fatal("xx-sdlc not found in lookup")
	}
	if got != "/repo/construct/local/sdlc/SKILL.md" {
		t.Errorf("BodyPath(xx-sdlc) = %q, want the downstream body", got)
	}
}

func TestBodyPath_Lookup(t *testing.T) {
	idx := Build([]Entry{
		{Name: "xx-fix", Description: "fix markers", BodyPath: "/base/construct/local/fix/SKILL.md"},
	})
	got, ok := idx.BodyPath("xx-fix")
	if !ok || got != "/base/construct/local/fix/SKILL.md" {
		t.Errorf("BodyPath(xx-fix) = (%q, %v), want the fix body", got, ok)
	}
	if _, ok := idx.BodyPath("does-not-exist"); ok {
		t.Error("BodyPath(does-not-exist) should report not-found")
	}
}

func TestBuild_Empty(t *testing.T) {
	idx := Build(nil)
	if len(idx.Menu()) != 0 {
		t.Errorf("empty index menu = %#v, want empty", idx.Menu())
	}
	if _, ok := idx.BodyPath("anything"); ok {
		t.Error("empty index should find nothing")
	}
}
