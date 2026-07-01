package processmanual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeProjectSlug(t *testing.T) {
	if got := claudeProjectSlug("/Users/x/workspace/ariadne"); got != "-Users-x-workspace-ariadne" {
		t.Errorf("claudeProjectSlug = %q, want -Users-x-workspace-ariadne", got)
	}
}

func TestMemorySources_AbsentThenPresent_WithInclude(t *testing.T) {
	home := t.TempDir()
	const repo = "/w/ariadne"

	// Absent (include=true) → exactly one best-effort "none" record.
	got := memorySources(home, repo, true)
	if len(got) != 1 || got[0].Kind != KindMemory || !strings.Contains(got[0].Body, "No persisted memories") {
		t.Fatalf("absent: want one 'none' memory record, got %+v", got)
	}
	if got[0].Link != "" {
		t.Errorf("absent record should have no link, got %q", got[0].Link)
	}

	// Present (include=true) → a record for MEMORY.md, with an absolute link.
	memDir := filepath.Join(home, ".claude", "projects", claudeProjectSlug(repo), "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte("# Memory Index\n\nrows.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got = memorySources(home, repo, true)
	var idx *InjectionSource
	for i := range got {
		if got[i].Kind != KindMemory {
			t.Errorf("non-memory kind %q", got[i].Kind)
		}
		if got[i].Title == "MEMORY.md" {
			idx = &got[i]
		}
	}
	if idx == nil {
		t.Fatalf("present: MEMORY.md record missing, got %+v", got)
	}
	if !strings.HasPrefix(idx.Link, "/") {
		t.Errorf("memory link should be absolute (outside-repo), got %q", idx.Link)
	}
}

// The default (include=false) MUST NOT leak private/machine-local data: no memory
// file contents, no absolute home paths — even when memories are present on disk.
func TestMemorySources_RedactedByDefault(t *testing.T) {
	home := t.TempDir()
	const repo = "/w/ariadne"
	memDir := filepath.Join(home, ".claude", "projects", claudeProjectSlug(repo), "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A memory whose content would be sensitive if inlined.
	if err := os.WriteFile(filepath.Join(memDir, "user_secret.md"), []byte("SECRET personal note.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := memorySources(home, repo, false)
	if len(got) != 1 || got[0].Kind != KindMemory {
		t.Fatalf("redacted: want one memory note, got %+v", got)
	}
	r := got[0]
	if r.Link != "" {
		t.Errorf("redacted note must have no (absolute) link, got %q", r.Link)
	}
	for _, leak := range []string{"SECRET", "user_secret", home, "/.claude/"} {
		if strings.Contains(r.Title+r.When+r.Body+r.Link, leak) {
			t.Errorf("redacted memory leaked %q:\n%+v", leak, r)
		}
	}
}
