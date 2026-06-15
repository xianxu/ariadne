package intent

import (
	"reflect"
	"testing"
)

// ParseManifest ports walk_manifest's line grammar (construct/setup.sh:290):
// '#'-comment + blank lines skipped, `read -r action source target` word-split,
// target defaults to source, unknown action warns+skips. Each ported file-op
// verb maps to its Kind; the new `prose`/`skill` semantic verbs parse too.
// Pure: it takes the manifest text, never the file (ARCH-PURE).

func TestParseManifestSymlink(t *testing.T) {
	got, err := ParseManifest("symlink AGENTS.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Symlink, Source: "AGENTS.md", Target: "AGENTS.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestTargetDefaultsToSource(t *testing.T) {
	// Single-column source ⇒ target defaults to source (walk_manifest:
	// target="${target:-$source}"). An explicit target overrides.
	got, err := ParseManifest("merge .claude/settings.ariadne.json .claude/settings.json\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestAllPortedVerbs(t *testing.T) {
	// Every file-op verb the live base.manifest uses maps to its Kind.
	content := `seed      bootstrap.sh
symlink   AGENTS.md
scaffold  .claude/skills
touch     workshop/lessons.md
`
	got, err := ParseManifest(content)
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{
		{Kind: Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
		{Kind: Symlink, Source: "AGENTS.md", Target: "AGENTS.md"},
		{Kind: Scaffold, Source: ".claude/skills", Target: ".claude/skills"},
		{Kind: Touch, Source: "workshop/lessons.md", Target: "workshop/lessons.md"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestSkipsRetiredToolVerb(t *testing.T) {
	// `tool` was retired in #95 M5 (Go-tool ownership is location-based via
	// construct/dev-aliases.sh, not a go.mod edit). A stale `tool` row must fall
	// through the unknown-verb skip (warn-and-ignore), exactly like `copy` —
	// never aborting the compile and never producing an Intent.
	got, err := ParseManifest("tool cmd/sdlc\nsymlink AGENTS.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Symlink, Source: "AGENTS.md", Target: "AGENTS.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v (tool row must be skipped)", got, want)
	}
}

func TestParseManifestNewSemanticVerbs(t *testing.T) {
	// The new `prose <relpath>` / `skill <relpath>` verbs parse too (per the
	// plan's prose grammar). These are not setup.sh verbs — weave adds them.
	got, err := ParseManifest("prose AGENTS.local.md\nskill construct/skills/xx-fix\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{
		{Kind: Prose, Source: "AGENTS.local.md", Target: "AGENTS.local.md"},
		{Kind: Skill, Source: "construct/skills/xx-fix", Target: "construct/skills/xx-fix"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestSkipsCommentsAndBlanks(t *testing.T) {
	// Whole-line '#' comments (with leading whitespace) and blank/whitespace
	// lines drop — matching walk_manifest's two `continue` guards. NOTE:
	// walk_manifest only skips WHOLE-line comments (^[[:space:]]*#); it does
	// NOT strip trailing comments (that's lib-deps' grammar, not this one).
	content := `# header comment
   # indented comment

seed bootstrap.sh

symlink AGENTS.md
`
	got, err := ParseManifest(content)
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{
		{Kind: Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
		{Kind: Symlink, Source: "AGENTS.md", Target: "AGENTS.md"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestVisibilityDefaultsExport(t *testing.T) {
	// A row with no leading visibility token defaults to Export (the zero value),
	// so every pre-visibility manifest row is unchanged. The visibility axis of
	// workshop/targets/weave-composition-algebra.md.
	got, err := ParseManifest("prose AGENTS.base.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Prose, Visibility: Export, Source: "AGENTS.base.md", Target: "AGENTS.base.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestExplicitExportToken(t *testing.T) {
	// A leading `export` token parses to Export and is stripped from the fields,
	// leaving `<type> <source> [<target>]` — identical to the bare row. A leading
	// export/internal is unambiguous: no type is named that.
	got, err := ParseManifest("export prose AGENTS.base.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Prose, Visibility: Export, Source: "AGENTS.base.md", Target: "AGENTS.base.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestInternalToken(t *testing.T) {
	// A leading `internal` token marks the artifact internal to the declaring
	// repo — selected only on its own self-walk, never in a derivative. The
	// target column still defaults to source after the token is consumed.
	got, err := ParseManifest("internal prose AGENTS.local.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Prose, Visibility: Internal, Source: "AGENTS.local.md", Target: "AGENTS.local.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestVisibilityWithExplicitTarget(t *testing.T) {
	// A visibility token + a three-column row: the token is consumed, then
	// `<type> <source> <target>` parses as before. Visibility composes with the
	// existing target-override grammar.
	got, err := ParseManifest("export merge .claude/settings.ariadne.json .claude/settings.json\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Merge, Visibility: Export, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v", got, want)
	}
}

func TestParseManifestUnknownActionSkips(t *testing.T) {
	// Unknown action mirrors walk_manifest's `*)` case: warn-and-skip, no
	// error (a stale `copy` row must not abort the whole compile).
	got, err := ParseManifest("copy old.txt\nsymlink AGENTS.md\n")
	if err != nil {
		t.Fatalf("ParseManifest: unexpected error: %v", err)
	}
	want := []Intent{{Kind: Symlink, Source: "AGENTS.md", Target: "AGENTS.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseManifest = %v, want %v (unknown `copy` skipped)", got, want)
	}
}
