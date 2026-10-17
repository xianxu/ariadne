package sidedoc

import (
	"slices"
	"strings"
	"testing"
)

func TestSideLinks(t *testing.T) {
	doc := "# D\n\nA → [discussion](d.side/01-a.md) and → [answer](d.side/2026-10-17-01-b.md).\n" +
		"Decided (operator): x → [notes](d.side/02-c.md) (folded 2026-10-17)\n" +
		"```\n→ [discussion](d.side/99-shown.md)\n```\n" +
		"→ [link](d.side/03-not-a-side-link.md)\n"
	want := []string{"d.side/01-a.md", "d.side/2026-10-17-01-b.md", "d.side/02-c.md"}
	if got := SideLinks(doc); !slices.Equal(got, want) {
		t.Errorf("SideLinks = %v, want %v", got, want)
	}
}

const linked = "\n# D\n\nOne → [discussion](d.side/01-a.md)\nTwo → [discussion](d.side/02-b.md)\n"

func doc(index string) string {
	return "---\ntype: design\nmarker: \"x: y\"\n" + index + "process-instruction: \"[free-form]\"\n---\n" + linked
}

// The one-way rule over a document's index (#424).
func TestReconcileIndex(t *testing.T) {
	for _, c := range []struct {
		name, index string
		add         []string
		want        string // the index block after; "" means the document is unchanged
		dropped     []string
	}{
		{"links kept, nothing to do", "discussions:\n  - doc: d.side/01-a.md\n", nil, "", nil},
		{"a deleted link drops its entry", "discussions:\n  - doc: d.side/01-a.md\n  - doc: d.side/09-gone.md\n", nil,
			"discussions:\n  - doc: d.side/01-a.md\n", []string{"d.side/09-gone.md"}},
		{"an add appends at the end", "discussions:\n  - doc: d.side/02-b.md\n", []string{"d.side/01-a.md"},
			"discussions:\n  - doc: d.side/02-b.md\n  - doc: d.side/01-a.md\n", nil},
		{"an add already listed keeps its place", "discussions:\n  - doc: d.side/01-a.md\n  - doc: d.side/02-b.md\n", []string{"./d.side/01-a.md"}, "", nil},
		{"a reviewed (removed) entry is not re-added for its link", "discussions:\n", nil, "", nil},
		{"no key gains one only when something is added", "", []string{"d.side/01-a.md"}, "discussions:\n  - doc: d.side/01-a.md\n", nil},
		{"no key and nothing added", "", nil, "", nil},
		{"an empty flow list", "discussions: []\n", []string{"d.side/02-b.md"}, "discussions:\n  - doc: d.side/02-b.md\n", nil},
		{"a quoted entry reads", "discussions:\n  - doc: \"d.side/01-a.md\"\n", nil, "", nil},
	} {
		in := doc(c.index)
		got, r, err := ReconcileIndex(in, c.add)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		want := in
		if c.want != "" {
			idx := c.index
			if idx == "" {
				// a new key lands at the end of the frontmatter
				want = strings.Replace(in, "\n---\n", "\n"+strings.TrimSuffix(c.want, "\n")+"\n---\n", 1)
			} else {
				want = strings.Replace(in, idx, c.want, 1)
			}
		}
		if got != want {
			t.Errorf("%s:\ngot  %q\nwant %q", c.name, got, want)
		}
		if !slices.Equal(r.Dropped, c.dropped) || (c.want == "") == r.Changed() {
			t.Errorf("%s: report %+v", c.name, r)
		}
	}
}

func TestReconcileIndexNoFrontmatter(t *testing.T) {
	got, _, err := ReconcileIndex(linked, []string{"d.side/01-a.md"})
	if err != nil || got != "---\ndiscussions:\n  - doc: d.side/01-a.md\n---\n"+linked {
		t.Errorf("got %q, %v", got, err)
	}
}

// A hand-edited index it can't read refuses and leaves the document as it was.
func TestReconcileIndexRefuses(t *testing.T) {
	for _, index := range []string{
		"discussions: d.side/01-a.md\n",
		"discussions:\n  - d.side/01-a.md\n",
		"discussions:\n  - doc: /abs/01-a.md\n",
		"discussions:\n  - doc: ../other.side/01-a.md\n",
		"discussions:\n  - doc:\n",
	} {
		in := doc(index)
		got, _, err := ReconcileIndex(in, []string{"d.side/02-b.md"})
		if err == nil || got != in {
			t.Errorf("%q accepted: %v", index, err)
		}
	}
	if _, _, err := ReconcileIndex(doc(""), []string{"../x.md"}); err == nil {
		t.Error("an add climbing out accepted")
	}
}

// Whatever the frontmatter holds, a reconcile either refuses with the document
// unchanged, or keeps every line outside the index byte for byte and leaves
// an index that reads back as the entries it reported.
func FuzzReconcileIndex(f *testing.F) {
	for _, s := range []string{"discussions:\n  - doc: d.side/01-a.md\n", "discussions: x\n", "discussions:\n- doc: d.side/02-b.md\n  - doc: \"q\n", "", "discussions:\n\n  - doc: z\n"} {
		f.Add(s, "d.side/02-b.md")
	}
	f.Fuzz(func(t *testing.T, index, add string) {
		in := doc(index)
		got, r, err := ReconcileIndex(in, []string{add})
		if err != nil {
			if got != in {
				t.Fatalf("refused but changed the document")
			}
			return
		}
		entries, err := Entries(got)
		if err != nil || !slices.Equal(entries, r.Entries) && len(r.Entries)+len(entries) > 0 {
			t.Fatalf("index reads back %v (%v), report says %v", entries, err, r.Entries)
		}
		for _, keep := range []string{"type: design\n", "marker: \"x: y\"\n", "process-instruction: \"[free-form]\"\n", linked} {
			if strings.Contains(in, keep) && !strings.Contains(got, keep) {
				t.Fatalf("lost %q", keep)
			}
		}
	})
}
