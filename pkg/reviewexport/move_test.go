package reviewexport

import (
	"fmt"
	"path"
	"strings"
	"testing"
)

// projectMove is the move contract of collaboration-modes.md §5, applied to a
// tree's files: the checker's statement of what a move of OLD to NEW leaves.
// It is a reference, not a writer: the agent follows the prose, and parley's
// mover is tested against the same fixtures (#424, parley.nvim#373).
//
//  1. A folder already named NEW.side/ is an orphan: it becomes NEW_N.side/,
//     N the first free number from 1.
//  2. The document moves to NEW and its folder to NEW.side/.
//  3. Its own links and index entries into the folder take the new name.
//  4. Each side doc whose parent: is the document points at NEW, and so does
//     its backlink.
//
// Nothing else changes: links from other documents into the folder are not
// rewritten.
func projectMove(files map[string]string, oldPath, newPath string) map[string]string {
	oldBase, newBase := strings.TrimSuffix(oldPath, ".md"), strings.TrimSuffix(newPath, ".md")
	oldDir, newDir := oldBase+".side/", newBase+".side/"
	oldName, newName := path.Base(oldBase), path.Base(newBase)
	under := func(dir string) bool {
		for p := range files {
			if strings.HasPrefix(p, dir) {
				return true
			}
		}
		return false
	}
	orphan := ""
	if under(newDir) {
		for n := 1; ; n++ {
			if d := fmt.Sprintf("%s_%d.side/", newBase, n); !under(d) {
				orphan = d
				break
			}
		}
	}
	out := map[string]string{}
	for p, text := range files {
		switch {
		case p == oldPath:
			text = strings.ReplaceAll(text, "]("+oldName+".side/", "]("+newName+".side/")
			text = strings.ReplaceAll(text, "doc: "+oldName+".side/", "doc: "+newName+".side/")
			out[newPath] = text
		case strings.HasPrefix(p, oldDir):
			text = strings.Replace(text, "\nparent: ../"+oldName+".md\n", "\nparent: ../"+newName+".md\n", 1)
			text = strings.ReplaceAll(text, "](../"+oldName+".md)", "](../"+newName+".md)")
			out[newDir+strings.TrimPrefix(p, oldDir)] = text
		case orphan != "" && strings.HasPrefix(p, newDir):
			out[orphan+strings.TrimPrefix(p, newDir)] = text
		default:
			out[p] = text
		}
	}
	return out
}

func TestProjectMove(t *testing.T) {
	side := "---\ntype: side-doc\nparent: ../design.md\nmarker: \"q\"\ncollaboration-mode: discussion\n---\n\n# Q\n\n_From [D](../design.md), section \"S\"._\n"
	nested := "---\ntype: side-doc\nparent: 01-q.md\n---\n\n_From [Q](01-q.md), section \"Q\"._\n"
	in := map[string]string{
		"design.md":           "---\ndiscussions:\n  - doc: design.side/01-q.md\n---\n# D\n\n→ [discussion](design.side/01-q.md)\n",
		"design.side/01-q.md": side,
		"design.side/02-n.md": nested,
		"spec.side/01-old.md": "orphan",
		"spec_1.side/01-x.md": "an earlier orphan",
		"other.md":            "→ [discussion](design.side/01-q.md)\n",
	}
	got := projectMove(in, "design.md", "spec.md")
	want := map[string]string{
		"spec.md":               "---\ndiscussions:\n  - doc: spec.side/01-q.md\n---\n# D\n\n→ [discussion](spec.side/01-q.md)\n",
		"spec.side/01-q.md":     strings.ReplaceAll(side, "../design.md", "../spec.md"),
		"spec.side/02-n.md":     nested,
		"spec_2.side/01-old.md": "orphan",
		"spec_1.side/01-x.md":   "an earlier orphan",
		"other.md":              "→ [discussion](design.side/01-q.md)\n",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d files, want %d: %v", len(got), len(want), got)
	}
	for p, w := range want {
		if got[p] != w {
			t.Errorf("%s:\ngot  %q\nwant %q", p, got[p], w)
		}
	}
}
