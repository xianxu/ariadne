package processmanual

import (
	"os"
	"path/filepath"
	"strings"
)

// claudeProjectSlug maps an absolute repo path to the Claude harness
// project-dir slug: every "/" becomes "-" (pure). e.g.
// /Users/x/workspace/ariadne → -Users-x-workspace-ariadne.
func claudeProjectSlug(absRepoRoot string) string {
	return strings.ReplaceAll(absRepoRoot, "/", "-")
}

// memorySources is best-effort: persisted agent memories are Claude-specific and
// live OUTSIDE the repo, at ~/.claude/projects/<slug>/memory. When present, each
// memory file becomes a record with an absolute (outside-repo) Link; when absent,
// a single note surfaces the documented blind spot rather than silently dropping
// it (#153).
func memorySources(homeDir, absRepoRoot string) []InjectionSource {
	memDir := filepath.Join(homeDir, ".claude", "projects", claudeProjectSlug(absRepoRoot), "memory")
	entries, err := os.ReadDir(memDir)
	if err == nil {
		var out []InjectionSource
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			full := filepath.Join(memDir, e.Name())
			content, _ := os.ReadFile(full)
			out = append(out, InjectionSource{
				Kind:  KindMemory,
				Title: e.Name(),
				When:  "persisted agent memory (Claude, injected at session start; outside the repo)",
				Link:  full, // absolute — renderManual leaves outside-repo links untouched
				Body:  firstParagraph(string(content)),
			})
		}
		if len(out) > 0 {
			return out
		}
	}
	return []InjectionSource{{
		Kind:  KindMemory,
		Title: "(none)",
		When:  "persisted agent memories are Claude-specific and live outside the repo",
		Body: "No persisted memories found for this repo (Claude harness project dir absent " +
			"or empty). Documented blind spot (#153): memories are agent-specific and outside " +
			"the repo tree, so they are located by convention, not parsed from the repo.",
	}}
}
