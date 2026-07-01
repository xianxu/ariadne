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

func TestMemorySources_AbsentThenPresent(t *testing.T) {
	home := t.TempDir()
	const repo = "/w/ariadne"

	// Absent → exactly one best-effort "none" record.
	got := memorySources(home, repo)
	if len(got) != 1 || got[0].Kind != KindMemory || !strings.Contains(got[0].Body, "No persisted memories") {
		t.Fatalf("absent: want one 'none' memory record, got %+v", got)
	}
	if got[0].Link != "" {
		t.Errorf("absent record should have no link, got %q", got[0].Link)
	}

	// Present → a record for MEMORY.md, with an absolute (outside-repo) link.
	memDir := filepath.Join(home, ".claude", "projects", claudeProjectSlug(repo), "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte("# Memory Index\n\nrows.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got = memorySources(home, repo)
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
