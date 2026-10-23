package sidedoc

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmtext "github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

// ChatSection is one chat section of a document (collaboration-modes.md §3,
// #575), in 1-based lines of the whole document: its heading's line (0 for a
// whole-document chat) through End, the line before the next heading of the
// same or a higher level. Source says what made it chat: "tag" (`@@chat@@`)
// or a legacy form, "legacy-scalar", "legacy-map" or "legacy-comment".
type ChatSection struct {
	Heading, End int
	Source       string
}

// ChatScan is what a round reads from a document's structure: its chat
// sections, the lines of `@@chat@@` tags that sit above no heading (ordinary
// text, named in the reply), and one notice per legacy or retired mode key,
// which a round reports and never fails on.
type ChatScan struct {
	Sections  []ChatSection
	Misplaced []int
	Notices   []string
}

// InChat reports whether a 1-based line lies inside a chat section.
func (s ChatScan) InChat(line int) bool {
	for _, c := range s.Sections {
		if line > c.Heading && line <= c.End {
			return true
		}
	}
	return false
}

// RetiredKeys are the mode keys a round reads around and never writes (#575).
var RetiredKeys = []string{"presentation", "presentation-mode", "process-instruction", "processing-instruction", "processing-mode"}

const chatTag = "@@chat@@"

var modeComment = regexp.MustCompile(`^<!--\s*(.*?)\s*-->\s*$`)

// ChatSections scans a document for its chat sections: a `@@chat@@` line
// directly above an ATX or setext heading, outside code, makes that section
// chat; so do the legacy forms that meant chat (a `collaboration-mode: chat`
// scalar, a section map's chat names, a heading comment). Every other mode
// key or value is retired: noticed, ignored. Headings come from goldmark, so a
// fenced "heading" is code. It is pure; a malformed legacy key degrades to a
// notice rather than an error, because a mode key never fails a round.
func ChatSections(doc string) ChatScan {
	var scan ChatScan
	lines := strings.Split(doc, "\n")
	total := len(lines)
	if strings.HasSuffix(doc, "\n") {
		total--
	}
	fm, body, hasFM := frontmatter(doc)
	bodyStart := 1 // the body's first line, 1-based
	if hasFM {
		bodyStart = len(fm) + 3 // the two fences and the frontmatter lines
	}
	scalarChat, mapNames := scan.readFrontmatter(fm)

	type heading struct {
		line, level int
		name        string
	}
	var headings []heading
	src := []byte(body)
	lineOf := func(off int) int { return bodyStart + strings.Count(body[:off], "\n") }
	root := goldmark.DefaultParser().Parse(gmtext.NewReader(src))
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok || h.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		seg := h.Lines().At(0)
		headings = append(headings, heading{lineOf(seg.Start), h.Level, string(seg.Value(src))})
		return ast.WalkSkipChildren, nil
	})
	isHeading := map[int]bool{}
	for _, h := range headings {
		isHeading[h.line] = true
	}
	end := func(i int) int {
		for _, next := range headings[i+1:] {
			if next.level <= headings[i].level {
				return next.line - 1
			}
		}
		return total
	}

	literal, _ := CodeLiteralMask(body)
	off := 0
	for i := bodyStart - 1; i < len(lines); i++ {
		l := lines[i]
		lead := len(l) - len(strings.TrimLeft(l, " \t"))
		if strings.TrimSpace(l) == chatTag && !(off+lead < len(literal) && literal[off+lead]) && !isHeading[i+2] {
			scan.Misplaced = append(scan.Misplaced, i+1)
		}
		off += len(l) + 1
	}

	if scalarChat {
		scan.Sections = append(scan.Sections, ChatSection{0, total, "legacy-scalar"})
	}
	matched := map[string]bool{}
	for i, h := range headings {
		source := ""
		if h.line >= 2 && strings.TrimSpace(lines[h.line-2]) == chatTag {
			source = "tag"
		}
		if h.line < len(lines) {
			if m := modeComment.FindStringSubmatch(lines[h.line]); m != nil {
				if chat, mode := scan.readComment(m[1], h.line+1, strings.TrimSpace(lines[h.line])); mode && chat && source == "" {
					source = "legacy-comment"
				}
			}
		}
		for _, n := range mapNames {
			if normalizeHeading(n) == normalizeHeading(h.name) {
				matched[n] = true
				if source == "" {
					source = "legacy-map"
				}
			}
		}
		if source != "" && !scalarChat {
			scan.Sections = append(scan.Sections, ChatSection{h.line, end(i), source})
		}
	}
	for _, n := range mapNames {
		if !matched[n] {
			scan.Notices = append(scan.Notices, fmt.Sprintf("collaboration-mode map name %q matches no heading", n))
		}
	}
	return scan
}

// readFrontmatter reads the legacy mode keys: whether a scalar or map default
// makes the whole document chat, and the map's chat names as written.
func (s *ChatScan) readFrontmatter(fm []string) (wholeChat bool, names []string) {
	if len(fm) == 0 {
		return false, nil
	}
	var front map[string]yaml.Node
	if yaml.Unmarshal([]byte(strings.Join(fm, "\n")), &front) != nil {
		return false, nil // not this scanner's to judge; the frontmatter readers report it
	}
	if cm, ok := front["collaboration-mode"]; ok {
		switch cm.Kind {
		case yaml.ScalarNode:
			if cm.Value == "chat" {
				wholeChat = true
				s.Notices = append(s.Notices, "collaboration-mode: chat is a legacy key; read as @@chat@@ for the whole document")
			} else {
				s.Notices = append(s.Notices, fmt.Sprintf("collaboration-mode: %s is retired and ignored: one collaborative mode applies", cm.Value))
			}
		case yaml.MappingNode:
			s.Notices = append(s.Notices, "collaboration-mode map is a legacy key; its chat sections are read as @@chat@@")
			for i := 0; i+1 < len(cm.Content); i += 2 {
				k, v := cm.Content[i].Value, cm.Content[i+1]
				switch {
				case k == "default" && v.Value == "chat":
					wholeChat = true
				case k == "chat":
					for _, item := range append([]*yaml.Node{v}, v.Content...) {
						if item.Kind == yaml.ScalarNode && item.Value != "" {
							names = append(names, item.Value)
						}
					}
				}
			}
		default:
			s.Notices = append(s.Notices, "collaboration-mode has an unreadable value; ignored")
		}
	}
	var retired []string
	for _, k := range RetiredKeys {
		if _, ok := front[k]; ok {
			retired = append(retired, k)
		}
	}
	if len(retired) > 0 {
		s.Notices = append(s.Notices, "retired keys ignored: "+strings.Join(retired, ", "))
	}
	return wholeChat, names
}

// readComment reads a legacy heading comment (`<!-- key: value; … -->`).
// mode is false for an ordinary comment, which stays text.
func (s *ChatScan) readComment(inner string, line int, raw string) (chat, mode bool) {
	for _, part := range strings.Split(inner, ";") {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k == "collaboration-mode" || slices.Contains(RetiredKeys, k) {
			mode = true
			chat = chat || k == "collaboration-mode" && v == "chat"
		}
	}
	if mode {
		if chat {
			s.Notices = append(s.Notices, fmt.Sprintf("%s at line %d is a legacy key; read as @@chat@@", raw, line))
		} else {
			s.Notices = append(s.Notices, fmt.Sprintf("%s at line %d is retired and ignored", raw, line))
		}
	}
	return chat, mode
}

var trailingAnnotation = regexp.MustCompile(`\s*\([^()]*\)$`)

// normalizeHeading is the legacy map's name match: lowercase, whitespace runs
// collapsed, trimmed, one trailing parenthesised annotation dropped.
func normalizeHeading(s string) string {
	s = strings.Join(strings.Fields(strings.ToLower(s)), " ")
	return strings.TrimSpace(trailingAnnotation.ReplaceAllString(s, ""))
}
