package layergraph

import (
	"os"
	"path/filepath"
	"testing"
)

// Walk follows each layer's construct/deps transitively from a repo root,
// resolves the DAG foundation-first (Resolve → root emitted last + self), and
// returns the ordered, deduped, absolute/canonicalized layer roots. It ports
// deps_substrate_targets' repo-root-relative + absolute + present-skip
// resolution and discover_ancestors' two _seen_or_add filters (a candidate is a
// layer ONLY if it ships construct/base.manifest; the root is never its own
// ancestor).
//
// The walk's physical-path canonicalization (filepath.EvalSymlinks ≈ pwd -P) is
// a real-disk concern — the present-skip of an unresolvable parent dir reads the
// live filesystem — so the test FS is a real-OS adapter over t.TempDir() (the
// same approach weave's walk_test uses), NOT a pure in-memory map. On macOS
// t.TempDir() hands back a /tmp/... logical path while /tmp → /private/tmp; Walk
// returns the physical /private/tmp/... form, so expectations are canon()-ed.

// testFS is the minimal FS the walk needs, backed by the real os package — the
// canonicalization seam genuinely depends on the live filesystem.
type testFS struct{}

func (testFS) ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func (testFS) ReadDir(path string) ([]os.DirEntry, error) { return os.ReadDir(path) }
func (testFS) Stat(path string) (os.FileInfo, error)      { return os.Stat(path) }

// canon canonicalizes a path to its physical form so expectations match the
// physical paths Walk returns.
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

func TestWalkFoundationFirst(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived := filepath.Join(parent, "derived")

	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	want := []string{canon(t, base), canon(t, derived)}
	if len(roots) != len(want) {
		t.Fatalf("Walk returned %d roots, want 2: %v", len(roots), roots)
	}
	for i := range want {
		if roots[i] != want[i] {
			t.Fatalf("roots[%d] = %q, want %q", i, roots[i], want[i])
		}
	}
}

func TestWalkTransitiveChainDepth3(t *testing.T) {
	// Depth-3 chain derived → mid → base: the BFS must enqueue the transitive
	// ancestor (mid → base), the chain the brain→nous→ariadne case realizes.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	mid := filepath.Join(parent, "mid")
	derived := filepath.Join(parent, "derived")

	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(mid, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(mid, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../mid\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	want := []string{canon(t, base), canon(t, mid), canon(t, derived)}
	if len(roots) != len(want) {
		t.Fatalf("Walk returned %d roots, want 3: %v", len(roots), roots)
	}
	for i := range want {
		if roots[i] != want[i] {
			t.Fatalf("roots[%d] = %q, want %q", i, roots[i], want[i])
		}
	}
}

func TestWalkDiamondAncestorAppliedOnce(t *testing.T) {
	// Diamond top → {left, right}; left → base; right → base. base is reached
	// via two paths but the BFS visited-set + Resolve dedup apply it exactly
	// once, before both left and right, with top (the root) last.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	left := filepath.Join(parent, "left")
	right := filepath.Join(parent, "right")
	top := filepath.Join(parent, "top")

	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(left, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(left, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(right, "construct", "deps"), "substrate ../base\n")
	writeFile(t, filepath.Join(right, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(top, "construct", "deps"), "substrate ../left\nsubstrate ../right\n")
	writeFile(t, filepath.Join(top, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, top)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 4 {
		t.Fatalf("Walk returned %d roots, want 4 unique (base once): %v", len(roots), roots)
	}
	pos := map[string]int{}
	for i, r := range roots {
		name := filepath.Base(r)
		if _, dup := pos[name]; dup {
			t.Fatalf("layer %q emitted twice: %v", name, roots)
		}
		pos[name] = i
		if r != canon(t, filepath.Join(parent, name)) {
			t.Fatalf("layer %q root = %q, want %q", name, r, canon(t, filepath.Join(parent, name)))
		}
	}
	if pos["base"] > pos["left"] || pos["base"] > pos["right"] {
		t.Fatalf("base not foundation-first (before left & right): %v", pos)
	}
	if pos["top"] != len(roots)-1 {
		t.Fatalf("top not last (root emitted last): %v", pos)
	}
}

func TestWalkAbsoluteSubstratePath(t *testing.T) {
	// deps_substrate_targets resolves an ABSOLUTE substrate target verbatim.
	parent := t.TempDir()
	base := filepath.Join(parent, "thebase")
	derived := filepath.Join(parent, "thederived")
	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate "+base+"\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 2 || roots[0] != canon(t, base) {
		t.Fatalf("absolute substrate not resolved: %v", roots)
	}
}

func TestWalkPresentSkipNonLayerDep(t *testing.T) {
	// _seen_or_add drops a substrate target that does NOT ship a
	// construct/base.manifest (it isn't a layer).
	parent := t.TempDir()
	notALayer := filepath.Join(parent, "notalayer")
	derived := filepath.Join(parent, "d")
	writeFile(t, filepath.Join(notALayer, "README.md"), "not a layer")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../notalayer\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 1 || roots[0] != canon(t, derived) {
		t.Fatalf("non-layer dep not skipped: %v", roots)
	}
}

func TestWalkSelfExclusion(t *testing.T) {
	// The root is never its own ancestor: a deps row pointing back at the root
	// is dropped (target-self-exclusion), leaving the root alone.
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	writeFile(t, filepath.Join(repo, "construct", "deps"), "substrate .\n")
	writeFile(t, filepath.Join(repo, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(testFS{}, repo)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 1 || roots[0] != canon(t, repo) {
		t.Fatalf("self-reference not excluded: %v", roots)
	}
}
