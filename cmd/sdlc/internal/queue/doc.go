package queue

import "strings"

// Doc is a parsed queue file: every source line in order, each either a queue
// entry or verbatim text (prose, blank lines, comments).
//
// The invariant is byte-exact round-tripping — Render(Parse(b)) == b — so an
// edit can never reformat the parts it did not touch. That matters because the
// file is co-authored by hand and by the verb, and a tool that silently
// reflows a human's prose stops being trusted with it.
type Doc struct {
	lines []Line // an unparsed source line is a Line with raw set
	// trailingNewline records whether the source ended with one, so Render can
	// reproduce a file with or without it.
	trailingNewline bool
}

// Parse reads a queue file.
func Parse(b []byte) *Doc {
	s := string(b)
	d := &Doc{}
	if s == "" {
		return d
	}
	d.trailingNewline = strings.HasSuffix(s, "\n")
	body := s
	if d.trailingNewline {
		body = body[:len(body)-1]
	}
	for _, raw := range strings.Split(body, "\n") {
		l, _ := ParseLine(raw) // on failure ParseLine returns the verbatim line
		d.lines = append(d.lines, l)
	}
	return d
}

// Render writes the document back out.
func (d *Doc) Render() []byte {
	if len(d.lines) == 0 {
		if d.trailingNewline {
			return []byte("\n")
		}
		return nil
	}
	parts := make([]string, len(d.lines))
	for i, l := range d.lines {
		parts[i] = l.String()
	}
	s := strings.Join(parts, "\n")
	if d.trailingNewline {
		s += "\n"
	}
	return []byte(s)
}

// Entries returns the parsed queue entries, in order.
func (d *Doc) Entries() []Line {
	var out []Line
	for _, l := range d.lines {
		if l.Parsed() {
			out = append(out, l)
		}
	}
	return out
}

// indexOf returns the position in d.lines of the entry with this ref, or -1.
func (d *Doc) indexOf(ref string) int {
	for i, l := range d.lines {
		if l.Parsed() && l.Ref == ref {
			return i
		}
	}
	return -1
}
