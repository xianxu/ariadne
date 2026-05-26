package helptext

import (
	"strings"
	"testing"
)

func TestRootEmbedded(t *testing.T) {
	s, ok := Get("root")
	if !ok {
		t.Fatal("root.md not found in embed FS")
	}
	if !strings.Contains(s, "WORKFLOW STAGES") {
		t.Errorf("root.md missing WORKFLOW STAGES section")
	}
	if !strings.HasSuffix(s, "\n") {
		t.Errorf("root.md content should end with a newline")
	}
}

func TestIndexEmbedded(t *testing.T) {
	s, ok := Get("index")
	if !ok {
		t.Fatal("index.md not found in embed FS")
	}
	// SKILL.md template must lead with frontmatter so it's loadable
	// by the skill loader without post-processing.
	if !strings.HasPrefix(s, "---\n") {
		t.Errorf("index.md should start with YAML frontmatter, got prefix %q", s[:min(40, len(s))])
	}
	if !strings.Contains(s, "name: sdlc") {
		t.Errorf("index.md frontmatter missing name: sdlc")
	}
}

func TestCloseEmbedded(t *testing.T) {
	s, ok := Get("close")
	if !ok {
		t.Fatal("close.md not found in embed FS")
	}
	if !strings.Contains(s, "--actual") || !strings.Contains(s, "--verified") {
		t.Errorf("close.md missing required-flag documentation")
	}
}

func TestMustGetPanicsOnMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet on missing entry should panic")
		}
	}()
	MustGet("definitely-not-a-real-entry")
}

func TestNamesIncludesCoreEntries(t *testing.T) {
	names := Names()
	found := make(map[string]bool, len(names))
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"root", "index", "close"} {
		if !found[want] {
			t.Errorf("Names() missing %q; got %v", want, names)
		}
	}
}
