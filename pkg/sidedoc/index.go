package sidedoc

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
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

// plainPathRE is an entry path that YAML reads back as the same plain scalar:
// no leading indicator (`!`, `&`, `*`, `-`, quotes…), no `: ` or ` #`.
var plainPathRE = regexp.MustCompile(`^[\p{L}\p{N}._][\p{L}\p{N}._/ -]*$`)

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
	if !ok || strings.ContainsRune(doc, '\r') {
		return nil, doc, false // YAML breaks lines at a bare CR; this parser doesn't
	}
	if after, empty := strings.CutPrefix(rest, "---\n"); empty {
		return nil, after, true
	}
	fm, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		if fm, ok = strings.CutSuffix(rest, "\n---"); !ok {
			return nil, doc, false
		}
		body = ""
	}
	lines = strings.Split(fm, "\n")
	for _, l := range lines {
		// A document marker or directive inside would hide what follows it
		// from YAML: refuse rather than write past it.
		if l == "..." || l == "---" || strings.HasPrefix(l, "--- ") || strings.HasPrefix(l, "... ") || strings.HasPrefix(l, "%") {
			return nil, doc, false
		}
	}
	return lines, body, true
}

// Fields returns a document's flat top-level frontmatter fields, `key: value`
// with the value raw (quotes kept): the one frontmatter reader for side-doc
// facts such as `parent:` and `marker:`. ok is false when it has none.
func Fields(doc string) (map[string]string, bool) {
	lines, _, ok := frontmatter(doc)
	if !ok {
		return nil, false
	}
	out := map[string]string{}
	for _, l := range lines {
		if k, v, ok := strings.Cut(l, ":"); ok && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			out[k] = strings.TrimSpace(v)
		}
	}
	return out, true
}

// indexBlock finds the `discussions:` key in frontmatter lines and returns
// the line span [start, end) it owns and its entries. start is -1 when the
// key is absent. A hand-edited index it can't read is an error, never an
// empty one.
func indexBlock(lines []string) (start, end int, entries []string, err error) {
	start = -1
	for i, l := range lines {
		if rest, ok := strings.CutPrefix(l, indexKey); ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
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
	// The block runs to its last item: a blank line between items is part of
	// it (YAML reads the list through it), and one before the next key is not.
	end = start + 1
	for i := start + 1; i < len(lines); i++ {
		l := lines[i]
		t := strings.TrimLeft(l, " \t")
		if t == "" {
			continue
		}
		if l == t && t != "-" && !strings.HasPrefix(t, "- ") {
			break // the next top-level key
		}
		end = i + 1
		// YAML needs the space: `- doc:x` is a string, not a doc: key.
		item, ok := strings.CutPrefix(t, "- doc: ")
		if t == "- doc:" {
			item, ok = "", true
		}
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
	case !plainPathRE.MatchString(p) || p != strings.TrimSpace(p):
		return fmt.Errorf("want a plain relative path: letters, digits, `.`, `_`, `-`, `/` and inner spaces, starting with a letter, digit, `.` or `_`")
	case path.IsAbs(p):
		return fmt.Errorf("an absolute path; want one relative to the document")
	case path.Clean(p) == ".." || strings.HasPrefix(path.Clean(p), "../"):
		return fmt.Errorf("climbs out of the document's directory; side docs live in its .side/ folder")
	}
	return nil
}

// errFrontmatter refuses a document that opens a frontmatter block it can't
// split (CRLF line endings, no closing fence): reading it as having none would
// put a second block in front of it.
func errFrontmatter(doc string) error {
	if strings.HasPrefix(doc, "---") {
		return fmt.Errorf("the frontmatter can't be read: it must open with `---` and close with a `---` line, with LF line endings")
	}
	return nil
}

// Entries returns a document's index entries, in order.
func Entries(doc string) ([]string, error) {
	lines, _, ok := frontmatter(doc)
	if !ok {
		return nil, errFrontmatter(doc)
	}
	_, _, entries, err := indexBlock(lines)
	if err == nil {
		err = agreesWithYAML(lines, entries)
	}
	if err != nil {
		return nil, err
	}
	return entries, nil
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
	if !hasFM {
		if err := errFrontmatter(doc); err != nil {
			return doc, Report{}, err
		}
	}
	start, end, entries, err := indexBlock(lines)
	if err != nil {
		return doc, Report{}, err
	}
	// The line parser locates the block so the rest of the frontmatter stays
	// byte for byte; YAML is the authority on what the index says. Any input
	// it reads differently (a comment, a tab, an indicator) is refused.
	if err := agreesWithYAML(lines, entries); err != nil {
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
		lines, body = block, doc
	case start < 0:
		lines = append(slices.Clone(lines), block...)
	default:
		lines = slices.Concat(lines[:start], block, lines[end:])
	}
	// And what it writes must read back as the entries it reports.
	if err := agreesWithYAML(lines, r.Entries); err != nil {
		return doc, Report{}, fmt.Errorf("the rewritten index would not read back: %v", err)
	}
	return "---\n" + strings.Join(lines, "\n") + "\n---\n" + body, r, nil
}

// agreesWithYAML decodes frontmatter lines with a YAML parser and refuses
// unless its discussions[].doc are exactly entries.
func agreesWithYAML(lines []string, entries []string) error {
	var fm struct {
		Discussions []struct {
			Doc string `yaml:"doc"`
		} `yaml:"discussions"`
	}
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &fm); err != nil {
		return fmt.Errorf("the frontmatter is not YAML this index can read: %v", err)
	}
	var got []string
	for _, d := range fm.Discussions {
		got = append(got, d.Doc)
	}
	if !slices.Equal(got, entries) {
		return fmt.Errorf("YAML reads the discussions: index as %q, not %q; rewrite it as `- doc: PATH` lines", got, entries)
	}
	return nil
}
