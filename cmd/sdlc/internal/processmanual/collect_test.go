package processmanual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestJudgeSources_CoversEveryCategoryIncludingEstimate(t *testing.T) {
	got := judgeSources()
	titles := map[string]InjectionSource{}
	for _, s := range got {
		if s.Kind != KindSDLCPrompt {
			t.Errorf("judgeSources produced non-prompt kind %q", s.Kind)
		}
		titles[s.Title] = s
	}
	// All 8 categories — AllCategories() omits estimate-quality, the catalog must not.
	want := []judge.Category{
		judge.DRY, judge.PURE, judge.Plan, judge.PlanQuality,
		judge.EstimateQuality, judge.Specs, judge.Lessons, judge.MilestoneReview,
	}
	for _, c := range want {
		s, ok := titles[string(c)]
		if !ok {
			t.Fatalf("judgeSources missing category %q", c)
		}
		if strings.TrimSpace(s.Body) == "" {
			t.Errorf("category %q has empty rendered body", c)
		}
		if !strings.Contains(s.Link, "prompts.go") {
			t.Errorf("category %q link should point at the builder, got %q", c, s.Link)
		}
	}
}

func TestHelptextSources_FromFakeFS(t *testing.T) {
	fsys := fstest.MapFS{
		"close.md": {Data: []byte("Close gate.\n\nSecond para.")},
		"root.md":  {Data: []byte("The workflow contract.")},
	}
	got := helptextSources(fsys)
	if len(got) != 2 {
		t.Fatalf("want 2 help-text sources, got %d", len(got))
	}
	byTitle := map[string]InjectionSource{}
	for _, s := range got {
		if s.Kind != KindHelpText || !strings.HasPrefix(s.Link, "cmd/sdlc/helptext/") {
			t.Errorf("bad help-text source: %+v", s)
		}
		byTitle[s.Title] = s
	}
	// root has no `sdlc root` verb — its help is bare `sdlc --help`.
	if w := byTitle["root"].When; !strings.Contains(w, "sdlc --help") || strings.Contains(w, "sdlc root") {
		t.Errorf("root When should name bare `sdlc --help`, got %q", w)
	}
	// The excerpt is the first paragraph only.
	if b := byTitle["close"].Body; strings.Contains(b, "Second para") {
		t.Errorf("close Body should be first paragraph only, got %q", b)
	}
}

func TestSkillSources_ParsesFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, ".claude", "skills", "xx-demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skill := "---\nname: xx-demo\ndescription: Use when demoing.\n---\n\nBody paragraph.\n\nSecond.\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skill), 0o644); err != nil {
		t.Fatal(err)
	}
	got := skillSources(filepath.Join(tmp, ".claude", "skills"), tmp)
	if len(got) != 1 {
		t.Fatalf("want 1 skill source, got %d: %+v", len(got), got)
	}
	s := got[0]
	if s.Kind != KindSkill || s.Title != "xx-demo" {
		t.Errorf("unexpected skill source: %+v", s)
	}
	if !strings.Contains(s.When, "Use when demoing") {
		t.Errorf("When should carry the description, got %q", s.When)
	}
	if !strings.HasSuffix(s.Link, "SKILL.md") {
		t.Errorf("Link should point at SKILL.md, got %q", s.Link)
	}
	if !strings.Contains(s.Body, "Body paragraph") || strings.Contains(s.Body, "Second") {
		t.Errorf("Body should be first paragraph only, got %q", s.Body)
	}
}

func TestFileSources_LessonsAndAgentsChain(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "workshop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workshop", "lessons.md"), []byte("# Lessons\n\nrule.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# Constitution\n\nbody.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// GEMINI.md / CLAUDE.md intentionally absent → must not error or appear.
	got := fileSources(root)
	kinds := map[Kind]int{}
	for _, s := range got {
		kinds[s.Kind]++
		if s.Link == "" {
			t.Errorf("fileSources record missing Link: %+v", s)
		}
	}
	if kinds[KindLessons] != 1 {
		t.Errorf("want 1 lessons record, got %d", kinds[KindLessons])
	}
	if kinds[KindAgentsChain] != 1 {
		t.Errorf("want 1 agents-chain record (only AGENTS.md present), got %d", kinds[KindAgentsChain])
	}
}

// TestCollect_SpansEveryKind is the ARCH-PURPOSE shadow-sweep: over a fixture
// wiring up every source, Collect must surface all six kinds — no injection
// source silently dropped.
func TestCollect_SpansEveryKind(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	skillDir := filepath.Join(root, ".claude", "skills", "xx-demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(skillDir, "SKILL.md"), "---\ndescription: demo.\n---\n\nbody.\n")

	if err := os.MkdirAll(filepath.Join(root, "workshop"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "workshop", "lessons.md"), "# Lessons\n\nx.\n")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "# Constitution\n\ny.\n")

	memDir := filepath.Join(home, ".claude", "projects", claudeProjectSlug(root), "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(memDir, "MEMORY.md"), "# Index\n")

	got := Collect(CollectOptions{
		RepoRoot:  root,
		SkillsDir: filepath.Join(root, ".claude", "skills"),
		HomeDir:   home,
	})
	seen := map[Kind]bool{}
	for _, s := range got {
		seen[s.Kind] = true
	}
	for _, k := range []Kind{KindSDLCPrompt, KindHelpText, KindSkill, KindLessons, KindAgentsChain, KindMemory} {
		if !seen[k] {
			t.Errorf("Collect missing kind %q (shadow-sweep)", k)
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
