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
		name, collaboration, presentation, processing, instruction string
	}{
		{"empty", "discussion", "highlight", "free-form", ""},
		{"existing", "discussion", "direct", "proofreading", ""},
		{"legacy", "discussion", "propose", "copy-editing", "Fix punctuation."},
		{"canonical-instruction", "chat", "propose", "copy-editing", "Keep the operator's exact brief."},
		{"canonical-mode", "collaboration", "highlight", "fact-check", "Keep this legacy brief."},
		{"presentation-unknown", "discussion", "typo", "free-form", ""},
		{"presentation-empty", "discussion", "", "free-form", ""},
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
				"collaboration-mode": tc.collaboration, "presentation-mode": tc.presentation,
				"processing-mode": tc.processing, "processing-instruction": tc.instruction,
			} {
				if got[k] != v {
					t.Errorf("%s = %q, want %q", k, got[k], v)
				}
			}
			for k, v := range old {
				if got[k] != v {
					t.Errorf("fill changed existing field %s: %q -> %q", k, v, got[k])
				}
				// Preserve actual source lines, including an empty value without
				// a post-colon space, rather than reconstructing their formatting.
				for _, line := range strings.Split(before, "\n") {
					if strings.HasPrefix(line, k+":") && !strings.Contains(after, "\n"+line+"\n") {
						t.Errorf("fill changed existing source line %q", line)
					}
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
