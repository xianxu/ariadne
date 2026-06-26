package transcripts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Select over the two real harnesses (temp-rooted) merges a Claude cwd dir and
// the cwd-matched Codex session files, excluding an unrelated Codex cwd. Migrated
// from cmd/sdlc/actual_test.go's TestSelectActualSourcesIncludesMatchingCodexSessions
// (#134) — the one test that exercises both harnesses together.
func TestSelectMergesClaudeDirsAndCodexFiles(t *testing.T) {
	claudeRoot := t.TempDir()
	codexRoot := t.TempDir()

	repo := "/w/repo"
	brain := "/w/brain"
	claudeRepo := filepath.Join(claudeRoot, cwdToClaudeDir(repo))
	if err := os.Mkdir(claudeRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	day := filepath.Join(codexRoot, "2026", "06", "26")
	if err := os.MkdirAll(day, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, cwd string) {
		line := `{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"` + cwd + `"}}`
		if err := os.WriteFile(filepath.Join(day, name), []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("repo.jsonl", "/w/repo")
	write("brain.jsonl", "/w/brain")
	write("other.jsonl", "/w/other") // unrelated cwd → excluded

	hs := []Harness{ClaudeHarness(claudeRoot), CodexHarness(codexRoot)}
	got := Select([]string{brain, repo}, hs)

	want := Sources{
		Dirs: []string{claudeRepo}, // only repo has a Claude folder
		Files: []string{
			filepath.Join(day, "brain.jsonl"),
			filepath.Join(day, "repo.jsonl"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Select = %+v, want %+v", got, want)
	}
}
