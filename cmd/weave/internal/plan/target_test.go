package plan

import (
	"reflect"
	"testing"
)

func TestParseTargetKnown(t *testing.T) {
	for _, s := range []string{"all", "claude", "codex", "gemini"} {
		got, err := ParseTarget(s)
		if err != nil {
			t.Fatalf("ParseTarget(%q) errored: %v", s, err)
		}
		if string(got) != s {
			t.Fatalf("ParseTarget(%q) = %q, want %q", s, got, s)
		}
	}
	// The empty string (no --target) is the Union default.
	if got, err := ParseTarget(""); err != nil || got != TargetAll {
		t.Fatalf("ParseTarget(\"\") = (%q,%v), want (all,nil)", got, err)
	}
}

func TestParseTargetUnknownErrors(t *testing.T) {
	_, err := ParseTarget("agy") // retired in Option B
	if err == nil {
		t.Fatal("ParseTarget on an unknown name should error")
	}
	// The error names the offending value and the valid set.
	for _, want := range []string{"agy", "all", "claude", "codex", "gemini"} {
		if !contains(err.Error(), want) {
			t.Errorf("ParseTarget error %q missing %q", err.Error(), want)
		}
	}
}

// Option B: each target's faces = (entry file, skill dir). The Union is every
// harness's face with entry files + skill dirs deduped (codex+gemini share
// .agents/skills); a lean target is exactly one face.
func TestTargetFaces(t *testing.T) {
	if got, want := TargetAll.EntryFiles(), []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Union EntryFiles = %v, want %v", got, want)
	}
	if got, want := TargetAll.SkillDirs(), []string{".claude/skills", ".agents/skills"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Union SkillDirs = %v, want %v (codex+gemini share .agents/skills)", got, want)
	}
	for _, c := range []struct {
		t          Target
		entry, dir string
	}{
		{TargetClaude, "CLAUDE.md", ".claude/skills"},
		{TargetCodex, "AGENTS.md", ".agents/skills"},
		{TargetGemini, "GEMINI.md", ".agents/skills"},
	} {
		if got, want := c.t.EntryFiles(), []string{c.entry}; !reflect.DeepEqual(got, want) {
			t.Errorf("%s EntryFiles = %v, want %v", c.t, got, want)
		}
		if got, want := c.t.SkillDirs(), []string{c.dir}; !reflect.DeepEqual(got, want) {
			t.Errorf("%s SkillDirs = %v, want %v", c.t, got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
