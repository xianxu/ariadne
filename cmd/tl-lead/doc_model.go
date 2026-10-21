package main

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
	"go.yaml.in/yaml/v3"
)

type settlementPointer struct{ Path, Commit string }
type workingdocMetadata struct {
	Owner, Topic                       string
	Pointer                            *settlementPointer
	pointerStart, pointerEnd, frontEnd int
}

var settlementObjectID = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func workingdocPath(p string) bool {
	return p != "" && path.Clean(p) == p && !strings.HasPrefix(p, "/") && p != ".." && !strings.HasPrefix(p, "../") && !strings.HasPrefix(p, "-") && !strings.ContainsAny(p, "\\\x00\r\n\t") && strings.HasSuffix(p, ".md") && !sidedoc.InSideDir(p)
}

func parseSettlementPointer(raw string) (settlementPointer, error) {
	p, sha, ok := strings.Cut(raw, " @ ")
	if !ok || !workingdocPath(p) || !settlementObjectID.MatchString(sha) {
		return settlementPointer{}, fmt.Errorf("invalid settled-into pointer %q", raw)
	}
	return settlementPointer{p, sha}, nil
}

func parseWorkingdoc(text string) (workingdocMetadata, error) {
	m := workingdocMetadata{pointerStart: -1}
	front, _, ok := sidedoc.FrontBlock(text)
	if !ok {
		return m, fmt.Errorf("working document requires closed YAML frontmatter")
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(front), &doc); err != nil {
		return m, fmt.Errorf("working document frontmatter: %w", err)
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return m, fmt.Errorf("working document frontmatter must be a mapping")
	}
	lines := strings.SplitAfter(front, "\n")
	offsets := make([]int, len(lines)+1)
	for i, l := range lines {
		offsets[i+1] = offsets[i] + len(l)
	}
	// FrontBlock's closing delimiter is the insertion point, independent of
	// a following blank line or the body's line-ending convention.
	m.frontEnd = strings.LastIndex(front, "---")
	fields := map[string]bool{}
	nodes := doc.Content[0].Content
	for i := 0; i < len(nodes); i += 2 {
		k, v := nodes[i], nodes[i+1]
		switch k.Value {
		case "type", "owner", "topic", "settled-into":
		default:
			continue
		}
		if fields[k.Value] {
			return m, fmt.Errorf("duplicate %s field", k.Value)
		}
		fields[k.Value] = true
		if v.Kind != yaml.ScalarNode || v.Tag != "!!str" {
			return m, fmt.Errorf("%s must be a string scalar", k.Value)
		}
		switch k.Value {
		case "type":
			if v.Value != "side-doc" && v.Value != "workingdoc" {
				return m, fmt.Errorf("type %q is not a side-doc or workingdoc", v.Value)
			}
		case "owner":
			m.Owner = v.Value
		case "topic":
			m.Topic = v.Value
		case "settled-into":
			p, err := parseSettlementPointer(v.Value)
			if err != nil {
				return m, err
			}
			if k.Line < 1 || k.Line > len(lines) {
				return m, fmt.Errorf("invalid settled-into location")
			}
			line := lines[k.Line-1]
			if strings.TrimRight(line, "\r\n") != "settled-into: "+v.Value || v.Style != 0 || k.Style != 0 || v.Anchor != "" || v.HeadComment != "" || v.LineComment != "" {
				return m, fmt.Errorf("settled-into must use one plain uncommented scalar line")
			}
			m.Pointer = &p
			m.pointerStart = offsets[k.Line-1]
			m.pointerEnd = offsets[k.Line]
		}
	}
	return m, nil
}

func stampWorkingdoc(text string, p settlementPointer) (string, error) {
	if _, err := parseSettlementPointer(p.Path + " @ " + p.Commit); err != nil {
		return "", err
	}
	m, err := parseWorkingdoc(text)
	if err != nil {
		return "", err
	}
	eol := "\n"
	if strings.HasPrefix(text, "---\r\n") {
		eol = "\r\n"
	}
	line := "settled-into: " + p.Path + " @ " + p.Commit + eol
	var stamped string
	if m.Pointer != nil {
		stamped = text[:m.pointerStart] + line + text[m.pointerEnd:]
	} else {
		stamped = text[:m.frontEnd] + line + text[m.frontEnd:]
	}
	// Paths valid for Git are not necessarily representable in the managed
	// plain YAML scalar. Validate the complete result with the same parser
	// that will consume it before the caller publishes or stamps anything.
	got, err := parseWorkingdoc(stamped)
	if err != nil {
		return "", fmt.Errorf("cannot represent settled-into pointer for %q: %w", p.Path, err)
	}
	if got.Pointer == nil || *got.Pointer != p {
		return "", fmt.Errorf("settled-into pointer for %q did not round-trip", p.Path)
	}
	return stamped, nil
}

type settlementPhase uint8

const (
	settlementPlanned settlementPhase = iota
	settlementPublicationConfirmed
	settlementDone
	settlementStopped
)

type settlementEvent uint8

const (
	settlementPublished settlementEvent = iota
	settlementStamped
	settlementFailed
	settlementUnconfirmed
)

type settlementState struct{ Phase settlementPhase }

func transitionSettlement(s settlementState, e settlementEvent) (settlementState, error) {
	if s.Phase >= settlementDone {
		return s, fmt.Errorf("settlement already terminal")
	}
	if e == settlementFailed || e == settlementUnconfirmed {
		s.Phase = settlementStopped
		return s, nil
	}
	switch {
	case s.Phase == settlementPlanned && e == settlementPublished:
		s.Phase = settlementPublicationConfirmed
	case s.Phase == settlementPublicationConfirmed && e == settlementStamped:
		s.Phase = settlementDone
	default:
		return s, fmt.Errorf("invalid settlement transition %d/%d", s.Phase, e)
	}
	return s, nil
}

func classifyWorkingdoc(source, baseline, target []byte) string {
	if !bytes.Equal(target, baseline) {
		return "conflict"
	}
	if !bytes.Equal(source, baseline) {
		return "drift"
	}
	return "settled"
}
