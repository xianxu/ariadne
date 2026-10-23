package reviewexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These instructions are composed together by hosted and standalone review.
// Checking only the canonical naming paragraph missed contradictory action rows.
func TestEditorialCreationPromptComposition(t *testing.T) {
	root := moduleRoot(t)
	for _, name := range []string{
		"construct/local/universal-review/collaboration-modes.md",
		"construct/local/fix/SKILL.md",
		GrammarSource,
	} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			for _, retired := range []string{"same flat folder", ".side/NN-slug.md", ".side/NN-lock-granularity.md", ".side/<date>-NN-s4-cost.md"} {
				if strings.Contains(s, retired) {
					t.Errorf("active prompt still prescribes %q", retired)
				}
			}
			if !strings.Contains(s, "collaboration-modes.md") && !strings.Contains(s, "<name>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md") {
				t.Error("creation prompt neither owns nor references the common side-doc contract")
			}
		})
	}
}
