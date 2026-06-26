package transcripts

import (
	"os"
	"path/filepath"
)

// DefaultHarnesses returns the production harness registry, rooted at the real
// per-user transcript stores. The one place the live roots are wired — tests
// construct harnesses with temp roots and call Select directly. Adding a new
// agent CLI is one entry here plus its <harness>.go file.
func DefaultHarnesses() []Harness {
	return []Harness{
		ClaudeHarness(defaultClaudeRoot()),
		CodexHarness(defaultCodexRoot()),
	}
}

// defaultClaudeRoot is ~/.claude/projects (Claude Code's per-cwd transcript store).
func defaultClaudeRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// defaultCodexRoot is ~/.codex/sessions (Codex's date-sharded transcript store).
func defaultCodexRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex", "sessions")
}
