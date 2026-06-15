package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
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
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, false, &out); err != nil {
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
	if err := run(weavefs.OSFS{}, leaf, plan.TargetClaude, false, &out); err != nil {
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

func TestCompileDryRunDoesNotMutate(t *testing.T) {
	_, _, derived := buildFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetClaude, true, &out); err != nil {
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
	if tf.DefValue != "claude" {
		t.Fatalf("--target default = %q, want claude", tf.DefValue)
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

// TestCompileTargetCodexEmitsMenuOnly asserts the codex backend (Approach-1):
// the `## Skills` menu composes into AGENTS.md and NO .claude/skills symlinks are
// emitted — the two skill backends are mutually exclusive per target. (The Claude
// target's mirror is TestCompileTargetClaudeEmitsSymlinksProseOnly.)
func TestCompileTargetCodexEmitsMenuOnly(t *testing.T) {
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

	// Prose first (foundation-first), then the always-on `## Skills` menu.
	if !strings.HasPrefix(body, "BASE PROSE\n\nDERIVED PROSE\n") {
		t.Errorf("AGENTS.md should start with foundation-first prose:\n%s", body)
	}
	if !strings.Contains(body, "## Skills") {
		t.Errorf("codex AGENTS.md missing the `## Skills` menu section:\n%s", body)
	}
	if !strings.Contains(body, "weave skill <name>") {
		t.Errorf("AGENTS.md missing the on-demand note:\n%s", body)
	}
	for _, line := range []string{
		"xx-sdlc — SDLC checkpoint gates",
		"superpowers-brainstorming — Brainstorm before building",
		"xx-issues — Issue files in workshop/issues",
	} {
		if !strings.Contains(body, line) {
			t.Errorf("AGENTS.md missing menu line %q:\n%s", line, body)
		}
	}
	// And ZERO .claude/skills symlinks under codex (the symlink backend is off).
	if entries, err := os.ReadDir(filepath.Join(derived, ".claude", "skills")); err == nil && len(entries) > 0 {
		t.Errorf("codex target wrote %d .claude/skills entries, want zero", len(entries))
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

	// AGENTS.md is PROSE-ONLY under claude — NO `## Skills` menu.
	agents, err := os.ReadFile(filepath.Join(derived, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	body := string(agents)
	if !strings.HasPrefix(body, "BASE PROSE\n\nDERIVED PROSE\n") {
		t.Errorf("claude AGENTS.md should be foundation-first prose:\n%s", body)
	}
	if strings.Contains(body, "## Skills") {
		t.Errorf("claude target should NOT compose a `## Skills` menu (prose-only):\n%s", body)
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
	// AGENTS.md: weave plans a composed WriteFile (base + derived prose); realize
	// it so the WriteFile classifies MATCH.
	if err := os.WriteFile(filepath.Join(derived, "AGENTS.md"), []byte("BASE PROSE\n\nDERIVED PROSE\n"), 0o644); err != nil {
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
