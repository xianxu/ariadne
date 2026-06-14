package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if err := run(weavefs.OSFS{}, derived, false, &out); err != nil {
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

func TestCompileDryRunDoesNotMutate(t *testing.T) {
	_, _, derived := buildFixture(t)

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, true, &out); err != nil {
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
	// The root command exists, is named weave, and exposes --dry-run.
	cmd := buildRoot()
	if cmd.Use != "weave" {
		t.Fatalf("root Use = %q, want weave", cmd.Use)
	}
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Fatalf("--dry-run flag not wired")
	}
}
