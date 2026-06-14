package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Walk is the IO seam that turns a repo root into a foundation-first
// []layer.Layer: it follows each layer's construct/deps (porting
// deps_substrate_targets' repo-root-relative + absolute + present-skip
// resolution and the two _seen_or_add filters), resolves the DAG
// (layer.Resolve → root last/self-included), then loads each layer's
// base.manifest into Intents and its prose files into ProseFragments, applying
// the self-reference filter. Tested against a real on-disk fixture under
// t.TempDir() (the seam exercised end-to-end, no mocks).

// fixture builds a base layer + a derived repo as siblings under a temp parent
// and returns (parent, baseDir, derivedDir).
func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived := filepath.Join(parent, "derived")

	// --- base layer ---
	writeFile(t, filepath.Join(base, "construct", "base.manifest"),
		"prose AGENTS.local.md\nscaffold .claude/skills\nsymlink shared.md\n")
	writeFile(t, filepath.Join(base, "AGENTS.local.md"), "BASE PROSE")
	writeFile(t, filepath.Join(base, "shared.md"), "SHARED")

	// --- derived repo: depends on base via construct/deps (relative path) ---
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "AGENTS.local.md"), "DERIVED PROSE")

	return parent, canon(t, base), canon(t, derived)
}

// canon canonicalizes a path to its physical form (EvalSymlinks ≈ pwd -P), so
// expectations match the physical paths Walk returns. On macOS t.TempDir()
// hands back a /tmp/... logical path while /tmp → /private/tmp; Walk (correctly,
// porting setup.sh's pwd -P) returns the physical /private/tmp/... form.
func canon(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWalkFoundationFirstWithProse(t *testing.T) {
	_, base, derived := fixture(t)

	layers, err := Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	// Foundation-first: base before derived; derived (root) last + self.
	if len(layers) != 2 {
		t.Fatalf("Walk returned %d layers, want 2: %+v", len(layers), layers)
	}
	if layers[0].Path != base {
		t.Fatalf("layers[0].Path = %q, want base %q", layers[0].Path, base)
	}
	if layers[1].Path != derived {
		t.Fatalf("layers[1].Path = %q, want derived (root last) %q", layers[1].Path, derived)
	}

	// Prose fragments loaded from each layer's prose file.
	if got := layers[0].ProseFragments; len(got) != 1 || got[0] != "BASE PROSE" {
		t.Fatalf("base ProseFragments = %v, want [BASE PROSE]", got)
	}
	if got := layers[1].ProseFragments; len(got) != 1 || got[0] != "DERIVED PROSE" {
		t.Fatalf("derived ProseFragments = %v, want [DERIVED PROSE]", got)
	}

	// base's intents loaded (prose + scaffold + symlink).
	wantKinds := []intent.Kind{intent.Prose, intent.Scaffold, intent.Symlink}
	if len(layers[0].Intents) != len(wantKinds) {
		t.Fatalf("base Intents = %+v, want %d", layers[0].Intents, len(wantKinds))
	}
	for i, k := range wantKinds {
		if layers[0].Intents[i].Kind != k {
			t.Fatalf("base Intents[%d].Kind = %v, want %v", i, layers[0].Intents[i].Kind, k)
		}
	}
}

func TestWalkTransitiveChainDepth3(t *testing.T) {
	// Depth-3 chain derived → mid → base: discoverEdges' BFS must enqueue the
	// transitive ancestor (mid → base), the chain the motivating
	// brain→nous→ariadne case realizes. Walk returns them foundation-first.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	mid := filepath.Join(parent, "mid")
	derived := filepath.Join(parent, "derived")

	// base: a real layer (ships construct/base.manifest), no deps.
	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(base, "AGENTS.local.md"), "BASE")
	// mid: depends on base, also a real layer.
	writeFile(t, filepath.Join(mid, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(mid, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(mid, "AGENTS.local.md"), "MID")
	// derived: depends on mid (which transitively depends on base).
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../mid\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "AGENTS.local.md"), "DERIVED")

	layers, err := Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	wantPaths := []string{canon(t, base), canon(t, mid), canon(t, derived)}
	wantNames := []string{"base", "mid", "derived"}
	if len(layers) != len(wantPaths) {
		t.Fatalf("Walk returned %d layers, want 3 (base, mid, derived): %+v", len(layers), layers)
	}
	for i := range wantPaths {
		if layers[i].Path != wantPaths[i] {
			t.Fatalf("layers[%d].Path = %q, want %q", i, layers[i].Path, wantPaths[i])
		}
		if layers[i].Name != wantNames[i] {
			t.Fatalf("layers[%d].Name = %q, want %q", i, layers[i].Name, wantNames[i])
		}
	}
}

func TestWalkDiamondAncestorAppliedOnce(t *testing.T) {
	// Diamond top → {left, right}; left → base; right → base. base is reached
	// via two paths but the BFS visited-set + Resolve dedup must apply it
	// exactly once, before both left and right, with top (the root) last.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	left := filepath.Join(parent, "left")
	right := filepath.Join(parent, "right")
	top := filepath.Join(parent, "top")

	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(base, "AGENTS.local.md"), "BASE")
	writeFile(t, filepath.Join(left, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(left, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(left, "AGENTS.local.md"), "LEFT")
	writeFile(t, filepath.Join(right, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(right, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(right, "AGENTS.local.md"), "RIGHT")
	writeFile(t, filepath.Join(top, "construct", "deps"), "substrate ../left\nsubstrate ../right\n")
	writeFile(t, filepath.Join(top, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(top, "AGENTS.local.md"), "TOP")

	layers, err := Walk(weavefs.OSFS{}, top)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	// Four unique layers (base applied once, not twice).
	if len(layers) != 4 {
		t.Fatalf("Walk returned %d layers, want 4 unique (base once): %+v", len(layers), layers)
	}
	pos := map[string]int{}
	for i, l := range layers {
		if _, dup := pos[l.Name]; dup {
			t.Fatalf("layer %q emitted twice: %+v", l.Name, layers)
		}
		pos[l.Name] = i
		if l.Path != canon(t, filepath.Join(parent, l.Name)) {
			t.Fatalf("layer %q Path = %q, want %q", l.Name, l.Path, canon(t, filepath.Join(parent, l.Name)))
		}
	}
	// base before both left and right; top (root) last.
	if pos["base"] > pos["left"] || pos["base"] > pos["right"] {
		t.Fatalf("base not foundation-first (before left & right): %v", pos)
	}
	if pos["top"] != len(layers)-1 {
		t.Fatalf("top not last (root emitted last): %v", pos)
	}
}

func TestWalkSkipsSelfReference(t *testing.T) {
	// A layer entry whose layerDir/Source == repoRoot/Target is a
	// self-reference (would symlink a file onto itself); the walk drops it
	// (ports walk_manifest:315). Build a single-layer repo (no deps) whose own
	// manifest declares `symlink AGENTS.md` — on a self-walk that resolves to
	// repoRoot/AGENTS.md == repoRoot/AGENTS.md, so it must be skipped, while a
	// non-self entry (`symlink Makefile.base Makefile`) survives... actually we
	// assert the self entry is gone and a prose entry stays.
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	writeFile(t, filepath.Join(repo, "construct", "base.manifest"),
		"symlink AGENTS.md\nprose AGENTS.local.md\n")
	writeFile(t, filepath.Join(repo, "AGENTS.md"), "SELF")
	writeFile(t, filepath.Join(repo, "AGENTS.local.md"), "PROSE")

	layers, err := Walk(weavefs.OSFS{}, repo)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(layers) != 1 {
		t.Fatalf("want 1 layer (self only), got %d", len(layers))
	}
	// The `symlink AGENTS.md` self-reference is filtered; only `prose` remains.
	for _, in := range layers[0].Intents {
		if in.Kind == intent.Symlink {
			t.Fatalf("self-reference symlink AGENTS.md was not filtered: %+v", layers[0].Intents)
		}
	}
	if len(layers[0].Intents) != 1 || layers[0].Intents[0].Kind != intent.Prose {
		t.Fatalf("want only [Prose] after self-filter, got %+v", layers[0].Intents)
	}
}

func TestWalkAbsoluteSubstratePath(t *testing.T) {
	// deps_substrate_targets resolves an ABSOLUTE substrate target verbatim
	// (raw="$target" when target starts with /). The derived repo points at an
	// absolute base path.
	parent := t.TempDir()
	base := filepath.Join(parent, "thebase")
	derived := filepath.Join(parent, "thederived")
	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(base, "AGENTS.local.md"), "B")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate "+base+"\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "AGENTS.local.md"), "D")

	layers, err := Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(layers) != 2 || layers[0].Path != canon(t, base) {
		t.Fatalf("absolute substrate not resolved: %+v", layers)
	}
}

func TestWalkPresentSkipNonLayerDep(t *testing.T) {
	// _seen_or_add drops a substrate target that does NOT ship a
	// construct/base.manifest (it isn't a layer). The derived repo names a peer
	// dir with no manifest; it must be filtered, leaving derived alone.
	parent := t.TempDir()
	notALayer := filepath.Join(parent, "notalayer")
	derived := filepath.Join(parent, "d")
	writeFile(t, filepath.Join(notALayer, "README.md"), "not a layer")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../notalayer\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "AGENTS.local.md"), "D")

	layers, err := Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(layers) != 1 || layers[0].Path != canon(t, derived) {
		t.Fatalf("non-layer dep not skipped: %+v", layers)
	}
}
