package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Integration test: a synthetic multi-layer repo on disk, compiled end-to-end
// through the real pipeline (walk → Plan → Apply over a real OSFS rooted at
// t.TempDir). Asserts the composed AGENTS.md ordering, the generic symlink, the
// scaffold dir, and the self-reference skip — the M2 IO seam working in anger.

func mkfile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildFixture lays down a `base` layer + a `derived` repo as siblings and
// returns the derived repo root. The base layer ships:
//   - prose AGENTS.local.md   (its prose fragment)
//   - scaffold .claude/skills (a directory to create)
//   - symlink shared.md       (a generic symlink — the dominant file-op case)
//   - symlink AGENTS.md       (a SELF-reference IN BASE? no — base's own; but on
//     the derived walk base/AGENTS.md != derived/AGENTS.md, so it's NOT a self-ref
//     there). To exercise the self-reference skip we add it to DERIVED's manifest.
//
// derived ships: construct/deps → ../base, prose AGENTS.local.md (its own
// fragment), and a self-reference `symlink selfdoc.md` (derived/selfdoc.md ==
// derived/selfdoc.md → must be skipped).
func buildFixture(t *testing.T) (parent, base, derived string) {
	t.Helper()
	parent = t.TempDir()
	base = filepath.Join(parent, "base")
	derived = filepath.Join(parent, "derived")

	mkfile(t, filepath.Join(base, "construct", "base.manifest"),
		"prose AGENTS.local.md\nscaffold .claude/skills\nsymlink shared.md\n")
	mkfile(t, filepath.Join(base, "AGENTS.local.md"), "BASE PROSE")
	mkfile(t, filepath.Join(base, "shared.md"), "SHARED CONTENT")

	mkfile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"),
		"prose AGENTS.local.md\nsymlink selfdoc.md\n")
	mkfile(t, filepath.Join(derived, "AGENTS.local.md"), "DERIVED PROSE")
	mkfile(t, filepath.Join(derived, "selfdoc.md"), "SELF DOC")

	return parent, base, derived
}

func TestCompileEndToEnd(t *testing.T) {
	_, _, derived := buildFixture(t)

	var out bytes.Buffer
	// The Union (default) writes the composed prose to every entry file (incl. AGENTS.md).
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}

	// AGENTS.md = base prose THEN derived prose (foundation-first).
	agents, err := os.ReadFile(filepath.Join(derived, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	want := "BASE PROSE\n\nDERIVED PROSE\n"
	if string(agents) != want {
		t.Fatalf("AGENTS.md = %q, want %q", agents, want)
	}
	// Union: the SAME composed prose is fanned to CLAUDE.md + GEMINI.md, byte-identical.
	for _, ef := range []string{"CLAUDE.md", "GEMINI.md"} {
		b, err := os.ReadFile(filepath.Join(derived, ef))
		if err != nil {
			t.Fatalf("Union did not write %s: %v", ef, err)
		}
		if string(b) != want {
			t.Errorf("%s = %q, want %q (same prose as AGENTS.md)", ef, b, want)
		}
	}

	// The generic symlink from base exists and resolves to base content.
	shared := filepath.Join(derived, "shared.md")
	if fi, err := os.Lstat(shared); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("shared.md is not a symlink: err=%v", err)
	}
	if c, err := os.ReadFile(shared); err != nil || string(c) != "SHARED CONTENT" {
		t.Fatalf("shared.md resolves to %q (err=%v), want SHARED CONTENT", c, err)
	}

	// The scaffold dir exists.
	if fi, err := os.Stat(filepath.Join(derived, ".claude", "skills")); err != nil || !fi.IsDir() {
		t.Fatalf(".claude/skills scaffold dir missing: err=%v", err)
	}

	// The self-reference (`symlink selfdoc.md` in derived's own manifest) was
	// skipped — selfdoc.md is still the ORIGINAL regular file, not a symlink.
	selfdoc := filepath.Join(derived, "selfdoc.md")
	fi, err := os.Lstat(selfdoc)
	if err != nil {
		t.Fatalf("selfdoc.md vanished: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("selfdoc.md became a symlink — self-reference was NOT skipped")
	}
	if c, _ := os.ReadFile(selfdoc); string(c) != "SELF DOC" {
		t.Fatalf("selfdoc.md content changed to %q — self-reference clobbered it", c)
	}
}

// TestCompileEnsuresGitignore proves weave OWNS ignoring its own generated-
// runtime artifacts: a `weave compile` on a fixture repo (which ships no
// .gitignore) leaves a .gitignore carrying every fixed generated-runtime entry,
// and a second compile is idempotent (no duplicate lines) — so a fresh compile
// on ANY derivative leaves a clean `git status` with no per-repo hand-edit.
func TestCompileEnsuresGitignore(t *testing.T) {
	_, _, derived := buildFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	gi := filepath.Join(derived, ".gitignore")
	got, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read .gitignore (compile should have created it): %v", err)
	}
	for _, entry := range plan.GeneratedRuntimeGitignoreEntries {
		if !strings.Contains(string(got), entry+"\n") {
			t.Fatalf(".gitignore missing generated-runtime entry %q:\n%s", entry, got)
		}
	}

	// Re-compile: idempotent, byte-identical .gitignore (no duplicated lines).
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, false, &out); err != nil {
		t.Fatalf("run (2nd): %v", err)
	}
	again, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read .gitignore after 2nd compile: %v", err)
	}
	if string(again) != string(got) {
		t.Fatalf("re-compile changed .gitignore:\n1st: %q\n2nd: %q", got, again)
	}
}

func TestCompileMultiLayerVisibility(t *testing.T) {
	// The 𝒜(R) invariant end-to-end (workshop/targets/weave-composition-
	// algebra.md, #99): a synthetic 3-layer stack — foundation with BOTH an
	// `export prose` and an `internal prose`; a middle with `export prose`; a leaf
	// with `internal prose` — compiled at the LEAF. The composed AGENTS.md must be
	// [foundation-export, middle-export, leaf-internal] and must NOT carry the
	// foundation's or middle's internal prose. This is the full walk → Plan →
	// Apply path, not just the pure planner.
	parent := t.TempDir()
	foundation := filepath.Join(parent, "foundation")
	middle := filepath.Join(parent, "middle")
	leaf := filepath.Join(parent, "leaf")

	// foundation: exports AGENTS.base.md, keeps AGENTS.local.md internal.
	mkfile(t, filepath.Join(foundation, "construct", "base.manifest"),
		"export prose AGENTS.base.md\ninternal prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(foundation, "AGENTS.base.md"), "FOUNDATION-EXPORT")
	mkfile(t, filepath.Join(foundation, "AGENTS.local.md"), "FOUNDATION-INTERNAL")

	// middle: depends on foundation; exports its own base prose.
	mkfile(t, filepath.Join(middle, "construct", "deps"), "substrate ../foundation\n")
	mkfile(t, filepath.Join(middle, "construct", "base.manifest"), "export prose AGENTS.base.md\n")
	mkfile(t, filepath.Join(middle, "AGENTS.base.md"), "MIDDLE-EXPORT")

	// leaf: depends on middle; declares only its own internal local prose.
	mkfile(t, filepath.Join(leaf, "construct", "deps"), "substrate ../middle\n")
	mkfile(t, filepath.Join(leaf, "construct", "base.manifest"), "internal prose AGENTS.local.md\n")
	mkfile(t, filepath.Join(leaf, "AGENTS.local.md"), "LEAF-INTERNAL")

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil { // Union writes AGENTS.md
		t.Fatalf("run: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(leaf, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	want := "FOUNDATION-EXPORT\n\nMIDDLE-EXPORT\n\nLEAF-INTERNAL\n"
	if string(agents) != want {
		t.Fatalf("AGENTS.md = %q, want %q", agents, want)
	}
	// The ancestor internals must be excluded — the 𝒜(R) exclusion proof.
	if strings.Contains(string(agents), "FOUNDATION-INTERNAL") {
		t.Errorf("leaf AGENTS.md leaked the foundation's INTERNAL prose:\n%s", agents)
	}
}

func TestBuildSkillIndexExcludesAncestorInternalSkill(t *testing.T) {
	// #104 M2 — the SKILL visibility 𝒜(R) end-to-end (the M1 review's deferred
	// leaf-index assertion, routed through walk → buildSkillIndex, NOT a direct
	// SelectVisible call — so the len(layers)-1 leaf-index plumbing is verified).
	// A base layer declares an `internal skill construct/skill` (a private skill);
	// a derived repo consumes base. Compiling the DERIVED must EXCLUDE base's
	// internal skill (ancestor-internal) but keep its export; base's OWN self-walk
	// INCLUDES the internal (leaf-internal).
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived := filepath.Join(parent, "derived")
	for _, r := range []string{base, derived} {
		mkfile(t, filepath.Join(r, "construct", "config.json"), `{"localPrefix": "xx-"}`+"\n")
	}
	mkfile(t, filepath.Join(base, "construct", "base.manifest"),
		"skill construct/local\ninternal skill construct/skill\n")
	writeSkillFile(t, filepath.Join(base, "construct", "local", "shared"), "shared", "exported", "S")
	writeSkillFile(t, filepath.Join(base, "construct", "skill", "secret"), "secret", "private", "X")
	mkfile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"), "skill construct/local\n")
	writeSkillFile(t, filepath.Join(derived, "construct", "local", "own"), "own", "derived's own", "O")

	// Compile the DERIVED (base is an ancestor): base's internal xx-secret excluded.
	layers, err := walk.Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatal(err)
	}
	idx, _, err := buildSkillIndex(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := idx.BodyPath("xx-secret"); ok {
		t.Error("ancestor's INTERNAL skill xx-secret leaked into the consumer's composition")
	}
	if _, ok := idx.BodyPath("xx-shared"); !ok {
		t.Error("ancestor's EXPORT skill xx-shared should reach the consumer")
	}
	if _, ok := idx.BodyPath("xx-own"); !ok {
		t.Error("derived's own skill xx-own missing")
	}

	// Base's OWN self-walk (base IS the leaf): its internal skill IS present.
	selfLayers, err := walk.Walk(weavefs.OSFS{}, base)
	if err != nil {
		t.Fatal(err)
	}
	selfIdx, _, err := buildSkillIndex(weavefs.OSFS{}, selfLayers)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := selfIdx.BodyPath("xx-secret"); !ok {
		t.Error("base's internal skill xx-secret should be present on its own self-walk")
	}
}

func TestCompileDryRunDoesNotMutate(t *testing.T) {
	_, _, derived := buildFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, true, &out); err != nil { // Union plans AGENTS.md
		t.Fatalf("run --dry-run: %v", err)
	}

	// Dry-run printed the actions...
	got := out.String()
	for _, want := range []string{"writefile AGENTS.md", "mkdir     .claude/skills", "symlink   shared.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, got)
		}
	}
	// ...and mutated NOTHING.
	if _, err := os.Stat(filepath.Join(derived, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote AGENTS.md (err=%v), want no mutation", err)
	}
	if _, err := os.Lstat(filepath.Join(derived, "shared.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created shared.md symlink, want no mutation")
	}
}

func TestFormatActions(t *testing.T) {
	got := formatActions([]plan.Action{
		plan.WriteFile{Path: "AGENTS.md", Content: "abc"},
		plan.Symlink{Src: "/up/x.md", Dst: "x.md"},
		plan.Mkdir{Path: ".claude/skills"},
	})
	want := "writefile AGENTS.md (3 bytes)\nsymlink   x.md -> /up/x.md\nmkdir     .claude/skills\n"
	if got != want {
		t.Fatalf("formatActions =\n%q\nwant\n%q", got, want)
	}
}

func TestBuildRootWiring(t *testing.T) {
	// The root command exists, is named weave, and no longer compiles itself:
	// RunE is nil (a bare `weave` prints help and mutates nothing — M5).
	cmd := buildRoot()
	if cmd.Use != "weave" {
		t.Fatalf("root Use = %q, want weave", cmd.Use)
	}
	if cmd.RunE != nil {
		t.Fatalf("root RunE should be nil (help-only); the compile moved to `weave compile`")
	}
	// The compile + golden + skills + skill subcommands are wired.
	want := map[string]bool{"compile": false, "golden": false, "skills": false, "skill": false}
	var compile *cobra.Command
	for _, c := range cmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
		if c.Name() == "compile" {
			compile = c
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("%s subcommand not wired", name)
		}
	}
	// compile carries --dry-run AND --target (default claude).
	if compile == nil {
		t.Fatal("compile subcommand missing")
	}
	if compile.Flags().Lookup("dry-run") == nil {
		t.Fatalf("--dry-run flag not wired on compile")
	}
	tf := compile.Flags().Lookup("target")
	if tf == nil {
		t.Fatalf("--target flag not wired on compile")
	}
	if tf.DefValue != "all" {
		t.Fatalf("--target default = %q, want all (the Union)", tf.DefValue)
	}
}

// writeSkillFile lays one SKILL.md (frontmatter name/description + body) under
// dir for the skill-server CLI tests.
func writeSkillFile(t *testing.T, dir, name, desc, body string) {
	t.Helper()
	mkfile(t, filepath.Join(dir, "SKILL.md"),
		"---\nname: "+name+"\ndescription: "+desc+"\n---\n\n"+body+"\n")
}

// buildSkillRepoFixture lays a base layer (a local skill `sdlc` + an adapted
// skill `superpowers-brainstorming`) and a derived repo (its own local skill
// `issues`) as siblings, each with a construct/config.json (xx- prefix) and a
// prose fragment, and returns the derived repo root. The derived repo depends
// on base via construct/deps, so the walk resolves [base, derived].
func buildSkillRepoFixture(t *testing.T) (derived string) {
	t.Helper()
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived = filepath.Join(parent, "derived")

	for _, root := range []string{base, derived} {
		mkfile(t, filepath.Join(root, "construct", "config.json"), `{"localPrefix": "xx-"}`+"\n")
	}
	// base: a manifest declaring prose + the two `skill` intents (so BOTH skill
	// backends fire — the menu scan and the .claude/skills symlink lowering),
	// matching the real base.manifest. Plus prose and two skills (local + adapted).
	mkfile(t, filepath.Join(base, "construct", "base.manifest"),
		"prose AGENTS.local.md\nskill construct/local\nskill construct/adapted\n")
	mkfile(t, filepath.Join(base, "AGENTS.local.md"), "BASE PROSE")
	writeSkillFile(t, filepath.Join(base, "construct", "local", "sdlc"),
		"sdlc", "SDLC checkpoint gates", "# sdlc\n\nBASE SDLC BODY")
	writeSkillFile(t, filepath.Join(base, "construct", "adapted", "superpowers-brainstorming"),
		"superpowers-brainstorming", "Brainstorm before building", "# Brainstorm\n\nBRAINSTORM BODY")

	// derived: depends on base, its own prose + a local skill, and the same two
	// `skill` intents so the derived layer's skill dir is lowered too.
	mkfile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"),
		"prose AGENTS.local.md\nskill construct/local\nskill construct/adapted\n")
	mkfile(t, filepath.Join(derived, "AGENTS.local.md"), "DERIVED PROSE")
	writeSkillFile(t, filepath.Join(derived, "construct", "local", "issues"),
		"xx-issues", "Issue files in workshop/issues", "# issues\n\nISSUES BODY")

	return derived
}

// TestCompileTargetCodexEmitsAgentsSkills asserts the codex face (Option B #107):
// a PROSE-ONLY AGENTS.md (NO `## Skills` menu — Codex auto-composes its own from
// .agents/skills) + .agents/skills/<name> symlinks, and ZERO .claude/skills (that's
// Claude's dir). Mirror: TestCompileTargetClaudeEmitsSymlinksProseOnly.
func TestCompileTargetCodexEmitsAgentsSkills(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetCodex, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	agents, err := os.ReadFile(filepath.Join(derived, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	body := string(agents)
	if !strings.HasPrefix(body, "BASE PROSE\n\nDERIVED PROSE\n") {
		t.Errorf("codex AGENTS.md should be foundation-first prose:\n%s", body)
	}
	if strings.Contains(body, "## Skills") {
		t.Errorf("codex AGENTS.md must NOT carry a `## Skills` menu (Option B):\n%s", body)
	}
	// Skills lower to .agents/skills/<name>, each resolving to its upstream body.
	for name, wantBody := range map[string]string{
		"xx-sdlc":                   "BASE SDLC BODY",
		"superpowers-brainstorming": "BRAINSTORM BODY",
		"xx-issues":                 "ISSUES BODY",
	} {
		link := filepath.Join(derived, ".agents", "skills", name)
		fi, err := os.Lstat(link)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("codex: expected .agents/skills/%s symlink: err=%v", name, err)
		}
		b, err := os.ReadFile(filepath.Join(link, "SKILL.md"))
		if err != nil || !strings.Contains(string(b), wantBody) {
			t.Errorf(".agents/skills/%s resolves to %q (err=%v), want %q", name, b, err, wantBody)
		}
	}
	// ZERO .claude/skills under codex (the Claude dir is not this harness's face).
	if entries, err := os.ReadDir(filepath.Join(derived, ".claude", "skills")); err == nil && len(entries) > 0 {
		t.Errorf("codex target wrote %d .claude/skills entries, want zero", len(entries))
	}
}

// TestCompileUnionMaterializesBothSkillDirs closes the "Done when: Union produces
// BOTH skill faces" loop e2e (#107): a single Union run() over a skill-bearing repo
// writes all three prose entry files AND lowers the same skills into BOTH
// .claude/skills and .agents/skills.
func TestCompileUnionMaterializesBothSkillDirs(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, ef := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		if _, err := os.Stat(filepath.Join(derived, ef)); err != nil {
			t.Errorf("Union missing entry file %s: %v", ef, err)
		}
	}
	for _, dir := range []string{".claude/skills", ".agents/skills"} {
		for _, name := range []string{"xx-sdlc", "superpowers-brainstorming", "xx-issues"} {
			link := filepath.Join(derived, dir, name)
			if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
				t.Errorf("Union: %s/%s missing or not a symlink: %v", dir, name, err)
			}
		}
	}
}

// TestCompileTargetClaudeEmitsSymlinksProseOnly is the M5 per-target assertion
// for the DEFAULT backend: `--target claude` lowers the .claude/skills/<name>
// symlink backend (the links every Claude harness reads) AND composes a
// PROSE-ONLY AGENTS.md — NO `## Skills` menu (the harness discovers skills from
// .claude/skills, so the always-on menu would be redundant). The two skill
// backends are mutually exclusive: symlinks here, no menu.
func TestCompileTargetClaudeEmitsSymlinksProseOnly(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}

	// Symlink backend: .claude/skills/<name> exists for each skill, each a real
	// symlink resolving to its upstream skill dir (local prefixed, adapted bare).
	skillLinks := map[string]string{
		"xx-sdlc":                   "BASE SDLC BODY",
		"superpowers-brainstorming": "BRAINSTORM BODY",
		"xx-issues":                 "ISSUES BODY",
	}
	for name, wantBody := range skillLinks {
		link := filepath.Join(derived, ".claude", "skills", name)
		fi, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("symlink backend: expected .claude/skills/%s: %v", name, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("symlink backend: .claude/skills/%s is not a symlink", name)
		}
		body, err := os.ReadFile(filepath.Join(link, "SKILL.md"))
		if err != nil {
			t.Fatalf("symlink backend: .claude/skills/%s/SKILL.md unreadable: %v", name, err)
		}
		if !strings.Contains(string(body), wantBody) {
			t.Errorf("symlink backend: .claude/skills/%s resolves to body %q, want it to contain %q", name, body, wantBody)
		}
	}

	// CLAUDE.md is the claude entry file — foundation-first prose, no menu. The
	// claude target writes NO AGENTS.md and NO .agents/skills (other harnesses' faces).
	claudeMD, err := os.ReadFile(filepath.Join(derived, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	body := string(claudeMD)
	if !strings.HasPrefix(body, "BASE PROSE\n\nDERIVED PROSE\n") {
		t.Errorf("claude CLAUDE.md should be foundation-first prose:\n%s", body)
	}
	if strings.Contains(body, "## Skills") {
		t.Errorf("claude CLAUDE.md should NOT carry a `## Skills` menu:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(derived, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("claude target wrote AGENTS.md (err=%v), want only CLAUDE.md", err)
	}
	if entries, err := os.ReadDir(filepath.Join(derived, ".agents", "skills")); err == nil && len(entries) > 0 {
		t.Errorf("claude target wrote %d .agents/skills entries, want zero", len(entries))
	}
}

// TestCompileTargetClaudeDryRunListsSkillSymlinks asserts the symlink backend
// shows up in --dry-run under claude: the planned actions include a `symlink
// .claude/skills/<name>` line per skill, so the operator sees the links weave
// WILL write (and the golden harness, which Plans the same actions, classifies
// them MATCH).
func TestCompileTargetClaudeDryRunListsSkillSymlinks(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, true, &out); err != nil {
		t.Fatalf("run --dry-run: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"symlink   .claude/skills/xx-sdlc",
		"symlink   .claude/skills/superpowers-brainstorming",
		"symlink   .claude/skills/xx-issues",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dry-run output missing skill symlink %q:\n%s", want, got)
		}
	}
	// And it stays a dry run: nothing landed on disk.
	if _, err := os.Lstat(filepath.Join(derived, ".claude", "skills", "xx-sdlc")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created a .claude/skills symlink (err=%v), want no mutation", err)
	}
}

// TestCompileTargetCodexDryRunNoSkillSymlinks is the codex mirror: the dry-run
// plan carries the AGENTS.md write (with the menu) but ZERO `.claude/skills`
// symlink lines — proving the symlink backend is suppressed for codex.
func TestCompileTargetCodexDryRunNoSkillSymlinks(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetCodex, true, &out); err != nil {
		t.Fatalf("run --dry-run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "writefile AGENTS.md") {
		t.Fatalf("codex dry-run missing the AGENTS.md write (with menu):\n%s", got)
	}
	if strings.Contains(got, ".claude/skills/") {
		t.Fatalf("codex dry-run lists a .claude/skills symlink, want zero:\n%s", got)
	}
}

// TestRootNoSubcommandDoesNotMutate confirms a bare `weave` (no subcommand) is
// help-only: it prints usage, returns nil, and mutates NOTHING — the compile is
// now the explicit `weave compile`.
func TestRootNoSubcommandDoesNotMutate(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	cmd := buildRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil) // bare `weave`
	if err := cmd.Execute(); err != nil {
		t.Fatalf("bare `weave` should print help and succeed, got: %v", err)
	}
	// Help text mentions the compile subcommand.
	if !strings.Contains(out.String(), "compile") {
		t.Errorf("bare `weave` help should mention `compile`:\n%s", out.String())
	}
	// No AGENTS.md, no .claude/skills — the bare command mutated nothing. (Run
	// from the fixture root so a stray compile would have touched these.)
	if _, err := os.Stat(filepath.Join(derived, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("bare `weave` wrote AGENTS.md (err=%v), want no mutation", err)
	}
	if _, err := os.Lstat(filepath.Join(derived, ".claude", "skills", "xx-sdlc")); !os.IsNotExist(err) {
		t.Errorf("bare `weave` created a .claude/skills symlink, want no mutation")
	}
}

// TestVerifyCompleteCoversSkillsBothTargets asserts a skill intent is satisfied
// by EITHER backend: --target claude (the .claude/skills symlinks) AND --target
// codex (the AGENTS.md menu) both report ZERO under-production on the fixture.
func TestVerifyCompleteCoversSkillsBothTargets(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	for _, target := range []plan.Target{plan.TargetClaude, plan.TargetCodex} {
		var out bytes.Buffer
		if err := runVerifyComplete(weavefs.OSFS{}, derived, []string{derived}, target, &out); err != nil {
			t.Fatalf("verify-complete --target %s reported under-production: %v\n%s", target, err, out.String())
		}
		if !strings.Contains(out.String(), "0 setup.sh-produced path(s) NOT planned") {
			t.Errorf("verify-complete --target %s missing zero-under verdict:\n%s", target, out.String())
		}
	}
}

func TestRunSkillsPrintsMenu(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	if err := runSkills(weavefs.OSFS{}, derived, &out); err != nil {
		t.Fatalf("runSkills: %v", err)
	}
	got := out.String()
	// Foundation-first order: base's skills (local then adapted) before derived.
	wantOrder := []string{
		"xx-sdlc — SDLC checkpoint gates",
		"superpowers-brainstorming — Brainstorm before building",
		"xx-issues — Issue files in workshop/issues",
	}
	last := -1
	for _, line := range wantOrder {
		i := strings.Index(got, line)
		if i < 0 {
			t.Fatalf("`weave skills` output missing %q:\n%s", line, got)
		}
		if i < last {
			t.Errorf("`weave skills` lines out of order around %q:\n%s", line, got)
		}
		last = i
	}
}

func TestRunSkillServesBody(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	// A base local skill, served by its namespaced name.
	var out bytes.Buffer
	if err := runSkill(weavefs.OSFS{}, derived, "xx-sdlc", &out); err != nil {
		t.Fatalf("runSkill xx-sdlc: %v", err)
	}
	if !strings.Contains(out.String(), "BASE SDLC BODY") {
		t.Errorf("`weave skill xx-sdlc` body = %q, want the sdlc SKILL.md body", out.String())
	}

	// An adapted skill, served by its bare name.
	out.Reset()
	if err := runSkill(weavefs.OSFS{}, derived, "superpowers-brainstorming", &out); err != nil {
		t.Fatalf("runSkill superpowers-brainstorming: %v", err)
	}
	if !strings.Contains(out.String(), "BRAINSTORM BODY") {
		t.Errorf("adapted body = %q, want the brainstorming SKILL.md body", out.String())
	}

	// The derived repo's own skill.
	out.Reset()
	if err := runSkill(weavefs.OSFS{}, derived, "xx-issues", &out); err != nil {
		t.Fatalf("runSkill xx-issues: %v", err)
	}
	if !strings.Contains(out.String(), "ISSUES BODY") {
		t.Errorf("derived body = %q, want the issues SKILL.md body", out.String())
	}
}

func TestRunSkillUnknownNameErrors(t *testing.T) {
	derived := buildSkillRepoFixture(t)

	var out bytes.Buffer
	err := runSkill(weavefs.OSFS{}, derived, "no-such-skill", &out)
	if err == nil {
		t.Fatal("runSkill on an unknown name should error (non-zero exit)")
	}
	if !strings.Contains(err.Error(), "unknown skill") || !strings.Contains(err.Error(), "weave skills") {
		t.Errorf("unknown-skill error = %q, want a helpful message pointing at `weave skills`", err)
	}
	if out.Len() != 0 {
		t.Errorf("unknown skill wrote a body %q, want nothing", out.String())
	}
}

func TestWorkspaceRoot(t *testing.T) {
	// A normal repo: workspace is the parent.
	if got := workspaceRoot("/ws/nous"); got != "/ws" {
		t.Fatalf("normal repo: workspaceRoot = %q, want /ws", got)
	}
	// A worktree …/workspace/worktree/<repo>/<branch>: climb past worktree.
	if got := workspaceRoot("/ws/worktree/ariadne/000095-weave"); got != "/ws" {
		t.Fatalf("worktree: workspaceRoot = %q, want /ws", got)
	}
}

func TestGoldenTargetsExplicitArgs(t *testing.T) {
	// Explicit args are made absolute against cwd; no auto-discovery.
	got := goldenTargets("/ws/cur", []string{"../nous", "/abs/brain"})
	want := []string{"/ws/nous", "/abs/brain"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("goldenTargets(args) = %v, want %v", got, want)
	}
}

func TestGoldenTargetsAutoDiscover(t *testing.T) {
	// No args → the canonical layer repos as siblings of the workspace root.
	got := goldenTargets("/ws/worktree/ariadne/000095-weave", nil)
	want := []string{"/ws/ariadne", "/ws/nous", "/ws/brain", "/ws/metis"}
	if len(got) != len(want) {
		t.Fatalf("goldenTargets(nil) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("goldenTargets(nil)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestGoldenCleanRepo runs the golden subcommand pipeline against a synthetic
// derivative whose live tree exactly realizes weave's plan (a correct symlink +
// scaffold dir) plus a deferred-verb (seed) output. The ledger must be clean:
// all MATCH/EXPECTED, zero UNEXPECTED → runGolden returns nil.
func TestGoldenCleanRepo(t *testing.T) {
	parent, base, derived := buildFixture(t)
	_ = parent

	// Realize weave's plan on `derived` exactly as Apply would: a relative
	// symlink to base/shared.md and the .claude/skills dir. (buildFixture's
	// derived has prose + a self-ref; the cross-layer symlink is base/shared.md.)
	rel, _ := filepath.Rel(derived, filepath.Join(base, "shared.md"))
	if err := os.Symlink(rel, filepath.Join(derived, "shared.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(derived, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	// CLAUDE.md: the claude target plans a composed WriteFile (base + derived prose);
	// realize it so the WriteFile classifies MATCH.
	if err := os.WriteFile(filepath.Join(derived, "CLAUDE.md"), []byte("BASE PROSE\n\nDERIVED PROSE\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runGolden(weavefs.OSFS{}, derived, []string{derived}, plan.TargetClaude, &out); err != nil {
		t.Fatalf("runGolden on a clean tree returned error: %v\nledger:\n%s", err, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "UNEXPECTED 0") {
		t.Fatalf("clean repo ledger missing UNEXPECTED 0:\n%s", got)
	}
}

// TestGoldenDetectsDivergence flips the symlink to point somewhere weave would
// NOT link; the harness must classify it UNEXPECTED and return a non-nil error.
func TestGoldenDetectsDivergence(t *testing.T) {
	_, _, derived := buildFixture(t)

	// A WRONG symlink for shared.md (points at a bogus target).
	if err := os.Symlink("../bogus/shared.md", filepath.Join(derived, "shared.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(derived, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(derived, "AGENTS.md"), []byte("BASE PROSE\n\nDERIVED PROSE\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runGolden(weavefs.OSFS{}, derived, []string{derived}, plan.TargetClaude, &out)
	if err == nil {
		t.Fatalf("runGolden on a divergent tree returned nil, want error\nledger:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "UNEXPECTED") {
		t.Fatalf("divergent ledger missing UNEXPECTED line:\n%s", out.String())
	}
}

// TestGoldenSkipsAbsent verifies skip-if-absent: a non-present repo is noted and
// skipped (no error from its absence alone).
func TestGoldenSkipsAbsent(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "not-there")
	var out bytes.Buffer
	if err := runGolden(weavefs.OSFS{}, t.TempDir(), []string{absent}, plan.TargetClaude, &out); err != nil {
		t.Fatalf("absent repo should be skipped, got error: %v", err)
	}
	if !strings.Contains(out.String(), "SKIP") {
		t.Fatalf("absent repo not noted as SKIP:\n%s", out.String())
	}
}

func TestLinkCreatesDepsVerbatim(t *testing.T) {
	// `weave link <path>` records `substrate <path>` VERBATIM in the repo's
	// construct/deps — the path exactly as given (not resolved/relativized), so a
	// test setup captures the real path it was handed (the directory-agnostic
	// establishment verb). Creates construct/deps when absent.
	root := t.TempDir()
	var out bytes.Buffer
	path := "/some/where/ariadne-checkout"
	if err := runLink(weavefs.OSFS{}, root, path, &out); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	deps, err := os.ReadFile(filepath.Join(root, "construct", "deps"))
	if err != nil {
		t.Fatalf("read construct/deps: %v", err)
	}
	want := "substrate " + path + "\n"
	if string(deps) != want {
		t.Fatalf("construct/deps = %q, want %q", deps, want)
	}
}

func TestLinkIdempotent(t *testing.T) {
	// A second link with the same path must NOT duplicate the row, and must
	// preserve existing content.
	root := t.TempDir()
	depsPath := filepath.Join(root, "construct", "deps")
	mkfile(t, depsPath, "data ../d git@x\nsubstrate ../existing\n")
	var out bytes.Buffer
	if err := runLink(weavefs.OSFS{}, root, "../existing", &out); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	deps, err := os.ReadFile(depsPath)
	if err != nil {
		t.Fatal(err)
	}
	want := "data ../d git@x\nsubstrate ../existing\n"
	if string(deps) != want {
		t.Fatalf("construct/deps = %q, want unchanged %q", deps, want)
	}
}

func TestLinkAppendsToExisting(t *testing.T) {
	// link a NEW path appends a second substrate row, preserving the first.
	root := t.TempDir()
	depsPath := filepath.Join(root, "construct", "deps")
	mkfile(t, depsPath, "substrate ../ariadne\n")
	var out bytes.Buffer
	if err := runLink(weavefs.OSFS{}, root, "/abs/other", &out); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	deps, err := os.ReadFile(depsPath)
	if err != nil {
		t.Fatal(err)
	}
	want := "substrate ../ariadne\nsubstrate /abs/other\n"
	if string(deps) != want {
		t.Fatalf("construct/deps = %q, want %q", deps, want)
	}
}

func TestLinkSeedsBaseManifest(t *testing.T) {
	// #155: `weave link` seeds a minimal construct/base.manifest in a fresh repo,
	// so it is a valid, traversable layer out of the box. The seed carries the
	// `internal prose AGENTS.local.md` row (marks it traversable) and names the
	// substrate in its header.
	root := t.TempDir()
	var out bytes.Buffer
	if err := runLink(weavefs.OSFS{}, root, "../ariadne", &out); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "construct", "base.manifest"))
	if err != nil {
		t.Fatalf("seed must create construct/base.manifest: %v", err)
	}
	if !strings.Contains(string(manifest), "internal  prose AGENTS.local.md") {
		t.Errorf("seeded manifest missing the traversable-layer row:\n%s", manifest)
	}
	if !strings.Contains(string(manifest), "../ariadne") {
		t.Errorf("seeded manifest header should name the substrate:\n%s", manifest)
	}
	if !strings.Contains(out.String(), "seeded construct/base.manifest") {
		t.Errorf("link should announce the seed, got: %q", out.String())
	}
}

func TestLinkSeedNeverClobbersExisting(t *testing.T) {
	// The seed is idempotent: a repo that already ships a (hand-authored) manifest
	// must keep it verbatim — link never overwrites it.
	root := t.TempDir()
	manifestPath := filepath.Join(root, "construct", "base.manifest")
	original := "# hand-authored\nexport  prose AGENTS.base.md\ninternal  prose AGENTS.local.md\n"
	mkfile(t, manifestPath, original)
	var out bytes.Buffer
	if err := runLink(weavefs.OSFS{}, root, "../ariadne", &out); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	got, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("existing manifest was clobbered:\ngot:  %q\nwant: %q", got, original)
	}
	if strings.Contains(out.String(), "seeded construct/base.manifest") {
		t.Errorf("link must not announce a seed when a manifest already exists, got: %q", out.String())
	}
}

func TestLinkSeedMakesRepoTraversable(t *testing.T) {
	// #155 end-to-end: after `weave link`-seeding a fresh MID repo, a downstream
	// consumer's walk traverses THROUGH it to the foundation — the exact chain
	// (kbench → kaggle → metis) that used to silently under-compile. Proves the
	// seed (not just the file) fixes the footgun.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	mid := filepath.Join(parent, "mid")
	derived := filepath.Join(parent, "derived")
	mkfile(t, filepath.Join(base, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	// Bootstrap mid the natural way: link it to base (seeds mid's own manifest).
	var out bytes.Buffer
	if err := runLink(weavefs.OSFS{}, mid, "../base", &out); err != nil {
		t.Fatalf("runLink(mid): %v", err)
	}
	// derived depends on the freshly-seeded mid.
	mkfile(t, filepath.Join(derived, "construct", "deps"), "substrate ../mid\n")
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"), "prose AGENTS.local.md\n")

	layers, err := walk.Walk(weavefs.OSFS{}, derived)
	if err != nil {
		t.Fatalf("walk through a seeded intermediate must reach the foundation, got error: %v", err)
	}
	if len(layers) != 3 {
		t.Fatalf("chain must compose fully (base, mid, derived), got %d: %+v", len(layers), layers)
	}
}

func TestLinkWired(t *testing.T) {
	// `link` is registered as a subcommand on the root.
	cmd := buildRoot()
	found := false
	for _, c := range cmd.Commands() {
		if c.Name() == "link" {
			found = true
		}
	}
	if !found {
		t.Fatalf("link subcommand not wired")
	}
}
