package sidedoc

import (
	"fmt"
	"path"
	"strings"
	"testing"
)

func TestRebaseIndexCoordinates(t *testing.T) {
	docs := []string{"note.md", "projects/p/design.md", "qa/op/a.md", "deep/nested/parent.md"}
	targets := []string{"note.side/a.md", "qa/op/b.md", "deep/shared.md"}
	for _, from := range docs {
		for _, to := range docs {
			for _, target := range targets {
				for _, moveTarget := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/%t", from, to, target, moveTarget), func(t *testing.T) {
						destination := target
						if moveTarget {
							destination = "new/" + target
						}
						in := "---\ndiscussions:\n  - doc: \"" + RelPath(path.Dir(from), target) + "\"\n---\nbody\n"
						moved := func(p string) (string, bool) {
							if p != target {
								t.Fatalf("resolved target = %q, want %q", p, target)
							}
							return destination, moveTarget
						}
						got, err := RebaseIndex(in, from, to, moved)
						want := "---\ndiscussions:\n  - doc: \"" + RelPath(path.Dir(to), destination) + "\"\n---\nbody\n"
						if err != nil || got != want {
							t.Fatalf("got %q, %v; want %q", got, err, want)
						}
					})
				}
			}
		}
	}
}

func TestRebaseIndexPreservesUnknownBytes(t *testing.T) {
	in := "---\n# heading\ncustom:\n  - doc: old/a.md\n    nested: { arbitrary: [1, true] }\ndiscussions:\n    - doc: 'old/a.md' # keep\n    - doc: \"old/b.md\"\n    - doc: old/c.md  # comment\nother: |\n  - doc: old/a.md\n---\n- doc: old/a.md\n"
	want := strings.Replace(in, "    - doc: 'old/a.md'", "    - doc: 'new/a.md'", 1)
	want = strings.Replace(want, "    - doc: \"old/b.md\"", "    - doc: \"new/b.md\"", 1)
	want = strings.Replace(want, "    - doc: old/c.md", "    - doc: new/c.md", 1)
	got, err := RebaseIndex(in, "parent.md", "parent.md", func(p string) (string, bool) { return MoveUnder(p, "old", "new") })
	if err != nil || got != want {
		t.Fatalf("got %q, %v; want %q", got, err, want)
	}
}

func TestRebaseIndexRejectsMalformedKnownFields(t *testing.T) {
	cases := []string{
		"discussions: true", "discussions: null", "discussions: {}", "discussions: [a.md]",
		"discussions:\n  - doc: 123", "discussions:\n  - doc: true", "discussions:\n  - doc: null",
		"discussions:\n  - doc: []", "discussions:\n  - doc: {}", "discussions:\n  - other: a.md",
		"discussions:\n  - doc: a.md\n    doc: b.md", "discussions: []\ndiscussions: []",
		"discussions:\n  - doc: ../outside.md", "discussions:\n  - doc: /absolute.md",
		"discussions:\n  - doc: |\n      a.md", "discussions: &list []\ncopy: *list\ndiscussions: []",
	}
	for _, fm := range cases {
		t.Run(fm, func(t *testing.T) {
			in := "---\n" + fm + "\n---\nbody\n"
			got, err := RebaseIndex(in, "parent.md", "new/parent.md", nil)
			if err == nil || got != in {
				t.Fatalf("got %q, %v; want unchanged document and error", got, err)
			}
		})
	}
}

func TestRebaseIndexAbsentAndEmpty(t *testing.T) {
	for _, in := range []string{"body", "---\ncustom:\n  - doc: old.md\n---\nbody", "---\ndiscussions: []\n---\n"} {
		got, err := RebaseIndex(in, "a.md", "b/a.md", nil)
		if err != nil || got != in {
			t.Fatalf("got %q, %v", got, err)
		}
	}
}

func TestRebaseIndexRejectsInheritedKnownFields(t *testing.T) {
	for _, fm := range []string{
		"defaults: &defaults\n  discussions:\n    - doc: a.md\n<<: *defaults",
		"defaults: &defaults\n  doc: a.md\ndiscussions:\n  - <<: *defaults",
	} {
		in := "---\n" + fm + "\n---\n"
		got, err := RebaseIndex(in, "a.md", "nested/a.md", nil)
		if err == nil || got != in {
			t.Fatalf("got %q, %v; want unchanged document and error", got, err)
		}
	}
}

func TestRebaseIndexFlowAndUnknownEntryMetadata(t *testing.T) {
	in := "---\ndiscussions: [{doc: 'old/a.md', extra: {doc: old/a.md}}, {doc: \"old/b.md\"}] # keep\n---\nbody"
	want := "---\ndiscussions: [{doc: 'new/a.md', extra: {doc: old/a.md}}, {doc: \"new/b.md\"}] # keep\n---\nbody"
	got, err := RebaseIndex(in, "parent.md", "parent.md", func(p string) (string, bool) { return MoveUnder(p, "old", "new") })
	if err != nil || got != want {
		t.Fatalf("got %q, %v; want %q", got, err, want)
	}
}
