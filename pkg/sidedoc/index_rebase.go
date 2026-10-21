package sidedoc

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// RebaseIndex relocates the top-level discussions index from document from to
// document to. Coordinates and moved's input/output are slash paths relative to
// one repository root. An entry can ascend from its document, but must resolve
// inside that root. Only entry scalar bytes change; other metadata stays intact.
// An unreadable index returns the original text and an error.
func RebaseIndex(text, from, to string, moved func(string) (string, bool)) (string, error) {
	if !indexRepoPath(from) || !indexRepoPath(to) {
		return text, fmt.Errorf("discussions: invalid document coordinates %q -> %q", from, to)
	}
	lines, _, ok := frontmatter(text)
	if !ok {
		return text, errFrontmatter(text)
	}
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &document); err != nil {
		return text, fmt.Errorf("discussions: invalid YAML: %w", err)
	}
	if len(document.Content) == 0 {
		return text, nil
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return text, fmt.Errorf("discussions: frontmatter must be a mapping")
	}
	var index *yaml.Node
	for i := 0; i < len(root.Content); i += 2 {
		key := root.Content[i]
		if key.Tag == "!!merge" {
			return text, fmt.Errorf("discussions: inherited top-level fields cannot be rebased")
		}
		if key.Value != "discussions" {
			continue
		}
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || index != nil {
			return text, fmt.Errorf("discussions: duplicate or invalid key")
		}
		index = root.Content[i+1]
	}
	if index == nil {
		return text, nil
	}
	if index.Kind != yaml.SequenceNode {
		return text, fmt.Errorf("discussions: want a list of doc mappings")
	}
	type edit struct {
		start, end int
		value      string
	}
	var edits []edit
	// YAML positions refer to the frontmatter without its opening fence.
	offsets := make([]int, len(lines))
	at := len("---\n")
	for i, line := range lines {
		offsets[i] = at
		at += len(line) + 1
	}
	for _, item := range index.Content {
		if item.Kind != yaml.MappingNode {
			return text, fmt.Errorf("discussions: want doc mappings")
		}
		var scalar *yaml.Node
		for i := 0; i < len(item.Content); i += 2 {
			key := item.Content[i]
			if key.Tag == "!!merge" {
				return text, fmt.Errorf("discussions: inherited doc fields cannot be rebased")
			}
			if key.Value != "doc" {
				continue
			}
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || scalar != nil {
				return text, fmt.Errorf("discussions: duplicate or invalid doc key")
			}
			scalar = item.Content[i+1]
		}
		if scalar == nil || scalar.Kind != yaml.ScalarNode || scalar.Tag != "!!str" {
			return text, fmt.Errorf("discussions: doc must be a path string")
		}
		p := scalar.Value
		if !plainPathRE.MatchString(p) || p != strings.TrimSpace(p) || path.IsAbs(p) {
			return text, fmt.Errorf("discussions: invalid relative path %q", p)
		}
		target := path.Join(path.Dir(from), p)
		if !indexRepoPath(target) {
			return text, fmt.Errorf("discussions: path %q escapes repository", p)
		}
		if moved != nil {
			if destination, ok := moved(target); ok {
				target = destination
			}
		}
		if !indexRepoPath(target) {
			return text, fmt.Errorf("discussions: invalid moved target %q", target)
		}
		replacement := RelPath(path.Dir(to), target)
		start, end, err := indexScalarSpan(lines, offsets, scalar)
		if err != nil {
			return text, err
		}
		if replacement == p {
			continue
		}
		switch scalar.Style {
		case yaml.SingleQuotedStyle:
			replacement = "'" + strings.ReplaceAll(replacement, "'", "''") + "'"
		case yaml.DoubleQuotedStyle:
			replacement = strconv.Quote(replacement)
		case 0:
			if !plainPathRE.MatchString(replacement) {
				return text, fmt.Errorf("discussions: rebased path %q cannot remain a plain scalar", replacement)
			}
		default:
			return text, fmt.Errorf("discussions: unsupported path scalar style")
		}
		edits = append(edits, edit{start, end, replacement})
	}
	// YAML visits entries in source order; apply backwards to keep offsets stable.
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		text = text[:e.start] + e.value + text[e.end:]
	}
	return text, nil
}

func indexRepoPath(p string) bool {
	clean := path.Clean(p)
	return p != "" && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../") && !path.IsAbs(p) && !strings.ContainsAny(p, "\\\x00\r\n")
}

// indexScalarSpan isolates one single-line scalar using YAML's source position,
// then checks that the isolated token has exactly the value YAML parsed. This
// rejects multiline scalars without touching neighboring comments or whitespace.
func indexScalarSpan(lines []string, offsets []int, n *yaml.Node) (int, int, error) {
	if n.Line < 1 || n.Line > len(lines) || n.Style != 0 && n.Style != yaml.SingleQuotedStyle && n.Style != yaml.DoubleQuotedStyle {
		return 0, 0, fmt.Errorf("discussions: doc must be a single-line path scalar")
	}
	line := lines[n.Line-1]
	column := 0
	for i := 1; i < n.Column && column < len(line); i++ {
		_, width := utf8.DecodeRuneInString(line[column:])
		column += width
	}
	raw := line[column:]
	end := len(raw)
	switch n.Style {
	case yaml.SingleQuotedStyle:
		end = 0
		for i := 1; i < len(raw); i++ {
			if raw[i] == '\'' {
				if i+1 < len(raw) && raw[i+1] == '\'' {
					i++
					continue
				}
				end = i + 1
				break
			}
		}
	case yaml.DoubleQuotedStyle:
		end = 0
		for i := 1; i < len(raw); i++ {
			if raw[i] == '\\' {
				i++
				continue
			}
			if raw[i] == '"' {
				end = i + 1
				break
			}
		}
	default:
		for i := 0; i < len(raw); i++ {
			if raw[i] == '#' && (i == 0 || raw[i-1] == ' ' || raw[i-1] == '\t') || raw[i] == ',' || raw[i] == ']' || raw[i] == '}' {
				end = i
				break
			}
		}
		end = len(strings.TrimRight(raw[:end], " \t"))
	}
	var token yaml.Node
	if end == 0 || yaml.Unmarshal([]byte(raw[:end]), &token) != nil || len(token.Content) != 1 || token.Content[0].Kind != yaml.ScalarNode || token.Content[0].Tag != "!!str" || token.Content[0].Value != n.Value {
		return 0, 0, fmt.Errorf("discussions: doc must be a single-line path scalar")
	}
	start := offsets[n.Line-1] + column
	return start, start + end, nil
}
