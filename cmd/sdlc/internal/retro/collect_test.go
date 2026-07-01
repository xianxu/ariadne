package retro

import (
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
