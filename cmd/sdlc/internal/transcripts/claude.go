package transcripts

import (
	"os"
	"path/filepath"
	"strings"
)

// claudeHarness reads Claude Code's per-cwd transcript store at
// ~/.claude/projects, where each repo cwd maps to one folder of *.jsonl session
// files. So it contributes one Dir per existing cwd folder; the engine globs each
// for sessions.
type claudeHarness struct{ root string }

// ClaudeHarness returns the Claude Code transcript harness rooted at root
// (~/.claude/projects in production; a temp dir in tests).
func ClaudeHarness(root string) Harness { return claudeHarness{root: root} }

func (claudeHarness) Name() string { return "claude" }

// Sources returns the transcript dir for each cwd whose encoded folder exists
// under root — brain + repo only, never an unrelated concurrently-edited folder
// (those inflate the count, #68). Deliberately not "all folders."
func (h claudeHarness) Sources(cwds []string) Sources {
	if h.root == "" {
		return Sources{}
	}
	var out Sources
	seen := map[string]bool{}
	for _, cwd := range cwds {
		if cwd == "" {
			continue
		}
		dir := filepath.Join(h.root, cwdToClaudeDir(cwd))
		if seen[dir] {
			continue
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			out.Dirs = append(out.Dirs, dir)
			seen[dir] = true
		}
	}
	return out
}

// claudePathEncoder mirrors Claude Code's cwd→folder encoding: every '/' and '.'
// becomes '-'. e.g. /Users/x/workspace/nous → -Users-x-workspace-nous.
var claudePathEncoder = strings.NewReplacer("/", "-", ".", "-")

// cwdToClaudeDir encodes an absolute cwd to its ~/.claude/projects folder name.
// Pure — table-tested directly.
func cwdToClaudeDir(absPath string) string { return claudePathEncoder.Replace(absPath) }
