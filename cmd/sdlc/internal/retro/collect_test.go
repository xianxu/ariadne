package retro

import (
	"strings"
	"testing"

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
