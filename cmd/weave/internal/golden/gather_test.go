package golden

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// The gatherer is the IO seam: it observes the live FS for each action/intent
// target and assembles the classifier Input. Tested against a real OSFS rooted
// at t.TempDir() (the seam exercised end-to-end, no mocks — same posture as
// plan.Apply's test).

func TestDeferredIntents(t *testing.T) {
	// DeferredIntents collects the seed intents across layers, de-duplicated by
	// target (a verb appearing in multiple layers ledgers once), and drops the
	// non-deferred kinds (symlink/scaffold/prose/skill — and now tool AND merge,
	// which lower to a ToolDep / MergeSettings and are classified directly).
	layers := []layer.Layer{
		{Name: "base", Intents: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Symlink, Source: "Makefile", Target: "Makefile"},
			{Kind: intent.Merge, Source: "s.json", Target: ".claude/settings.json"}, // now lowered, not deferred
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},             // now lowered, not deferred
		}},
		{Name: "self", Intents: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"}, // dup target
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}},
	}
	got := DeferredIntents(layers)
	if len(got) != 1 {
		t.Fatalf("got %d deferred intents, want 1 (seed, dedup; tool+merge no longer deferred): %+v", len(got), got)
	}
	for _, in := range got {
		if !IsDeferred(in.Kind) {
			t.Fatalf("non-deferred intent leaked: %+v", in)
		}
		if in.Kind == intent.Tool || in.Kind == intent.Merge {
			t.Fatalf("tool/merge must NOT be ledgered as deferred anymore: %+v", in)
		}
	}
	if got[0].Target != "bootstrap.sh" {
		t.Fatalf("deferred target = %q, want bootstrap.sh", got[0].Target)
	}
}

func TestGatherObservesMergeSettingsTriple(t *testing.T) {
	// A MergeSettings action makes the gatherer observe THREE files WITH content:
	// the base (Source), the live target (Target), and the sibling
	// settings.local.json — exactly the probe classifyMergeSettings reads.
	root := t.TempDir()
	claude := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claude, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claude, "settings.ariadne.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claude, "settings.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claude, "settings.local.json"), []byte(`{"b":2}`), 0o644); err != nil {
		t.Fatal(err)
	}

	actions := []plan.Action{
		plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
	}
	in := Gather(weavefs.OSFS{}, root, actions, nil)

	for rel, wantContent := range map[string]string{
		".claude/settings.ariadne.json": `{"a":1}`,
		".claude/settings.json":         `{"a":1}`,
		".claude/settings.local.json":   `{"b":2}`,
	} {
		o := in.Observed[filepath.Join(root, rel)]
		if !o.Exists || o.Content != wantContent {
			t.Fatalf("%s observed = %+v, want content %q", rel, o, wantContent)
		}
	}
}

func TestGatherMergeFollowsSymlinkedBase(t *testing.T) {
	// The derivative case: .claude/settings.ariadne.json is a SYMLINK to the
	// upstream base. The merge probe must read its RESOLVED content (follow the
	// link), like merge-settings.sh — otherwise the base reads empty and the merge
	// spuriously fails to parse. And a Symlink action probing the SAME path must
	// still see it as a symlink (the two observations coexist, regardless of the
	// order they run in — symlink-before-merge here, matching the manifest).
	root := t.TempDir()
	claude := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claude, 0o755); err != nil {
		t.Fatal(err)
	}
	// Upstream real base, then a symlink to it inside the derivative's .claude.
	upstream := filepath.Join(t.TempDir(), "settings.ariadne.json")
	if err := os.WriteFile(upstream, []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	linkRel, _ := filepath.Rel(claude, upstream)
	if err := os.Symlink(linkRel, filepath.Join(claude, "settings.ariadne.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claude, "settings.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Symlink action FIRST (manifest order), then the MergeSettings action — both
	// probe .claude/settings.ariadne.json.
	actions := []plan.Action{
		plan.Symlink{Src: upstream, Dst: ".claude/settings.ariadne.json"},
		plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
	}
	in := Gather(weavefs.OSFS{}, root, actions, nil)

	base := in.Observed[filepath.Join(claude, "settings.ariadne.json")]
	if !base.Exists {
		t.Fatalf("base settings.ariadne.json observed absent: %+v", base)
	}
	if !base.IsSymlink {
		t.Errorf("base should still be seen as a symlink (for the Symlink action): %+v", base)
	}
	if base.Content != `{"a":1}` {
		t.Errorf("merge probe should read RESOLVED content through the symlink, got %q", base.Content)
	}

	// And the whole thing classifies clean (symlink MATCH + merge MATCH).
	divs := Classify(in)
	if HasUnexpected(divs) {
		t.Fatalf("symlinked-base merge classified UNEXPECTED: %+v", divs)
	}
}

func TestGatherObservesLiveState(t *testing.T) {
	root := t.TempDir()
	// Lay down a live tree: a correct symlink, a scaffold dir, a seed file.
	upstream := filepath.Join(t.TempDir(), "up")
	if err := os.MkdirAll(upstream, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(upstream, "Makefile"), []byte("M"), 0o644); err != nil {
		t.Fatal(err)
	}
	// symlink Makefile -> rel(root, upstream/Makefile)
	rel, _ := filepath.Rel(root, filepath.Join(upstream, "Makefile"))
	if err := os.Symlink(rel, filepath.Join(root, "Makefile")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bootstrap.sh"), []byte("boot"), 0o755); err != nil {
		t.Fatal(err)
	}

	actions := []plan.Action{
		plan.Symlink{Src: filepath.Join(upstream, "Makefile"), Dst: "Makefile"},
		plan.Mkdir{Path: ".claude/skills"},
	}
	deferred := []intent.Intent{
		{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
	}

	in := Gather(weavefs.OSFS{}, root, actions, deferred)

	// The symlink target was observed with its real link.
	mk := in.Observed[filepath.Join(root, "Makefile")]
	if !mk.Exists || !mk.IsSymlink || mk.LinkTarget != rel {
		t.Fatalf("Makefile observed = %+v, want symlink -> %s", mk, rel)
	}
	// The scaffold dir.
	sk := in.Observed[filepath.Join(root, ".claude", "skills")]
	if !sk.Exists || !sk.IsDir {
		t.Fatalf(".claude/skills observed = %+v, want dir", sk)
	}
	// The seed file (deferred intent target).
	bs := in.Observed[filepath.Join(root, "bootstrap.sh")]
	if !bs.Exists {
		t.Fatalf("bootstrap.sh observed = %+v, want present", bs)
	}

	// And the whole thing classifies clean (all MATCH or EXPECTED).
	divs := Classify(in)
	if HasUnexpected(divs) {
		t.Fatalf("unexpected divergence in a clean live tree: %+v", divs)
	}
}
