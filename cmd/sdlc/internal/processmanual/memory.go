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
// live OUTSIDE the repo, at ~/.claude/projects/<slug>/memory. They are also
// PRIVATE + machine-specific (absolute home paths, personal content), so they are
// REDACTED by default — a written/committed manual must never carry them. `include`
// (the `--include-memory` flag, local-only) inlines them for the operator's own view.
func memorySources(homeDir, absRepoRoot string, include bool) []InjectionSource {
	if !include {
		return []InjectionSource{{
			Kind:  KindMemory,
			Title: "(persisted memories — not inlined)",
			When:  "persisted agent memories (Claude) — private, machine-local, outside the repo",
			Body: "Redacted by default: memories carry absolute home paths + personal content, " +
				"so they are never written into a shareable/committed manual. Run " +
				"`sdlc process-manual --include-memory` locally to inspect them — do not commit that output.",
		}}
	}
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
