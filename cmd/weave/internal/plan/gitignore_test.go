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

func TestEnsureGitignoreTextCreatesBlock(t *testing.T) {
	got, changed, err := ensureGitignoreText("", []string{"/AGENTS.md", "/.colima/"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false on an empty .gitignore, want true")
	}
	want := ignoreBegin + "\n/.colima/\n/AGENTS.md\n" + ignoreEnd + "\n"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
}

func TestEnsureGitignoreTextMigratesExactLegacyEntries(t *testing.T) {
	// Authored entries remain verbatim after the block; exact legacy entries move
	// into the generated block without duplication.
	current := "# existing comment\n/AGENTS.md\nbin/\n"
	got, changed, err := ensureGitignoreText(current, []string{"/AGENTS.md", "/.claude/skills/"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false, want true (one entry was absent)")
	}
	want := ignoreBegin + "\n/.claude/skills/\n/AGENTS.md\n" + ignoreEnd + "\n# existing comment\nbin/\n"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "/AGENTS.md") != 1 {
		t.Fatalf("/AGENTS.md duplicated:\n%s", got)
	}
}

func TestEnsureGitignoreTextIdempotentWhenAllPresent(t *testing.T) {
	// After initial legacy migration, another reconciliation is byte-identical.
	current, _, firstErr := ensureGitignoreText(strings.Join(GeneratedRuntimeGitignoreEntries, "\n")+"\n", GeneratedRuntimeGitignoreEntries)
	if firstErr != nil {
		t.Fatal(firstErr)
	}
	got, changed, err := ensureGitignoreText(current, GeneratedRuntimeGitignoreEntries)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("changed = true when all entries present, want false; got:\n%s", got)
	}
	if got != current {
		t.Fatalf("content mutated when all present:\n got %q\nwant %q", got, current)
	}
}

// TestGeneratedRuntimeGitignoreCoversConstructGenerated locks the #115 M3
// addition: the per-repo dynamic-skill materialization tree construct/generated/ is
// in the owned ignore set (gitignored EVERYWHERE — it's regenerated every compile
// and must never be tracked; the body lived committed under construct/local before).
func TestGeneratedRuntimeGitignoreCoversConstructGenerated(t *testing.T) {
	found := false
	for _, e := range GeneratedRuntimeGitignoreEntries {
		if e == "/construct/generated/" {
			found = true
		}
	}
	if !found {
		t.Fatalf("GeneratedRuntimeGitignoreEntries missing /construct/generated/: %v", GeneratedRuntimeGitignoreEntries)
	}
}

func TestEnsureGitignoreTextPreservesAuthoredMissingFinalNewline(t *testing.T) {
	// The generated block precedes authored content without rewriting its newline.
	got, changed, err := ensureGitignoreText("bin/", []string{"/AGENTS.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	want := ignoreBegin + "\n/AGENTS.md\n" + ignoreEnd + "\nbin/"
	if got != want {
		t.Fatalf("ensureGitignoreText = %q, want %q", got, want)
	}
}

func TestEnsureGitignoreTextDedupsRepeatedInputEntry(t *testing.T) {
	// A duplicate in the INPUT entry list is appended only once.
	got, _, err := ensureGitignoreText("", []string{"/AGENTS.md", "/AGENTS.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "/AGENTS.md") != 1 {
		t.Fatalf("repeated input entry duplicated:\n%s", got)
	}
}

func TestApplyEnsureGitignoreCreatesBlock(t *testing.T) {
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
