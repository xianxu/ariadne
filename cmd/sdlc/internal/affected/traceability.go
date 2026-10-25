package affected

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

// TraceabilityFile is one declared traceability file's base copy (#616): its repository path and content.
type TraceabilityFile struct {
	Path    string
	Content []byte
}

// TraceEntry is one feature entry of a traceability file: a change to a code path counts as a change to every test.
type TraceEntry struct {
	Key   string
	Code  []string
	Tests []string
}

// ParseTraceability parses the parley-style format (#616):
//
//	version: 1
//	atlas:
//	  <key>: {code: [path or glob, ...], tests: [path, ...]}
//
// It walks a yaml.Node tree, so every check is the walk's own: decoding into a Node skips yaml.v3's decode-time
// checks, duplicate keys included. One document, exact keys, string scalars, clean paths; aliases, anchors, merge
// keys and duplicates refuse. An entry with no code or no tests declares nothing (its owner's tooling gives it meaning).
func ParseTraceability(data []byte) ([]TraceEntry, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var doc yaml.Node
	if err := dec.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty document")
		}
		return nil, err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("more than one document")
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 {
		return nil, errors.New("not one document")
	}
	root, err := mappingPairs(doc.Content[0], "the document")
	if err != nil {
		return nil, err
	}
	var version, atlas *yaml.Node
	for _, kv := range root {
		switch kv[0].Value {
		case "version":
			version = kv[1]
		case "atlas":
			atlas = kv[1]
		default:
			return nil, fmt.Errorf("unknown key %q", kv[0].Value)
		}
	}
	if version == nil || version.Kind != yaml.ScalarNode || version.ShortTag() != "!!int" || version.Value != "1" {
		return nil, errors.New("version must be 1")
	}
	if atlas == nil {
		return nil, errors.New("atlas is missing")
	}
	entries, err := mappingPairs(atlas, "atlas")
	if err != nil {
		return nil, err
	}
	var out []TraceEntry
	for _, kv := range entries {
		key := kv[0].Value
		fields, err := mappingPairs(kv[1], "atlas."+key)
		if err != nil {
			return nil, err
		}
		e := TraceEntry{Key: key}
		for _, f := range fields {
			var dst *[]string
			glob := false
			switch f[0].Value {
			case "code":
				dst, glob = &e.Code, true
			case "tests":
				dst = &e.Tests
			default:
				return nil, fmt.Errorf("atlas.%s: unknown key %q", key, f[0].Value)
			}
			if *dst, err = pathList(f[1], "atlas."+key+"."+f[0].Value, glob); err != nil {
				return nil, err
			}
		}
		if len(e.Code) > 0 && len(e.Tests) > 0 {
			out = append(out, e)
		}
	}
	return out, nil
}

// mappingPairs yields a mapping node's key/value pairs, refusing anything the format does not allow at any level: a
// non-mapping, an alias or anchor, a non-string or merge key, and a duplicate key.
func mappingPairs(n *yaml.Node, where string) ([][2]*yaml.Node, error) {
	if err := plainNode(n, where); err != nil {
		return nil, err
	}
	if n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s must be a mapping", where)
	}
	seen := map[string]bool{}
	var out [][2]*yaml.Node
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		if err := plainNode(k, where); err != nil {
			return nil, err
		}
		if k.Kind != yaml.ScalarNode || k.ShortTag() != "!!str" || k.Value == "" {
			return nil, fmt.Errorf("%s: keys must be non-empty strings", where)
		}
		if seen[k.Value] {
			return nil, fmt.Errorf("%s: duplicate key %q", where, k.Value)
		}
		seen[k.Value] = true
		out = append(out, [2]*yaml.Node{k, v})
	}
	return out, nil
}

// pathList is a sequence of clean repository-relative path strings; globs only where the field allows them.
func pathList(n *yaml.Node, where string, glob bool) ([]string, error) {
	if err := plainNode(n, where); err != nil {
		return nil, err
	}
	if n.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s must be a list", where)
	}
	out := make([]string, 0, len(n.Content))
	for _, s := range n.Content {
		if err := plainNode(s, where); err != nil {
			return nil, err
		}
		if s.Kind != yaml.ScalarNode || s.ShortTag() != "!!str" {
			return nil, fmt.Errorf("%s: entries must be strings", where)
		}
		ok := cleanRelPath(s.Value) && (glob && validGlob(s.Value) || !glob && !hasGlobMeta(s.Value))
		if !ok {
			return nil, fmt.Errorf("%s: %q must be a clean repository-relative path", where, s.Value)
		}
		out = append(out, s.Value)
	}
	return out, nil
}

// plainNode refuses aliases and anchors: an anchored subtree would let one entry silently restate another.
func plainNode(n *yaml.Node, where string) error {
	if n.Kind == yaml.AliasNode || n.Anchor != "" {
		return fmt.Errorf("%s: aliases and anchors are not allowed", where)
	}
	if n.ShortTag() == "!!merge" {
		return fmt.Errorf("%s: merge keys are not allowed", where)
	}
	return nil
}
