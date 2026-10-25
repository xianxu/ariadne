package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
	"go.yaml.in/yaml/v3"
)

// classifySideDoc distinguishes untyped evidence from a typed declaration
// whose syntax cannot reach the normal vocabulary validator. Parse failure
// must never turn a recognizable declaration into an attachment exemption.
func classifySideDoc(data []byte) (bool, error) {
	doc := string(data)
	first, rest, _ := strings.Cut(doc, "\n")
	if strings.TrimSpace(first) != "---" {
		return false, nil
	}
	block, _, closed := sidedoc.FrontBlock(doc)
	header := rest
	if closed {
		_, header, _ = strings.Cut(block, "\n")
		header = header[:strings.LastIndex(header, "---")]
	}
	var node yaml.Node
	parseErr := yaml.Unmarshal([]byte(header), &node)
	typed := false
	if parseErr == nil && len(node.Content) > 0 {
		fields := node.Content[0]
		if fields.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(fields.Content); i += 2 {
				key := fields.Content[i]
				if key.Kind == yaml.ScalarNode && key.Tag == "!!str" && key.Value == "type" {
					typed = true
				}
			}
		}
	} else {
		typed = declaresSideType(header, sideTypeKey)
	}
	if !typed {
		return false, nil
	}
	if !closed {
		return true, fmt.Errorf("unterminated typed frontmatter; add its closing --- delimiter")
	}
	if parseErr != nil {
		return true, fmt.Errorf("malformed typed frontmatter: %w", parseErr)
	}
	return true, nil
}

// sideTypeKey delegates quoted/escaped scalar syntax to YAML, independently
// of the malformed value. It does not validate the document or its schema.
func sideTypeKey(raw string) bool {
	var node yaml.Node
	if yaml.Unmarshal([]byte(strings.TrimSpace(raw)), &node) != nil || len(node.Content) == 0 {
		return false
	}
	n := node.Content[0]
	return n.Kind == yaml.ScalarNode && n.Tag == "!!str" && n.Value == "type"
}

// declaresSideType recognizes top-level block or flow mapping keys even if
// their values cannot parse. Each block key is examined once; nested keys and
// block-scalar text never opt an untyped evidence document into conformance.
func declaresSideType(header string, key func(string) bool) bool {
	flow := ""
	rootIndent, offset := -1, 0
	for _, raw := range strings.Split(header, "\n") {
		lineStart := offset
		offset += len(raw) + 1
		line := strings.TrimLeft(raw, " \t")
		indent := len(raw) - len(line)
		if line == "" || line[0] == '#' {
			continue
		}
		if rootIndent < 0 {
			if !strings.Contains(line, ":") && !strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "? ") {
				continue
			}
			rootIndent = indent
		}
		if indent != rootIndent {
			continue
		}
		if strings.HasPrefix(line, "{") {
			flow = header[lineStart+indent:]
			break
		}
		if explicit, ok := strings.CutPrefix(line, "? "); ok && key(explicit) {
			return true
		}
		if candidate, _, ok := strings.Cut(line, ":"); ok && key(candidate) {
			return true
		}
	}
	depth, start, wantKey := 0, 0, false
	var quote byte
	for i := 0; i < len(flow); i++ {
		c := flow[i]
		if quote != 0 {
			if quote == '"' && c == '\\' {
				i++
			} else if c == quote {
				if quote == '\'' && i+1 < len(flow) && flow[i+1] == '\'' {
					i++
				} else {
					quote = 0
				}
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '#':
			if i == 0 || strings.ContainsRune(" \t\r\n", rune(flow[i-1])) {
				for i < len(flow) && flow[i] != '\n' {
					i++
				}
			}
		case '{', '[':
			depth++
			if depth == 1 {
				start, wantKey = i+1, true
			}
		case '}', ']':
			depth--
			if depth == 0 {
				return false
			}
		case ',':
			if depth == 1 {
				start, wantKey = i+1, true
			}
		case ':':
			if depth == 1 && wantKey {
				if key(flow[start:i]) {
					return true
				}
				wantKey = false
			}
		}
	}
	return false
}
