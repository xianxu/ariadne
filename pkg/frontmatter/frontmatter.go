// Package frontmatter parses the one field weave's skill menu (and any
// DAG-aware consumer's prototype reader) needs out of a markdown file's leading
// YAML frontmatter block: the flat `description:` value. It lives at module
// level — not in cmd/weave/internal/walk — for the same import-boundary reason
// pkg/layergraph does: cmd/datatype's merge/list must read each prototype's
// `description:` and cannot reach into weave's internal packages, so the parser
// is shared once (ARCH-DRY — one source of truth for "parse a flat-YAML
// description:"). Pure string → string; no IO.
package frontmatter

import (
	"errors"
	"regexp"
	"strings"
)

// frontmatterRE matches "---\n<fm>\n---\n<body>" with <fm> + <body> captured.
// Multiline DOTALL via (?s). One source for the split (ARCH-DRY) —
// cmd/sdlc/internal/issue.Parse delegates here.
var frontmatterRE = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)

// Split separates a markdown document into its leading YAML frontmatter and its
// body. Returns an error when the document doesn't open with a "---\n…\n---\n"
// fence — the contract cmd/sdlc/internal/issue.Parse relies on (it now delegates
// here, so the parse lives once). Pure string → (string, string, error); no IO.
func Split(content string) (fm, body string, err error) {
	m := frontmatterRE.FindStringSubmatch(content)
	if m == nil {
		return "", "", errors.New("no YAML frontmatter")
	}
	return m[1], m[2], nil
}

// Description extracts the `description:` field from content's leading YAML
// frontmatter block (the `---` … `---` fence). It reads the one field the
// menu/merge need; a full YAML parse is overkill (frontmatter here is flat
// key: value). A surrounding pair of quotes on the value is stripped (some
// skills quote the description). Returns "" if no frontmatter fence or no
// description is present.
func Description(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "" // no frontmatter fence
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break // end of frontmatter
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "description" {
			continue
		}
		return unquote(strings.TrimSpace(val))
	}
	return ""
}

// unquote strips one symmetric pair of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
