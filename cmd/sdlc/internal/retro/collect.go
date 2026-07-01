package retro

import (
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// catalogCategories is the COMPLETE injected set — judge.AllCategories() omits
// EstimateQuality (scoped to standalone `sdlc judge` validity), but that prompt
// IS injected at change-code (changecode.go), so the manual must list it
// (ARCH-PURPOSE: deliver every injection point, not the easy subset).
func catalogCategories() []judge.Category {
	return append(append([]judge.Category{}, judge.AllCategories()...), judge.EstimateQuality)
}

// categoryBody is what the manual shows for a judge category. BuildPrompt is a
// pure function, so an empty PromptInput yields the deterministic prompt
// skeleton — EXCEPT `lessons`, which BuildPrompt renders as "" because it is a
// reminder ping, not a full agent prompt. For lessons the injected content IS
// judge.LessonsReminder. (This judge `lessons` category is a DIFFERENT injection
// point from the workshop/lessons.md file, which fileSources emits under Kind
// `lessons`.)
func categoryBody(c judge.Category) string {
	if body := judge.BuildPrompt(c, judge.PromptInput{}); strings.TrimSpace(body) != "" {
		return body
	}
	if c == judge.Lessons {
		return judge.LessonsReminder
	}
	return ""
}

// whenForCategory maps each category to its injection-trigger prose (pure).
func whenForCategory(c judge.Category) string {
	switch c {
	case judge.PlanQuality, judge.EstimateQuality:
		return "plan-quality gate at `sdlc change-code`"
	case judge.MilestoneReview:
		return "boundary review at `sdlc close` / `sdlc milestone-close`"
	case judge.Lessons:
		return "lessons reminder emitted at the review boundary"
	default:
		return "`sdlc judge " + string(c) + "`"
	}
}

// judgeSources is pure (judge.BuildPrompt + a constant do no IO).
func judgeSources() []InjectionSource {
	var out []InjectionSource
	for _, c := range catalogCategories() {
		out = append(out, InjectionSource{
			Kind:  KindSDLCPrompt,
			Title: string(c),
			When:  whenForCategory(c),
			Link:  "cmd/sdlc/internal/judge/prompts.go",
			Body:  categoryBody(c),
		})
	}
	return out
}
