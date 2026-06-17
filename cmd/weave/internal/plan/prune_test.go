package plan

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// prune_test.go is where the #96 SAFETY lives: a managed-location fixture
// exercising every KEEP/PRUNE class against a real t.TempDir-rooted OSFS (no
// mocks — the scan + unlink seam end-to-end). The decision (shouldPrune /
// PrunePlan) is pure and tested through the IO seam (PruneOrphans), so a
// regression that would delete a repo's own content fails here, not in prod.

// pruneFixture builds, under a fresh temp repoRoot, a substrate dir + a managed
// lowered location (.claude/skills) populated with one of each class:
//
//	(a) produced symlink            → KEEP (it's in the produced set)
//	(b) orphaned weave-symlink      → PRUNE (old prefix, target under substrate)
//	(c) dangling weave-symlink      → PRUNE (target deleted, e.g. setup.sh)
//	(d) repo's own REAL dir         → KEEP
//	(e) repo's own REAL file        → KEEP
//	(f) non-weave symlink           → KEEP (points somewhere unrelated)
//
// It returns the repoRoot, the substrate root (the lowering source root), and
// the produced []Action (one Symlink for (a) plus the substrate links it stands
// in for). The substrate lives OUTSIDE the repo (a sibling), exactly like a real
// ../ariadne dependency.
func pruneFixture(t *testing.T) (repoRoot, substrate string, actions []Action) {
	t.Helper()
	parent := t.TempDir()
	repoRoot = filepath.Join(parent, "repo")
	substrate = filepath.Join(parent, "substrate")

	// Substrate skill dirs the live links point at.
	for _, d := range []string{"construct/local/keep", "construct/local/orphan-target", "construct/adapted/unrelated-real"} {
		if err := os.MkdirAll(filepath.Join(substrate, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	skillsDir := filepath.Join(repoRoot, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	mklink := func(name, target string) {
		if err := os.Symlink(target, filepath.Join(skillsDir, name)); err != nil {
			t.Fatalf("symlink %s: %v", name, err)
		}
	}

	// (a) produced symlink → KEEP. Its target is a live substrate skill dir, and
	// it is in the produced action set below.
	mklink("xx-keep", "../../../substrate/construct/local/keep")
	// (b) orphaned weave-symlink (old prefix) → PRUNE. Target still EXISTS under
	// the substrate (a re-prefix leaves the old name pointing at a live dir), but
	// weave did NOT produce it this run.
	mklink("yy-orphan", "../../../substrate/construct/local/orphan-target")
	// (c) dangling weave-symlink → PRUNE. Models the dead cutover symlink: target
	// under the substrate but DELETED (setup.sh removed from ariadne).
	mklink("xx-dead", "../../../substrate/construct/scripts/setup.sh")
	// (f) non-weave symlink → KEEP. Points somewhere UNRELATED (outside any source
	// root) — a link the repo authored for its own reasons.
	mklink("vendor-thing", filepath.Join(parent, "elsewhere", "thing"))

	// (d) repo's own REAL dir → KEEP. A real directory sitting in the managed
	// location (e.g. a hand-authored skill dir).
	if err := os.MkdirAll(filepath.Join(skillsDir, "my-real-skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "my-real-skill", "SKILL.md"), []byte("real"), 0o644); err != nil {
		t.Fatal(err)
	}
	// (e) repo's own REAL file → KEEP.
	if err := os.WriteFile(filepath.Join(skillsDir, "README.md"), []byte("authored"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The produced action set: weave produced (a) this run (Src = the absolute
	// substrate dir, matching skillSymlinks' lowering). That makes .claude/skills
	// a managed location AND xx-keep a produced symlink.
	actions = []Action{
		Symlink{Src: filepath.Join(substrate, "construct/local/keep"), Dst: ".claude/skills/xx-keep"},
	}
	return repoRoot, substrate, actions
}

func TestPruneOrphansRemovesExactlyOrphanAndDangling(t *testing.T) {
	repoRoot, substrate, actions := pruneFixture(t)
	sourceRoots := SourceRootsFromPaths([]string{substrate})

	pruned, err := PruneOrphans(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots)
	if err != nil {
		t.Fatalf("PruneOrphans: %v", err)
	}

	want := []string{
		filepath.Join(".claude/skills", "xx-dead"),   // (c) dangling
		filepath.Join(".claude/skills", "yy-orphan"), // (b) orphan prefix
	}
	sort.Strings(want)
	if !reflect.DeepEqual(pruned, want) {
		t.Fatalf("pruned = %v, want %v", pruned, want)
	}

	// KEEP assertions — the safety contract.
	skills := filepath.Join(repoRoot, ".claude", "skills")
	assertExists := func(name string) {
		t.Helper()
		if _, err := os.Lstat(filepath.Join(skills, name)); err != nil {
			t.Errorf("expected %s to be KEPT, but it's gone: %v", name, err)
		}
	}
	assertGone := func(name string) {
		t.Helper()
		if _, err := os.Lstat(filepath.Join(skills, name)); err == nil {
			t.Errorf("expected %s to be PRUNED, but it still exists", name)
		}
	}
	assertExists("xx-keep")       // (a) produced
	assertExists("my-real-skill") // (d) real dir
	assertExists("README.md")     // (e) real file
	assertExists("vendor-thing")  // (f) non-weave symlink
	assertGone("yy-orphan")       // (b)
	assertGone("xx-dead")         // (c)
}

func TestPruneOrphansIdempotent(t *testing.T) {
	repoRoot, substrate, actions := pruneFixture(t)
	sourceRoots := SourceRootsFromPaths([]string{substrate})

	if _, err := PruneOrphans(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots); err != nil {
		t.Fatalf("first PruneOrphans: %v", err)
	}
	// Second run: the orphans are gone, the produced + real + non-weave entries
	// remain — nothing left to prune.
	pruned, err := PruneOrphans(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots)
	if err != nil {
		t.Fatalf("second PruneOrphans: %v", err)
	}
	if len(pruned) != 0 {
		t.Fatalf("second run pruned %v, want nothing (idempotent)", pruned)
	}
}

func TestPrunePreviewMatchesApplyButMutatesNothing(t *testing.T) {
	repoRoot, substrate, actions := pruneFixture(t)
	sourceRoots := SourceRootsFromPaths([]string{substrate})

	preview, err := PrunePreview(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots)
	if err != nil {
		t.Fatalf("PrunePreview: %v", err)
	}
	// The orphans must STILL be on disk after a preview (read-only).
	skills := filepath.Join(repoRoot, ".claude", "skills")
	for _, name := range []string{"yy-orphan", "xx-dead"} {
		if _, err := os.Lstat(filepath.Join(skills, name)); err != nil {
			t.Errorf("preview removed %s, but it must be read-only", name)
		}
	}
	// And the preview list must equal what an apply would prune.
	pruned, err := PruneOrphans(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots)
	if err != nil {
		t.Fatalf("PruneOrphans: %v", err)
	}
	if !reflect.DeepEqual(preview, pruned) {
		t.Fatalf("preview %v != apply %v", preview, pruned)
	}
}

func TestPruneKeepsWriteFileTargetStillSymlinked(t *testing.T) {
	// The #95 cutover edge: before Apply rewrites it, a derivative's AGENTS.md is
	// still a SYMLINK into the ancestor (the source root) — so the prune scan sees
	// an orphan-shaped symlink at a slot weave actually PRODUCES this run (as a
	// WriteFile, not a Symlink). The broadened ProducedPathSet must exclude it, so
	// a dry-run preview does NOT falsely list `prune AGENTS.md`. (A real apply is
	// safe regardless — Apply converts it to a regular file before the prune — but
	// the preview must be honest.)
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "repo")
	substrate := filepath.Join(parent, "substrate")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(substrate, 0o755); err != nil {
		t.Fatal(err)
	}
	// The ancestor's AGENTS.md (the symlink's target, under the source root).
	if err := os.WriteFile(filepath.Join(substrate, "AGENTS.md"), []byte("ANCESTOR"), 0o644); err != nil {
		t.Fatal(err)
	}
	// repo/AGENTS.md → ../substrate/AGENTS.md (the pre-cutover shape).
	if err := os.Symlink("../substrate/AGENTS.md", filepath.Join(repoRoot, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	// weave produces AGENTS.md as a WriteFile this run, plus a sibling symlink at
	// the repo root so "." is a managed location (the prune scans it).
	actions := []Action{
		WriteFile{Path: "AGENTS.md", Content: "COMPOSED"},
		Symlink{Src: filepath.Join(substrate, "CLAUDE.md"), Dst: "CLAUDE.md"},
	}
	sourceRoots := SourceRootsFromPaths([]string{substrate})

	preview, err := PrunePreview(weavefs.OSFS{}, repoRoot, actions, actions, sourceRoots)
	if err != nil {
		t.Fatalf("PrunePreview: %v", err)
	}
	for _, p := range preview {
		if p == "AGENTS.md" {
			t.Fatalf("preview falsely prunes AGENTS.md (a WriteFile target): %v", preview)
		}
	}
}

func TestManagedLocationsOnlyWhereWeaveProducedSymlinks(t *testing.T) {
	// A dir weave does NOT emit a symlink into is NOT managed (so a self-walk
	// that owns construct/scripts/ as real files never has it scanned).
	actions := []Action{
		Symlink{Src: "/sub/a", Dst: ".claude/skills/xx-a"},
		Symlink{Src: "/sub/b", Dst: "construct/local"},
		Mkdir{Path: "workshop/issues"}, // not a symlink ⇒ not managed
		WriteFile{Path: "AGENTS.md"},   // not a symlink ⇒ not managed
	}
	got := ManagedLocations(actions)
	want := []string{".claude/skills", "construct"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ManagedLocations = %v, want %v", got, want)
	}
}

func TestShouldPruneKeepsNonWeaveAndProduced(t *testing.T) {
	roots := []string{"/ws/substrate"}
	produced := map[string]bool{".claude/skills/xx-keep": true}

	cases := []struct {
		name string
		c    PruneCandidate
		want bool
	}{
		{
			name: "produced symlink under substrate → KEEP",
			c:    PruneCandidate{RelPath: ".claude/skills/xx-keep", ResolvedTarget: "/ws/substrate/construct/local/keep", IsSymlink: true},
			want: false,
		},
		{
			name: "orphan symlink under substrate → PRUNE",
			c:    PruneCandidate{RelPath: ".claude/skills/yy-orphan", ResolvedTarget: "/ws/substrate/construct/local/orphan", IsSymlink: true},
			want: true,
		},
		{
			name: "dangling symlink under substrate → PRUNE",
			c:    PruneCandidate{RelPath: ".claude/skills/xx-dead", ResolvedTarget: "/ws/substrate/construct/scripts/setup.sh", IsSymlink: true},
			want: true,
		},
		{
			name: "non-weave symlink (unrelated target) → KEEP",
			c:    PruneCandidate{RelPath: ".claude/skills/vendor", ResolvedTarget: "/ws/elsewhere/thing", IsSymlink: true},
			want: false,
		},
		{
			name: "not a symlink (real file/dir slipped in) → KEEP",
			c:    PruneCandidate{RelPath: ".claude/skills/real", ResolvedTarget: "/ws/substrate/x", IsSymlink: false},
			want: false,
		},
		{
			name: "target is a sibling sharing a string prefix → KEEP",
			c:    PruneCandidate{RelPath: ".claude/skills/sib", ResolvedTarget: "/ws/substrate-evil/x", IsSymlink: true},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldPrune(tc.c, produced, roots); got != tc.want {
				t.Fatalf("shouldPrune = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPruneCrossTargetBidirectional pins the Option B (#107) cross-target prune: a
// lean compile SCANS the Union's managed locations (both skill dirs) but PRODUCES
// only its own face, so the OTHER face's stale symlinks are pruned — both ways. The
// Union compile (scan == produced) prunes neither.
func TestPruneCrossTargetBidirectional(t *testing.T) {
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "repo")
	substrate := filepath.Join(parent, "substrate")
	if err := os.MkdirAll(filepath.Join(substrate, "construct/local/fix"), 0o755); err != nil {
		t.Fatal(err)
	}
	mkFace := func(dir string) { // a prior Union compile left BOTH faces on disk
		d := filepath.Join(repoRoot, dir)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("../../../substrate/construct/local/fix", filepath.Join(d, "xx-fix")); err != nil && !os.IsExist(err) {
			t.Fatal(err)
		}
	}
	mkFace(".claude/skills")
	mkFace(".agents/skills")

	sourceRoots := SourceRootsFromPaths([]string{substrate})
	src := filepath.Join(substrate, "construct/local/fix")
	claudeLink := Symlink{Src: src, Dst: ".claude/skills/xx-fix"}
	agentsLink := Symlink{Src: src, Dst: ".agents/skills/xx-fix"}
	union := []Action{claudeLink, agentsLink} // the scan set: both faces' dirs
	exists := func(rel string) bool { _, err := os.Lstat(filepath.Join(repoRoot, rel)); return err == nil }

	// Lean codex: produces .agents/skills only → prunes the .claude/skills face.
	pruned, err := PruneOrphans(weavefs.OSFS{}, repoRoot, union, []Action{agentsLink}, sourceRoots)
	if err != nil {
		t.Fatalf("codex prune: %v", err)
	}
	if !reflect.DeepEqual(pruned, []string{".claude/skills/xx-fix"}) {
		t.Errorf("codex pruned %v, want [.claude/skills/xx-fix]", pruned)
	}
	if exists(".claude/skills/xx-fix") || !exists(".agents/skills/xx-fix") {
		t.Errorf("codex: want .claude/skills pruned + .agents/skills kept")
	}

	// Lean claude: the mirror — produces .claude/skills only → prunes .agents/skills.
	mkFace(".claude/skills") // restore the just-pruned link
	// SAFETY in the cross-target scan context: a hand-authored REAL dir + a non-weave
	// symlink sitting in the OTHER face's now-scanned location must SURVIVE the prune.
	if err := os.MkdirAll(filepath.Join(repoRoot, ".agents/skills/hand-authored"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(parent, "elsewhere"), filepath.Join(repoRoot, ".agents/skills/vendor")); err != nil {
		t.Fatal(err)
	}
	pruned, err = PruneOrphans(weavefs.OSFS{}, repoRoot, union, []Action{claudeLink}, sourceRoots)
	if err != nil {
		t.Fatalf("claude prune: %v", err)
	}
	if !reflect.DeepEqual(pruned, []string{".agents/skills/xx-fix"}) {
		t.Errorf("claude pruned %v, want only [.agents/skills/xx-fix]", pruned)
	}
	if !exists(".claude/skills/xx-fix") || exists(".agents/skills/xx-fix") {
		t.Errorf("claude: want .agents/skills/xx-fix pruned + .claude/skills/xx-fix kept")
	}
	if !exists(".agents/skills/hand-authored") || !exists(".agents/skills/vendor") {
		t.Errorf("cross-target prune destroyed a hand-authored dir / non-weave symlink in the scanned .agents/skills")
	}

	// Union compile (scan == produced) prunes neither face.
	mkFace(".agents/skills")
	pruned, err = PruneOrphans(weavefs.OSFS{}, repoRoot, union, union, sourceRoots)
	if err != nil {
		t.Fatalf("union prune: %v", err)
	}
	if len(pruned) != 0 {
		t.Errorf("union pruned %v, want none", pruned)
	}
}
