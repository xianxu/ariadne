package issue

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// ParseDetails reads a details file for `sdlc issue new --details` (#371):
// `## <Section>` blocks only, each one of the model's sections, at most once.
// It returns each section's body, trimmed. Fenced code is body text, so a
// `## ` line inside a fence is not a heading. The input is untrusted, so any
// other structure is refused rather than guessed at: frontmatter, a `# `
// title (the title is the positional argument), text before the first
// section, an unknown or repeated section, an unclosed fence, an empty file.
func ParseDetails(text string) (map[string]string, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("the details file is empty")
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	fenced := FenceSpans(lines, UnterminatedIsFenced)
	if !slices.Equal(fenced, FenceSpans(lines, UnterminatedIsProse)) {
		return nil, fmt.Errorf("the details file has a code fence that never closes")
	}
	if strings.TrimSpace(lines[0]) == "---" {
		return nil, fmt.Errorf("the details file has frontmatter; it holds only `## <Section>` blocks (sdlc writes the frontmatter)")
	}
	known := map[string]bool{}
	var names []string
	for _, s := range vocab.Issue().Sections() {
		known[s.Name] = true
		names = append(names, s.Name)
	}
	out := map[string]string{}
	var cur string
	var body []string
	flush := func() {
		if cur != "" {
			out[cur] = strings.TrimSpace(strings.Join(body, "\n"))
		}
	}
	for i, line := range lines {
		if fenced[i] {
			if cur == "" {
				return nil, fmt.Errorf("line %d: text before the first `## <Section>` heading", i+1)
			}
			body = append(body, line)
			continue
		}
		switch {
		case strings.HasPrefix(line, "# "):
			return nil, fmt.Errorf("line %d: a `# ` title; the title is the positional argument, and the file holds only `## <Section>` blocks", i+1)
		case strings.HasPrefix(line, "## "):
			flush()
			name := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if !known[name] {
				return nil, fmt.Errorf("line %d: `## %s` is not an issue section (one of: %s)", i+1, name, strings.Join(names, ", "))
			}
			if _, dup := out[name]; dup {
				return nil, fmt.Errorf("line %d: `## %s` appears twice", i+1, name)
			}
			cur, body = name, nil
			out[name] = ""
		case cur == "" && strings.TrimSpace(line) != "":
			return nil, fmt.Errorf("line %d: text before the first `## <Section>` heading", i+1)
		default:
			body = append(body, line)
		}
	}
	flush()
	return out, nil
}
