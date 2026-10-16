package reviewexport

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

// The examples (#412) are reference rounds for collaboration-modes.md: one
// directory per case, each a before/ and an after/ tree plus case.md. The
// agent and parley read the prose; this checker only keeps the committed
// fixtures true to the side-doc contract of collaboration-modes.md §4.

var (
	datedNameRE    = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(\d{2,})-([a-z0-9]+(?:-[a-z0-9]+)*)\.md$`)
	numberedNameRE = regexp.MustCompile(`^(\d{2,})-([a-z0-9]+(?:-[a-z0-9]+)*)\.md$`)
	sideLinkRE     = regexp.MustCompile(`→ \[(?:discussion|answer)\]\(([^)\s]+)\)`)
)

// sideDocName classifies a side doc's file name. A dated answer doc is tried
// first: its date would otherwise read as a number (2026).
func sideDocName(name string) (date string, n int, slug string, ok bool) {
	if m := datedNameRE.FindStringSubmatch(name); m != nil {
		n, _ = strconv.Atoi(m[2])
		return m[1], n, m[3], len(m[3]) <= 40
	}
	if m := numberedNameRE.FindStringSubmatch(name); m != nil {
		n, _ = strconv.Atoi(m[1])
		return "", n, m[2], len(m[2]) <= 40
	}
	return "", 0, "", false
}

// nextSideDocName is the name a round gives a new side doc in a folder
// holding names: the next free number, two digits widening as needed, among
// the numbered side docs (date "") or among that date's answer docs.
func nextSideDocName(names []string, date, slug string) string {
	max := 0
	for _, nm := range names {
		if d, n, _, ok := sideDocName(nm); ok && d == date && n > max {
			max = n
		}
	}
	if date == "" {
		return fmt.Sprintf("%02d-%s.md", max+1, slug)
	}
	return fmt.Sprintf("%s-%02d-%s.md", date, max+1, slug)
}

// frontmatter returns a document's flat `key: value` lines, values raw.
func frontmatter(text string) (map[string]string, bool) {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return nil, false
	}
	fm, _, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return nil, false
	}
	out := map[string]string{}
	for _, l := range strings.Split(fm, "\n") {
		if k, v, ok := strings.Cut(l, ":"); ok && !strings.HasPrefix(l, " ") {
			out[k] = strings.TrimSpace(v)
		}
	}
	return out, true
}

// quotedMarker checks a marker's YAML form: double-quoted, one line, with
// only the \\ and \" escapes, so operator text can't become frontmatter.
func quotedMarker(raw string) (string, error) {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return "", fmt.Errorf("marker %s is not double-quoted", raw)
	}
	var sb strings.Builder
	in := raw[1 : len(raw)-1]
	for i := 0; i < len(in); i++ {
		switch c := in[i]; c {
		case '\\':
			if i+1 == len(in) || (in[i+1] != '\\' && in[i+1] != '"') {
				return "", fmt.Errorf("marker %s has an escape other than \\\\ or \\\"", raw)
			}
			i++
			sb.WriteByte(in[i])
		case '"':
			return "", fmt.Errorf("marker %s has an unescaped quote", raw)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String(), nil
}

// checkTree checks one example tree against the side-doc contract and
// returns its side docs' paths.
func checkTree(fsys fs.FS) (sides []string, errs []string) {
	bad := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	read := func(p string) string { b, _ := fs.ReadFile(fsys, p); return string(b) }
	owners := map[string]string{} // folder + parent + marker → the side doc holding it
	_ = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			bad("%s: %v", p, err)
			return nil
		}
		dir := path.Dir(p)
		if d.IsDir() {
			if strings.HasSuffix(p, ".side") && strings.Contains(dir, ".side") {
				bad("%s: a .side/ inside a .side/", p)
			}
			return nil
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		text := read(p)
		for _, m := range sideLinkRE.FindAllStringSubmatch(text, -1) {
			if _, err := fs.Stat(fsys, path.Join(dir, m[1])); err != nil {
				bad("%s: link %s does not resolve", p, m[1])
			}
		}
		if !strings.HasSuffix(dir, ".side") {
			return nil
		}
		sides = append(sides, p)
		if _, _, _, ok := sideDocName(path.Base(p)); !ok {
			bad("%s: not an NN-slug or YYYY-MM-DD-NN-slug name", p)
		}
		fm, ok := frontmatter(text)
		if !ok {
			bad("%s: no frontmatter", p)
			return nil
		}
		if fm["type"] != "side-doc" {
			bad("%s: type is %q, want side-doc", p, fm["type"])
		}
		if fm["collaboration-mode"] != "discussion" {
			bad("%s: collaboration-mode is %q, want discussion", p, fm["collaboration-mode"])
		}
		parent := fm["parent"]
		if parent == "" || strings.Contains(parent, "#") || path.IsAbs(parent) {
			bad("%s: parent %q must be a plain relative path", p, parent)
			return nil
		}
		pp := path.Join(dir, parent)
		ptext := read(pp)
		if ptext == "" {
			bad("%s: parent %s does not resolve", p, parent)
			return nil
		}
		pfm, _ := frontmatter(ptext)
		if fm["process-instruction"] != pfm["process-instruction"] {
			bad("%s: process-instruction %q, the parent's is %q (copied only when it has one)", p, fm["process-instruction"], pfm["process-instruction"])
		}
		marker, err := quotedMarker(fm["marker"])
		if err != nil {
			bad("%s: %v", p, err)
			return nil
		}
		key := dir + "\x00" + pp + "\x00" + marker
		if other, dup := owners[key]; dup {
			bad("%s: same parent and marker as %s", p, other)
		}
		owners[key] = p
		return nil
	})
	return sides, errs
}

// caseFields reads case.md's `submitted:` and `outcome:` lines.
func caseFields(text string) (submitted, outcome string) {
	for _, l := range strings.Split(text, "\n") {
		if v, ok := strings.CutPrefix(l, "submitted: "); ok {
			submitted = v
		}
		if v, ok := strings.CutPrefix(l, "outcome: "); ok {
			outcome = v
		}
	}
	return submitted, outcome
}

// checkCase checks one case directory: both trees, what the round may change
// between them, and its case.md.
func checkCase(fsys fs.FS) []string {
	var errs []string
	bad := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	cm, err := fs.ReadFile(fsys, "case.md")
	if err != nil {
		return []string{"no case.md"}
	}
	submitted, outcome := caseFields(string(cm))
	if outcome != "ok" && outcome != "fails" {
		bad("case.md: outcome %q, want ok or fails", outcome)
	}
	trees := map[string][]string{}
	for _, t := range []string{"before", "after"} {
		sub, err := fs.Sub(fsys, t)
		if err != nil {
			bad("%s: %v", t, err)
			continue
		}
		if _, err := fs.Stat(sub, submitted); err != nil {
			bad("%s: the submitted document %q is missing", t, submitted)
		}
		sides, tErrs := checkTree(sub)
		for _, e := range tErrs {
			bad("%s/%s", t, e)
		}
		trees[t] = sides
	}
	if outcome == "fails" {
		if !sameTree(fsys) {
			bad("a failed round changes nothing, but after/ differs from before/")
		}
		return errs
	}
	// Side docs are never pruned, and new ones take the next free numbers.
	byDir := map[string][]string{}
	for _, p := range trees["before"] {
		byDir[path.Dir(p)] = append(byDir[path.Dir(p)], path.Base(p))
		if !slices.Contains(trees["after"], p) {
			bad("after/ dropped the side doc %s (side docs are never pruned)", p)
		}
	}
	var added []string
	for _, p := range trees["after"] {
		if !slices.Contains(trees["before"], p) {
			added = append(added, p)
		}
	}
	slices.Sort(added)
	for _, p := range added {
		dir, name := path.Dir(p), path.Base(p)
		date, _, slug, _ := sideDocName(name)
		if want := nextSideDocName(byDir[dir], date, slug); name != want {
			bad("after/%s: a new side doc here is named %s", p, want)
		}
		byDir[dir] = append(byDir[dir], name)
	}
	return errs
}

func sameTree(fsys fs.FS) bool {
	files := func(t string) map[string]string {
		out := map[string]string{}
		_ = fs.WalkDir(fsys, t, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				b, _ := fs.ReadFile(fsys, p)
				out[strings.TrimPrefix(p, t+"/")] = string(b)
			}
			return nil
		})
		return out
	}
	b, a := files("before"), files("after")
	if len(a) != len(b) {
		return false
	}
	for k, v := range b {
		if a[k] != v {
			return false
		}
	}
	return true
}

func TestSideDocNames(t *testing.T) {
	for _, c := range []struct {
		names      []string
		date, want string
	}{
		{nil, "", "01-q.md"},
		{[]string{"01-a.md", "02-b.md"}, "", "03-q.md"},
		{[]string{"98-a.md"}, "", "99-q.md"},
		{[]string{"99-a.md"}, "", "100-q.md"},
		{[]string{"100-a.md", "99-b.md"}, "", "101-q.md"},
		// A dated answer doc's year is not a number, and dates count apart.
		{[]string{"2026-10-16-01-a.md", "01-b.md"}, "", "02-q.md"},
		{[]string{"2026-10-16-01-a.md", "01-b.md"}, "2026-10-16", "2026-10-16-02-q.md"},
		{[]string{"2026-10-15-04-a.md"}, "2026-10-16", "2026-10-16-01-q.md"},
	} {
		if got := nextSideDocName(c.names, c.date, "q"); got != c.want {
			t.Errorf("next(%v, %q) = %s, want %s", c.names, c.date, got, c.want)
		}
	}
	for _, bad := range []string{"1-a.md", "01-A.md", "01-a--b.md", "01-.md", "01-" + strings.Repeat("a", 41) + ".md", "x.md"} {
		if _, _, _, ok := sideDocName(bad); ok {
			t.Errorf("%s accepted as a side doc name", bad)
		}
	}
}

func TestQuotedMarker(t *testing.T) {
	if got, err := quotedMarker(`"say \"hi\" \\ <br> ok"`); err != nil || got != `say "hi" \ <br> ok` {
		t.Errorf("got %q, %v", got, err)
	}
	for _, bad := range []string{`why?`, `'why?'`, `"a"b"`, `"a\n"`, `"`} {
		if _, err := quotedMarker(bad); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

// goodCase is a minimal valid case the mutation test breaks one way at a time.
func goodCase() fstest.MapFS {
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	side := "---\ntype: side-doc\nparent: ../d.md\nmarker: \"why?\"\ncollaboration-mode: discussion\n---\n\n# Why?\n"
	return fstest.MapFS{
		"case.md":                f("submitted: d.md\noutcome: ok\n"),
		"before/d.md":            f("# D\n\nText. 🤖[why?]\n"),
		"after/d.md":             f("# D\n\nText. → [discussion](d.side/01-why.md)\n"),
		"after/d.side/01-why.md": f(side),
	}
}

func TestCheckCaseCatches(t *testing.T) {
	if errs := checkCase(goodCase()); len(errs) != 0 {
		t.Fatalf("the good case fails: %v", errs)
	}
	side := "after/d.side/01-why.md"
	edit := func(p, old, new string) func(fstest.MapFS) {
		return func(m fstest.MapFS) {
			m[p] = &fstest.MapFile{Data: []byte(strings.Replace(string(m[p].Data), old, new, 1))}
		}
	}
	for _, c := range []struct {
		name   string
		mutate func(fstest.MapFS)
		want   string
	}{
		{"bad name", func(m fstest.MapFS) {
			m["after/d.side/1-why.md"] = m[side]
			delete(m, side)
			edit("after/d.md", "01-why", "1-why")(m)
		}, "name"},
		{"anchored parent", edit(side, "../d.md", "../d.md#x"), "plain relative path"},
		{"unresolved parent", edit(side, "../d.md", "../e.md"), "does not resolve"},
		{"unquoted marker", edit(side, `"why?"`, `why?`), "double-quoted"},
		{"wrong mode", edit(side, "mode: discussion", "mode: chat"), "want discussion"},
		{"copied process-instruction from nowhere", edit(side, "---\n\n", "process-instruction: \"[free-form]\"\n---\n\n"), "process-instruction"},
		{"dangling link", edit("after/d.md", "01-why", "02-why"), "does not resolve"},
		{"skipped number", func(m fstest.MapFS) {
			m["after/d.side/02-why.md"] = m[side]
			delete(m, side)
			edit("after/d.md", "01-why", "02-why")(m)
		}, "named 01-why.md"},
		{"nested .side", func(m fstest.MapFS) { m["after/d.side/x.side/01-a.md"] = m[side] }, ".side/ inside"},
		{"duplicate parent and marker", func(m fstest.MapFS) {
			m["after/d.side/02-again.md"] = m[side]
		}, "same parent and marker"},
		{"a pruned side doc", func(m fstest.MapFS) { m["before/d.side/01-why.md"] = m[side]; delete(m, side) }, "never pruned"},
		{"a failed round that changed the document", edit("case.md", "outcome: ok", "outcome: fails"), "changes nothing"},
	} {
		m := goodCase()
		c.mutate(m)
		errs := strings.Join(checkCase(m), "\n")
		if !strings.Contains(errs, c.want) {
			t.Errorf("%s: errors %q, want one naming %q", c.name, errs, c.want)
		}
	}
}

func skillDir(t *testing.T) string {
	return filepath.Join(moduleRoot(t), "construct", "local", "universal-review")
}

// Every committed case satisfies the contract.
func TestExamples(t *testing.T) {
	root := filepath.Join(skillDir(t), "examples")
	cases, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range cases {
		if !c.IsDir() {
			continue
		}
		n++
		for _, e := range checkCase(os.DirFS(filepath.Join(root, c.Name()))) {
			t.Errorf("%s: %s", c.Name(), e)
		}
	}
	if n < 10 {
		t.Errorf("%d cases, want the ten the plan's test list names", n)
	}
}

// The state table has its rows, and the rules a static case can't show are
// stated (rows 9, 15 and 17: a real effect, a closed doc, the lock).
func TestCollaborationModesProse(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(skillDir(t), "collaboration-modes.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	flat := strings.Join(strings.Fields(doc), " ")
	for i := 1; i <= 18; i++ {
		if !strings.Contains(doc, fmt.Sprintf("\n| %d | ", i)) {
			t.Errorf("the state table lacks row %d", i)
		}
	}
	for _, phrase := range []string{
		"add a closing record at the end of the side doc",    // row 9
		"never reopen or extend it",                          // row 15
		"the adapter's exit 3",                               // row 17
		"Never create a `.side/` inside a `.side/`",          // the flat folder
		"Swapping a marker for its link is applied directly", // propose-strike
		"don't fold again",                                   // a refused side-doc write after the fold landed
		"a plain relative path with no `#anchor`",            // the body's backlink
	} {
		// Compared with whitespace collapsed, so rewrapping the prose doesn't break it.
		if !strings.Contains(flat, phrase) {
			t.Errorf("collaboration-modes.md lacks %q", phrase)
		}
	}
}
