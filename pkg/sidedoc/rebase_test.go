package sidedoc

import (
	"path"
	"strings"
	"testing"
)

// The tree the table moves within: qa/op/ moves under s.side/ (#455).
var rebaseTree = map[string]bool{
	"s.md": true, "qa/op/a.md": true, "qa/op/b.md": true, "qa/op/a.side/01-x.md": true,
	"qa/op/assets/i.png": true, "projects/p/project.md": true, "projects/p/design.md": true,
}

func qaMoved(p string) (string, bool) { return MoveUnder(p, "qa/op", "s.side") }

func rebaseExists(p string) bool { return rebaseTree[p] }

// Each case moves the document (from → to), what it links, or both.
func TestRebaseLinks(t *testing.T) {
	for _, c := range []struct {
		name, from, to, in, want string
		dangling                 []string
	}{
		{"backlink one level shallower", "qa/op/a.md", "s.side/a.md",
			"_From [S](../../s.md)._", "_From [S](../s.md)._", nil},
		{"up-tree link keeps its fragment", "qa/op/a.md", "s.side/a.md",
			"see [p](../../projects/p/project.md#tasks)", "see [p](../projects/p/project.md#tasks)", nil},
		{"sibling stays a sibling", "qa/op/a.md", "s.side/a.md",
			"and [b](b.md) too", "and [b](b.md) too", nil},
		{"nested side doc", "qa/op/a.side/01-x.md", "s.side/a.side/01-x.md",
			"[up](../a.md) [s](../../../s.md)", "[up](../a.md) [s](../../s.md)", nil},
		{"image and title", "qa/op/a.md", "s.side/a.md",
			`![i](assets/i.png "shot") [p](../../projects/p/design.md "d")`, `![i](assets/i.png "shot") [p](../projects/p/design.md "d")`, nil},
		{"the parent's link into the moved tree", "s.md", "s.md",
			"1. ask → [answer](qa/op/a.md)", "1. ask → [answer](s.side/a.md)", nil},
		{"outside file into the moved tree", "projects/p/project.md", "projects/p/project.md",
			"Review: [qa](../../qa/op/b.md).", "Review: [qa](../../s.side/b.md).", nil},
		{"reference definition", "projects/p/project.md", "projects/p/project.md",
			"[r]: ../../qa/op/b.md", "[r]: ../../s.side/b.md", nil},
		{"parent frontmatter, plain and quoted", "qa/op/a.side/01-x.md", "s.side/a.side/01-x.md",
			"---\ntype: side-doc\nparent: ../a.md\n---\n", "---\ntype: side-doc\nparent: ../a.md\n---\n", nil},
		{"parent rebased", "qa/op/a.md", "s.side/a.md",
			"---\nparent: \"../../s.md\"\n---\nx", "---\nparent: \"../s.md\"\n---\nx", nil},
		{"fenced and code span untouched", "qa/op/a.md", "s.side/a.md",
			"```\n[s](../../s.md)\n```\n`[s](../../s.md)` [s](../../s.md)", "```\n[s](../../s.md)\n```\n`[s](../../s.md)` [s](../s.md)", nil},
		{"urls, absolute, fragment-only, escaping root", "qa/op/a.md", "s.side/a.md",
			"[u](https://x.y/a) [abs](/etc/x) [f](#top) [esc](../../../x.md)", "[u](https://x.y/a) [abs](/etc/x) [f](#top) [esc](../../../x.md)", nil},
		{"dangling is reported and left", "qa/op/a.md", "s.side/a.md",
			"[gone](projects/p/project.md) [s](../../s.md)", "[gone](projects/p/project.md) [s](../s.md)", []string{"projects/p/project.md"}},
		{"crlf", "qa/op/a.md", "s.side/a.md",
			"---\r\nparent: ../../s.md\r\n---\r\n[s](../../s.md)\r\n", "---\r\nparent: ../s.md\r\n---\r\n[s](../s.md)\r\n", nil},
	} {
		got, dangling := RebaseLinks(c.in, c.from, c.to, qaMoved, rebaseExists)
		if got != c.want || strings.Join(dangling, ",") != strings.Join(c.dangling, ",") {
			t.Errorf("%s:\n got %q %q\nwant %q %q", c.name, got, dangling, c.want, c.dangling)
		}
	}
}

// Over arbitrary Markdown: nothing changes when nothing moves, and every
// target a move rewrote resolves, from the new place, to an existing path.
func FuzzRebaseLinks(f *testing.F) {
	for _, s := range []string{
		"[s](../../s.md) [b](b.md#x)", "```\n[s](../../s.md)\n```", "`[s](../../s.md)` [a](a.side/01-x.md)",
		"---\nparent: ../../s.md\n---\n[p](../../projects/p/project.md)\r\n", "[x]( ) [y](a(b).md) ![i](assets/i.png \"t\")",
		"[r]: ../../s.md\n``` ` [s](../../s.md)",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, text string) {
		none := func(string) (string, bool) { return "", false }
		if got, _ := RebaseLinks(text, "qa/op/a.md", "qa/op/a.md", none, rebaseExists); got != text {
			t.Fatalf("changed with nothing moved:\n%q\n%q", text, got)
		}
		got, _ := RebaseLinks(text, "qa/op/a.md", "s.side/a.md", qaMoved, rebaseExists)
		before, after := targets(text), targets(got)
		if len(before) != len(after) {
			t.Fatalf("link count changed: %q → %q", before, after)
		}
		for i, tgt := range after {
			if tgt == before[i] {
				continue
			}
			p, _, _ := strings.Cut(tgt, "#")
			at := path.Join("s.side", p)
			if !rebaseExists(at) && !rebaseExists(mustMoved(at)) {
				t.Fatalf("rewritten %q → %q resolves to nothing (%s)", before[i], tgt, at)
			}
		}
	})
}

func mustMoved(p string) string {
	if to, ok := qaMoved(p); ok {
		return to
	}
	return p
}

// targets is every inline link target, as the rebase sees them.
func targets(text string) []string {
	var out []string
	for _, m := range inlineRE.FindAllStringSubmatch(text, -1) {
		out = append(out, m[1])
	}
	return out
}

func TestAnswerName(t *testing.T) {
	for name, want := range map[string]bool{
		"2026-10-18-24-finder-answer-docs.md": true, "2026-10-18-100-x.md": true,
		"01-x.md": false, "2026-10-18-x.md": false, "2026-10-18-01-Big.md": false,
		"2026-10-18-01-a--b.md": false, "2026-10-18-01-" + strings.Repeat("a", 41) + ".md": false,
	} {
		if AnswerName(name) != want {
			t.Errorf("AnswerName(%s) = %v", name, !want)
		}
	}
}
