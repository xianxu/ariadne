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
	// DeferredIntents collects the seed/merge/tool intents across layers,
	// de-duplicated by target (a verb appearing in multiple layers ledgers once),
	// and drops the non-deferred kinds (symlink/scaffold/prose/skill).
	layers := []layer.Layer{
		{Name: "base", Intents: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Symlink, Source: "Makefile", Target: "Makefile"},
			{Kind: intent.Merge, Source: "s.json", Target: ".claude/settings.json"},
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
		}},
		{Name: "self", Intents: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"}, // dup target
			{Kind: intent.Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		}},
	}
	got := DeferredIntents(layers)
	if len(got) != 3 {
		t.Fatalf("got %d deferred intents, want 3 (seed/merge/tool, dedup): %+v", len(got), got)
	}
	seen := map[string]bool{}
	for _, in := range got {
		seen[in.Target] = true
		if !IsDeferred(in.Kind) {
			t.Fatalf("non-deferred intent leaked: %+v", in)
		}
	}
	for _, want := range []string{"bootstrap.sh", ".claude/settings.json", "cmd/sdlc"} {
		if !seen[want] {
			t.Fatalf("missing deferred target %q", want)
		}
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
