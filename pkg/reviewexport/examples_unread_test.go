package reviewexport

import (
	"strings"
	"testing"
	"testing/fstest"
)

// #581 review (BR-1): the checker excuses an index error only for an unread
// header in a document with no side link; a readable header with a malformed
// index, and an unread header beside a side link, still fail.
func TestCheckCaseExcusesOnlyAnUnreadHeader(t *testing.T) {
	caseMD := "# x\n\nsubmitted: note.md\noutcome: ok\n\n## Reply\n\n- Retries: edited.\n"
	for _, tc := range []struct {
		name, doc string
		fails     bool
	}{
		{"unread header, no side link", "---\ntitle: x\n ---\n\nText.\n", false},
		{"readable header, malformed index", "---\ndiscussions: [\n---\n\nText.\n", true},
		{"unread header beside a side link", "---\ntitle: x\n ---\n\nText. → [discussion](note.side/2026-10/2026-10-23-01-x.md)\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"case.md":        {Data: []byte(caseMD)},
				"before/note.md": {Data: []byte(tc.doc)},
				"after/note.md":  {Data: []byte(tc.doc)},
			}
			var hit []string
			for _, e := range checkCase(fsys) {
				if strings.Contains(e, "note.md:") && (strings.Contains(e, "frontmatter") || strings.Contains(e, "discussions")) {
					hit = append(hit, e)
				}
			}
			if (len(hit) > 0) != tc.fails {
				t.Errorf("index errors %q, want failure %v", hit, tc.fails)
			}
		})
	}
}
