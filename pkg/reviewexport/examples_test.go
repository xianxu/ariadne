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

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// The examples (#412) are reference rounds for collaboration-modes.md: one
// directory per case, each a before/ and an after/ tree plus case.md. The
// agent and parley read the prose; this checker only keeps the committed
// fixtures true to the side-doc contract of collaboration-modes.md §4.

var (
	datedNameRE    = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(\d{2,6})-([a-z0-9]+(?:-[a-z0-9]+)*)\.md$`)
	numberedNameRE = regexp.MustCompile(`^(\d{2,6})-([a-z0-9]+(?:-[a-z0-9]+)*)\.md$`)
	backlinkRE     = regexp.MustCompile(`(?m)^_From \[[^\]]*\]\([^)#\s]+\), section "([^"]+)"\._`)
	foldedRE       = regexp.MustCompile(`^ \(folded \d{4}-\d{2}-\d{2}\)`)
	foldNoteRE     = regexp.MustCompile(`(?m)^(?:Folded into|Proposed in) the parent on \d{4}-\d{2}-\d{2}: `)
	multiTurnRE    = regexp.MustCompile(`🤖(?:<[^>\n]*>|~[^~\n]*~)?(?:\[[^]\n]*\]|\{[^}\n]*\}){2,}`)
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
var frontmatter = sidedoc.Fields

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

// sideDoc is one side doc as a tree holds it. Key is what reuse matches on
// (collaboration-modes.md §4 step 1): its folder, parent, marker and section.
// Linked says a link in the parent points to it.
type sideDoc struct {
	Key    string
	Linked bool
}

// checkTree checks one example tree against the side-doc contract and
// returns its side docs by path.
func checkTree(fsys fs.FS) (sides map[string]sideDoc, errs []string) {
	bad := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	read := func(p string) string { b, _ := fs.ReadFile(fsys, p); return string(b) }
	sides = map[string]sideDoc{}
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
		// Every reuse invariant of collaboration-modes.md §4 step 1 that a
		// tree can show: a side doc is the target of at most one link (an
		// already-linked doc is never reused), and only from its own parent.
		targets := map[string]int{}
		for _, link := range sidedoc.SideLinks(text) {
			t := path.Join(dir, link)
			if _, err := fs.Stat(fsys, t); err != nil {
				bad("%s: link %s does not resolve", p, link)
				continue
			}
			if !strings.HasSuffix(path.Dir(t), ".side") {
				continue // an older qa/ answer file, not a side doc
			}
			if targets[t]++; targets[t] == 2 {
				bad("%s: two links point to %s; a linked side doc is never reused", p, link)
			}
			if tfm, _ := frontmatter(read(t)); path.Join(path.Dir(t), tfm["parent"]) != p {
				bad("%s: links %s, whose parent is %q", p, link, tfm["parent"])
			}
		}
		// Every index entry names a side doc whose parent: is this document
		// (collaboration-modes.md §4, "The index").
		entries, err := sidedoc.Entries(text)
		if err != nil {
			bad("%s: %v", p, err)
		}
		for _, e := range entries {
			t := path.Join(dir, e)
			ttext := read(t)
			if ttext == "" {
				bad("%s: index entry %s does not resolve", p, e)
				continue
			}
			if tfm, _ := frontmatter(ttext); path.Join(path.Dir(t), tfm["parent"]) != p {
				bad("%s: index entry %s names a side doc whose parent is %q", p, e, tfm["parent"])
			}
		}
		if !strings.HasSuffix(dir, ".side") {
			return nil
		}
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
		section := backlinkRE.FindStringSubmatch(text)
		if section == nil {
			bad("%s: no backlink line naming the section", p)
			return nil
		}
		rel, _ := filepath.Rel(path.Dir(pp), p)
		link := "](" + filepath.ToSlash(rel) + ")"
		at := strings.Index(ptext, link)
		sides[p] = sideDoc{Key: dir + "\x00" + pp + "\x00" + marker + "\x00" + section[1], Linked: at >= 0}
		// An approval's fold marks the parent's link to the side doc it folded.
		if foldNoteRE.MatchString(text) && (at < 0 || !foldedRE.MatchString(ptext[at+len(link):])) {
			bad("%s: records a fold, but the parent's link to it is not marked (folded <date>)", p)
		}
		return nil
	})
	return sides, errs
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
	submitted, outcome, move := ParseCase(string(cm))
	if outcome != "ok" && outcome != "fails" {
		bad("case.md: outcome %q, want ok or fails", outcome)
	}
	trees := map[string]map[string]sideDoc{}
	for _, t := range []string{"before", "after"} {
		sub, err := fs.Sub(fsys, t)
		if err != nil {
			bad("%s: %v", t, err)
			continue
		}
		doc := submitted
		if t == "after" && len(move) == 2 && submitted == move[0] {
			doc = move[1] // the move renamed it
		}
		if _, err := fs.Stat(sub, doc); err != nil {
			bad("%s: the submitted document %q is missing", t, doc)
		}
		sides, tErrs := checkTree(sub)
		for _, e := range tErrs {
			// A move's orphan folder has no document by definition, so its
			// side docs' parent: does not resolve.
			if len(move) == 2 && isOrphanPath(t, e, move[1]) {
				continue
			}
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
	if move != nil {
		// A move is checked whole: after/ is exactly the move contract applied
		// to before/, so parley's mover and the agent share one fixture.
		if len(move) != 2 {
			bad("case.md: move %v, want `move: OLD NEW`", move)
			return errs
		}
		want := projectMove(treeFiles(fsys, "before"), move[0], move[1])
		got := treeFiles(fsys, "after")
		for p, b := range want {
			if g, ok := got[p]; !ok {
				bad("after/%s: missing; the move produces it", p)
			} else if g != b {
				bad("after/%s: differs from the move's projection:\n%s", p, b)
			}
		}
		for p := range got {
			if _, ok := want[p]; !ok {
				bad("after/%s: the move does not produce it", p)
			}
		}
		return errs
	}
	// Readers keep old chains; a normal round must never write a new one.
	// A move has already been checked against the exact projection above,
	// including renamed parents, side docs and displaced orphan folders.
	// This regex checks the simple markers in these fixtures, not the full
	// nested/escaped grammar owned by the editor.
	beforeText, afterText := treeFiles(fsys, "before"), treeFiles(fsys, "after")
	for p, body := range afterText {
		for _, chain := range multiTurnRE.FindAllString(body, -1) {
			if !strings.Contains(beforeText[p], chain) {
				bad("after/%s: agent wrote a second turn in %q", p, chain)
			}
		}
	}
	errs = append(errs, checkIndexTransitions(fsys)...)
	// Side docs are never pruned; new ones take the next free numbers; and an
	// unlinked doc for the same ask is relinked rather than duplicated.
	before, after := trees["before"], trees["after"]
	// Numbering counts every file in a folder, a malformed side doc's too:
	// its number is taken whether or not it meets the contract.
	beforeFiles := treeFiles(fsys, "before")
	byDir := map[string][]string{}
	for p := range beforeFiles {
		if strings.HasSuffix(path.Dir(p), ".side") {
			byDir[path.Dir(p)] = append(byDir[path.Dir(p)], path.Base(p))
		}
	}
	for p := range before {
		if _, ok := after[p]; !ok {
			bad("after/ dropped the side doc %s (side docs are never pruned)", p)
		}
	}
	var added []string
	for p := range after {
		if _, ok := beforeFiles[p]; !ok {
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
		for q, d := range before {
			if d.Key == after[p].Key && !d.Linked && !after[q].Linked {
				bad("after/%s duplicates %s, an unlinked side doc for the same ask that should have been relinked", p, q)
			}
		}
	}
	return errs
}

// treeFiles reads one tree's files by their path inside it.
func treeFiles(fsys fs.FS, t string) map[string]string {
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

// checkIndexTransitions checks what a round may do to the links and index
// entries between before/ and after/ (collaboration-modes.md §4, "The
// index"): it never removes a link; it drops an entry only when its link is
// gone; every entry it leaves is linked; and it adds an entry only for a side
// doc it wrote or linked this round (a relink adds the entry back), and never
// for a fold.
func checkIndexTransitions(fsys fs.FS) []string {
	var errs []string
	bad := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	before, after := treeFiles(fsys, "before"), treeFiles(fsys, "after")
	for p, a := range after {
		if !strings.HasSuffix(p, ".md") {
			continue
		}
		dir := path.Dir(p)
		linked := map[string]bool{}
		for _, l := range sidedoc.SideLinks(a) {
			linked[path.Join(dir, l)] = true
		}
		b, existed := before[p]
		wasLinked := map[string]bool{}
		if existed {
			for _, l := range sidedoc.SideLinks(b) {
				wasLinked[path.Join(dir, l)] = true
				if !linked[path.Join(dir, l)] {
					bad("after/%s: the link to %s is gone; a round never removes a link", p, l)
				}
			}
		}
		was := map[string]bool{}
		if existed {
			old, _ := sidedoc.Entries(b)
			for _, e := range old {
				was[path.Join(dir, e)] = true
				if linked[path.Join(dir, e)] && !slices.Contains(mustEntries(a), e) {
					bad("after/%s: dropped the entry %s while its link remains", p, e)
				}
			}
		}
		for _, e := range mustEntries(a) {
			t := path.Join(dir, e)
			if !linked[t] {
				bad("after/%s: the entry %s has no link; a deleted link drops its entry", p, e)
			}
			if !was[t] && before[t] == after[t] && wasLinked[t] {
				bad("after/%s: added the entry %s, but the round neither wrote that side doc nor linked it", p, e)
			}
			// A fold's closing note is a record that needs nothing from the
			// operator: the fold adds no entry.
			if !was[t] && len(foldNoteRE.FindAllString(after[t], -1)) > len(foldNoteRE.FindAllString(before[t], -1)) {
				bad("after/%s: added the entry %s for a fold; a fold adds no index entry", p, e)
			}
		}
	}
	return errs
}

// isOrphanPath reports whether a tree error is about a file in the orphan
// folder a move to newPath displaces: NEW.side/ before the move, NEW_N.side/
// after it.
func isOrphanPath(tree, e, newPath string) bool {
	base := strings.TrimSuffix(newPath, ".md")
	if tree == "before" {
		return strings.HasPrefix(e, base+".side/")
	}
	rest, ok := strings.CutPrefix(e, base+"_")
	n, _, _ := strings.Cut(rest, ".side/")
	_, err := strconv.Atoi(n)
	return ok && err == nil && strings.Contains(rest, ".side/")
}

func mustEntries(doc string) []string {
	e, _ := sidedoc.Entries(doc)
	return e
}

func sameTree(fsys fs.FS) bool {
	b, a := treeFiles(fsys, "before"), treeFiles(fsys, "after")
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
	side := "---\ntype: side-doc\nparent: ../d.md\nmarker: \"why?\"\ncollaboration-mode: discussion\n---\n\n# Why?\n\n_From [D](../d.md), section \"D\"._\n"
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
	legacy := goodCase()
	for _, p := range []string{"before/d.md", "after/d.md"} {
		legacy[p] = &fstest.MapFile{Data: append(append([]byte(nil), legacy[p].Data...), []byte("\nLegacy 🤖[first]{answer}[follow-up]\n")...)}
	}
	if errs := checkCase(legacy); len(errs) != 0 {
		t.Fatalf("an unchanged legacy chain must stay readable: %v", errs)
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
		{"an unlinked doc for the same ask not relinked", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["after/d.side/02-why.md"] = m[side]
			edit("after/d.md", "01-why", "02-why")(m)
		}, "should have been relinked"},
		{"a linked doc reused", func(m fstest.MapFS) {
			edit("before/d.md", "🤖[why?]", "🤖[why?] Again. 🤖[why?]")(m)
			edit("after/d.md", "(d.side/01-why.md)", "(d.side/01-why.md) Again. → [discussion](d.side/01-why.md)")(m)
		}, "two links point to"},
		{"a link from a document that is not the parent", func(m fstest.MapFS) {
			m["after/e.md"] = &fstest.MapFile{Data: []byte("# E\n\n→ [discussion](d.side/01-why.md)\n")}
		}, "whose parent is"},
		{"no backlink", edit(side, "_From [D](../d.md), section \"D\"._", "From D."), "no backlink"},
		{"a fold whose link is not marked", edit(side, "# Why?\n", "# Why?\n\nFolded into the parent on 2026-10-17: the reason.\n"), "not marked (folded"},
		{"an entry naming a side doc of another document", func(m fstest.MapFS) {
			m["after/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: d.side/01-why.md\n---\n# D\n\nText. → [discussion](d.side/01-why.md)\n")}
			edit(side, "../d.md", "../e.md")(m)
			m["after/e.md"] = &fstest.MapFile{Data: []byte("# E\n")}
		}, "names a side doc whose parent is"},
		{"an entry that does not resolve", func(m fstest.MapFS) {
			edit("after/d.md", "# D", "---\ndiscussions:\n  - doc: d.side/09-gone.md\n---\n# D")(m)
		}, "does not resolve"},
		{"an entry left without its link", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["before/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: d.side/01-why.md\n---\n# D\n\nText.\n")}
			m["after/d.md"] = m["before/d.md"]
		}, "has no link"},
		{"an entry dropped while its link remains", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["before/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: d.side/01-why.md\n---\n# D\n\nText. → [discussion](d.side/01-why.md) 🤖[tighten]\n")}
			m["after/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n---\n# D\n\nText. → [discussion](d.side/01-why.md)\n")}
		}, "while its link remains"},
		{"a link removed by the round", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["before/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. → [discussion](d.side/01-why.md)\n")}
			m["after/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText.\n")}
		}, "never removes a link"},
		{"an entry added for a side doc the round neither wrote nor linked", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["before/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. → [discussion](d.side/01-why.md)\n")}
			m["after/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: d.side/01-why.md\n---\n# D\n\nText. → [discussion](d.side/01-why.md)\n")}
		}, "neither wrote"},
		{"a number taken by a malformed side doc", func(m fstest.MapFS) {
			m["before/d.side/01-old.md"] = &fstest.MapFile{Data: []byte("---\ntype: side-doc\n---\nno backlink\n")}
			m["after/d.side/01-old.md"] = m["before/d.side/01-old.md"]
		}, "named 02-why.md"},
		{"an entry added by a fold", func(m fstest.MapFS) {
			m["before/d.side/01-why.md"] = m[side]
			m["before/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. → [discussion](d.side/01-why.md) 🤖[fold in]\n")}
			m["after/d.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: d.side/01-why.md\n---\n# D\n\nFolded. → [discussion](d.side/01-why.md) (folded 2026-10-17)\n")}
			edit(side, "section \"D\"._\n", "section \"D\"._\n\nFolded into the parent on 2026-10-17: the reason.\n")(m)
		}, "a fold adds no index entry"},
		{"a move whose after/ is not the projection", func(m fstest.MapFS) {
			for k := range m {
				delete(m, k)
			}
			f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
			m["case.md"] = f("submitted: d.md\noutcome: ok\nmove: d.md e.md\n")
			m["before/d.md"] = f("# D\n\n→ [discussion](d.side/01-why.md)\n")
			m["before/d.side/01-why.md"] = f(string(goodCase()[side].Data))
			m["after/e.md"] = f("# D\n\n→ [discussion](d.side/01-why.md)\n")
			m["after/e.side/01-why.md"] = m["before/d.side/01-why.md"]
		}, "differs from the move's projection"},
		{"a pruned side doc", func(m fstest.MapFS) { m["before/d.side/01-why.md"] = m[side]; delete(m, side) }, "never pruned"},
		{"a failed round that changed the document", edit("case.md", "outcome: ok", "outcome: fails"), "changes nothing"},
		{"an agent appends a second turn", func(m fstest.MapFS) {
			m["after/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. 🤖[why?]{which case?}\n")}
		}, "second turn"},
		{"an agent appends to a legacy chain", func(m fstest.MapFS) {
			m["before/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. 🤖[why?]{answer}[follow-up]\n")}
			m["after/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. 🤖[why?]{answer}[follow-up]{more}\n")}
		}, "second turn"},
	} {
		m := goodCase()
		c.mutate(m)
		errs := strings.Join(checkCase(m), "\n")
		if !strings.Contains(errs, c.want) {
			t.Errorf("%s: errors %q, want one naming %q", c.name, errs, c.want)
		}
	}
}

func TestMovePreservesLegacyChains(t *testing.T) {
	root := os.DirFS(filepath.Join(skillDir(t), "examples", "move-orphan"))
	m := fstest.MapFS{}
	if err := fs.WalkDir(root, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(root, p)
		if err == nil {
			m[p] = &fstest.MapFile{Data: b}
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		"design.md",                      // renamed parent
		"design.side/01-why-per-file.md", // moved side doc
		"design.side/02-sections.md",     // another moved side doc
		"spec.side/01-older.md",          // displaced orphan
	} {
		name := "before/" + p
		m[name] = &fstest.MapFile{Data: append(append([]byte(nil), m[name].Data...), []byte("\nLegacy 🤖[ask]{reply}\n")...)}
	}
	for p := range treeFiles(m, "after") {
		delete(m, "after/"+p)
	}
	for p, body := range projectMove(treeFiles(m, "before"), "design.md", "spec.md") {
		m["after/"+p] = &fstest.MapFile{Data: []byte(body)}
	}
	if errs := checkCase(m); len(errs) != 0 {
		t.Fatalf("unchanged legacy chains across all move paths: %v", errs)
	}
	m["after/spec.md"] = &fstest.MapFile{Data: []byte(strings.Replace(string(m["after/spec.md"].Data), "🤖[ask]{reply}", "🤖[ask]{reply}{extra}", 1))}
	if errs := strings.Join(checkCase(m), "\n"); !strings.Contains(errs, "differs from the move's projection") {
		t.Fatalf("an appended turn during a move was not refused: %s", errs)
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
		if !c.IsDir() || strings.HasPrefix(c.Name(), ".") {
			continue // README.md, and a tool's hidden state directory
		}
		n++
		for _, e := range checkCase(os.DirFS(filepath.Join(root, c.Name()))) {
			t.Errorf("%s: %s", c.Name(), e)
		}
	}
	if n < 12 {
		t.Errorf("%d cases, want the twelve the plan's test list names", n)
	}
}

// The state table has its rows, and the rules a static case can't show are
// stated (rows 9, 15 and 17: a real effect, a deleted item's doc, the lock).
func TestCollaborationModesProse(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(skillDir(t), "collaboration-modes.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	flat := strings.Join(strings.Fields(doc), " ")
	for i := 1; i <= 29; i++ {
		if !strings.Contains(doc, fmt.Sprintf("\n| %d | ", i)) {
			t.Errorf("the state table lacks row %d", i)
		}
	}
	for _, phrase := range []string{
		"add a closing record at the end of the side doc",                           // row 9
		"leaves its answer doc in place",                                            // row 15
		"Deleting a link drops its entry; removing an entry never touches the link", // the one-way rule
		"Rename the orphan to `spec_1.side/`",                                       // the orphan suffix
		"Folded into the parent on <date>:",                                         // the fold's closing note the checker reads
		"the adapter's exit 3",                                                      // row 17
		"Never create a `.side/` inside a `.side/`",                                 // the flat folder
		"Swapping a marker for its link is an edit through `apply`",                 // presentation owns proposals
		"don't fold again",                                                          // a refused side-doc write after the fold landed
		"a plain relative path with no `#anchor`",                                   // the body's backlink
	} {
		// Compared with whitespace collapsed, so rewrapping the prose doesn't break it.
		if !strings.Contains(flat, phrase) {
			t.Errorf("collaboration-modes.md lacks %q", phrase)
		}
	}
}
