package layergraph

import (
	"os"
	"path/filepath"
	"strings"
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

// The walk is tested through the production OSFS (the real-OS FS) rather than a
// duplicate test double — the canonicalization seam genuinely depends on the live
// filesystem, so an in-memory fake can't exercise it anyway, and using OSFS gives
// it direct coverage (ARCH-DRY: one FS impl).

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

	roots, err := Walk(OSFS{}, derived)
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

	roots, err := Walk(OSFS{}, derived)
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

	roots, err := Walk(OSFS{}, top)
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

	roots, err := Walk(OSFS{}, derived)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 2 || roots[0] != canon(t, base) {
		t.Fatalf("absolute substrate not resolved: %v", roots)
	}
}

func TestWalkPresentSubstrateMissingManifestErrors(t *testing.T) {
	// #155: a declared substrate whose target is PRESENT on disk but ships no
	// construct/base.manifest is a broken layer edge — Walk must fail LOUD (the
	// pre-#155 silent skip dropped the whole transitive chain under it with no
	// signal). The error must name the missing base.manifest.
	parent := t.TempDir()
	notALayer := filepath.Join(parent, "notalayer")
	derived := filepath.Join(parent, "d")
	writeFile(t, filepath.Join(notALayer, "README.md"), "present, but no base.manifest")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../notalayer\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	_, err := Walk(OSFS{}, derived)
	if err == nil {
		t.Fatal("Walk must error on a present substrate lacking base.manifest, got nil")
	}
	if !strings.Contains(err.Error(), "base.manifest") {
		t.Errorf("error must name the missing base.manifest, got: %v", err)
	}
}

func TestWalkAbsentSubstrateSilentlySkipped(t *testing.T) {
	// #155: the loud failure is scoped to a PRESENT-but-manifest-less substrate.
	// A substrate target that is simply not checked out (absent dir) keeps the
	// silent present-skip — a partial checkout must not hard-fail the walk.
	parent := t.TempDir()
	derived := filepath.Join(parent, "d")
	// ../absent is never created on disk.
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../absent\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(OSFS{}, derived)
	if err != nil {
		t.Fatalf("absent substrate must be silently skipped, got error: %v", err)
	}
	if len(roots) != 1 || roots[0] != canon(t, derived) {
		t.Fatalf("absent substrate not skipped cleanly: %v", roots)
	}
}

func TestWalkChainBrokenByManifestlessIntermediate(t *testing.T) {
	// #155 exact repro: kbench → kaggle → metis, where the MID layer (kaggle) is
	// present but lacks base.manifest. The old walk silently stopped at kaggle and
	// under-compiled kbench to a no-op; now the transitive walk errors when it
	// reaches the broken intermediate, naming it.
	parent := t.TempDir()
	base := filepath.Join(parent, "metis")
	mid := filepath.Join(parent, "kaggle")
	derived := filepath.Join(parent, "kbench")

	writeFile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	// mid is present (has construct/deps) but ships NO base.manifest.
	writeFile(t, filepath.Join(mid, "construct", "deps"), "substrate ../metis\n")
	writeFile(t, filepath.Join(derived, "construct", "deps"), "substrate ../kaggle\n")
	writeFile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	_, err := Walk(OSFS{}, derived)
	if err == nil {
		t.Fatal("Walk must error when a manifest-less intermediate breaks the chain, got nil")
	}
	if !strings.Contains(err.Error(), "kaggle") {
		t.Errorf("error must name the broken intermediate (kaggle), got: %v", err)
	}
}

func TestWalkSelfExclusion(t *testing.T) {
	// The root is never its own ancestor: a deps row pointing back at the root
	// is dropped (target-self-exclusion), leaving the root alone.
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	writeFile(t, filepath.Join(repo, "construct", "deps"), "substrate .\n")
	writeFile(t, filepath.Join(repo, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	roots, err := Walk(OSFS{}, repo)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(roots) != 1 || roots[0] != canon(t, repo) {
		t.Fatalf("self-reference not excluded: %v", roots)
	}
}
