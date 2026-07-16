package vocab

import (
	"os"
	"strings"
	"testing"
)

// TestProjectProseCitesModel binds construct/datatype/project.md to the model:
// every status token and scaffold section appears in the prose, the prose cites
// the cue as schema authority, and the retired hand-maintained enum is absent.
func TestProjectProseCitesModel(t *testing.T) {
	prose, err := os.ReadFile("../../construct/datatype/project.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(prose)
	for _, status := range Project().AllStatuses() {
		if !strings.Contains(text, "`"+status+"`") {
			t.Errorf("project prose omits modeled status %q", status)
		}
	}
	for _, section := range Project().Sections() {
		if !strings.Contains(text, "## "+section.Name) {
			t.Errorf("project prose omits modeled scaffold section %q", section.Name)
		}
	}
	if !strings.Contains(text, "construct/vocabulary/project.cue") || !strings.Contains(text, "schema authority") {
		t.Error("project prose does not cite project.cue as schema authority")
	}
	for _, retired := range []string{"`active`", "status: active"} {
		if strings.Contains(text, retired) {
			t.Errorf("project prose retains retired status form %q", retired)
		}
	}
}
