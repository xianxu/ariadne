// Package linkguard judges an issue's process documents before their text
// reaches main (#579): a relative link to a file that is not committed, or
// that git ignores, would publish a link that resolves nowhere. Side docs of
// process artifacts are committed with them; a private session side doc is
// folded in, never linked. Everything here is pure over the injected state.
package linkguard

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// LinkState is what the repository holds at a link's target.
type LinkState int

const (
	Committed   LinkState = iota // in the tree the verb publishes
	Uncommitted                  // on disk, not in that tree, not ignored
	Ignored                      // on disk, not in that tree, ignored by git
	Absent                       // nowhere: a dangling link, not this guard's
)

// Doc is one process document: its root-relative slash path and its text.
type Doc struct {
	Path string
	Text string
}

// Refusal is one link the guard refuses.
type Refusal struct {
	Doc    string
	Line   int
	Target string // as written
	Path   string // root-relative
	State  LinkState
}

// Census walks docs' local links, asking each target's state once, and
// refuses those Uncommitted or Ignored. A linked Markdown file inside a side
// folder is walked in turn (read from disk), so a nested side doc's links
// count too; a visited set ends cycles. Absent targets pass.
func Census(docs []Doc, state func(path string) (LinkState, error), read func(path string) (string, error)) ([]Refusal, error) {
	asked := map[string]LinkState{}
	walked := map[string]bool{}
	queue := append([]Doc(nil), docs...)
	for _, d := range docs {
		walked[d.Path] = true
	}
	var out []Refusal
	for len(queue) > 0 {
		d := queue[0]
		queue = queue[1:]
		for _, l := range sidedoc.LocalLinks(d.Text, d.Path) {
			st, ok := asked[l.Path]
			if !ok {
				var err error
				if st, err = state(l.Path); err != nil {
					return nil, fmt.Errorf("%s:%d: %s: %w", d.Path, l.Line, l.Target, err)
				}
				asked[l.Path] = st
			}
			if st == Uncommitted || st == Ignored {
				out = append(out, Refusal{Doc: d.Path, Line: l.Line, Target: l.Target, Path: l.Path, State: st})
			}
			if st == Absent || walked[l.Path] || !strings.HasSuffix(l.Path, ".md") || !sidedoc.InSideDir(l.Path) {
				continue
			}
			walked[l.Path] = true
			text, err := read(l.Path)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", l.Path, err)
			}
			queue = append(queue, Doc{Path: l.Path, Text: text})
		}
	}
	return out, nil
}

// FormatRefusals is one line per refusal, each naming its fix.
func FormatRefusals(rs []Refusal) string {
	var b strings.Builder
	for _, r := range rs {
		fix := fmt.Sprintf("commit %q with the artifact, or fold what it settled in and drop the link", r.Path)
		what := "is not committed"
		if r.State == Ignored {
			what = "is git-ignored (a private session side doc?)"
			fix = "fold what it settled into the artifact and drop the link"
		}
		fmt.Fprintf(&b, "  %s:%d links %q, which %s: %s\n", r.Doc, r.Line, r.Target, what, fix)
	}
	return b.String()
}
