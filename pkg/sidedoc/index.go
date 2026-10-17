package sidedoc

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

// The `discussions:` index (#424, collaboration-modes.md "The index") lists
// the side docs waiting for the operator's review, in the document that links
// them:
//
//	discussions:
//	  - doc: design.side/07-s3-side-docs.md
//
// The rule is one-way. Deleting a link drops its entry; removing an entry
// means reviewed and leaves the link. An entry is only ever added by a write
// for the operator, so reconciling never adds one for a link by itself.

const indexKey = "discussions:"

// Report is what ReconcileIndex changed, and the entries it kept.
type Report struct {
	Dropped, Added, Entries []string
}

// Changed reports whether the reconcile rewrote the document.
func (r Report) Changed() bool { return len(r.Dropped)+len(r.Added) > 0 }

// frontmatter splits a document into its frontmatter lines and the body. ok
// is false when it has none.
func frontmatter(doc string) (lines []string, body string, ok bool) {
	rest, ok := strings.CutPrefix(doc, "---\n")
	if !ok {
		return nil, doc, false
	}
	if after, empty := strings.CutPrefix(rest, "---\n"); empty {
		return nil, after, true
	}
	fm, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		if fm2, found := strings.CutSuffix(rest, "\n---"); found {
			return strings.Split(fm2, "\n"), "", true
		}
		return nil, doc, false
	}
	return strings.Split(fm, "\n"), body, true
}

// indexBlock finds the `discussions:` key in frontmatter lines and returns
// the line span [start, end) it owns and its entries. start is -1 when the
// key is absent. A hand-edited index it can't read is an error, never an
// empty one.
func indexBlock(lines []string) (start, end int, entries []string, err error) {
	start = -1
	for i, l := range lines {
		if l == indexKey || strings.HasPrefix(l, indexKey+" ") {
			start = i
			break
		}
	}
	if start < 0 {
		return -1, -1, nil, nil
	}
	switch v := strings.TrimSpace(strings.TrimPrefix(lines[start], indexKey)); v {
	case "", "[]":
	default:
		return 0, 0, nil, fmt.Errorf("discussions: has the value %q; want a list of `- doc: PATH` items", v)
	}
	end = start + 1
	for ; end < len(lines); end++ {
		l := lines[end]
		t := strings.TrimLeft(l, " \t")
		if l == t && !strings.HasPrefix(t, "- ") {
			break // the next top-level key
		}
		if t == "" {
			continue
		}
		item, ok := strings.CutPrefix(t, "- doc:")
		if !ok {
			return 0, 0, nil, fmt.Errorf("discussions: line %q is not a `- doc: PATH` item", l)
		}
		p, err := entryPath(strings.TrimSpace(item))
		if err != nil {
			return 0, 0, nil, err
		}
		entries = append(entries, p)
	}
	return start, end, entries, nil
}

// entryPath reads one entry's path: bare or double-quoted, relative, and
// inside the document's directory.
func entryPath(raw string) (string, error) {
	p := raw
	if len(p) >= 2 && p[0] == '"' && p[len(p)-1] == '"' {
		p = p[1 : len(p)-1]
	}
	if err := checkEntry(p); err != nil {
		return "", fmt.Errorf("discussions: entry %q: %v", raw, err)
	}
	return p, nil
}

func checkEntry(p string) error {
	switch {
	case p == "" || p != strings.TrimSpace(p) || strings.ContainsAny(p, "\"'") || strings.Contains(p, " #") ||
		strings.ContainsFunc(p, func(r rune) bool { return r < 0x20 || r == 0x7f }):
		return fmt.Errorf("want a plain relative path")
	case path.IsAbs(p):
		return fmt.Errorf("an absolute path; want one relative to the document")
	case path.Clean(p) == ".." || strings.HasPrefix(path.Clean(p), "../"):
		return fmt.Errorf("climbs out of the document's directory; side docs live in its .side/ folder")
	}
	return nil
}

// Entries returns a document's index entries, in order.
func Entries(doc string) ([]string, error) {
	lines, _, ok := frontmatter(doc)
	if !ok {
		return nil, nil
	}
	_, _, entries, err := indexBlock(lines)
	return entries, err
}

// ReconcileIndex applies the one-way rule to a document: it drops each entry
// no side link in the body points to (the operator deleted the link), then
// appends each of add not already listed, at the end. An entry already listed
// keeps its place. It never adds an entry for a link by itself, and it leaves
// an entry whose file is missing to the caller. An index it can't read is an
// error, and the document comes back unchanged.
func ReconcileIndex(doc string, add []string) (string, Report, error) {
	for _, a := range add {
		if err := checkEntry(a); err != nil {
			return doc, Report{}, fmt.Errorf("add %q: %v", a, err)
		}
	}
	lines, body, hasFM := frontmatter(doc)
	start, end, entries, err := indexBlock(lines)
	if err != nil {
		return doc, Report{}, err
	}
	linked := linkedSet(body)
	var r Report
	for _, e := range entries {
		if linked[path.Clean(e)] {
			r.Entries = append(r.Entries, e)
		} else {
			r.Dropped = append(r.Dropped, e)
		}
	}
	for _, a := range add {
		if !slices.ContainsFunc(r.Entries, func(e string) bool { return path.Clean(e) == path.Clean(a) }) {
			r.Entries = append(r.Entries, a)
			r.Added = append(r.Added, a)
		}
	}
	if !r.Changed() {
		return doc, r, nil
	}
	block := []string{indexKey}
	for _, e := range r.Entries {
		block = append(block, "  - doc: "+e)
	}
	switch {
	case !hasFM:
		return "---\n" + strings.Join(block, "\n") + "\n---\n" + doc, r, nil
	case start < 0:
		lines = append(lines, block...)
	default:
		lines = slices.Concat(lines[:start], block, lines[end:])
	}
	return "---\n" + strings.Join(lines, "\n") + "\n---\n" + body, r, nil
}
