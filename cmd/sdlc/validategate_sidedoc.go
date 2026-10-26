package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
	"go.yaml.in/yaml/v3"
)

// classifySideDoc exempts untyped evidence only after its leading frontmatter
// has parsed. Malformed input cannot prove the absence of a type declaration;
// one YAML decoder owns syntax, including tags, explicit keys and duplicates.
func classifySideDoc(data []byte) (bool, error) {
	doc := string(data)
	first, _, _ := strings.Cut(doc, "\n")
	if strings.TrimSpace(first) != "---" {
		return false, nil
	}
	block, _, closed := sidedoc.FrontBlock(doc)
	if !closed {
		return false, fmt.Errorf("ambiguous frontmatter: missing closing --- delimiter; fix the header or remove its delimiters for an untyped attachment")
	}
	_, header, _ := strings.Cut(block, "\n")
	header = header[:strings.LastIndex(header, "---")]
	decoder := yaml.NewDecoder(strings.NewReader(header))
	var value any
	if err := decoder.Decode(&value); err != nil && err != io.EOF {
		return false, fmt.Errorf("ambiguous malformed frontmatter: %w; fix the header or remove its delimiters for an untyped attachment", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return false, fmt.Errorf("ambiguous frontmatter: expected one YAML document; fix the header or remove its delimiters for an untyped attachment")
	}
	switch fields := value.(type) {
	case map[string]any:
		_, typed := fields["type"]
		return typed, nil
	case map[any]any:
		_, typed := fields["type"]
		return typed, nil
	default:
		return false, nil
	}
}
