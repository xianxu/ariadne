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
	// root.md folds in the start-of-work runbook: `sdlc --help` is the
	// single workflow contract (the old `--agents.md` flag was merged
	// in), so these load-bearing anchors must survive edits.
	for _, want := range []string{"BEFORE WORK", "WHEN A VERB ERRORS", "checkpoint guard"} {
		if !strings.Contains(s, want) {
			t.Errorf("root.md missing %q", want)
		}
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

func TestPushEmbedded(t *testing.T) {
	s, ok := Get("push")
	if !ok {
		t.Fatal("push.md not found in embed FS")
	}
	// Regression guard for #54: push.md must frame the in-place branch
	// flow (change-code → pr → merge) as the default close path, not
	// present `sdlc push` as a co-equal "lighter path".
	if !strings.Contains(s, "change-code") {
		t.Errorf("push.md missing reference to the default change-code (in-place branch) flow")
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
