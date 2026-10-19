package main

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// The one-time move of the hub's answer docs from `qa/<operator>/` into the
// session file's side folder (#455). planMigration is pure over a snapshot of
// the tree; migrate_apply.go reads the snapshot and runs the plan. Every path
// here is a slash path relative to the session file's directory (the root).

// entryClass is what the migration does with one entry under qa/<operator>/.
type entryClass int

const (
	classDoc        entryClass = iota // *.md: an answer doc or a nested side doc, moved and rewritten
	classAttachment                   // any other regular file, moved byte for byte
	classForeign                      // a dot path another tool keeps: left in place, listed
	classLock                         // *.md.lock: the operator holds that doc
	classClobbered                    // *.md.clobbered: couch's clobber evidence, resolved by hand
	classOdd                          // a symlink or other non-regular entry, never followed
)

// classifyEntry sorts an entry under qa/<operator>/ by its path relative to
// it and whether it is a regular file. Couch's session-sync sidecars are the
// dot names and the two suffixes, the set the ops .gitignore names for status
// files.
func classifyEntry(rel string, regular bool) entryClass {
	for _, seg := range strings.Split(rel, "/") {
		if strings.HasPrefix(seg, ".") {
			return classForeign
		}
	}
	switch {
	case !regular:
		return classOdd
	case strings.HasSuffix(rel, ".md.lock"):
		return classLock
	case strings.HasSuffix(rel, ".md.clobbered"):
		return classClobbered
	case strings.HasSuffix(rel, ".md"):
		return classDoc
	}
	return classAttachment
}

// migEntry is one entry under qa/<operator>/ as the snapshot read it.
type migEntry struct {
	Rel     string // relative to qa/<operator>/
	Regular bool
	Bytes   []byte
	Mtime   time.Time
}

// migSnapshot is everything the plan depends on.
type migSnapshot struct {
	Session     string // the session file's name, e.g. tl-status-op.md
	SessionText string
	QA          string            // qa/<operator>, relative to the root
	Entries     []migEntry        // every entry under QA
	Dest        map[string][]byte // files already in the side folder tree, by root path
	Outside     map[string][]byte // other *.md under the root, by root path
	OutsideTime map[string]time.Time
	Registered  map[string]bool // couch-registered files, by root path
	Locked      map[string]bool // registered files the operator holds, by root path
	Exists      func(rel string) bool
	// Warnings are facts the shell found that the TL should act on, such as
	// a side folder git would track where qa/ was ignored.
	Warnings []string
}

// fileMove is one file's move.
type fileMove struct {
	Old, New   string
	Bytes      []byte
	Mtime      time.Time
	Registered bool
	Resume     bool   // New already holds Bytes: a run stopped after creating it
	Marker     string // for an answer doc: "bullet", "title" or "kept"
}

type migConflict struct{ Path, Why string }

type migNote struct{ Path, Text string }

type migrationPlan struct {
	Side      string // the side folder, e.g. tl-status-op.side
	Moves     []fileMove
	Left      []string // foreign entries, left in place
	Conflicts []migConflict
	Session   []editRecord
	Outside   map[string][]byte // rebased outside files, by root path
	// OutsideEdits are the apply records for outside files couch tracks.
	OutsideEdits map[string][]editRecord
	Warnings     []string
	Mentions     []migNote // bare references to qa/ left as written
	Dangling     []migNote // link targets that resolve to nothing, left as written
}

const migrationExplain = "#455: the answer moved from qa/ into the session file's side folder"

var (
	bulletPrefixRE = regexp.MustCompile(`^\s*(?:\d+\.|[-*])\s+`)
	answerLinkRE   = regexp.MustCompile(`\s*→ \[answer\]\([^)\s]*\)`)
)

// planMigration plans the move of s.QA's entries into the side folder.
func planMigration(s migSnapshot) migrationPlan {
	side := sidedoc.SideDir(s.Session)
	p := migrationPlan{Side: side, Outside: map[string][]byte{}, OutsideEdits: map[string][]editRecord{}, Warnings: s.Warnings}
	moved := func(rel string) (string, bool) { return sidedoc.MoveUnder(rel, s.QA, side) }
	exists := func(rel string) bool {
		if s.Exists != nil && s.Exists(rel) {
			return true
		}
		_, ok := s.Dest[rel]
		return ok
	}
	note := func(list *[]migNote, file string, items []string) {
		for _, it := range items {
			*list = append(*list, migNote{file, it})
		}
	}
	if s.Locked[s.Session] {
		p.Conflicts = append(p.Conflicts, migConflict{s.Session, "the operator holds the session file's lock; wait for their submit, then rerun"})
	}
	markers := bulletMarkers(s.SessionText, s.QA)
	title := docTitle(s.SessionText, s.Session)

	for _, e := range s.Entries {
		old := s.QA + "/" + e.Rel
		switch classifyEntry(e.Rel, e.Regular) {
		case classForeign:
			p.Left = append(p.Left, old)
			continue
		case classLock:
			p.Conflicts = append(p.Conflicts, migConflict{old, "the operator holds this doc's lock; wait for their submit, then rerun"})
			continue
		case classClobbered:
			p.Conflicts = append(p.Conflicts, migConflict{old, "couch's clobber evidence; resolve it by hand first"})
			continue
		case classOdd:
			p.Conflicts = append(p.Conflicts, migConflict{old, "not a regular file (a symlink?); move it by hand"})
			continue
		}
		next, _ := moved(old)
		// A stopped run may have moved the registration already: the state
		// lives at either path, so both are read (BR-1).
		m := fileMove{Old: old, New: next, Bytes: e.Bytes, Mtime: e.Mtime, Registered: s.Registered[old] || s.Registered[next]}
		if s.Locked[old] || s.Locked[next] {
			p.Conflicts = append(p.Conflicts, migConflict{old, "the operator holds this doc's lock; wait for their submit, then rerun"})
			continue
		}
		dest, atDest := s.Dest[next]
		if strings.HasSuffix(e.Rel, ".md") {
			text, dangling := sidedoc.RebaseLinks(string(e.Bytes), old, next, moved, exists)
			note(&p.Dangling, old, dangling)
			if !strings.Contains(e.Rel, "/") {
				marker, how := markers[old], "bullet"
				if atDest {
					if f, ok := sidedoc.Fields(string(dest)); ok && f["marker"] != "" {
						marker, how = f["marker"], "kept" // pinned on resume
					}
				}
				if marker == "" {
					marker, how = yamlQuote(answerAsk(text, path.Base(e.Rel))), "title"
				}
				rewritten, err := answerFrontmatter(text, "../"+s.Session, marker, title)
				if err != nil {
					p.Conflicts = append(p.Conflicts, migConflict{old, err.Error()})
					continue
				}
				text, m.Marker = rewritten, how
			}
			m.Bytes = []byte(text)
			note(&p.Mentions, next, mentions(text, s.QA))
		}
		if atDest {
			if string(dest) != string(m.Bytes) {
				p.Conflicts = append(p.Conflicts, migConflict{next, "already exists with other content than " + old + " would move there; resolve by hand, then rerun"})
				continue
			}
			m.Resume = true
		}
		p.Moves = append(p.Moves, m)
	}

	// The session file: its links and its discussions: index, in one apply.
	after, dangling := sidedoc.RebaseLinks(s.SessionText, s.Session, s.Session, moved, exists)
	after = rebaseIndex(after, s.Session, moved)
	note(&p.Dangling, s.Session, dangling)
	recs, err := lineRecords(s.SessionText, after, migrationExplain)
	if err != nil {
		p.Conflicts = append(p.Conflicts, migConflict{s.Session, err.Error()})
	}
	p.Session = recs
	note(&p.Mentions, s.Session, mentions(after, s.QA))

	// Every other Markdown file under the root that links into qa/.
	for _, f := range sortedKeys(s.Outside) {
		before := string(s.Outside[f])
		after, dangling := sidedoc.RebaseLinks(before, f, f, moved, exists)
		after = rebaseIndex(after, f, moved)
		if after != before {
			p.Outside[f] = []byte(after)
			note(&p.Dangling, f, dangling)
			if s.Registered[f] {
				if s.Locked[f] {
					p.Conflicts = append(p.Conflicts, migConflict{f, "the operator holds this file's lock; wait for their submit, then rerun"})
				}
				recs, err := lineRecords(before, after, migrationExplain)
				if err != nil {
					p.Conflicts = append(p.Conflicts, migConflict{f, err.Error()})
				}
				p.OutsideEdits[f] = recs
			}
		}
		note(&p.Mentions, f, mentions(after, s.QA))
	}
	return p
}

// bulletMarkers maps each answer doc an Operator instructions bullet links to
// the ask, as §4 writes `marker:`: the bullet's first line without its
// number or bullet and without its link, YAML-quoted.
func bulletMarkers(session, qa string) map[string]string {
	out := map[string]string{}
	_, after, ok := strings.Cut(session, "\n## "+instructionsHeading+"\n")
	if !ok {
		return out
	}
	section, _, _ := strings.Cut(after, "\n## ")
	for _, b := range instructionBullets(section) {
		ask := strings.TrimSpace(answerLinkRE.ReplaceAllString(bulletPrefixRE.ReplaceAllString(b.Lines[0], ""), ""))
		for _, l := range b.Lines {
			for _, t := range sidedoc.SideLinks(l) {
				if c := path.Clean(t); strings.HasPrefix(c, qa+"/") {
					out[c] = yamlQuote(ask)
				}
			}
		}
	}
	return out
}

// answerAsk is the ask of an answer doc whose bullet is gone: its `# ` title
// without a leading `Q: `, or its file name when it has none.
func answerAsk(text, name string) string {
	for _, l := range strings.Split(text, "\n") {
		if t, ok := strings.CutPrefix(strings.TrimRight(l, "\r"), "# "); ok {
			return strings.TrimSpace(strings.TrimPrefix(t, "Q: "))
		}
	}
	return strings.TrimSuffix(name, ".md")
}

// docTitle is a document's `# ` title, or its file name when it has none.
func docTitle(text, name string) string {
	_, body, _ := sidedoc.FrontBlock(text)
	for _, l := range strings.Split(body, "\n") {
		if t, ok := strings.CutPrefix(strings.TrimRight(l, "\r"), "# "); ok {
			return strings.TrimSpace(t)
		}
	}
	return name
}

// yamlQuote writes s as a YAML double-quoted string on one line (§4).
func yamlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}

// answerFrontmatter gives an answer doc the side-doc frontmatter (§4) in its
// order: type, parent, marker, collaboration-mode, then process-instruction
// when the doc has one (#455 decision 3), then any other keys as they were.
// Under the title it puts the backlink, keeping the line that was there
// (decision 6). A doc that already has both is returned unchanged.
func answerFrontmatter(text, parent, markerYAML, parentTitle string) (string, error) {
	block, body, ok := sidedoc.FrontBlock(text)
	if !ok {
		if strings.HasPrefix(text, "---") {
			return "", fmt.Errorf("its frontmatter block is not closed; fix it by hand")
		}
		block, body = "", "\n"+text
	}
	if strings.Contains(block, "\r") {
		return "", fmt.Errorf("its frontmatter has CRLF line endings; fix it by hand")
	}
	var keep, pi []string
	lines := strings.Split(strings.TrimPrefix(strings.TrimSuffix(block, "\n---"), "---\n"), "\n")
	if block == "" {
		lines = nil
	}
	current := ""
	for _, l := range lines {
		if l != "" && l[0] != ' ' && l[0] != '\t' && l[0] != '-' {
			k, _, ok := strings.Cut(l, ":")
			if !ok {
				return "", fmt.Errorf("its frontmatter line %q is not `key: value`; fix it by hand", l)
			}
			current = k
		}
		switch current {
		case "type", "parent", "marker", "collaboration-mode":
		case "process-instruction":
			pi = append(pi, l)
		default:
			keep = append(keep, l)
		}
	}
	front := append([]string{"---", "type: side-doc", "parent: " + parent, "marker: " + markerYAML, "collaboration-mode: discussion"}, pi...)
	front = append(append(front, keep...), "---")
	backlink := fmt.Sprintf(`_From [%s](%s), section "%s"._`, escapeLinkText(parentTitle), parent, instructionsHeading)
	return strings.Join(front, "\n") + withBacklink(body, backlink), nil
}

// withBacklink puts backlink on its own line under body's `# ` title, unless
// the first non-blank line under it already is the backlink.
func withBacklink(body, backlink string) string {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "# ") {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				continue
			}
			if lines[j] == backlink {
				return body
			}
			break
		}
		out := append(slices.Clone(lines[:i+1]), "", backlink)
		return strings.Join(append(out, lines[i+1:]...), "\n")
	}
	return body // no title: §4's backlink sits under one, so there is nowhere to put it
}

func escapeLinkText(s string) string {
	return strings.NewReplacer("[", `\[`, "]", `\]`).Replace(s)
}

// rebaseIndex maps the discussions: entries of file's frontmatter that name
// moved files (`  - doc: qa/op/x.md`). Entries are relative to file, like
// its links.
func rebaseIndex(text, file string, moved func(string) (string, bool)) string {
	block, rest, ok := sidedoc.FrontBlock(text)
	if !ok {
		return text
	}
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		v, ok := strings.CutPrefix(strings.TrimSpace(l), "- doc:")
		if !ok {
			continue
		}
		at := path.Join(path.Dir(file), strings.TrimSpace(v))
		if to, ok := moved(at); ok {
			lines[i] = l[:strings.Index(l, "- doc:")] + "- doc: " + sidedoc.RelPath(path.Dir(file), to)
		}
	}
	return strings.Join(lines, "\n") + rest
}

// mentions are the lines that still name qa after the rebase: bare paths in
// prose, code spans or `sources:` lists, listed for the TL rather than
// rewritten (#455 decision 5).
func mentions(text, qa string) []string {
	var out []string
	for i, l := range strings.Split(text, "\n") {
		if strings.Contains(l, qa) {
			t := strings.TrimSpace(l)
			if len(t) > 120 {
				t = t[:117] + "..."
			}
			out = append(out, fmt.Sprintf("%d: %s", i+1, t))
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
