package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// gitignore.go is weave's generated-runtime ignore mechanism. The pure transform
// (ensureGitignoreText) is unit-tested directly; the IO seam
// (applyEnsureGitignore, via Apply) is tested against a real t.TempDir-rooted
// OSFS (ARCH: faithful over mocked).

func TestEnsureGitignoreTextAppendsToEmpty(t *testing.T) {
	got, changed := ensureGitignoreText("", []string{"/AGENTS.md", "/.colima/"})
	if !changed {
		t.Fatal("changed = false on an empty .gitignore, want true")
	}
	want := "/AGENTS.md\n/.colima/\n"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
}

func TestEnsureGitignoreTextPreservesExistingAndAppendsAbsent(t *testing.T) {
	// Existing entries + a comment are preserved verbatim; only the truly absent
	// entry is appended (the present one is NOT duplicated — grep -qxF semantics).
	current := "# existing comment\n/AGENTS.md\nbin/\n"
	got, changed := ensureGitignoreText(current, []string{"/AGENTS.md", "/.claude/skills/"})
	if !changed {
		t.Fatal("changed = false, want true (one entry was absent)")
	}
	want := "# existing comment\n/AGENTS.md\nbin/\n/.claude/skills/\n"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
	if strings.Count(got, "/AGENTS.md") != 1 {
		t.Fatalf("/AGENTS.md duplicated:\n%s", got)
	}
}

func TestEnsureGitignoreTextIdempotentWhenAllPresent(t *testing.T) {
	// Every entry already present ⇒ no change, byte-identical (running weave twice
	// never duplicates lines).
	current := "/AGENTS.md\n/CLAUDE.md\n/GEMINI.md\n/.claude/skills/\n/.agents/skills/\n/.claude/settings.json\n/.colima/\n/construct/scripts/vm-log.sh\n"
	got, changed := ensureGitignoreText(current, GeneratedRuntimeGitignoreEntries)
	if changed {
		t.Fatalf("changed = true when all entries present, want false; got:\n%s", got)
	}
	if got != current {
		t.Fatalf("content mutated when all present:\n got %q\nwant %q", got, current)
	}
}

func TestEnsureGitignoreTextAddsTrailingNewlineBeforeAppend(t *testing.T) {
	// A non-empty file NOT ending in a newline gets one before the appended entry,
	// so the new entry never glues onto the last existing line.
	got, changed := ensureGitignoreText("bin/", []string{"/AGENTS.md"})
	if !changed {
		t.Fatal("changed = false, want true")
	}
	want := "bin/\n/AGENTS.md\n"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
}

func TestEnsureGitignoreTextDedupsRepeatedInputEntry(t *testing.T) {
	// A duplicate in the INPUT entry list is appended only once.
	got, _ := ensureGitignoreText("", []string{"/AGENTS.md", "/AGENTS.md"})
	if strings.Count(got, "/AGENTS.md") != 1 {
		t.Fatalf("repeated input entry duplicated:\n%s", got)
	}
}

func TestApplyEnsureGitignoreCreatesAndAppends(t *testing.T) {
	// Apply on a repo with no .gitignore creates it carrying the fixed entries.
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, entry := range GeneratedRuntimeGitignoreEntries {
		if !strings.Contains(string(got), entry+"\n") {
			t.Fatalf(".gitignore missing %q:\n%s", entry, got)
		}
	}
}

func TestApplyEnsureGitignoreIdempotent(t *testing.T) {
	// Two applies in a row leave the .gitignore byte-identical (no churn / no dup).
	root := t.TempDir()
	gi := filepath.Join(root, ".gitignore")
	act := []Action{EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries}}

	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply (1st): %v", err)
	}
	first, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read after 1st: %v", err)
	}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply (2nd): %v", err)
	}
	second, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read after 2nd: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("re-weave changed .gitignore:\n1st: %q\n2nd: %q", first, second)
	}
}

func TestApplyEnsureGitignorePreservesExisting(t *testing.T) {
	// Apply preserves pre-existing entries/comments and appends only the absent.
	root := t.TempDir()
	gi := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gi, []byte("# hand notes\nbin/\n/AGENTS.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		EnsureGitignore{Entries: GeneratedRuntimeGitignoreEntries},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, keep := range []string{"# hand notes", "bin/"} {
		if !strings.Contains(string(got), keep) {
			t.Fatalf(".gitignore dropped pre-existing %q:\n%s", keep, got)
		}
	}
	if strings.Count(string(got), "/AGENTS.md") != 1 {
		t.Fatalf("/AGENTS.md duplicated:\n%s", got)
	}
}
