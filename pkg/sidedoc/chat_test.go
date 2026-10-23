package sidedoc

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

// The chat carve-out grammar of collaboration-modes.md §3 (#575), with the
// legacy forms of §7 that still mean chat.
func TestChatSections(t *testing.T) {
	type sec struct {
		Heading, End int
		Source       string
		Legacy       bool
	}
	for _, tc := range []struct {
		name      string
		doc       string
		want      []sec
		misplaced []int
		notices   []string // substrings, one per notice, in order
	}{
		{
			name: "tag above a heading: through the next same-level heading, subsections included",
			doc:  "# T\n\n@@chat@@\n## Ops\n1. a\n### Sub\n2. b\n## Next\ntext\n",
			want: []sec{{4, 7, "tag", false}},
		},
		{
			name: "tag with surrounding spaces",
			doc:  "  @@chat@@  \n## Ops\n1. a\n",
			want: []sec{{2, 3, "tag", false}},
		},
		{
			name:      "blank line between tag and heading: ordinary text, reported",
			doc:       "@@chat@@\n\n## Ops\n1. a\n",
			misplaced: []int{1},
		},
		{
			name:      "tag above prose: ordinary text, reported",
			doc:       "## A\n@@chat@@\nprose\n",
			misplaced: []int{2},
		},
		{
			name: "tag inside a fence: code, not a tag",
			doc:  "```\n@@chat@@\n## Ops\n```\n## Real\n",
		},
		{
			name:      "tag above a fenced heading: no heading follows it",
			doc:       "@@chat@@\n```\n## Ops\n```\n",
			misplaced: []int{1},
		},
		{
			name:      "tag above a setext heading: CommonMark folds it into the heading, so it is misplaced",
			doc:       "@@chat@@\nOps\n---\n1. a\n",
			misplaced: []int{1},
		},
		{
			name: "tag in indented code directly above a heading: code, not a tag",
			doc:  "para\n\n    @@chat@@\n## Ops\n1. a\n",
		},
		{
			name:    "a map default of a retired or unknown value is named",
			doc:     "---\ncollaboration-mode: {default: debate, chat: [Q]}\n---\n## Q\n- a\n",
			want:    []sec{{4, 5, "legacy-map", true}},
			notices: []string{"collaboration-mode map", "map default: debate"},
		},
		{
			name:    "CRLF frontmatter is not read, and says so",
			doc:     "---\r\ncollaboration-mode: chat\r\n---\r\n- a\r\n",
			notices: []string{"frontmatter could not be read"},
		},
		{
			name: "a leading thematic break with no closing fence is not frontmatter: no notice",
			doc:  "---\nJust a rule above prose.\n",
		},
		{
			name:    "an empty collaboration-mode is empty or malformed, not retired",
			doc:     "---\ncollaboration-mode:\n---\n",
			notices: []string{"collaboration-mode is empty or malformed; ignored"},
		},
		{
			name:    "LF frontmatter with a CR in the body is not read, and says so",
			doc:     "---\ncollaboration-mode: chat\n---\n- a\r\n",
			notices: []string{"frontmatter could not be read"},
		},
		{
			name:    "a document marker inside frontmatter: not read, and says so",
			doc:     "---\ncollaboration-mode: chat\n...\n---\n- a\n",
			notices: []string{"frontmatter could not be read"},
		},
		{
			name:    "YAML that does not parse: not read, and says so",
			doc:     "---\ncollaboration-mode: {chat: [Q\n---\n## Q\n- a\n",
			notices: []string{"frontmatter could not be read"},
		},
		{
			name: "other @@tag@@ lines are parley's, not review's",
			doc:  "@@label@@\n## Ops\n",
		},
		{
			name:    "legacy scalar chat: the whole document",
			doc:     "---\ncollaboration-mode: chat\n---\n# T\n1. a\n",
			want:    []sec{{0, 5, "legacy-scalar", true}},
			notices: []string{"collaboration-mode: chat"},
		},
		{
			name:    "legacy block map: named sections by normalized name",
			doc:     "---\ncollaboration-mode:\n  default: discussion\n  chat: [\"operator  Instructions\"]\n---\n# T\n### Operator instructions (open)\n- a\n## Next\n",
			want:    []sec{{7, 8, "legacy-map", true}},
			notices: []string{"collaboration-mode map", "map default: discussion"},
		},
		{
			name:    "legacy flow map",
			doc:     "---\ncollaboration-mode: {chat: [Notes]}\n---\n## Notes\n- a\n",
			want:    []sec{{4, 5, "legacy-map", true}},
			notices: []string{"collaboration-mode map"},
		},
		{
			name:    "legacy map naming no heading: noticed, no section",
			doc:     "---\ncollaboration-mode: {chat: [Missing]}\n---\n## Notes\n",
			notices: []string{"collaboration-mode map", "\"Missing\" matches no heading"},
		},
		{
			name:    "legacy heading comment",
			doc:     "## Q\n<!-- collaboration-mode: chat -->\n- a\n",
			want:    []sec{{1, 3, "legacy-comment", true}},
			notices: []string{"<!-- collaboration-mode: chat -->"},
		},
		{
			name:    "tag and legacy comment on one heading: one section, the tag's, keeping legacy provenance",
			doc:     "@@chat@@\n## Q\n<!-- collaboration-mode: chat -->\n1. a\n",
			want:    []sec{{2, 4, "tag", true}},
			notices: []string{"<!-- collaboration-mode: chat -->"},
		},
		{
			name:    "retired values and keys: notices only",
			doc:     "---\ncollaboration-mode: collaboration\npresentation-mode: propose\nprocessing-mode: copy-editing\n---\n## Q\n<!-- processing-mode: proofreading -->\n",
			notices: []string{"collaboration-mode: collaboration", "presentation-mode, processing-mode", "<!-- processing-mode: proofreading -->"},
		},
		{
			name:    "an unknown value never fails: a notice",
			doc:     "---\ncollaboration-mode: debate\n---\n",
			notices: []string{"collaboration-mode: debate"},
		},
		{
			name:    "every retired key is named, legacy aliases included",
			doc:     "---\npresentation: direct\nprocess-instruction: \"[copy-edit] x\"\nprocessing-instruction: y\n---\n",
			notices: []string{"presentation, process-instruction, processing-instruction"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ChatSections(tc.doc)
			var secs []sec
			for _, s := range got.Sections {
				secs = append(secs, sec{s.Heading, s.End, s.Source, s.Legacy})
			}
			if !reflect.DeepEqual(secs, tc.want) {
				t.Errorf("sections %+v, want %+v", secs, tc.want)
			}
			if !reflect.DeepEqual(got.Misplaced, tc.misplaced) {
				t.Errorf("misplaced %v, want %v", got.Misplaced, tc.misplaced)
			}
			if len(got.Notices) != len(tc.notices) {
				t.Fatalf("notices %q, want %d matching %q", got.Notices, len(tc.notices), tc.notices)
			}
			for i, want := range tc.notices {
				if !strings.Contains(got.Notices[i], want) {
					t.Errorf("notice %d %q lacks %q", i, got.Notices[i], want)
				}
			}
		})
	}
}

func TestChatSectionContains(t *testing.T) {
	s := ChatSections("@@chat@@\n## Ops\n1. a\n## Next\nb\n")
	if !s.InChat(3) || s.InChat(5) || s.InChat(1) {
		t.Errorf("InChat over %+v", s.Sections)
	}
}

// pair#838 calls ChatSections once per snapshot, so its work must grow
// linearly in the document: N and 2N sections, each with a tag, a nested
// subsection and items, and N legacy map names. A per-heading prefix count
// or a rescan of later headings would roughly quadruple; linear doubles.
func TestChatSectionsWorkIsLinear(t *testing.T) {
	doc := func(n int) string {
		var b strings.Builder
		b.WriteString("---\ncollaboration-mode:\n  chat:\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "    - Named %d\n", i)
		}
		b.WriteString("---\n# Doc\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "@@chat@@\n## Ops %d\n1. why %d?\n### Sub %d\n2. how?\n## Named %d\ntext\n", i, i, i, i)
		}
		return b.String()
	}
	scan, w1 := chatSections(doc(200))
	if len(scan.Sections) != 400 {
		t.Fatalf("%d sections at N=200, want 400", len(scan.Sections))
	}
	_, w2 := chatSections(doc(400))
	if w2 > 2*w1+w1/10 {
		t.Errorf("work %d at N=200, %d at N=400: superlinear (ratio %.2f)", w1, w2, float64(w2)/float64(w1))
	}
}

func TestChatSectionsTagOverLegacyMap(t *testing.T) {
	scan := ChatSections("---\ncollaboration-mode: {chat: [Q]}\n---\n@@chat@@\n## Q\n- a\n")
	if len(scan.Sections) != 1 || scan.Sections[0].Source != "tag" || !scan.Sections[0].Legacy {
		t.Fatalf("a tagged heading the map also names: %+v, want one tag section with Legacy", scan.Sections)
	}
}

// Sections are laminar and Innermost picks the deepest: checked against an
// independent reading of random documents (heading levels, tags, map names,
// comments), not against the implementation's own stack.
func TestChatSectionsLaminar(t *testing.T) {
	rng := rand.New(rand.NewPCG(575, 576))
	for round := 0; round < 500; round++ {
		var b strings.Builder
		b.WriteString("---\ncollaboration-mode: {chat: [H1, H4]}\n---\n")
		n := 1 + rng.IntN(12)
		for i := 0; i < n; i++ {
			if rng.IntN(3) == 0 {
				b.WriteString("@@chat@@\n")
			}
			fmt.Fprintf(&b, "%s H%d\n", strings.Repeat("#", 1+rng.IntN(4)), i)
			if rng.IntN(4) == 0 {
				b.WriteString("<!-- collaboration-mode: chat -->\n")
			}
			b.WriteString("1. item\n")
		}
		doc := b.String()
		scan := ChatSections(doc)
		owners := map[int]string{}
		for _, c := range scan.Sections {
			owners[c.Heading] = c.Source
		}
		for _, tag := range scan.Tags {
			if owners[tag+1] != "tag" {
				t.Fatalf("tag at line %d owns no section on line %d:\n%s", tag, tag+1, doc)
			}
		}
		seen := map[int]bool{}
		for i, a := range scan.Sections {
			if seen[a.Heading] {
				t.Fatalf("heading line %d is two sections:\n%s", a.Heading, doc)
			}
			seen[a.Heading] = true
			for _, c := range scan.Sections[i+1:] {
				disjoint := a.End < c.Heading || c.End < a.Heading
				nested := (a.Heading <= c.Heading && c.End <= a.End) || (c.Heading <= a.Heading && a.End <= c.End)
				if !disjoint && !nested {
					t.Fatalf("sections %+v and %+v overlap without nesting:\n%s", a, c, doc)
				}
			}
		}
		for line := 1; line <= strings.Count(doc, "\n"); line++ {
			got, ok := scan.Innermost(line)
			deepest, any := ChatSection{}, false
			for _, c := range scan.Sections {
				if line > c.Heading && line <= c.End && (!any || c.End-c.Heading < deepest.End-deepest.Heading) {
					deepest, any = c, true
				}
			}
			if ok != any || ok && got != deepest {
				t.Fatalf("line %d: Innermost %+v/%v, want %+v/%v\n%s", line, got, ok, deepest, any, doc)
			}
		}
	}
}

// Tags lists every valid tag, including those a whole-document legacy chat
// suppresses as sections, and never a misplaced one.
func TestChatSectionsTags(t *testing.T) {
	for _, tc := range []struct {
		name            string
		doc             string
		tags, misplaced []int
		sections        int
	}{
		{"tag sections", "@@chat@@\n## A\n1. a\n@@chat@@\n\n## B\n", []int{1}, []int{4}, 1},
		{"whole-document legacy keeps its tags", "---\ncollaboration-mode: chat\n---\n@@chat@@\n## A\n1. a\n", []int{4}, nil, 1},
		{"code is no tag", "```\n@@chat@@\n## A\n```\n", nil, nil, 0},
	} {
		scan := ChatSections(tc.doc)
		if !reflect.DeepEqual(scan.Tags, tc.tags) || !reflect.DeepEqual(scan.Misplaced, tc.misplaced) || len(scan.Sections) != tc.sections {
			t.Errorf("%s: tags %v misplaced %v sections %d, want %v %v %d", tc.name, scan.Tags, scan.Misplaced, len(scan.Sections), tc.tags, tc.misplaced, tc.sections)
		}
	}
}

