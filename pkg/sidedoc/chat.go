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
//
// Legacy is true when any legacy form (scalar, map name, heading comment)
// also makes the section chat, whatever Source names; it selects the legacy
// item grammar (`-`, `*` and `N.` items) over the tag's `N.` items alone
// (collaboration-modes.md §3, §7). Sections are laminar: two sections are
// disjoint or one contains the other, and the innermost one containing a line
// governs it (Innermost).
type ChatSection struct {
	Heading, End int
	Source       string
	Legacy       bool
}

// ChatScan is what a round reads from a document's structure: its chat
// sections, the lines of `@@chat@@` tags that sit above no heading (ordinary
// text, named in the reply), and one notice per legacy or retired mode key,
// which a round reports and never fails on.
//
// Tags are the lines of every valid `@@chat@@` tag (outside code, directly
// above an ATX heading; its heading is the next line), reported even when a
// whole-document legacy chat emits no section of its own for it, so a
// consumer can protect each tag and its heading as operator-owned without
// recognizing tags itself. Misplaced and Tags never share a line.
type ChatScan struct {
	Sections  []ChatSection
	Tags      []int
	Misplaced []int
	Notices   []string
}

// Innermost returns the chat section governing a 1-based line: the
// innermost one whose extent holds it.
func (s ChatScan) Innermost(line int) (ChatSection, bool) {
	var best ChatSection
	found := false
	for _, c := range s.Sections {
		if line > c.Heading && line <= c.End && (!found || c.Heading > best.Heading) {
			best, found = c, true
		}
	}
	return best, found
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

const unreadFrontmatter = "frontmatter could not be read, so any legacy mode keys in it were not read"

var modeComment = regexp.MustCompile(`^<!--\s*(.*?)\s*-->\s*$`)

// ChatSections scans a document for its chat sections: a `@@chat@@` line
// directly above an ATX heading, outside fenced or indented code, makes that section
// chat; so do the legacy forms that meant chat (a `collaboration-mode: chat`
// scalar, a section map's chat names, a heading comment). Every other mode
// key or value is retired: noticed, ignored. Headings come from goldmark, so a
// fenced "heading" is code. It is pure; a malformed legacy key degrades to a
// notice rather than an error, because a mode key never fails a round.
func ChatSections(doc string) ChatScan {
	scan, _ := chatSections(doc)
	return scan
}

// chatSections is ChatSections with its work count: lines and bytes scanned,
// headings and map names visited, and section-stack operations, plus
// CodeLiteralMask's own count. It is linear in the document (pair#838 runs it
// once per snapshot); goldmark's parse is outside the count, as in
// CodeLiteralMask.
func chatSections(doc string) (ChatScan, int) {
	var scan ChatScan
	lines := strings.Split(doc, "\n")
	work := len(lines)
	total := len(lines)
	if strings.HasSuffix(doc, "\n") {
		total--
	}
	fm, body, hasFM := frontmatter(doc)
	// One rule for every way frontmatter goes unread (CR line endings, a
	// document marker or directive inside, YAML that does not parse): when a
	// closed leading fence block exists but was not read, say so, so a legacy
	// chat never vanishes silently. A document that merely opens with a
	// thematic break has no such block and gets no notice.
	if _, _, fenced := FrontBlock(doc); fenced && !hasFM {
		scan.Notices = append(scan.Notices, unreadFrontmatter)
	}
	bodyStart := 1 // the body's first line, 1-based
	if hasFM {
		bodyStart = len(fm) + 3 // the two fences and the frontmatter lines
	}
	scalarChat, mapNames := scan.readFrontmatter(fm)
	work += len(mapNames)
	wanted := map[string][]string{} // normalized name → the names as written
	for _, n := range mapNames {
		k := normalizeHeading(n)
		wanted[k] = append(wanted[k], n)
	}

	type heading struct {
		line, level int
		name        string
	}
	var headings []heading
	src := []byte(body)
	// Headings arrive in source order, so one cursor counts each newline once.
	cursor, cursorLine := 0, bodyStart
	lineOf := func(off int) int {
		cursorLine += strings.Count(body[cursor:off], "\n")
		work += off - cursor
		cursor = off
		return cursorLine
	}
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
	// Each section ends before the next heading of the same or a higher
	// level: one stack pass, each heading pushed and popped once.
	ends := make([]int, len(headings))
	var open []int
	for i, h := range headings {
		for len(open) > 0 && headings[open[len(open)-1]].level >= h.level {
			ends[open[len(open)-1]] = h.line - 1
			open = open[:len(open)-1]
			work++
		}
		open = append(open, i)
		work++
	}
	for _, i := range open {
		ends[i] = total
	}

	// A tag is a line holding only @@chat@@ outside code (fenced or indented),
	// and it counts only directly above an ATX heading: CommonMark folds a
	// line above a setext heading into that heading's text.
	literal, maskWork := CodeLiteralMask(body)
	work += maskWork
	tagLine := map[int]bool{}
	off := 0
	for i := bodyStart - 1; i < len(lines); i++ {
		l := lines[i]
		lead := len(l) - len(strings.TrimLeft(l, " \t"))
		if strings.TrimSpace(l) == chatTag && !(off+lead < len(literal) && literal[off+lead]) {
			tagLine[i+1] = true
		}
		off += len(l) + 1
	}
	atx := func(line int) bool {
		return isHeading[line] && strings.HasPrefix(strings.TrimLeft(lines[line-1], " "), "#")
	}
	for i := bodyStart; i <= len(lines); i++ {
		switch {
		case tagLine[i] && atx(i+1):
			scan.Tags = append(scan.Tags, i)
		case tagLine[i]:
			scan.Misplaced = append(scan.Misplaced, i)
		}
	}

	if scalarChat {
		scan.Sections = append(scan.Sections, ChatSection{0, total, "legacy-scalar", true})
	}
	matched := map[string]bool{}
	for i, h := range headings {
		source, legacy := "", false
		if tagLine[h.line-1] && atx(h.line) {
			source = "tag"
		}
		if h.line < len(lines) {
			if m := modeComment.FindStringSubmatch(lines[h.line]); m != nil {
				if chat, mode := scan.readComment(m[1], h.line+1, strings.TrimSpace(lines[h.line])); mode && chat {
					legacy = true
					if source == "" {
						source = "legacy-comment"
					}
				}
			}
		}
		if names, ok := wanted[normalizeHeading(h.name)]; ok {
			for _, n := range names {
				matched[n] = true
			}
			legacy = true
			if source == "" {
				source = "legacy-map"
			}
		}
		if source != "" && !scalarChat {
			scan.Sections = append(scan.Sections, ChatSection{h.line, ends[i], source, legacy})
		}
	}
	for _, n := range mapNames {
		if !matched[n] {
			scan.Notices = append(scan.Notices, fmt.Sprintf("collaboration-mode map name %q matches no heading", n))
		}
	}
	return scan, work
}

// readFrontmatter reads the legacy mode keys: whether a scalar or map default
// makes the whole document chat, and the map's chat names as written.
func (s *ChatScan) readFrontmatter(fm []string) (wholeChat bool, names []string) {
	if len(fm) == 0 {
		return false, nil
	}
	var front map[string]yaml.Node
	if yaml.Unmarshal([]byte(strings.Join(fm, "\n")), &front) != nil {
		s.Notices = append(s.Notices, unreadFrontmatter)
		return false, nil
	}
	if cm, ok := front["collaboration-mode"]; ok {
		switch cm.Kind {
		case yaml.ScalarNode:
			if cm.Value == "chat" {
				wholeChat = true
				s.Notices = append(s.Notices, "collaboration-mode: chat is a legacy key; read as @@chat@@ for the whole document")
			} else if cm.Value == "" {
				s.Notices = append(s.Notices, "collaboration-mode is empty or malformed; ignored")
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
				case k == "default":
					s.Notices = append(s.Notices, fmt.Sprintf("collaboration-mode map default: %s is retired and ignored", v.Value))
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

// ChatItem is one item of a chat section (collaboration-modes.md §3). Pos is
// its 1-based position in the section: the operator reuses numbers, so the
// number never identifies an item. Start indexes its first line in the
// section text.
type ChatItem struct {
	Pos      int
	Start    int
	Lines    []string
	Answered bool
}

var (
	numberedItemStart = regexp.MustCompile(`^\d+\.\s`)
	legacyItemStart   = regexp.MustCompile(`^(?:[-*]|\d+\.)\s`)
	answeredItem      = regexp.MustCompile(`→\s*(\[answer\]\([^)]*\)\s*$|done:)`)
)

// ChatItems splits a chat section's text into its items, by the grammar of
// collaboration-modes.md §3: a `N.` line at the section's shallowest list
// indent starts one (also `-` and `*` lines when legacy is set, as
// ChatSection.Legacy reports), and it runs over its sub-items and
// continuation lines until the next such line, or until a blank line
// followed by prose indented no deeper than that indent (prose belongs to no
// item). Fenced code starts no item; a tab indents to the next multiple of
// four. Trailing blank lines are dropped. An item is answered by a link
// ending any of its lines outside fences, so a follow-up added after the link
// keeps it answered; `→ done:` is the TL's older form. It is one linear pass
// over the section, the shared grammar of the review checker, tl-lead and
// pair#838.
func ChatItems(section string, legacy bool) []ChatItem {
	items, _ := chatItems(section, legacy)
	return items
}

// chatItems is ChatItems with its work count: one unit per line per pass.
func chatItems(section string, legacy bool) ([]ChatItem, int) {
	itemStart := numberedItemStart
	if legacy {
		itemStart = legacyItemStart
	}
	lines := strings.Split(section, "\n")
	fenced := make([]bool, len(lines))
	in := false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fenced[i], in = true, !in
			continue
		}
		fenced[i] = in
	}
	start := func(i int) (int, bool) {
		if fenced[i] || !itemStart.MatchString(strings.TrimLeft(lines[i], " \t")) {
			return 0, false
		}
		return IndentOf(lines[i]), true
	}
	shallow := -1
	for i := range lines {
		if n, ok := start(i); ok && (shallow < 0 || n < shallow) {
			shallow = n
		}
	}
	var out []ChatItem
	open, blank := false, false
	for i, l := range lines {
		n, isItem := start(i)
		switch {
		case strings.TrimSpace(l) == "" && !fenced[i]:
			blank = true
			if open {
				it := &out[len(out)-1]
				it.Lines = append(it.Lines, l)
			}
			continue
		case isItem && n == shallow:
			out = append(out, ChatItem{Pos: len(out) + 1, Start: i, Lines: []string{l}})
			open = true
		case open && blank && !fenced[i] && IndentOf(l) <= shallow:
			open = false
		case open:
			it := &out[len(out)-1]
			it.Lines = append(it.Lines, l)
		}
		blank = false
	}
	work := 3 * len(lines) // the fence, indent and item passes
	for i := range out {
		it := &out[i]
		work += len(it.Lines)
		for len(it.Lines) > 1 && strings.TrimSpace(it.Lines[len(it.Lines)-1]) == "" {
			it.Lines = it.Lines[:len(it.Lines)-1]
		}
		for j, l := range it.Lines {
			if !fenced[it.Start+j] && answeredItem.MatchString(l) {
				it.Answered = true
			}
		}
	}
	return out, work
}

// IndentOf is a line's leading indent in columns, a tab reaching the next
// multiple of four.
func IndentOf(l string) int {
	n := 0
	for _, r := range l {
		switch r {
		case ' ':
			n++
		case '\t':
			n += 4 - n%4
		default:
			return n
		}
	}
	return n
}
