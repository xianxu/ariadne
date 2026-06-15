package plan

import "testing"

func TestParseTargetKnown(t *testing.T) {
	for _, s := range []string{"claude", "codex", "agy"} {
		got, err := ParseTarget(s)
		if err != nil {
			t.Fatalf("ParseTarget(%q) errored: %v", s, err)
		}
		if string(got) != s {
			t.Fatalf("ParseTarget(%q) = %q, want %q", s, got, s)
		}
	}
}

func TestParseTargetUnknownErrors(t *testing.T) {
	_, err := ParseTarget("gemini")
	if err == nil {
		t.Fatal("ParseTarget on an unknown name should error")
	}
	// The error names the offending value and the valid set.
	for _, want := range []string{"gemini", "claude", "codex", "agy"} {
		if !contains(err.Error(), want) {
			t.Errorf("ParseTarget error %q missing %q", err.Error(), want)
		}
	}
}

func TestTargetSkillBackendsMutuallyExclusive(t *testing.T) {
	// claude: symlinks, NO menu.
	if !TargetClaude.EmitSkillSymlinks() {
		t.Error("claude should emit skill symlinks")
	}
	if TargetClaude.IncludeSkillMenu() {
		t.Error("claude should NOT include the skill menu")
	}
	// codex / agy: menu, NO symlinks. The two backends are complementary.
	for _, tg := range []Target{TargetCodex, TargetAgy} {
		if tg.EmitSkillSymlinks() {
			t.Errorf("%s should NOT emit skill symlinks", tg)
		}
		if !tg.IncludeSkillMenu() {
			t.Errorf("%s should include the skill menu", tg)
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
