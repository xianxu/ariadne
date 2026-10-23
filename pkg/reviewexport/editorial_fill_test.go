package reviewexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are shared writer conformance oracles, not an implementation of fill.
// Pair/parley own the runtime operation; the skill's fresh-context scenarios
// also apply the prose contract to these inputs.
func TestEditorialFrontmatterFillExamples(t *testing.T) {
	for _, tc := range []struct {
		name, presentation, processing, instruction string
	}{
		{"empty", "highlight", "free-form", ""},
		{"existing", "direct", "proofreading", ""},
		{"legacy", "propose", "copy-editing", "Fix punctuation."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			read := func(tree string) string {
				b, err := os.ReadFile(filepath.Join(skillDir(t), "examples", "editorial-fill-"+tc.name, tree, "note.md"))
				if err != nil {
					t.Fatal(err)
				}
				return string(b)
			}
			before, after := read("before"), read("after")
			old, _ := frontmatter(before)
			got, ok := frontmatter(after)
			if !ok {
				t.Fatal("filled document lacks frontmatter")
			}
			for k, v := range map[string]string{
				"collaboration-mode": "discussion", "presentation-mode": tc.presentation,
				"processing-mode": tc.processing, "processing-instruction": tc.instruction,
			} {
				if got[k] != v {
					t.Errorf("%s = %q, want %q", k, got[k], v)
				}
			}
			for k, v := range old {
				if got[k] != v || !strings.Contains(after, k+": "+v+"\n") {
					t.Errorf("fill changed existing field %s: %q -> %q", k, v, got[k])
				}
			}
			if before[strings.Index(before, "# Note"):] != after[strings.Index(after, "# Note"):] {
				t.Error("fill changed document body")
			}
		})
	}
	root := filepath.Join(skillDir(t), "examples", "editorial-fill-multimode")
	if !sameTree(os.DirFS(root)) {
		t.Error("legacy multi-mode fill must refuse without changing any bytes")
	}
}
